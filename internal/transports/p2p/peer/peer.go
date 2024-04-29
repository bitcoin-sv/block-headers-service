package peer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
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

// Manager is peer manager.
type Manager interface {
	AddAddrs([]*wire.NetAddress)
	SignalError(*Peer, error)
}

type Peer struct {
	// conn is the current connection to peer
	conn net.Conn
	// addr is the net.Addr of peer, used for connection
	addr  *net.TCPAddr
	addrS string
	// inbound is specifying if the peer is inbound (incoming) or outbound (outgoing)
	inbound bool
	// cfg is the P2P configuration specified by the user
	cfg *config.P2PConfig
	// chainParams are the params of the main network
	chainParams *chaincfg.Params
	// checkpoint is used to validate block headers when syncing
	checkpoint *checkpoint
	// headersService is the service allowing us to retrieve e.g. the current tip height
	headersService service.Headers
	// chainService is used for adding new headers to the database
	chainService service.Chains
	// log is a zerolog logger used in the p2p package
	log *zerolog.Logger
	// services is a flag that specifies whether the peer is a full node or an SPV
	services wire.ServiceFlag
	// protocolVersion is the negotiated protocol version between us and the peer
	protocolVersion uint32
	// nonce is used when negotiating protocol version for another peer to identify us
	nonce uint64
	// latestHeight is the latest header's height of the peer,
	// used for checking if we're synced with the peer
	latestHeight int32
	// latestHash is the latest header's hash from the peer,
	// used for checking if we're synced with the peer
	latestHash *chainhash.Hash
	// latestStatsMutex is used to make latestHeight and latestHash thread-safe
	latestStatsMutex sync.RWMutex
	// timeOffset is the offset between peer sending the version message and us receiving it
	timeOffset int64
	// userAgent is the 'name' of the peer used for identification
	userAgent string
	// syncedCheckpoints is specifying whether we have synced all checkpoints
	syncedCheckpoints bool
	// sendHeadersMode is specifying whether we already sent the sendheaders message
	// to the peer and we're expecting to get just headers and no inv msg
	sendHeadersMode bool
	// lastSeen is timestamp of last pong message from peer
	lastSeen int64

	// wg is a WaitGroup used to properly handle peer disconnection
	wg sync.WaitGroup
	// msgChan is a buffered channel used as a queue for sending messages to peer
	msgChan chan wire.Message
	// quitting is a flag used to properly handle peer disconnection
	quitting bool
	// quit is a channel used to properly handle peer disconnection
	quit chan struct{}

	// manager is a peer  manager
	manager Manager
	// prevents errors if client invoke Disconect() multiple times
	disconnected bool
}

func NewPeer(
	conn net.Conn,
	inbound bool,
	cfg *config.P2PConfig,
	chainParams *chaincfg.Params,
	headersService service.Headers,
	chainService service.Chains,
	log *zerolog.Logger,
	manager Manager,
) *Peer {
	peer := &Peer{
		conn:            conn,
		inbound:         inbound,
		cfg:             cfg,
		chainParams:     chainParams,
		headersService:  headersService,
		chainService:    chainService,
		log:             log,
		services:        wire.SFspv,
		protocolVersion: initialProtocolVersion,
		wg:              sync.WaitGroup{},
		msgChan:         make(chan wire.Message, writeMsgChannelBufferSize),
		quitting:        false,
		quit:            make(chan struct{}),
		manager:         manager,
	}
	return peer
}

func (p *Peer) Connect() error {
	err := p.updatePeerAddr()
	if err != nil {
		return err
	}

	p.logInfo("connected to peer")

	err = p.negotiateProtocol()
	if err != nil {
		p.logError("error negotiating protocol with peer, reason: %v", err)
		return err
	}

	p.lastSeen = time.Now().Unix()
	go p.writeMsgHandler()
	go p.readMsgHandler()
	go p.pingHandler()

	return nil
}

func (p *Peer) Disconnect() {
	if p.disconnected {
		return
	}

	p.logInfo("disconnecting peer")

	p.quitting = true
	close(p.quit)
	err := p.conn.Close()
	if err != nil {
		p.logError("error disconnecting peer, reason %v", err)
	}

	p.wg.Wait()
	p.logInfo("successfully disconnected peer")

	p.disconnected = true
}

func (p *Peer) StartHeadersSync() error {
	currentTipHeight := p.headersService.GetTipHeight()
	p.checkpoint = newCheckpoint(p.chainParams.Checkpoints, currentTipHeight, p.log)
	p.sendHeadersMode = false

	err := p.requestHeaders()
	if err != nil {
		return err
	}
	return nil
}

