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
		s.log.Debug().Msgf("got peer addr: %s", seed.String())
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
	peer := peer.NewPeer(conn, inbound, s.config, s.chainParams, s.headersService, s.chainService, s.log, s)

	err := peer.Connect()
	if err != nil {
		peer.Disconnect()
		s.log.Error().Str("peer", peer.String()).Msgf("error connecting with peer, reason: %v", err)
		return err
	}

	s.peers = append(s.peers, peer)

	if !inbound {
		err = peer.StartHeadersSync()
		if err != nil {
			peer.Disconnect()
			return err
		}
	}

	return nil
}

func (s *server) SignalError(p *peer.Peer, err error) {
	//TODO: handle error and decide what to do with the peer
	p.Disconnect()

}
