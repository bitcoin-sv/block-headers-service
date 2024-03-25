package p2pexp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/rs/zerolog"
)

type Peer struct {
	conn            net.Conn
	addr            *net.TCPAddr
	cfg             *config.P2PConfig
	chainParams     *chaincfg.Params
	headersService  service.Headers
	log             *zerolog.Logger
	services        wire.ServiceFlag
	protocolVersion uint32
	nonce           uint64
	lastBlock       int32
	startingHeight  int32
	timeOffset      int64
	userAgent       string
	verAckReceived  bool
}

func NewPeer(addr string, cfg *config.P2PConfig, chainParams *chaincfg.Params, headersService service.Headers, log *zerolog.Logger) (*Peer, error) {
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

	peer := &Peer{
		addr:            netAddr,
		cfg:             cfg,
		chainParams:     chainParams,
		headersService:  headersService,
		log:             log,
		services:        wire.SFspv,
		protocolVersion: initialProtocolVersion,
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
	p.log.Info().Msgf("connected to peer: %s", p.addr.String())
	return nil
}

func (p *Peer) Disconnect() error {
	p.log.Info().Msgf("disconnecting peer: %s", p.addr.String())
	p.conn.Close()
	return nil
}

func (p *Peer) NegotiateProtocol() error {
	err := p.writeOurVersionMsg()
	if err != nil {
		return err
	}
	p.log.Info().Msgf("version sent to peer: %s", p.addr.String())

	err = p.readVerOrVerAck()
	if err != nil {
		return err
	}

	err = p.readVerOrVerAck()
	if err != nil {
		return err
	}
	p.log.Info().Msgf("received version and verack from peer: %s", p.addr.String())

	err = p.writeMessage(wire.NewMsgVerAck())
	if err != nil {
		return err
	}

	p.log.Info().Msgf("protocol negotiated successfully with peer: %s", p.addr.String())
	return nil
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