func (p *Peer) SendGetAddrInfo() {
	p.queueMessage(wire.NewMsgGetAddr())
}

func (p *Peer) GetPeerAddr() *wire.NetAddress {
	return &wire.NetAddress{
		Timestamp: time.Unix(p.lastSeen, 0),
		Services:  p.services,
		IP:        p.addr.IP,
		Port:      uint16(p.addr.Port),
	}
}

func (p *Peer) Inbound() bool {
	return p.inbound
}

func (p *Peer) updatePeerAddr() error {
	remoteAddr, addrIsTcp := p.conn.RemoteAddr().(*net.TCPAddr)

	if remoteAddr != nil && addrIsTcp {
		p.addr = remoteAddr
		p.addrS = p.addr.String()
	} else {
		errMsg := "error retreiving address from peer"
		p.logError(errMsg)
		return errors.New(errMsg)
	}
	return nil
}

func (p *Peer) requestHeaders() error {
	var err error
	if p.checkpoint.LastReached() {
		p.logInfo("checkpoints synced, requesting headers up to end of chain from peer %s", p)
		err = p.writeGetHeadersMsg(&zeroHash)
	} else {
		p.logInfo("requesting next headers batch from peer %s, up to height %d", p, p.checkpoint.Height())
		err = p.writeGetHeadersMsg(p.checkpoint.Hash())
	}

	if err != nil {
		p.logError("error requesting headers, reason: %v", err)
		return err
	}
	return nil
}

