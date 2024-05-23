// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bitcoin-sv/block-headers-service/logging"
	"github.com/rs/zerolog"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/service"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/addrmgr"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/connmgr"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/p2psync"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/p2putil"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
	"github.com/kr/pretty"
)

const (
	// defaultServices describes the default services that are supported by
	// the server.
	defaultServices = wire.SFspv

	// defaultRequiredServices describes the default services that are
	// required to be supported by outbound peers.
	defaultRequiredServices = wire.SFspv

	// connectionRetryInterval is the base amount of time to wait in between
	// retries when connecting to persistent peers.  It is adjusted by the
	// number of retries such that there is a retry backoff.
	connectionRetryInterval = time.Second * 5
)

// ServerAlreadyStarted represents starting error when a p2p server is already started.
var ServerAlreadyStarted = errors.New("p2p server already started")

type sampledLoggers struct {
	query     *zerolog.Logger
	donePeers *zerolog.Logger
}

// makeSampledLoggers creates special loggers that reduce the number of logs.
func makeSampledLoggers(baseLogger *zerolog.Logger) sampledLoggers {
	return sampledLoggers{
		query:     logging.SampledLogger(baseLogger, 2000),
		donePeers: logging.SampledLogger(baseLogger, 100),
	}
}

// addrMe specifies the server address to send peers.
var addrMe *wire.NetAddress

// simpleAddr implements the net.Addr interface with two struct fields.
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}

// broadcastMsg provides the ability to house a bitcoin message to be broadcast
// to all connected peers except specified excluded peers.
type broadcastMsg struct {
	message      wire.Message
	excludePeers []*serverPeer
}

// relayMsg packages an inventory vector along with the newly discovered
// inventory so the relay has access to that information.
type relayMsg struct {
	invVect *wire.InvVect
	data    interface{}
}

// updatePeerHeightsMsg is a message sent from the blockmanager to the server
// after a new block has been accepted. The purpose of the message is to update
// the heights of peers that were known to announce the block before we
// connected it to the main chain or recognized it as an orphan. With these
// updates, peer heights will be kept up to date, allowing for fresh data when
// selecting sync peer candidacy.
type updatePeerHeightsMsg struct {
	newHash    *chainhash.Hash
	newHeight  int32
	originPeer *peer.Peer
}

// server provides a bitcoin server for handling communications to and from
// bitcoin peers.
type server struct {
	started     int32
	shutdown    int32
	startupTime int64

	chainParams       *chaincfg.Params
	addrManager       *addrmgr.AddrManager
	connManager       *connmgr.ConnManager
	syncManager       *p2psync.SyncManager
	newPeers          chan *serverPeer
	donePeers         chan *serverPeer
	banPeers          chan *peer.Peer
	query             chan interface{}
	relayInv          chan relayMsg
	broadcast         chan broadcastMsg
	peerHeightsUpdate chan updatePeerHeightsMsg
	wg                sync.WaitGroup
	quit              chan struct{}
	nat               NAT
	timeSource        config.MedianTimeSource
	wireServices      wire.ServiceFlag
	p2pConfig         *config.P2PConfig
	log               *zerolog.Logger
}

// handleUpdatePeerHeight updates the heights of all peers who were known to
// announce a block we recently accepted.
func (s *server) handleUpdatePeerHeights(state *peerState, umsg updatePeerHeightsMsg) {
	state.forAllPeers(func(sp *serverPeer) {
		// The origin peer should already have the updated height.
		if sp.Peer == umsg.originPeer {
			return
		}

		// This is a pointer to the underlying memory which doesn't
		// change.
		latestBlkHash := sp.LastAnnouncedBlock()

		// Skip this peer if it hasn't recently announced any new blocks.
		if latestBlkHash == nil {
			return
		}

		// If the peer has recently announced a block, and this block
		// matches our newly accepted block, then update their block
		// height.
		if *latestBlkHash == *umsg.newHash {
			sp.UpdateLastBlockHeight(umsg.newHeight)
			sp.UpdateLastAnnouncedBlock(nil)
		}
	})
}

