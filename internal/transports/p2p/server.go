package p2pexp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/network"
	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/rs/zerolog"
)

type server struct {
	config         *config.P2PConfig
	chainParams    *chaincfg.Params
	headersService service.Headers
	chainService   service.Chains
	log            *zerolog.Logger

	outboundPeers *peer.PeersCollection
	inboundPeers  *peer.PeersCollection
	addresses     *network.AddressBook

	ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewServer(
	config *config.P2PConfig,
	chainParams *chaincfg.Params,
	headersService service.Headers,
	chainService service.Chains,
	log *zerolog.Logger,
) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	ctx, ctxCancel := context.WithCancel(context.Background())
	server := &server{
		config:         config,
		chainParams:    chainParams,
		headersService: headersService,
		chainService:   chainService,
		log:            &serverLogger,

		outboundPeers: peer.NewPeersCollection(config.MaxOutboundConnections),
		inboundPeers:  peer.NewPeersCollection(config.MaxInboundConnections),
		addresses:     network.NewAdressbook(time.Hour*time.Duration(config.BanDuration), config.AcceptLocalPeers),

		ctx:       ctx,
		ctxCancel: ctxCancel,
	}

	return server
}

func (s *server) Start() error {
	err := s.connectOutboundPeers()
	if err != nil {
		return err
	}

	err = s.listenInboundPeers()
	if err != nil {
		s.log.Info().Msg(" shutdown p2p server")
		s.Shutdown()
	}
	return nil
}

func (s *server) Shutdown() error {
	s.ctxCancel()

	for _, p := range s.outboundPeers.Enumerate() {
		p.Disconnect()
	}

	for _, p := range s.inboundPeers.Enumerate() {
		p.Disconnect()
	}

	return nil
}

func (s *server) connectOutboundPeers() error {
	seeds := seedFromDNS(s.chainParams.DNSSeeds, s.log)
	if len(seeds) == 0 {
		return errors.New("no seeds found")
	}

	if len(seeds) > int(s.config.MaxOutboundConnections) {
		seeds = seeds[:s.config.MaxOutboundConnections]
	}

	for _, seed := range seeds {
		s.log.Debug().Msgf("got peer addr: %s", seed.String())
	}

	peersCounter := 0
	for _, addr := range seeds {
		if err := s.connectToAddr(addr, s.chainParams.DefaultPort); err != nil {
			continue
		}

		peersCounter++
	}

	if peersCounter == 0 {
		return errors.New("cannot connect to any peers from seed")
	}

	s.log.Info().Msgf("connected to %d peers", peersCounter)

	go s.observeOutboundPeers()
	return nil
}

func (s *server) listenInboundPeers() error {
	ourAddr := net.JoinHostPort("", fmt.Sprintf("%d", s.chainParams.DefaultPort))
	listener, err := net.Listen("tcp", ourAddr)
	if err != nil {
		s.log.Error().Msgf("error creating listener, reason: %v", err)
		return err
	}

	go s.observeInboundPeers(listener)
	return nil
}

func (s *server) connectToAddr(addr net.IP, port uint16) error {
	netAddr := &net.TCPAddr{
		IP:   addr,
		Port: int(port),
	}

	conn, err := net.Dial(netAddr.Network(), netAddr.String())
	if err != nil {
		s.log.Error().Str("peer", netAddr.String()).
			Bool("inbound", false).
			Msgf("error connecting with peer, reason: %v", err)

		s.log.Info().Str("peer", netAddr.String()).
			Bool("inbound", false).
			Msgf("peer banned, reason: %v", err)

		s.addresses.BanAddr(&wire.NetAddress{IP: netAddr.IP, Port: uint16(netAddr.Port)})
		return err
	}

	peer, err := s.connectPeer(conn, false)
	if err != nil {
		s.log.Info().Str("peer", netAddr.String()).
			Bool("inbound", false).
			Msgf("peer banned, reason: %v", err)

		s.addresses.BanAddr(&wire.NetAddress{IP: netAddr.IP, Port: uint16(netAddr.Port)})
		return err
	}

	s.outboundPeers.AddPeer(peer)
	s.addresses.UpsertPeerAddr(peer)
	return nil
}

