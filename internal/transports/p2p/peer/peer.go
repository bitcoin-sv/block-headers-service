package peer

import (
	"crypto/rand"
	"errors"
	"fmt"
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
	conn            net.Conn
	addr            *net.TCPAddr
	cfg             *config.P2PConfig
	chainParams     *chaincfg.Params
	checkpoints     []chaincfg.Checkpoint
	nextCheckpoint  *chaincfg.Checkpoint
	headersService  service.Headers
	chainService    service.Chains
	log             *zerolog.Logger
	services        wire.ServiceFlag
	protocolVersion uint32
	nonce           uint64
	lastBlock       int32
	startingHeight  int32
	timeOffset      int64
	userAgent       string
	verAckReceived  bool
	synced          bool

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
		return nil, err
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
	nextCheckpoint := findNextHeaderCheckpoint(chainParams.Checkpoints, currentTipHeight)

	peer := &Peer{
		addr:            netAddr,
		cfg:             cfg,
		chainParams:     chainParams,
		headersService:  headersService,
		chainService:    chainService,
		checkpoints:     chainParams.Checkpoints,
		nextCheckpoint:  nextCheckpoint,
		log:             log,
		services:        wire.SFspv,
		protocolVersion: initialProtocolVersion,
		synced:          nextCheckpoint == nil,
		wg:              sync.WaitGroup{},
		msgChan:         make(chan wire.Message),
		quitting:        false,
		quit:            make(chan struct{}),
	}
	return peer, nil
}

func (p *Peer) Connect() error {
	conn, err := net.Dial(p.addr.Network(), p.addr.String())
	if err != nil {
		p.log.Error().Msgf("error connecting to the peer, reason: %v", err)
		return err
	}
	p.conn = conn
	p.log.Info().Msgf("connected to peer: %s", p)
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

func (p *Peer) Start() error {
	err := p.negotiateProtocol()
	if err != nil {
		p.log.Error().Msgf("error negotiating protocol with peer %s, reason: %v", p, err)
		return err
	}

	go p.writeMsgHandler()
	go p.readMsgHandler()
	go p.pingHandler()

	err = p.startSync()
	if err != nil {
		p.log.Error().Msgf("error starting sync from peer %s, reason: %v", p, err)
		return err
	}
	return nil
}

func (p *Peer) startSync() error {
	p.log.Info().Msgf("requesting headers from peer %s", p)

	var err error
	if p.nextCheckpoint != nil {
		err = p.writeGetHeadersMsg(p.nextCheckpoint.Hash)
	} else {
		err = p.writeGetHeadersMsg(&zeroHash)
	}

	if err != nil {
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

	err = p.readVerOrVerAck()
	if err != nil {
		return err
	}

	err = p.readVerOrVerAck()
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
	p.wg.Add(1)
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

out:
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
			break out
		}
	}

	p.log.Info().Msgf("ping handler shutdown for peer %s", p)
	p.wg.Done()
}