// handleAddPeerMsg deals with adding new peers.  It is invoked from the
// peerHandler goroutine.
func (s *server) handleAddPeerMsg(state *peerState, sp *serverPeer) bool {
	if sp == nil {
		return false
	}

	// Ignore new peers if we're shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		sp.log.Info().Msgf("New peer %s ignored - server is shutting down", sp)
		sp.Disconnect()
		return false
	}

	// Disconnect banned peers.
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		sp.log.Debug().Msgf("can't split hostport %v", err)
		sp.Disconnect()
		return false
	}
	if banEnd, ok := state.banned[host]; ok {
		if time.Now().Before(banEnd) {
			sp.log.Debug().Msgf("Peer %s is banned for another %v - disconnecting", host, time.Until(banEnd))
			sp.Disconnect()
			return false
		}

		sp.log.Info().Msgf("Peer %s is no longer banned", host)
		delete(state.banned, host)
	}

	// Limit max number of total peers per ip.
	if state.CountIP(host) >= config.MaxPeersPerIP {
		sp.log.Info().Msgf("Max peers per IP reached [%d] - disconnecting peer %s", config.MaxPeersPerIP, sp)
		sp.Disconnect()

		return false
	}

	// Limit max number of total peers.
	if state.Count() >= config.MaxPeers {
		sp.log.Info().Msgf("Max peers reached [%d] - disconnecting peer %s", config.MaxPeers, sp)
		sp.Disconnect()
		// TODO: how to handle permanent peers here?
		// they should be rescheduled.
		return false
	}

	// Add the new peer and start it.
	sp.log.Debug().Msgf("New peer %s", sp)

	if sp.Inbound() {
		state.inboundPeers[sp.ID()] = sp
		state.connectionCount[host]++
	} else {
		state.outboundGroups[addrmgr.GroupKey(sp.NA())]++

		if sp.persistent {
			state.persistentPeers[sp.ID()] = sp
		} else {
			state.outboundPeers[sp.ID()] = sp
			state.connectionCount[host]++
		}
	}

	return true
}

// handleDonePeerMsg deals with peers that have signaled they are done.  It is
// invoked from the peerHandler goroutine.
func (s *server) handleDonePeerMsg(state *peerState, sp *serverPeer) {
	var list map[int32]*serverPeer

	if sp.persistent {
		list = state.persistentPeers
	} else if sp.Inbound() {
		list = state.inboundPeers
	} else {
		list = state.outboundPeers
	}

	if _, ok := list[sp.ID()]; ok {
		if !sp.Inbound() && sp.VersionKnown() {
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		}

		if !sp.Inbound() && sp.connReq != nil {
			s.connManager.Disconnect(sp.connReq.ID())
		}

		delete(list, sp.ID())

		host, _, err := net.SplitHostPort(sp.Addr())
		if err == nil && !sp.persistent {
			state.connectionCount[host]--
		}

		sp.log.Debug().Msgf("Removed peer %s", sp)
		return
	}

	if sp.connReq != nil {
		s.connManager.Disconnect(sp.connReq.ID())
	}

	// Update the address' last seen time if the peer has acknowledged
	// our version and has sent us its version as well.
	if sp.VerAckReceived() && sp.VersionKnown() && sp.NA() != nil {
		s.addrManager.Connected(sp.NA())
	}

	// If we get here it means that either we didn't know about the peer
	// or we purposefully deleted it.
}

// handleBanPeerMsg deals with banning peers.  It is invoked from the
// peerHandler goroutine.
func (s *server) handleBanPeerMsg(state *peerState, p *peer.Peer) {
	host, _, err := net.SplitHostPort(p.Addr())
	if err != nil {
		s.log.Debug().Msgf("can't split ban peer %s %v", p.Addr(), err)
		return
	}
	direction := logging.DirectionString(p.Inbound())
	s.log.Info().Msgf("Banned peer %s (%s) for %v", host, direction, s.p2pConfig.BanDuration)
	state.banned[host] = time.Now().Add(s.p2pConfig.BanDuration)
}

// handleBroadcastMsg deals with broadcasting messages to peers.  It is invoked
// from the peerHandler goroutine.
func (s *server) handleBroadcastMsg(state *peerState, bmsg *broadcastMsg) {
	state.forAllPeers(func(sp *serverPeer) {
		if !sp.Connected() {
			return
		}

		for _, ep := range bmsg.excludePeers {
			if sp == ep {
				return
			}
		}

		sp.QueueMessage(bmsg.message, nil)
	})
}

type getConnCountMsg struct {
	reply chan int32
}

type getPeersMsg struct {
	reply chan []*serverPeer
}

type getOutboundGroup struct {
	key   string
	reply chan int
}

type getAddedNodesMsg struct {
	reply chan []*serverPeer
}

type disconnectNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}

type connectNodeMsg struct {
	addr      string
	permanent bool
	reply     chan error
}

type removeNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}

