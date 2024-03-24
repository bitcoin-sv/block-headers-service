package p2pexp

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net"
	"strconv"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/rs/zerolog"
)

type Peer struct {
	conn            net.Conn
	addr            *net.TCPAddr
	cfg             *config.P2PConfig
	chainParams     *chaincfg.Params
	log             *zerolog.Logger
	services        wire.ServiceFlag
	protocolVersion uint32
}

func NewPeer(addr string, cfg *config.P2PConfig, chainParams *chaincfg.Params, log *zerolog.Logger) (*Peer, error) {
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
	return nil
}

func (p *Peer) Disconnect() error {
	p.conn.Close()
	return nil
}

func (p *Peer) writeOurVersionMsg() error {
	n, err := rand.Int(rand.Reader, big.NewInt(9223372036854775807))
	if err != nil {
		panic(err)
	}
	nonce := n.Uint64()

	ourNA := &wire.NetAddress{
		Services:  wire.SFspv,
		Timestamp: time.Now(),
	}

	// TODO: how to get peer's services?
	theirNA := wire.NewNetAddress(p.addr, 0)

	// TODO: get newest block from DB.
	blockNum := int32(0)

	msg := wire.NewMsgVersion(ourNA, theirNA, nonce, blockNum)
	err = msg.AddUserAgent(p.cfg.UserAgentName, p.cfg.UserAgentVersion, userAgentComments)
	if err != nil {
		p.log.Error().Msgf("could not add user agent to version message, reason: %v", err)
		return err
	}

	// NOTE: it's 0 by default, so in theory we don't need
	// to set that, but it's better to be explicit. Later
	// we will negotiate the lowest common protocol.
	msg.Services = p.services

	err = wire.WriteMessage(p.conn, msg, p.protocolVersion, p.chainParams.Net)
	if err != nil {
		p.log.Error().Msgf("could not write version message to peer, reason: %v", err)
		return err
	}

	return nil
}
