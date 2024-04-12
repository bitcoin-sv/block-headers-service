package peer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/rs/zerolog"
)

type Peer struct {
	conn              net.Conn
	addr              *net.TCPAddr
	cfg               *config.P2PConfig
	chainParams       *chaincfg.Params
	checkpoint        *checkpoint
	headersService    service.Headers
	chainService      service.Chains
	log               *zerolog.Logger
	services          wire.ServiceFlag
	protocolVersion   uint32
	nonce             uint64
	latestHeight      int32
	latestHash        *chainhash.Hash
	latestStatsMutex  sync.RWMutex
	timeOffset        int64
	userAgent         string
	syncedCheckpoints bool
	sendHeadersMode   bool

	wg       sync.WaitGroup
	msgChan  chan wire.Message
	quitting bool
	quit     chan struct{}
}

func NewPeer(
	addr string,
	cfg *config.P2PConfig,
	chainParams *chaincfg.Params,
	headersService service.Headers,
	chainService service.Chains,
	log *zerolog.Logger,
) (*Peer, error) {
	port, err := strconv.Atoi(config.ActiveNetParams.DefaultPort)
	if err != nil {
		return nil, fmt.Errorf("could not parse default port: %w", err)
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, errors.New("could not parse peer IP")
	}

	netAddr := &net.TCPAddr{
		IP:   ip,
		Port: port,
	}

	currentTipHeight := headersService.GetTipHeight()

	peer := &Peer{
		addr:            netAddr,
		cfg:             cfg,
		chainParams:     chainParams,
		headersService:  headersService,
		chainService:    chainService,
		checkpoint:      newCheckpoint(chainParams.Checkpoints, currentTipHeight, log),
		log:             log,
		services:        wire.SFspv,
		protocolVersion: initialProtocolVersion,
		wg:              sync.WaitGroup{},
		msgChan:         make(chan wire.Message, writeMsgChannelBufferSize),
		quitting:        false,
		quit:            make(chan struct{}),
	}
	return peer, nil
}

func (p *Peer) Connect() error {
	conn, err := net.Dial(p.addr.Network(), p.addr.String())
	if err != nil {
		p.log.Error().Msgf("error connecting to peer %s, reason: %v", p, err)
		return err
	}
	p.conn = conn
	p.log.Info().Msgf("connected to peer: %s", p)

	err = p.negotiateProtocol()
	if err != nil {
		p.log.Error().Msgf("error negotiating protocol with peer %s, reason: %v", p, err)
		_ = p.Disconnect()
		return err
	}

	go p.pingHandler()

	return nil
}

func (p *Peer) Disconnect() error {
	p.log.Info().Msgf("disconnecting peer: %s", p)

	p.quitting = true
	close(p.quit)
	err := p.conn.Close()
	if err != nil {
		return err
	}

	p.wg.Wait()
	p.log.Info().Msgf("successfully disconnected peer %s", p)
	return nil
}

func (p *Peer) StartHeadersSync() error {
	go p.writeMsgHandler()
	go p.readMsgHandler()

	currentTipHeight := p.headersService.GetTipHeight()
	p.checkpoint.UseTip(currentTipHeight)
	p.sendHeadersMode = false

	err := p.requestHeaders()
	if err != nil {
		// TODO: lower peer sync score
		_ = p.Disconnect()
		return err
	}
	return nil
}

func (p *Peer) requestHeaders() error {
	var err error
	if p.checkpoint.LastReached() {
		p.log.Info().Msgf("requesting next headers batch from peer %s, up to height %d", p, p.checkpoint.Height())
		err = p.writeGetHeadersMsg(p.checkpoint.Hash())
	} else {
		p.log.Info().Msgf("checkpoints synced, requesting headers up to end of chain from peer %s", p)
		err = p.writeGetHeadersMsg(&zeroHash)
	}

	if err != nil {
		p.log.Error().Msgf("error requesting headers from peer %s, reason: %v", p, err)
		return err
	}
	return nil
}

