package p2pexp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
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
	outboundPeers  *peer.PeersCollection
	inboundPeers   *peer.PeersCollection
	listener       net.Listener
	addresses      *network.AddressBook

	// lifecycle properties
	ctx       context.Context
	ctxCancel context.CancelFunc
	ctxWg     sync.WaitGroup
}

// NewServer creates and initializes a new P2P server instance.
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
		outboundPeers:  peer.NewPeersCollection(config.MaxOutboundConnections),
		inboundPeers:   peer.NewPeersCollection(config.MaxInboundConnections),
		addresses:      network.NewAddressBook(config.BanDuration, config.AcceptLocalPeers),
		ctx:            ctx,
		ctxCancel:      ctxCancel,
	}

	return server
}

// Start starts the P2P server by connecting to outbound peers and listening for inbound connections.
func (s *server) Start() error {
	err := s.connectOutboundPeers()
	if err != nil {
		return err
	}

	err = s.listenInboundPeers()
	if err != nil {
		s.log.Error().Msgf("error during server start. Shutdown p2p server. Reason: %v", err)
		s.Shutdown()

		return err
	}
	return nil
}

// Shutdown gracefully shuts down the P2P server by disconnecting all peers.
func (s *server) Shutdown() {
	// Stop listening for incoming connections
	if s.listener != nil {
		_ = s.listener.Close()
	}

	// Cancel all child goroutines
	s.ctxCancel()
	s.ctxWg.Wait()

	// Disconnect active peers
	for _, p := range s.outboundPeers.Enumerate() {
		p.Disconnect()
	}
	for _, p := range s.inboundPeers.Enumerate() {
		p.Disconnect()
	}
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

	s.listener = listener
	go s.observeInboundPeers()
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

	_ = s.outboundPeers.AddPeer(peer) // don't need to check error here
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
	sleeDuration := 1 * time.Minute
	time.Sleep(sleeDuration)

	for {
		select {
		case <-s.ctx.Done(): // Exit if context was canceled
			s.log.Info().Msg("[observeOutboundPeers] context canceled -> exit")
			return
		default:
			s.ctxWg.Add(1) // Wait on shutdown

			freeSlots := s.outboundPeers.Space()
			if freeSlots == 0 {
				s.log.Debug().Msg("[observeOutboundPeers] nothing to do")
				s.noWaitingSleep(sleeDuration)
				continue
			}

			s.log.Info().Msgf("try connect with a new peer. Free slots: %d", freeSlots)
			s.connectToRandomAddr() // Connect one-by-one to gracefully handle shutdown

			s.ctxWg.Done()
		}
	}
}

func (s *server) connectToRandomAddr() {
	const tries = 20
	addr := s.addresses.GetRandUnusedAddr(tries)
	if addr == nil {
		s.log.Warn().Msgf("[observeOutboundPeers] coudnt find random unused/unbanned peer address with %d tries", tries)
		return
	}

	_ = s.connectToAddr(addr.IP, addr.Port)
}

func (s *server) observeInboundPeers() {
	sleeDuration := 1 * time.Minute
	time.Sleep(sleeDuration)

	for {
		select {
		case <-s.ctx.Done(): // Exit if context was canceled
			s.log.Info().Msg("[observeInboundPeers] context canceled -> exit")
			return
		default:
			s.ctxWg.Add(1) // Wait on shutdown

			if s.inboundPeers.Space() == 0 {
				s.log.Debug().Msg("[observeInboundPeers] nothing to do")
				s.noWaitingSleep(sleeDuration)
				continue
			}

			s.log.Info().Msgf("listening for inbound connections on port %d", s.chainParams.DefaultPort)
			s.waitForIncomingConnection() // Accept connection one-by-one to gracefully handle shutdown

			s.ctxWg.Done()
		}
	}
}

// usage MUST be preceded by `s.ctx.Wg.Add(1)`.
func (s *server) noWaitingSleep(duration time.Duration) {
	s.ctxWg.Done() // We are sleeping -> no need to wait
	time.Sleep(duration)
	s.ctxWg.Add(1) // Wake up. Wait for us
}

func (s *server) waitForIncomingConnection() {
	conn, err := s.listener.Accept()
	if err != nil {
		s.log.Error().Msgf("error accepting connection, reason: %v", err)
		return
	}

	peer, err := s.connectPeer(conn, true)
	if err != nil {
		return
	}

	_ = s.inboundPeers.AddPeer(peer)
	s.addresses.UpsertPeerAddr(peer)
}

// AddAddrs adds addresses to the address book of the P2P server. It's peer.Manager functionality.
func (s *server) AddAddrs(address []*wire.NetAddress) {
	s.addresses.UpsertAddrs(address)
}

// SignalError signals an error with a peer and takes appropriate actions such as banning the peer and disconnecting it. It's peer.Manager functionality.
func (s *server) SignalError(p *peer.Peer, err error) {
	// Handle error and decide what to do with the peer

	s.log.Info().Str("peer", p.String()).
		Bool("inbound", p.Inbound()).
		Msgf("peer banned, reason: %v", err)

	s.addresses.BanAddr(p.GetPeerAddr())

	// Disconnection here must be non-blocking to prevent deadlock within the peer logic.
	go func() {
		p.Disconnect()
		s.outboundPeers.RmPeer(p)
		s.inboundPeers.RmPeer(p)
	}()
}
