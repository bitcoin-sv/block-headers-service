package p2pexp

import (
	"errors"
	"fmt"
	"net"
	"strconv"
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

	////
	outboundPeers *peer.PeersCollection
	addresses *network.AddressBook
}

func NewServer(
	config *config.P2PConfig,
	chainParams *chaincfg.Params,
	headersService service.Headers,
	chainService service.Chains,
	log *zerolog.Logger,
) *server {
	serverLogger := log.With().Str("service", "p2p-experimental").Logger()
	server := &server{
		config:         config,
		chainParams:    chainParams,
		headersService: headersService,
		chainService:   chainService,
		log:            &serverLogger,
		addresses: network.NewAdressbook(24 * time.Hour),
	}
	server.outboundPeers = peer.NewPeersCollection(server.config.MaxOutboundConnections)
	return server
}

func (s *server) Start() error {
	err := s.connectOutboundPeers()
	if err != nil {
		return err
	}

	//return s.listenAndConnect()
	return nil
}

func (s *server) Shutdown() error {
	for _, p := range s.outboundPeers.Enumerate() {
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

	portInt, err := strconv.Atoi(s.chainParams.DefaultPort)
	if err != nil {
		return fmt.Errorf("could not parse port: %w", err)
	}

	peersCounter := 0
	for _, addr := range seeds {
		if err = s.connectToAddr(addr, portInt); err != nil {
			continue
		}

		peersCounter++
	}

	if peersCounter == 0 {
		return errors.New("cannot connect to any peers from seed")
	}

	s.log.Info().Msgf("connected to %d peers", peersCounter)
	return nil
}

// func (s *server) listenAndConnect() error {
// 	s.log.Info().Msgf("listening for inbound connections on port %s", s.chainParams.DefaultPort)

// 	ourAddr := net.JoinHostPort("", s.chainParams.DefaultPort)
// 	listener, err := net.Listen("tcp", ourAddr)
// 	if err != nil {
// 		s.log.Error().Msgf("error creating listener, reason: %v", err)
// 		return err
// 	}

// 	conn, err := listener.Accept()
// 	if err != nil {
// 		s.log.Error().Msgf("error accepting connection, reason: %v", err)
// 		return err
// 	}

// 	inbound := true
// 	return s.connectPeer(conn, inbound)
// }

func (s *server) connectToAddr(addr net.IP, port int) error {
	netAddr := &net.TCPAddr{
		IP:   addr,
		Port: port,
	}

	conn, err := net.Dial(netAddr.Network(), netAddr.String())
	if err != nil {
		s.log.Error().Str("peer", netAddr.String()).
			Msgf("error connecting with peer, reason: %v", err)

		s.log.Info().Str("peer", netAddr.String()).
			Msgf("peer banned, reason: %v", err)

		s.addresses.BanAddr(&wire.NetAddress{IP: netAddr.IP, Port: uint16(netAddr.Port)})
		return err
	}

	peer, err := s.connectPeer(conn, false)
	if err != nil {
		s.log.Info().Str("peer", netAddr.String()).
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
		s.log.Error().Str("peer", peer.String()).Msgf("error connecting with peer, reason: %v", err)
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


func (s *server) AddAddrs(address []*wire.NetAddress) {
	s.addresses.AddAddrs(address)
}

func (s *server) SignalError(p *peer.Peer, err error) {
	// handle error and decide what to do with the peer

	s.log.Info().Str("peer", p.String()).Msgf("peer banned, reason: %v", err)
	s.addresses.BanAddr(p.GetPeerAddr())

	p.Disconnect()
	s.outboundPeers.RmPeer(p)
}