func (p *Peer) negotiateProtocol() error {
	err := p.writeOurVersionMsg()
	if err != nil {
		return err
	}
	p.log.Info().Msgf("version sent to peer: %s", p)

	err = p.readVerAndVerAck()
	if err != nil {
		return err
	}
	p.log.Info().Msgf("received version and verack from peer: %s", p)

	err = p.writeMessage(wire.NewMsgVerAck())
	if err != nil {
		return err
	}

	p.log.Info().Msgf("protocol negotiated successfully with peer: %s", p)
	return nil
}

// PingHandler is a handler for sending ping messages to peers.
// Must be run as a goroutine.
func (p *Peer) pingHandler() {
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case <-pingTicker.C:
			nonce, err := wire.RandomUint64()
			if err != nil {
				p.log.Error().Msgf("error generating nonce for ping msg, reason: %v", err)
				continue
			}
			p.log.Info().Msgf("sending ping to peer %s with nonce: %d", p, nonce)
			p.queueMessage(wire.NewMsgPing(nonce))

		case <-p.quit:
			p.log.Info().Msgf("ping handler shutdown for peer %s", p)
			return
		}
	}
}

// MsgHandler is a message handler for incoming messages.
// Must be run as a goroutine.
func (p *Peer) readMsgHandler() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case <-p.quit:
			p.log.Info().Msgf("read msg handler shutdown for peer %s", p)
			return

		default:
			remoteMsg, _, err := wire.ReadMessage(p.conn, p.protocolVersion, p.chainParams.Net)
			if err != nil {
				if !p.quitting {
					p.log.Error().Msgf("cannot read message from peer %s, reason: %v", p, err)
				}
				continue
			}

			switch msg := remoteMsg.(type) {
			case *wire.MsgPing:
				p.handlePingMsg(msg)
			case *wire.MsgPong:
				p.log.Info().Msgf("received pong from peer %s with nonce: %d", p, msg.Nonce)
			case *wire.MsgHeaders:
				p.handleHeadersMsg(msg)
			case *wire.MsgInv:
				p.handleInvMsg(msg)
			default:
				p.log.Info().Msgf("received msg of type: %T", msg)
			}
		}
	}
}

func (p *Peer) writeMessage(msg wire.Message) error {
	return wire.WriteMessage(p.conn, msg, p.protocolVersion, p.chainParams.Net)
}

func (p *Peer) writeRejectMessage(msg wire.Message, reason string) {
	rejectMsg := wire.NewMsgReject(msg.Command(), wire.RejectObsolete, reason)
	err := p.writeMessage(rejectMsg)
	if err != nil {
		p.log.Error().Msgf("could not write reject message to peer, reason: %v", err)
	}
}

func (p *Peer) queueMessage(msg wire.Message) {
	p.msgChan <- msg
}

// writeMsgHandler serves as a queue for writing messages to peers,
// must be run as a goroutine.
func (p *Peer) writeMsgHandler() {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case msg := <-p.msgChan:
			err := p.writeMessage(msg)
			if err != nil {
				p.log.Error().Msgf("error writing msg %T to peer %s", msg, p)
			}
		case <-p.quit:
			// draining the channels for cleanup
			for {
				select {
				case <-p.msgChan:
				default:
					p.log.Info().Msgf("write msg handler shutdown for peer %s", p)
					return
				}
			}
		}
	}
}

func (p *Peer) writeOurVersionMsg() error {
	randomVal, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return fmt.Errorf("could not generate random nonce: %v", err)
	}

	nonce := randomVal.Uint64()
	p.nonce = nonce

	ourNA := &wire.NetAddress{
		Services:  ourServices,
		Timestamp: time.Now(),
	}
	theirNA := wire.NewNetAddress(p.addr, 0)

	lastBlock := p.headersService.GetTip().Height

	msg := wire.NewMsgVersion(ourNA, theirNA, nonce, lastBlock)
	err = msg.AddUserAgent(p.cfg.UserAgentName, p.cfg.UserAgentVersion, userAgentComments)
	if err != nil {
		p.log.Error().Msgf("could not add user agent to version message, reason: %v", err)
		return err
	}

	msg.Services = ourServices

	err = p.writeMessage(msg)
	if err != nil {
		p.log.Error().Msgf("could not write version message to peer, reason: %v", err)
		return err
	}

	return nil
}