// handleQuery is the central handler for all queries and commands from other
// goroutines related to peer state.
func (s *server) handleQuery(state *peerState, querymsg interface{}) {
	switch msg := querymsg.(type) {
	case getConnCountMsg:
		nconnected := int32(0)
		state.forAllPeers(func(sp *serverPeer) {
			if sp.Connected() {
				nconnected++
			}
		})
		msg.reply <- nconnected

	case getPeersMsg:
		peers := make([]*serverPeer, 0, state.Count())
		state.forAllPeers(func(sp *serverPeer) {
			if !sp.Connected() {
				return
			}
			peers = append(peers, sp)
		})
		fmt.Printf("[Server] getPeersMsg: %#v\n", pretty.Formatter(peers))
		msg.reply <- peers

	case connectNodeMsg:
		// TODO: duplicate oneshots?
		// Limit max number of total peers.
		fmt.Print("[Server] connectNodeMsg")
		if state.Count() >= config.MaxPeers {
			msg.reply <- errors.New("max peers reached")
			return
		}
		for _, peer := range state.persistentPeers {
			if peer.Addr() == msg.addr {
				if msg.permanent {
					msg.reply <- errors.New("peer already connected")
				} else {
					msg.reply <- errors.New("peer exists as a permanent peer")
				}
				return
			}
		}

		netAddr, err := p2putil.AddrStringToNetAddr(msg.addr, s.p2pConfig.BsvdLookup)
		if err != nil {
			msg.reply <- err
			return
		}

		// TODO: if too many, nuke a non-perm peer.
		go s.connManager.Connect(&connmgr.ConnReq{
			Addr:      netAddr,
			Permanent: msg.permanent,
		})
		msg.reply <- nil
	case removeNodeMsg:
		found := disconnectPeer(state.persistentPeers, msg.cmp, func(sp *serverPeer) {
			// Keep group counts ok since we remove from
			// the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})

		if found {
			msg.reply <- nil
		} else {
			msg.reply <- errors.New("peer not found")
		}
	case getOutboundGroup:
		count, ok := state.outboundGroups[msg.key]
		if ok {
			msg.reply <- count
		} else {
			msg.reply <- 0
		}
	// Request a list of the persistent (added) peers.
	case getAddedNodesMsg:
		// Respond with a slice of the relevant peers.
		peers := make([]*serverPeer, 0, len(state.persistentPeers))
		for _, sp := range state.persistentPeers {
			peers = append(peers, sp)
		}
		msg.reply <- peers
	case disconnectNodeMsg:
		// Check inbound peers. We pass a nil callback since we don't
		// require any additional actions on disconnect for inbound peers.
		found := disconnectPeer(state.inboundPeers, msg.cmp, nil)
		if found {
			msg.reply <- nil
			return
		}

		// Check outbound peers.
		found = disconnectPeer(state.outboundPeers, msg.cmp, func(sp *serverPeer) {
			// Keep group counts ok since we remove from
			// the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})
		if found {
			// If there are multiple outbound connections to the same
			// ip:port, continue disconnecting them all until no such
			// peers are found.
			for found {
				found = disconnectPeer(state.outboundPeers, msg.cmp, func(sp *serverPeer) {
					state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
				})
			}
			msg.reply <- nil
			return
		}

		msg.reply <- errors.New("peer not found")
	}
}

// disconnectPeer attempts to drop the connection of a targeted peer in the
// passed peer list. Targets are identified via usage of the passed
// `compareFunc`, which should return `true` if the passed peer is the target
// peer. This function returns true on success and false if the peer is unable
// to be located. If the peer is found, and the passed callback: `whenFound'
// isn't nil, we call it with the peer as the argument before it is removed
// from the peerList, and is disconnected from the server.
func disconnectPeer(peerList map[int32]*serverPeer, compareFunc func(*serverPeer) bool, whenFound func(*serverPeer)) bool {
	for addr, peer := range peerList {
		if compareFunc(peer) {
			if whenFound != nil {
				whenFound(peer)
			}

			// This is ok because we are not continuing
			// to iterate so won't corrupt the loop.
			delete(peerList, addr)
			peer.Disconnect()
			return true
		}
	}
	return false
}

func initUserAgentComments(excessiveBlockSize uint32) []string {
	var userAgentComments []string
	if excessiveBlockSize != 0 {
		userAgentComments = append(userAgentComments, fmt.Sprintf("EB%v", excessiveBlockSize))
	}

	return userAgentComments
}

// inboundPeerConnected is invoked by the connection manager when a new inbound
// connection is established.  It initializes a new inbound server peer
// instance, associates it with the connection, and starts a goroutine to wait
// for disconnection.
func (s *server) inboundPeerConnected(conn net.Conn, log *zerolog.Logger) {
	sp := newServerPeer(s, false, log)
	sp.Peer = peer.NewInboundPeer(newPeerConfig(sp))
	// Peer log
	// srvrconfigs.Log.Infof("[Server] inboundPeer: %#v", sp.Peer)
	sp.AssociateConnection(conn)
	go s.peerDoneHandler(sp)
}

// outboundPeerConnected is invoked by the connection manager when a new
// outbound connection is established.  It initializes a new outbound server
// peer instance, associates it with the relevant state such as the connection
// request instance and the connection itself, and finally notifies the address
// manager of the attempt.
func (s *server) outboundPeerConnected(c *connmgr.ConnReq, conn net.Conn, log *zerolog.Logger) {
	sp := newServerPeer(s, c.Permanent, log)
	p, err := peer.NewOutboundPeer(newPeerConfig(sp), c.Addr.String())
	if err != nil {
		s.log.Debug().Msgf("Cannot create outbound peer %s: %v", c.Addr, err)
		s.connManager.Disconnect(c.ID())
	}
	// Peer log
	// srvrconfigs.Log.Infof("[Server] outboundPeer: %#v", p)
	sp.Peer = p
	sp.connReq = c
	sp.AssociateConnection(conn)
	go s.peerDoneHandler(sp)
	s.addrManager.Attempt(sp.NA())
}

// peerDoneHandler handles peer disconnects by notifiying the server that it's
// done along with other performing other desirable cleanup.
func (s *server) peerDoneHandler(sp *serverPeer) {
	sp.WaitForDisconnect()
	s.donePeers <- sp

	// Only tell sync manager we are gone if we ever told it we existed.
	if sp.VersionKnown() {
		s.syncManager.DonePeer(sp.Peer, nil)
	}
	close(sp.quit)
}

// peerHandler is used to handle peer operations such as adding and removing
// peers to and from the server, banning peers, and broadcasting messages to
// peers.  It must be run in a goroutine.
func (s *server) peerHandler() {
	// Start the address manager and sync manager, both of which are needed
	// by peers.  This is done here since their lifecycle is closely tied
	// to this handler and rather than adding more channels to sychronize
	// things, it's easier and slightly faster to simply start and stop them
	// in this handler.
	s.addrManager.Start()
	s.syncManager.Start()

	s.log.Trace().Msg("Starting peer handler")
	sLogger := makeSampledLoggers(s.log)
	state := &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
		connectionCount: make(map[string]int),
	}

	// Add peers discovered through DNS to the address manager.
	s.log.Info().Msgf("[Server] configs.ActiveNetParams.Params: %#v", pretty.Formatter(config.ActiveNetParams))
	connmgr.SeedFromDNS(config.ActiveNetParams, defaultRequiredServices,
		s.p2pConfig.BsvdLookup, func(addrs []*wire.NetAddress) {
			// Bitcoind uses a lookup of the dns seeder here. This
			// is rather strange since the values looked up by the
			// DNS seed lookups will vary quite a lot.
			// to replicate this behavior we put all addresses as
			// having come from the first one.
			s.addrManager.AddAddresses(addrs, addrs[0])
		}, s.log)
	go s.connManager.Start()

out:
	for {
		select {
		// New peers connected to the server.
		case p := <-s.newPeers:
			s.log.Info().Msg("[Server] newPeers")
			s.handleAddPeerMsg(state, p)

		// Disconnected peers.
		case p := <-s.donePeers:
			sLogger.donePeers.Info().Msg("[Server] donePeers")
			s.handleDonePeerMsg(state, p)

		// Block accepted in mainchain or orphan, update peer height.
		case umsg := <-s.peerHeightsUpdate:
			s.log.Info().Msg("[Server] peerHeightsUpdate")
			s.handleUpdatePeerHeights(state, umsg)

		// Peer to ban.
		case p := <-s.banPeers:
			s.handleBanPeerMsg(state, p)

		// Message to broadcast to all connected peers except those
		// which are excluded by the message.
		case bmsg := <-s.broadcast:
			s.log.Info().Msg("[Server] broadcast")
			s.handleBroadcastMsg(state, &bmsg)

		case qmsg := <-s.query:
			sLogger.query.Info().Msg("[Server] query")
			s.handleQuery(state, qmsg)

		case <-s.quit:
			// Disconnect all peers on server shutdown.
			state.forAllPeers(func(sp *serverPeer) {
				s.log.Trace().Msgf("Shutdown peer %s", sp)
				sp.Disconnect()
			})
			break out
		}
	}

	s.connManager.Stop()
	s.syncManager.Stop()
	s.addrManager.Stop()

	// Drain channels before exiting so nothing is left waiting around
	// to send.
cleanup:
	for {
		select {
		case <-s.newPeers:
		case <-s.donePeers:
		case <-s.peerHeightsUpdate:
		case <-s.relayInv:
		case <-s.broadcast:
		case <-s.query:
		default:
			break cleanup
		}
	}
	s.wg.Done()
	s.log.Trace().Msg("Peer handler done")
}

// AddPeer adds a new peer that has already been connected to the server.
func (s *server) AddPeer(sp *serverPeer) {
	s.newPeers <- sp
}

// BanPeer bans a peer that has already been connected to the server by ip.
func (s *server) BanPeer(p *peer.Peer) {
	s.banPeers <- p
}

// RelayInventory relays the passed inventory vector to all connected peers
// that are not already known to have it.
func (s *server) RelayInventory(invVect *wire.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}

// BroadcastMessage sends msg to all peers currently connected to the server
// except those in the passed peers to exclude.
func (s *server) BroadcastMessage(msg wire.Message, exclPeers ...*serverPeer) {
	bmsg := broadcastMsg{message: msg, excludePeers: exclPeers}
	s.broadcast <- bmsg
}

// ConnectedCount returns the number of currently connected peers.
func (s *server) ConnectedCount() int32 {
	replyChan := make(chan int32)

	s.query <- getConnCountMsg{reply: replyChan}

	return <-replyChan
}

// OutboundGroupCount returns the number of peers connected to the given
// outbound group key.
func (s *server) OutboundGroupCount(key string) int {
	replyChan := make(chan int)
	s.query <- getOutboundGroup{key: key, reply: replyChan}
	return <-replyChan
}

// UpdatePeerHeights updates the heights of all peers who have have announced
// the latest connected main chain block, or a recognized orphan. These height
// updates allow us to dynamically refresh peer heights, ensuring sync peer
// selection has access to the latest block heights for each peer.
func (s *server) UpdatePeerHeights(latestBlkHash *chainhash.Hash, latestHeight int32, updateSource *peer.Peer) {
	s.peerHeightsUpdate <- updatePeerHeightsMsg{
		newHash:    latestBlkHash,
		newHeight:  latestHeight,
		originPeer: updateSource,
	}
}

// Start begins accepting connections from peers.
func (s *server) Start() error {
	// Already started?
	if atomic.AddInt32(&s.started, 1) != 1 {
		return ServerAlreadyStarted
	}

	s.log.Trace().Msg("Starting server")

	// Start the peer handler which in turn starts the address and block
	// managers.
	s.wg.Add(1)
	go s.peerHandler()

	if s.nat != nil {
		s.wg.Add(1)
		go s.upnpUpdateThread()
	}

	return nil
}

// Stop gracefully shuts down the server by stopping and disconnecting all
// peers and the main listener.
func (s *server) Stop() {
	// Make sure this only happens once.
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		s.log.Info().Msg("Server is already in the process of shutting down")
	}

	s.log.Warn().Msg("P2P Server shutting down")

	// Signal the remaining goroutines to quit.
	close(s.quit)
}

// Shutdown gracefully shuts down the server by stopping and disconnecting all
// peers and the main listener and waits for server to stop.
func (s *server) Shutdown() {
	s.log.Info().Msg("Gracefully shutting down the P2P server...")
	s.Stop()
	s.WaitForShutdown()
	s.log.Info().Msg("P2P Server shutdown complete")
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *server) WaitForShutdown() {
	s.wg.Wait()
}

func (s *server) upnpUpdateThread() {
	// Go off immediately to prevent code duplication, thereafter we renew
	// lease every 15 minutes.
	timer := time.NewTimer(0 * time.Second)
	lport := config.ActiveNetParams.DefaultPort
	first := true
out:
	for {
		select {
		case <-timer.C:
			// TODO: pick external port  more cleverly
			// TODO: know which ports we are listening to on an external net.
			// TODO: if specific listen port doesn't work then ask for wildcard
			// listen port?
			// XXX this assumes timeout is in seconds.
			listenPort, err := s.nat.AddPortMapping("tcp", int(lport), int(lport),
				"bsvd listen port", 20*60)
			if err != nil {
				s.log.Warn().Msgf("can't add UPnP port mapping: %v", err)
			}
			if first && err == nil {
				// TODO: look this up periodically to see if upnp domain changed
				// and so did ip.
				externalip, err := s.nat.GetExternalAddress()
				if err != nil {
					s.log.Warn().Msgf("UPnP can't get external address: %v", err)
					continue out
				}
				na := wire.NewNetAddressIPPort(externalip, uint16(listenPort),
					s.wireServices)
				err = s.addrManager.AddLocalAddress(na, addrmgr.UpnpPrio)
				if err != nil {
					s.log.Warn().Msgf("can't add local address: %v", err)
				}
				s.log.Warn().Msgf("Successfully bound via UPnP to %s", addrmgr.NetAddressKey(na))
				first = false
			}
			timer.Reset(time.Minute * 15)
		case <-s.quit:
			break out
		}
	}

	timer.Stop()

	if err := s.nat.DeletePortMapping("tcp", int(lport), int(lport)); err != nil {
		s.log.Warn().Msgf("unable to remove UPnP port mapping: %v", err)
	} else {
		s.log.Debug().Msg("successfully disestablished UPnP port mapping")
	}

	s.wg.Done()
}

// newServer returns a new bsvd server configured to listen on addr for the
// bitcoin network type specified by chainParams.  Use start to begin accepting
// connections from peers.
func newServer(chainParams *chaincfg.Params, services *service.Services,
	peers map[*peer.Peer]*peer.PeerSyncState, p2pCfg *config.P2PConfig, log *zerolog.Logger,
) (*server, error) {
	wireServices := defaultServices

	amgr := addrmgr.New(p2pCfg.BsvdLookup, log)

	var listeners []net.Listener
	var err error
	listeners, err = p2putil.InitListeners(log)
	if err != nil {
		return nil, err
	}
	if len(listeners) == 0 {
		return nil, errors.New("no valid listen address")
	}

	s := server{
		startupTime:       time.Now().Unix(),
		chainParams:       chainParams,
		addrManager:       amgr,
		newPeers:          make(chan *serverPeer, config.MaxPeers),
		donePeers:         make(chan *serverPeer, config.MaxPeers),
		banPeers:          make(chan *peer.Peer, config.MaxPeers),
		query:             make(chan interface{}),
		relayInv:          make(chan relayMsg, config.MaxPeers),
		broadcast:         make(chan broadcastMsg, config.MaxPeers),
		quit:              make(chan struct{}),
		peerHeightsUpdate: make(chan updatePeerHeightsMsg),
		nat:               nil,
		timeSource:        config.TimeSource,
		wireServices:      wireServices,
		p2pConfig:         p2pCfg,
		log:               log,
	}

	s.syncManager, err = p2psync.New(&p2psync.Config{
		PeerNotifier:              &s,
		ChainParams:               s.chainParams,
		DisableCheckpoints:        p2pCfg.DisableCheckpoints,
		MaxPeers:                  config.MaxPeers,
		MinSyncPeerNetworkSpeed:   config.MinSyncPeerNetworkSpeed,
		BlocksForForkConfirmation: p2pCfg.BlocksForForkConfirmation,
		Logger:                    log,
		Services:                  services,
		Checkpoints:               config.Checkpoints,
	}, peers)
	if err != nil {
		return nil, err
	}

	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners:     listeners,
		OnAccept:      s.inboundPeerConnected,
		RetryDuration: connectionRetryInterval,
		Dial:          p2pCfg.BsvdDial,
		OnConnection:  s.outboundPeerConnected,
		GetNewAddress: p2putil.NewAddressFunc(s.addrManager.GetAddress, s.OutboundGroupCount, p2pCfg.BsvdLookup),
		BanAddress:    s.addrManager.BanAddress,
		Logger:        log,
	})
	if err != nil {
		return nil, err
	}
	s.connManager = cmgr

	return &s, nil
}

// NewServer creates and return p2p server.
func NewServer(services *service.Services, peers map[*peer.Peer]*peer.PeerSyncState, p2pCfg *config.P2PConfig, log *zerolog.Logger) (*server, error) {
	serverLogger := log.With().Str("service", "p2p").Logger()
	server, err := newServer(config.ActiveNetParams, services, peers, p2pCfg, &serverLogger)
	if err != nil {
		serverLogger.Error().Msgf("Unable to start server: %v", err)
		return nil, err
	}
	return server, nil
}