func (p *Peer) negotiateProtocol() error {
	err := p.writeOurVersionMsg()
	if err != nil {
		return err
	}
	p.logInfo("version sent to peer")

	err = p.readVerAndVerAck()
	if err != nil {
		return err
	}
	p.logInfo("received version and verack")

	err = p.writeMessage(wire.NewMsgVerAck())
	if err != nil {
		return err
	}

	p.logInfo("protocol negotiated successfully with peer")
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
				p.logError("error generating nonce for ping msg, reason: %v", err)
				continue
			}
			p.logDebug("sending ping to peer with nonce: %d", nonce)
			p.queueMessage(wire.NewMsgPing(nonce))

		case <-p.quit:
			p.logInfo("ping handler shutdown for peer")
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
			p.logInfo("read msg handler shutdown")
			return

		default:
			remoteMsg, _, err := wire.ReadMessage(p.conn, p.protocolVersion, p.chainParams.Net)
			if err != nil {
				if !p.quitting {
					err = fmt.Errorf("cannot read message, reason: %w", err)
					p.logError(err.Error())
					p.manager.SignalError(p, err)
				}
				continue
			}

			switch msg := remoteMsg.(type) {
			case *wire.MsgPing:
				p.handlePingMsg(msg)
			case *wire.MsgPong:
				p.logDebug("received pong with nonce: %d", msg.Nonce)
				p.lastSeen = time.Now().Unix()
			case *wire.MsgHeaders:
				p.handleHeadersMsg(msg)
			case *wire.MsgInv:
				p.handleInvMsg(msg)
			case *wire.MsgAddr:
				p.handleAddrMsg(msg)
			default:
				p.logDebug("received msg of type: %T", msg)
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
		p.logError("could not write reject message to peer, reason: %v", err)
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
				p.logError("error writing msg %T to peer", msg)
			}
		case <-p.quit:
			// draining the channels for cleanup
			for {
				select {
				case <-p.msgChan:
				default:
					p.logInfo("write msg handler shutdown")
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

	lastBlock := p.headersService.GetTipHeight()

	msg := wire.NewMsgVersion(ourNA, theirNA, nonce, lastBlock)
	err = msg.AddUserAgent(p.cfg.UserAgentName, p.cfg.UserAgentVersion, userAgentComments)
	if err != nil {
		p.logError("could not add user agent to version message, reason: %v", err)
		return err
	}

	msg.Services = ourServices

	err = p.writeMessage(msg)
	if err != nil {
		p.logError("could not write version message to peer, reason: %v", err)
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

	p.logDebug("queue GetHeaders msg: %v", *msg)
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
			p.logError("could not read message from peer, reason: %v", err)
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
		p.logError(err.Error())
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
		p.logError("%s, rejecting connection", reason)
		return errors.New(reason)
	}

	return nil
}

func (p *Peer) requireVerAckReceived(remoteMsg wire.Message) (*wire.MsgVerAck, error) {
	msg, ok := remoteMsg.(*wire.MsgVerAck)
	if !ok {
		reason := "missing-verack message, misbehaving peer"
		p.writeRejectMessage(msg, reason)
		p.logError("%s, rejecting connection", reason)
		return nil, errors.New(reason)
	}
	return msg, nil
}

func (p *Peer) handlePingMsg(msg *wire.MsgPing) {
	p.logDebug("received ping with nonce: %d", msg.Nonce)
	if p.protocolVersion > wire.BIP0031Version {
		p.logDebug("sending pong to peer with nonce: %d", msg.Nonce)
		p.queueMessage(wire.NewMsgPong(msg.Nonce))
	}
}

func (p *Peer) handleInvMsg(msg *wire.MsgInv) {
	p.logInfo("received inv msg")
	if !p.syncedCheckpoints {
		p.logInfo("we are still syncing, ignoring inv msg")
		return
	}

	lastBlock := searchForFinalBlockIndex(msg.InvList)
	if lastBlock == -1 {
		p.logInfo("no blocks in inv msg from peer")
		return
	}

	lastBlockHash := &msg.InvList[lastBlock].Hash
	_, err := p.headersService.GetHeightByHash(lastBlockHash)
	if err == nil {
		p.logInfo("blocks from inv msg already existsing in db")
		return
	}
	p.updateLatestStats(0, lastBlockHash)

	p.logDebug("requesting new headers")
	err = p.writeGetHeadersMsg(lastBlockHash)
	if err != nil {
		p.logError("error while requesting new headers, reason: %v", err)
	}
}

func (p *Peer) handleHeadersMsg(msg *wire.MsgHeaders) {
	p.logInfo("received headers msg")

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
				p.logError("received rejected header %v", h)
				p.manager.SignalError(p, err)
				return
			}

			if service.HeaderSaveFail.Is(err) {
				p.logError("couldn't save header %v in database, because of %+v", h, err)
				continue
			}

			if service.HeaderCreationFail.Is(err) {
				p.logError("couldn't create header from %v because of error %+v", header, err)
				continue
			}

			if service.ChainUpdateFail.Is(err) {
				p.logError("when adding header %v couldn't update chains state because of error %+v", header, err)
				continue
			}
		}

		if !h.IsLongestChain() {
			// TODO: ban peer or lower sync score
			p.logWrn("received header with hash: %s that's not a part of the longest chain", h.Hash)
			continue
		}

		err = p.checkpoint.VerifyAndAdvance(h)
		if err != nil {
			p.logError("error when checking checkpoint, reason: %v", err)
			p.manager.SignalError(p, err)
			return
		}

		lastHeight = h.Height
		lastHash = &h.Hash
		headersReceived += 1
	}

	if headersReceived == 0 {
		p.logDebug("received only existing headers")
		return
	}

	p.logInfo("successfully received %d headers up to height %d", headersReceived, lastHeight)
	p.updateLatestStats(lastHeight, lastHash)

	if p.sendHeadersMode {
		return
	}

	if p.isSynced() {
		p.logInfo("synced with the tip of chain from peer")
		p.switchToSendHeadersMode()
		return
	}

	err := p.requestHeaders()
	if err != nil {
		p.manager.SignalError(p, err)
	}
}

func (p *Peer) handleAddrMsg(msg *wire.MsgAddr) {
	p.logInfo("received addr msg with %d addresses", len(msg.AddrList))
	p.manager.AddAddrs(msg.AddrList)
}

func (p *Peer) switchToSendHeadersMode() {
	if !p.sendHeadersMode && p.protocolVersion >= wire.SendHeadersVersion {
		p.logInfo("switching to send headers mode - requesting peer %s to send us headers directly instead of inv msg", p)
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
	return p.addrS
}

func (p *Peer) logDebug(format string, v ...any) {
	p.log.Debug().Str("peer", p.String()).Bool("inbound", p.inbound).Msgf(format, v...)
}

func (p *Peer) logInfo(format string, v ...any) {
	p.log.Info().Str("peer", p.String()).Bool("inbound", p.inbound).Msgf(format, v...)
}

func (p *Peer) logWrn(format string, v ...any) {
	p.log.Warn().Str("peer", p.String()).Bool("inbound", p.inbound).Msgf(format, v...)
}

func (p *Peer) logError(format string, v ...any) {
	p.log.Error().Str("peer", p.String()).Bool("inbound", p.inbound).Msgf(format, v...)
}