func (s *server) connectPeer(conn net.Conn, inbound bool) (*peer.Peer, error) {
	peer := peer.NewPeer(conn, inbound, s.config, s.chainParams, s.headersService, s.chainService, s.log, s)

	err := peer.Connect()
	if err != nil {
		peer.Disconnect()
		s.log.Error().Str("peer", peer.String()).
			Bool("inbound", inbound).
			Msgf("error connecting with peer, reason: %v", err)
		return nil, err
	}

	peer.SendGetAddrInfo()

	if !inbound {
		err = peer.StartHeadersSync()
		if err != nil {
			peer.Disconnect()
			return nil, err
		}
	}

	return peer, nil
}

func (s *server) observeOutboundPeers() {
	const sleepMinutes = 5
	time.Sleep(sleepMinutes * time.Minute)

	for {
		select {
		case <-s.ctx.Done(): // exit if context was cancaled
			s.log.Info().Msg("[observeOutboundPeers] exit")
			return

		default:
			peersToConnect := s.outboundPeers.Space()
			if peersToConnect == 0 {
				s.log.Debug().Msg("[observeOutboundPeers] nothing to do")
				time.Sleep(sleepMinutes * time.Minute)
				continue
			}

			s.log.Info().Msgf("try connect with %d new peers", peersToConnect)

			for ; peersToConnect > 0; peersToConnect-- {
				s.connectToRandomAddr()
			}
		}
	}
}

func (s *server) connectToRandomAddr() {
	const tries = 20

	for i := 0; i < tries; i++ {
		addr := s.addresses.GetRndUnusedAddr(tries)
		if addr == nil {
			s.log.Warn().Msgf("[observeOutboundPeers] coudnt find random unused/unbanned peer address with %d tries", tries)
			continue
		}

		if err := s.connectToAddr(addr.IP, addr.Port); err != nil {
			continue // try one more time
		}

		return // success
	}
}

func (s *server) observeInboundPeers(listener net.Listener) {
	const sleepMinutes = 5
	time.Sleep(sleepMinutes * time.Minute)

	for {
		select {
		case <-s.ctx.Done(): // exit if context was canceled
			s.log.Info().Msg("[observeInboundPeers] exit")
			return

		default:
			if s.inboundPeers.Space() == 0 {
				s.log.Debug().Msg("[observeInboundPeers] nothing to do")
				time.Sleep(sleepMinutes * time.Minute)
				continue
			}

			s.log.Info().Msgf("listening for inbound connections on port %d", s.chainParams.DefaultPort)
			s.waitForIncomingConnection(listener)
		}
	}
}

func (s *server) waitForIncomingConnection(listener net.Listener) {
	conn, err := listener.Accept() // should we check if is the same adress already connected? or it's outbounded peer?
	if err != nil {
		s.log.Error().Msgf("error accepting connection, reason: %v", err)
		return
	}

	peer, err := s.connectPeer(conn, true)
	if err != nil {
		return
	}

	s.inboundPeers.AddPeer(peer)
	s.addresses.UpsertPeerAddr(peer)
}

func (s *server) AddAddrs(address []*wire.NetAddress) {
	s.addresses.UpsertAddrs(address)
}

func (s *server) SignalError(p *peer.Peer, err error) {
	// handle error and decide what to do with the peer

	s.log.Info().Str("peer", p.String()).
		Bool("inbound", p.Inbound()).
		Msgf("peer banned, reason: %v", err)

	s.addresses.BanAddr(p.GetPeerAddr())

	p.Disconnect()
	s.outboundPeers.RmPeer(p)
	s.inboundPeers.RmPeer(p)
}