func (p *Peer) writeGetHeadersMsg(stopHash *chainhash.Hash) error {
	msg := wire.NewMsgGetHeaders()
	msg.HashStop = *stopHash

	locator := p.headersService.LatestHeaderLocator()
	for _, hash := range locator {
		err := msg.AddBlockLocatorHash(hash)
		if err != nil {
			return err
		}
	}

	p.queueMessage(msg)
	return nil
}

func (p *Peer) readVerAndVerAck() error {
	versionReceived := false
	verAckReceived := false
	numberOfMsgsExpected := 2

	for i := 0; i < numberOfMsgsExpected; i++ {
		remoteMsg, _, err := wire.ReadMessage(p.conn, p.protocolVersion, p.chainParams.Net)
		if err != nil {
			p.log.Error().Msgf("could not read message from peer, reason: %v", err)
			return err
		}

		switch msg := remoteMsg.(type) {
		case *wire.MsgVersion:
			err = p.handleVersionMessage(msg)
			if err != nil {
				return err
			}
			versionReceived = true
		case *wire.MsgVerAck:
			verAckReceived = true
		default:
			_, err = p.requireVerAckReceived(remoteMsg)
			if err != nil {
				return err
			}
		}
	}

	if !versionReceived || !verAckReceived {
		return fmt.Errorf("did not receive version and verack correctly, peer %s misbehaving", p)
	}

	return nil
}

func (p *Peer) handleVersionMessage(msg *wire.MsgVersion) error {
	// Detect self connections.
	if msg.Nonce == p.nonce {
		err := errors.New("disconnecting peer connected to self")
		p.log.Error().Msg(err.Error())
		return err
	}

	p.protocolVersion = min(p.protocolVersion, uint32(msg.ProtocolVersion))
	p.services = msg.Services
	p.latestHeight = msg.LastBlock
	p.timeOffset = msg.Timestamp.Unix() - time.Now().Unix()
	p.userAgent = msg.UserAgent

	if uint32(msg.ProtocolVersion) < minAcceptableProtocolVersion {
		reason := fmt.Sprintf("protocol version must be %d or greater", minAcceptableProtocolVersion)
		p.writeRejectMessage(msg, reason)
		p.log.Error().Msgf("%s, rejecting connection", reason)
		return errors.New(reason)
	}

	return nil
}

func (p *Peer) requireVerAckReceived(remoteMsg wire.Message) (*wire.MsgVerAck, error) {
	msg, ok := remoteMsg.(*wire.MsgVerAck)
	if !ok {
		reason := "missing-verack message, misbehaving peer"
		p.writeRejectMessage(msg, reason)
		p.log.Error().Msgf("%s, rejecting connection", reason)
		return nil, errors.New(reason)
	}
	return msg, nil
}

func (p *Peer) handlePingMsg(msg *wire.MsgPing) {
	p.log.Info().Msgf("received ping from peer %s with nonce: %d", p, msg.Nonce)
	if p.protocolVersion > wire.BIP0031Version {
		p.log.Info().Msgf("sending pong to peer %s with nonce: %d", p, msg.Nonce)
		p.queueMessage(wire.NewMsgPong(msg.Nonce))
	}
}

func (p *Peer) handleInvMsg(msg *wire.MsgInv) {
	p.log.Info().Msgf("received inv msg from peer %s", p)
	if !p.syncedCheckpoints {
		p.log.Info().Msgf("we are still syncing, ignoring inv msg from peer %s", p)
		return
	}

	lastBlock := searchForFinalBlockIndex(msg.InvList)
	if lastBlock == -1 {
		p.log.Info().Msgf("no blocks in inv msg from peer %s", p)
		return
	}

	lastBlockHash := &msg.InvList[lastBlock].Hash
	_, err := p.headersService.GetHeightByHash(lastBlockHash)
	if err == nil {
		p.log.Info().Msgf("blocks from inv msg from peer %s already existsing in db", p)
		return
	}
	p.updateLatestStats(0, lastBlockHash)

	p.log.Info().Msgf("requesting new headers from peer %s", p)
	err = p.writeGetHeadersMsg(lastBlockHash)
	if err != nil {
		p.log.Error().Msgf("error while requesting new headers from peer %s, reason: %v", p, err)
	}
}

