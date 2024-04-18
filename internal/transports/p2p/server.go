package p2pexp

import (
	"errors"
	"net"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/transports/p2p/peer"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/rs/zerolog"
)

type server struct {
	config         *config.P2PConfig
	chainParams    *chaincfg.Params
	headersService service.Headers
	chainService   service.Chains
	log            *zerolog.Logger

	peers []*peer.Peer
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
		peers:          make([]*peer.Peer, 0),
	}
	return server
}

func (s *server) Start() error {
	err := s.seedAndConnect()
	if err != nil {
		return err
	}

	return s.listenAndConnect()
}

func (s *server) Shutdown() error {
	for _, p := range s.peers {
		p.Disconnect()
	}
	return nil
}

func (s *server) seedAndConnect() error {
	seeds := seedFromDNS(s.chainParams.DNSSeeds, s.log)
	if len(seeds) == 0 {
		return errors.New("no seeds found")
	}

	for _, seed := range seeds {
		s.log.Info().Msgf("Got peer addr: %s", seed.String())
	}

	firstPeerSeed := seeds[0].String()
	firstPeerAddr, err := parseAddress(firstPeerSeed, s.chainParams.DefaultPort)
	if err != nil {
		s.log.Error().Msgf("error parsing peer %s address, reason: %v", firstPeerAddr.String(), err)
		return err
	}

	inbound := false
	conn, err := net.Dial(firstPeerAddr.Network(), firstPeerAddr.String())
	if err != nil {
		s.log.Error().Msgf("error connecting to peer %s, reason: %v", firstPeerAddr.String(), err)
		return err
	}

	return s.connectPeer(conn, inbound)
}

func (s *server) listenAndConnect() error {
	s.log.Info().Msgf("listening for inbound connections on port %s", s.chainParams.DefaultPort)

	ourAddr := net.JoinHostPort("", s.chainParams.DefaultPort)
	listener, err := net.Listen("tcp", ourAddr)
	if err != nil {
		s.log.Error().Msgf("error creating listener, reason: %v", err)
		return err
	}

	inbound := true
	conn, err := listener.Accept()
	if err != nil {
		s.log.Error().Msgf("error accepting connection, reason: %v", err)
		return err
	}

	return s.connectPeer(conn, inbound)
}

func (s *server) connectPeer(conn net.Conn, inbound bool) error {
	peer, err := peer.NewPeer(conn, inbound, s.config, s.chainParams, s.headersService, s.chainService, s.log)
	if err != nil {
		return err
	}

	s.peers = append(s.peers, peer)

	err = peer.Connect()
	if err != nil {
		s.log.Error().Msgf("error connecting with peer %s, reason: %v", peer, err)
		return err
	}

	err = peer.StartHeadersSync()
	if err != nil {
		s.log.Error().Msgf("error starting sync with peer %s, reason: %v", peer, err)
		return err
	}

	return nil
}