// MsgHandler is a message handler for incoming messages.
// Must be run as a goroutine.
func (p *Peer) readMsgHandler() {
	p.wg.Add(1)

out:
	for {
		select {
		case <-p.quit:
			break out

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

	p.log.Info().Msgf("read msg handler shutdown for peer %s", p)
	p.wg.Done()
}

func (p *Peer) writeMessage(msg wire.Message) error {
	err := wire.WriteMessage(p.conn, msg, p.protocolVersion, p.chainParams.Net)
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) writeRejectMessage(msg wire.Message, reason string) {
	rejectMsg := wire.NewMsgReject(msg.Command(), wire.RejectObsolete, reason)
	err := p.writeMessage(rejectMsg)
	if err != nil {
		p.log.Error().Msgf("could not write reject message to peer, reason: %v", err)
	}
}

func (p *Peer) queueMessage(msg wire.Message) {
	// running in goroutine here, because writing
	// to channel is blocking until read
	go func() {
		p.msgChan <- msg
	}()
}

// writeMsgHandler serves as a queue for writing messages to peers,
// must be run as a goroutine.
func (p *Peer) writeMsgHandler() {
	p.wg.Add(1)

out:
	for {
		select {
		case msg := <-p.msgChan:
			err := p.writeMessage(msg)
			if err != nil {
				p.log.Error().Msgf("error writing msg %T to peer %s", msg, p)
			}
		case <-p.quit:
			break out
		}
	}

cleanup:
	for {
		select {
		case <-p.msgChan:
		default:
			break cleanup
		}
	}

	p.log.Info().Msgf("write msg handler shutdown for peer %s", p)
	p.wg.Done()
}

func (p *Peer) writeOurVersionMsg() error {
	n, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
	if err != nil {
		panic(err)
	}
	nonce := n.Uint64()
	p.nonce = nonce

	ourNA := &wire.NetAddress{
		Services:  wire.SFspv,
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

	// NOTE: it's 0 by default, so in theory we don't need
	// to set that, but it's better to be explicit.
	msg.Services = p.services

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

func (p *Peer) readVerOrVerAck() error {
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
	case *wire.MsgVerAck:
		p.verAckReceived = true
	default:
		_, err = p.requireVerAckReceived(remoteMsg)
		if err != nil {
			return err
		}
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

	p.protocolVersion = minUint32(p.protocolVersion, uint32(msg.ProtocolVersion))
	p.services = msg.Services
	p.lastBlock = msg.LastBlock
	p.startingHeight = msg.LastBlock
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
		p.queueMessage(wire.NewMsgPong(msg.Nonce))
	}
}

func (p *Peer) handleInvMsg(msg *wire.MsgInv) {
	p.log.Info().Msgf("received inv msg from peer %s", p)
	if !p.synced {
		p.log.Info().Msgf("we are still syncing, ignoring inv msg from peer %s", p)
		return
	}

	lastBlock := searchForFinalBlock(msg.InvList)
	if lastBlock == -1 {
		p.log.Info().Msgf("no blocks in inv msg from peer %s", p)
		return
	}

	// if last header from inv msg is already in database, ignore
	lastBlockHash := &msg.InvList[lastBlock].Hash
	_, err := p.headersService.GetHeightByHash(lastBlockHash)
	if err == nil {
		p.log.Info().Msgf("blocks from inv msg from peer %s already existsing in db", p)
		return
	}

	p.log.Info().Msgf("requesting new headers from peer %s", p)
	err = p.writeGetHeadersMsg(lastBlockHash)
	if err != nil {
		p.log.Error().Msgf("error while requesting new headers from peer %s, reason: %v", p, err)
	}
}

func (p *Peer) handleHeadersMsg(msg *wire.MsgHeaders) {
	receivedCheckpoint := false
	lastHeight := int32(0)
	headersReceived := 0

	for _, header := range msg.Headers {
		h, addErr := p.chainService.Add(domains.BlockHeaderSource(*header))

		if service.HeaderAlreadyExists.Is(addErr) {
			continue
		}

		if service.BlockRejected.Is(addErr) {
			// TODO: ban peer
			p.log.Error().Msgf("received rejected header %v from peer %s", h, p)
			_ = p.Disconnect()
			return
		}

		if service.HeaderSaveFail.Is(addErr) {
			p.log.Error().Msgf("couldn't save header %v in database, because of %+v", h, addErr)
			continue
		}

		if service.HeaderCreationFail.Is(addErr) {
			p.log.Error().Msgf("couldn't create header from %v because of error %+v", header, addErr)
			continue
		}

		if service.ChainUpdateFail.Is(addErr) {
			p.log.Error().Msgf("when adding header %v couldn't update chains state because of error %+v", header, addErr)
			continue
		}

		var err error
		receivedCheckpoint, err = p.verifyCheckpointReached(h, receivedCheckpoint)
		if err != nil {
			// TODO: ban peer or lower peer sync score
			_ = p.Disconnect()
			return
		}

		lastHeight = h.Height
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

	if receivedCheckpoint {
		p.nextCheckpoint = findNextHeaderCheckpoint(p.checkpoints, p.nextCheckpoint.Height)
		if p.nextCheckpoint == nil {
			p.synced = true
		}
	}

	if p.nextCheckpoint != nil {
		p.log.Info().Msgf(
			"requesting next headers batch from peer %s, height range %d - %d",
			p, lastHeight, p.nextCheckpoint.Height,
		)
		err := p.writeGetHeadersMsg(p.nextCheckpoint.Hash)
		if err != nil {
			p.log.Error().Msgf("error requesting headers from peer %s, reason: %v", p, err)
		}
		return
	}

	p.log.Info().Msgf(
		"checkpoints synced, requesting headers from height %d up to end of chain (zero hash) from peer %s",
		lastHeight, p,
	)
	err := p.writeGetHeadersMsg(&zeroHash)
	if err != nil {
		p.log.Error().Msgf("error requesting headers from peer %s, reason: %v", p, err)
	}
}

func (p *Peer) verifyCheckpointReached(h *domains.BlockHeader, receivedCheckpoint bool) (bool, error) {
	if p.nextCheckpoint != nil && h.Height == p.nextCheckpoint.Height {
		if h.Hash == *p.nextCheckpoint.Hash {
			receivedCheckpoint = true
			p.log.Info().Msgf(
				"verified downloaded block header against checkpoint at height %d / hash %s",
				h.Height, h.Hash,
			)
		} else {
			p.log.Error().Msgf(
				"block header at height %d/hash %s from peer %s does NOT match expected checkpoint hash of %s -- disconnecting",
				h.Height, h.Hash, p, p.nextCheckpoint.Hash,
			)
			return false, fmt.Errorf("corresponding checkpoint height does not match got: %v, exp: %v", h.Height, p.nextCheckpoint.Height)
		}
	}
	return receivedCheckpoint, nil
}

func (p *Peer) String() string {
	return p.addr.String()
}