func (p *Peer) handleHeadersMsg(msg *wire.MsgHeaders) {
	p.log.Info().Msgf("received headers msg from peer %s", p)

	lastHeight := int32(0)
	headersReceived := 0
	var lastHash *chainhash.Hash

	for _, header := range msg.Headers {
		h, err := p.chainService.Add(domains.BlockHeaderSource(*header))
		if err != nil {
			if service.HeaderAlreadyExists.Is(err) {
				continue
			}

			if service.BlockRejected.Is(err) {
				// TODO: ban peer
				p.log.Error().Msgf("received rejected header %v from peer %s", h, p)
				_ = p.Disconnect()
				return
			}

			if service.HeaderSaveFail.Is(err) {
				p.log.Error().Msgf("couldn't save header %v in database, because of %+v", h, err)
				continue
			}

			if service.HeaderCreationFail.Is(err) {
				p.log.Error().Msgf("couldn't create header from %v because of error %+v", header, err)
				continue
			}

			if service.ChainUpdateFail.Is(err) {
				p.log.Error().Msgf("when adding header %v couldn't update chains state because of error %+v", header, err)
				continue
			}
		}

		if !h.IsLongestChain() {
			// TODO: ban peer or lower sync score
			p.log.Warn().Msgf(
				"received header with hash: %s that's not a part of the longest chain, from peer %s",
				h.Hash.String(), p,
			)
			continue
		}

		err = p.checkpoint.Check(h)
		if err != nil {
			// TODO: ban peer or lower peer sync score
			p.log.Error().Msgf("error when checking checkpoint, reason: %v", err)
			_ = p.Disconnect()
			return
		}

		lastHeight = h.Height
		lastHash = &h.Hash
		headersReceived += 1
	}

	if headersReceived == 0 {
		p.log.Debug().Msgf("received only existing headers from peer: %s", p)
		return
	}

	p.log.Info().Msgf(
		"successfully received %d headers from peer %s, up to height %d",
		headersReceived, p, lastHeight,
	)

	p.updateLatestStats(lastHeight, lastHash)

	if p.sendHeadersMode {
		return
	}

	if p.isSynced() {
		p.log.Info().Msgf("synced with the tip of chain from peer %s", p)
		p.switchToSendHeadersMode()
		return
	}

	err := p.requestHeaders()
	if err != nil {
		// TODO: lower peer sync score
		_ = p.Disconnect()
	}
}

func (p *Peer) switchToSendHeadersMode() {
	if !p.sendHeadersMode && p.protocolVersion >= wire.SendHeadersVersion {
		p.log.Info().Msgf("switching to send headers mode - requesting peer %s to send us headers directly instead of inv msg", p)
		p.queueMessage(wire.NewMsgSendHeaders())
		p.sendHeadersMode = true
	}
}

func (p *Peer) updateLatestStats(lastHeight int32, lastHash *chainhash.Hash) {
	p.latestStatsMutex.Lock()
	defer p.latestStatsMutex.Unlock()

	if lastHeight > p.latestHeight {
		p.latestHeight = lastHeight
	}
	if lastHash.IsEqual(p.latestHash) {
		p.latestHash = nil
	}
}

func (p *Peer) getLatestStats() (lastHeight int32, lastHash *chainhash.Hash) {
	p.latestStatsMutex.RLock()
	defer p.latestStatsMutex.RUnlock()

	return p.latestHeight, p.latestHash
}

func (p *Peer) isSynced() bool {
	tipHeight := p.headersService.GetTipHeight()
	latestHeight, latestHash := p.getLatestStats()
	noNewHash := latestHash.IsEqual(nil)

	return noNewHash && latestHeight == tipHeight
}

func (p *Peer) String() string {
	return p.addr.String()
}
