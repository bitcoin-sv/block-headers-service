// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2p

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
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
	peerpkg "github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
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

	// userAgentName is the user agent name and is used to help identify
	// ourselves to other bitcoin peers.
	userAgentName = "/block-headers-service"

	// userAgentVersion is the user agent version and is used to help
	// identify ourselves to other bitcoin peers.
	// userAgentVersion = fmt.Sprintf("%d.%d.%d", version.AppMajor, version.AppMinor, version.AppPatch).
	userAgentVersion = "0.3.0"
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

// onionAddr implements the net.Addr interface and represents a tor address.
type onionAddr struct {
	addr string
}

// String returns the onion address.
//
// This is part of the net.Addr interface.
func (oa *onionAddr) String() string {
	return oa.addr
}

// Network returns "onion".
//
// This is part of the net.Addr interface.
func (oa *onionAddr) Network() string {
	return "onion"
}

// Ensure onionAddr implements the net.Addr interface.
var _ net.Addr = (*onionAddr)(nil)

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

// peerState maintains state of inbound, persistent, outbound peers as well
// as banned peers and outbound groups.
type peerState struct {
	inboundPeers    map[int32]*serverPeer
	outboundPeers   map[int32]*serverPeer
	persistentPeers map[int32]*serverPeer
	banned          map[string]time.Time
	outboundGroups  map[string]int
	connectionCount map[string]int
}

// Count returns the count of all known peers.
func (ps *peerState) Count() int {
	return len(ps.inboundPeers) + len(ps.outboundPeers) +
		len(ps.persistentPeers)
}

// CountIP returns the count of all peers matching the IP.
func (ps *peerState) CountIP(host string) int {
	return ps.connectionCount[host]
}

// forAllOutboundPeers is a helper function that runs closure on all outbound
// peers known to peerState.
func (ps *peerState) forAllOutboundPeers(closure func(sp *serverPeer)) {
	for _, e := range ps.outboundPeers {
		closure(e)
	}
	for _, e := range ps.persistentPeers {
		closure(e)
	}
}

// forAllPeers is a helper function that runs closure on all peers known to
// peerState.
func (ps *peerState) forAllPeers(closure func(sp *serverPeer)) {
	for _, e := range ps.inboundPeers {
		closure(e)
	}
	ps.forAllOutboundPeers(closure)
}

// server provides a bitcoin server for handling communications to and from
// bitcoin peers.
type server struct {
	started       int32
	shutdown      int32
	shutdownSched int32
	startupTime   int64

	chainParams          *chaincfg.Params
	addrManager          *addrmgr.AddrManager
	connManager          *connmgr.ConnManager
	syncManager          *p2psync.SyncManager
	modifyRebroadcastInv chan interface{}
	newPeers             chan *serverPeer
	donePeers            chan *serverPeer
	banPeers             chan *peer.Peer
	query                chan interface{}
	relayInv             chan relayMsg
	broadcast            chan broadcastMsg
	peerHeightsUpdate    chan updatePeerHeightsMsg
	wg                   sync.WaitGroup
	quit                 chan struct{}
	nat                  NAT
	timeSource           config.MedianTimeSource
	wireServices         wire.ServiceFlag
	p2pConfig            *config.P2PConfig
	log                  *zerolog.Logger
}

// serverPeer extends the peer to maintain state shared by the server and
// the blockmanager.
type serverPeer struct {
	numberOfFields       uint64
	maxRecvPayloadLength uint32

	*peer.Peer

	connReq        *connmgr.ConnReq
	server         *server
	persistent     bool
	sentAddrs      bool
	addrMtx        sync.RWMutex
	knownAddresses map[string]struct{}
	quit           chan struct{}
	log            *zerolog.Logger
}

// newServerPeer returns a new serverPeer instance. The peer needs to be set by
// the caller.
func newServerPeer(s *server, isPersistent bool, log *zerolog.Logger) *serverPeer {
	serverPeerLogger := log.With().Str("p2pModule", "server-peer").Logger()
	return &serverPeer{
		server:         s,
		persistent:     isPersistent,
		knownAddresses: make(map[string]struct{}),
		quit:           make(chan struct{}),
		log:            &serverPeerLogger,
	}
}

// newestBlock returns the current best block hash and height using the format
// required by the configuration for the peer package.
func (sp *serverPeer) newestBlock() (*chainhash.Hash, int32, error) {
	header := sp.server.syncManager.Services.Headers.GetTip()
	if header == nil {
		return nil, 0, nil
	}

	return &header.Hash, header.Height, nil
}

// addKnownAddresses adds the given addresses to the set of known addresses to
// the peer to prevent sending duplicate addresses.
func (sp *serverPeer) addKnownAddresses(addresses []*wire.NetAddress) {
	sp.addrMtx.Lock()
	defer sp.addrMtx.Unlock()
	for _, na := range addresses {
		sp.knownAddresses[addrmgr.NetAddressKey(na)] = struct{}{}
	}
}

// addressKnown true if the given address is already known to the peer.
func (sp *serverPeer) addressKnown(na *wire.NetAddress) bool {
	sp.addrMtx.RLock()
	defer sp.addrMtx.RUnlock()
	_, exists := sp.knownAddresses[addrmgr.NetAddressKey(na)]
	return exists
}

// pushAddrMsg sends an addr message to the connected peer using the provided
// addresses.
func (sp *serverPeer) pushAddrMsg(addresses []*wire.NetAddress) {
	// Filter addresses already known to the peer.
	addrs := make([]*wire.NetAddress, 0, len(addresses))
	for _, addr := range addresses {
		if !sp.addressKnown(addr) {
			addrs = append(addrs, addr)
		}
	}
	known, err := sp.PushAddrMsg(addrs)
	if err != nil {
		sp.server.log.Error().Msgf("Can't push address message to %s: %v", sp.Peer, err)
		sp.Disconnect()
		return
	}
	sp.addKnownAddresses(known)
}

// OnVersion is invoked when a peer receives a version bitcoin message
// and is used to negotiate the protocol version details as well as kick start
// the communications.
func (sp *serverPeer) OnVersion(_ *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
	sp.log.Info().Msgf("[Server] msg.ProtocolVersion: %d", msg.ProtocolVersion)
	// Update the address manager with the advertised services for outbound
	// connections in case they have changed.  This is not done for inbound
	// connections to help prevent malicious behavior and is skipped when
	// running on the simulation test network since it is only intended to
	// connect to specified peers and actively avoids advertising and
	// connecting to discovered peers.
	//
	// NOTE: This is done before rejecting peers that are too old to ensure
	// it is updated regardless in the case a new minimum protocol version is
	// enforced and the remote node has not upgraded yet.
	isInbound := sp.Inbound()
	remoteAddr := sp.NA()
	addrManager := sp.server.addrManager
	if !isInbound {
		addrManager.SetServices(remoteAddr, msg.Services)
	}

	// Ignore peers that have a protcol version that is too old.  The peer
	// negotiation logic will disconnect it after this callback returns.
	if msg.ProtocolVersion < int32(peer.MinAcceptableProtocolVersion) {
		return nil
	}

	// Ignore peers that aren't running Bitcoin
	sp.log.Info().Msgf("[Server] OnVersion msg.UserAgent: %s", msg.UserAgent)
	if strings.Contains(msg.UserAgent, "ABC") || strings.Contains(msg.UserAgent, "BUCash") || strings.Contains(msg.UserAgent, "Cash") {
		sp.log.Debug().Msgf("Rejecting peer %s for not running Bitcoin", sp.Peer)
		reason := "Sorry, you are not running Bitcoin"
		return wire.NewMsgReject(msg.Command(), wire.RejectNonstandard, reason)
	}

	// Update the address manager and request known addresses from the
	// remote peer for outbound connections.  This is skipped when running
	// on the simulation test network since it is only intended to connect
	// to specified peers and actively avoids advertising and connecting to
	// discovered peers.
	if !isInbound {
		// Advertise the local address when the server accepts incoming
		// connections and it believes itself to be close to the best known tip.
		if sp.server.syncManager.IsCurrent() {
			// Get address that best matches.
			lna := addrManager.GetBestLocalAddress(remoteAddr)
			if addrmgr.IsRoutable(lna) {
				// Filter addresses the peer already knows about.
				addresses := []*wire.NetAddress{lna}
				sp.pushAddrMsg(addresses)
			}
		}

		// Request known addresses if the server address manager needs
		// more and the peer has a protocol version new enough to
		// include a timestamp with addresses.
		hasTimestamp := sp.ProtocolVersion() >= wire.NetAddressTimeVersion
		if addrManager.NeedMoreAddresses() && hasTimestamp {
			sp.QueueMessage(wire.NewMsgGetAddr(), nil)
		}

		// Mark the address as a known good address.
		addrManager.Good(remoteAddr)
	}

	// Add the remote peer time as a sample for creating an offset against
	// the local clock to keep the network time in sync.
	sp.server.timeSource.AddTimeSample(sp.Addr(), msg.Timestamp)

	// Signal the sync manager this peer is a new sync candidate.
	sp.server.syncManager.NewPeer(sp.Peer, nil)

	// Add valid peer to the server.
	sp.server.AddPeer(sp)
	return nil
}

// OnInv is invoked when a peer receives an inv bitcoin message and is
// used to examine the inventory being advertised by the remote peer and react
// accordingly.  We pass the message down to blockmanager which will call
// QueueMessage with any appropriate responses.
func (sp *serverPeer) OnInv(_ *peer.Peer, msg *wire.MsgInv) {
	if len(msg.InvList) > 0 {
		sp.server.syncManager.QueueInv(msg, sp.Peer)
	}
}

// OnHeaders is invoked when a peer receives a headers bitcoin
// message.  The message is passed down to the sync manager.
func (sp *serverPeer) OnHeaders(_ *peer.Peer, msg *wire.MsgHeaders) {
	sp.server.syncManager.QueueHeaders(msg, sp.Peer)
}

// OnGetHeaders is invoked when a peer receives a getheaders bitcoin
// message.
func (sp *serverPeer) OnGetHeaders(_ *peer.Peer, msg *wire.MsgGetHeaders) {
	sp.log.Info().Msg("[Server] OnGetHeaders")
	// Ignore getheaders requests if not in sync.
	if !sp.server.syncManager.IsCurrent() {
		return
	}

	// Find the most recent known block in the best chain based on the block
	// locator and fetch all of the headers after it until either
	// wire.MaxBlockHeadersPerMsg have been fetched or the provided stop
	// hash is encountered.
	//
	// Use the block after the genesis block if no other blocks in the
	// provided locator are known.  This does mean the client will start
	// over with the genesis block if unknown block locators are provided.
	//
	// This mirrors the behavior in the reference implementation.
	headers := sp.server.syncManager.Services.Headers.LocateHeaders(msg.BlockLocatorHashes, &msg.HashStop)

	// Send found headers to the requesting peer.
	blockHeaders := make([]*wire.BlockHeader, len(headers))
	for i := range headers {
		blockHeaders[i] = &headers[i]
	}
	sp.QueueMessage(&wire.MsgHeaders{Headers: blockHeaders}, nil)
}

// OnProtoconf is invoked when a peer receives a protoconf bitcoin message and
// is used by remote peers to confirm protocol parameters .
func (sp *serverPeer) OnProtoconf(_ *peer.Peer, msg *wire.MsgProtoconf) {
	// Check that the num of fields.
	if msg.NumberOfFields == 0 {
		atomic.StoreUint64(&sp.numberOfFields, msg.NumberOfFields)
		return
	}
	if msg.NumberOfFields == 1 {
		atomic.StoreUint64(&sp.numberOfFields, msg.NumberOfFields)
		atomic.StoreUint32(&sp.maxRecvPayloadLength, msg.MaxRecvPayloadLength)
		return
	}

	sp.log.Debug().Msgf("Peer %v sent an invalid protoconf '%v' -- "+
		"disconnecting", sp, p2putil.Amount(msg.NumberOfFields))
	sp.Disconnect()
}

// OnGetAddr is invoked when a peer receives a getaddr bitcoin message
// and is used to provide the peer with known addresses from the address
// manager.
func (sp *serverPeer) OnGetAddr(_ *peer.Peer, msg *wire.MsgGetAddr) {
	sp.log.Info().Msgf("[Server] OnGetAddr msg: %#v", msg)

	// Do not accept getaddr requests from outbound peers.  This reduces
	// fingerprinting attacks.
	if !sp.Inbound() {
		sp.log.Debug().Msgf("Ignoring getaddr request from outbound peer %v", sp)
		return
	}

	// Only allow one getaddr request per connection to discourage
	// address stamping of inv announcements.
	if sp.sentAddrs {
		sp.log.Debug().Msgf("Ignoring repeated getaddr request from peer %v", sp)
		return
	}
	sp.sentAddrs = true

	// Get the current known addresses from the address manager.
	addrCache := sp.server.addrManager.AddressCache()

	// Add our best net address for peers to discover us. If the port
	// is 0 that indicates no worthy address was found, therefore
	// we do not broadcast it. We also must trim the cache by one
	// entry if we insert a record to prevent sending past the max
	// send size.
	bestAddress := sp.server.addrManager.GetBestLocalAddress(sp.NA())
	if bestAddress.Port != 0 {
		if len(addrCache) > 0 {
			addrCache = addrCache[1:]
		}
		addrCache = append(addrCache, bestAddress)
	}

	// Push the addresses.
	sp.pushAddrMsg(addrCache)
}

// OnAddr is invoked when a peer receives an addr bitcoin message and is
// used to notify the server about advertised addresses.
func (sp *serverPeer) OnAddr(_ *peer.Peer, msg *wire.MsgAddr) {
	// Ignore old style addresses which don't include a timestamp.
	if sp.ProtocolVersion() < wire.NetAddressTimeVersion {
		return
	}

	// A message that has no addresses is invalid.
	if len(msg.AddrList) == 0 {
		sp.log.Error().Msgf("Command [%s] from %s does not contain any addresses", msg.Command(), sp.Peer)
		sp.Disconnect()
		return
	}

	for _, na := range msg.AddrList {
		// Don't add more address if we're disconnecting.
		if !sp.Connected() {
			return
		}

		// Set the timestamp to 5 days ago if it's more than 24 hours
		// in the future so this address is one of the first to be
		// removed when space is needed.
		now := time.Now()
		if na.Timestamp.After(now.Add(time.Minute * 10)) {
			na.Timestamp = now.Add(-1 * time.Hour * 24 * 5)
		}

		// Add address to known addresses for this peer.
		sp.addKnownAddresses([]*wire.NetAddress{na})
		sp.log.Debug().Msgf("Added address %s from host %s", na.IP.String(), sp.NA().IP.String())
	}

	// Add addresses to server address manager.  The address manager handles
	// the details of things such as preventing duplicate addresses, max
	// addresses, and last seen updates.
	// XXX bitcoind gives a 2 hour time penalty here, do we want to do the
	// same?
	sp.server.addrManager.AddAddresses(msg.AddrList, sp.NA())
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

		netAddr, err := addrStringToNetAddr(msg.addr, s.p2pConfig)
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

// newPeerConfig returns the configuration for the given serverPeer.
func newPeerConfig(sp *serverPeer) *peer.Config {
	return &peer.Config{
		Listeners: peer.MessageListeners{
			OnVersion:    sp.OnVersion,
			OnInv:        sp.OnInv,
			OnHeaders:    sp.OnHeaders,
			OnGetHeaders: sp.OnGetHeaders,
			OnGetAddr:    sp.OnGetAddr,
			OnAddr:       sp.OnAddr,
			OnProtoconf:  sp.OnProtoconf,
		},
		Log:               sp.log,
		AddrMe:            addrMe,
		NewestBlock:       sp.newestBlock,
		HostToNetAddress:  sp.server.addrManager.HostToNetAddress,
		UserAgentName:     userAgentName,
		UserAgentVersion:  userAgentVersion,
		UserAgentComments: initUserAgentComments(config.ExcessiveBlockSize),
		ChainParams:       sp.server.chainParams,
		Services:          sp.server.wireServices,
		ProtocolVersion:   uint32(70013),
		TrickleInterval:   config.TrickleLinterval,
	}
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
	fmt.Printf("[Server] configs.ActiveNetParams.Params: %#v", pretty.Formatter(config.ActiveNetParams.Params))
	connmgr.SeedFromDNS(config.ActiveNetParams.Params, defaultRequiredServices,
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
func (s *server) Shutdown() error {
	s.log.Info().Msg("Gracefully shutting down the P2P server...")
	s.Stop()
	s.WaitForShutdown()
	s.log.Info().Msg("P2P Server shutdown complete")
	return nil
}

// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (s *server) WaitForShutdown() {
	s.wg.Wait()
}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalised.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

func (s *server) upnpUpdateThread() {
	// Go off immediately to prevent code duplication, thereafter we renew
	// lease every 15 minutes.
	timer := time.NewTimer(0 * time.Second)
	lport, _ := strconv.ParseInt(config.ActiveNetParams.DefaultPort, 10, 16)
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
	peers map[*peerpkg.Peer]*peerpkg.PeerSyncState, p2pCfg *config.P2PConfig, log *zerolog.Logger,
) (*server, error) {
	wireServices := defaultServices

	amgr := addrmgr.New(p2pCfg.BsvdLookup, log)

	var listeners []net.Listener
	var err error
	listeners, err = initListeners(log)
	if err != nil {
		return nil, err
	}
	if len(listeners) == 0 {
		return nil, errors.New("no valid listen address")
	}

	initErr := services.Headers.InsertGenesisHeaderInDatabase()
	if initErr != nil {
		return nil, initErr
	}

	s := server{
		startupTime:          time.Now().Unix(),
		chainParams:          chainParams,
		addrManager:          amgr,
		newPeers:             make(chan *serverPeer, config.MaxPeers),
		donePeers:            make(chan *serverPeer, config.MaxPeers),
		banPeers:             make(chan *peer.Peer, config.MaxPeers),
		query:                make(chan interface{}),
		relayInv:             make(chan relayMsg, config.MaxPeers),
		broadcast:            make(chan broadcastMsg, config.MaxPeers),
		quit:                 make(chan struct{}),
		modifyRebroadcastInv: make(chan interface{}),
		peerHeightsUpdate:    make(chan updatePeerHeightsMsg),
		nat:                  nil,
		timeSource:           config.TimeSource,
		wireServices:         wireServices,
		p2pConfig:            p2pCfg,
		log:                  log,
	}

	if err != nil {
		return nil, err
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
	// Peer log
	// srvrconfigs.Log.Infof("[Server] syncManager: %#v", s.syncManager)
	// Only setup a function to return new addresses to connect to when
	// not running in connect-only mode.  The simulation network is always
	// in connect-only mode since it is only intended to connect to
	// specified peers and actively avoid advertising and connecting to
	// discovered peers in order to prevent it from becoming a public test
	// network.
	newAddressFunc := func() (net.Addr, error) {
		for tries := 0; tries < 100; tries++ {
			addr := s.addrManager.GetAddress()
			// Peer log
			// srvrconfigs.Log.Infof("[Server] newServer addr: %#v", addr)
			if addr == nil {
				break
			}

			// Address will not be invalid, local or unroutable
			// because addrmanager rejects those on addition.
			// Just check that we don't already have an address
			// in the same group so that we are not connecting
			// to the same network segment at the expense of
			// others.
			key := addrmgr.GroupKey(addr.NetAddress())
			if s.OutboundGroupCount(key) != 0 {
				continue
			}

			// only allow recent nodes (10mins) after we failed 30
			// times
			if tries < 30 && time.Since(addr.LastAttempt()) < 10*time.Minute {
				continue
			}

			// allow nondefault ports after 50 failed tries.
			if tries < 50 && fmt.Sprintf("%d", addr.NetAddress().Port) !=
				config.ActiveNetParams.DefaultPort {
				continue
			}

			addrString := addrmgr.NetAddressKey(addr.NetAddress())
			return addrStringToNetAddr(addrString, s.p2pConfig)
		}

		return nil, errors.New("no valid connect address")
	}

	cmgr, err := connmgr.New(&connmgr.Config{
		Listeners:     listeners,
		OnAccept:      s.inboundPeerConnected,
		RetryDuration: connectionRetryInterval,
		Dial:          p2pCfg.BsvdDial,
		OnConnection:  s.outboundPeerConnected,
		GetNewAddress: newAddressFunc,
		BanAddress:    s.addrManager.BanAddress,
		Logger:        log,
	})
	if err != nil {
		return nil, err
	}
	s.connManager = cmgr

	return &s, nil
}

// initListeners initializes the configured net listeners and adds any bound
// addresses to the address manager. Returns the listeners and a NAT interface,
// which is non-nil if UPnP is in use.
func initListeners(log *zerolog.Logger) ([]net.Listener, error) {
	listenAddrs := prepareListeners()

	// Listen for TCP connections at the configured addresses
	netAddrs, err := parseListeners(listenAddrs)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := net.Listen(addr.Network(), addr.String())
		if err != nil {
			log.Warn().Msgf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	if err != nil {
		log.Error().Msgf("Can not parse default port %s for active chain: %v",
			config.ActiveNetParams.DefaultPort, err)
		return nil, err
	}

	return listeners, nil
}

func prepareListeners() []string {
	listeners := []string{
		net.JoinHostPort("", config.ActiveNetParams.DefaultPort),
	}

	return listeners
}

// addrStringToNetAddr takes an address in the form of 'host:port' and returns
// a net.Addr which maps to the original address with any host names resolved
// to IP addresses.  It also handles tor addresses properly by returning a
// net.Addr that encapsulates the address.
func addrStringToNetAddr(addr string, p2pCfg *config.P2PConfig) (net.Addr, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}

	// Skip if host is already an IP address.
	if ip := net.ParseIP(host); ip != nil {
		return &net.TCPAddr{
			IP:   ip,
			Port: port,
		}, nil
	}

	// Tor addresses cannot be resolved to an IP, so just return an onion
	// address instead.
	if strings.HasSuffix(host, ".onion") {
		return nil, errors.New("tor option is not allowed")
	}

	// Attempt to look up an IP address associated with the parsed host.
	ips, err := p2pCfg.BsvdLookup(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}

	return &net.TCPAddr{
		IP:   ips[0],
		Port: port,
	}, nil
}

// NewServer creates and return p2p server.
func NewServer(services *service.Services, peers map[*peerpkg.Peer]*peerpkg.PeerSyncState, p2pCfg *config.P2PConfig, log *zerolog.Logger) (*server, error) {
	serverLogger := log.With().Str("service", "p2p").Logger()
	server, err := newServer(config.ActiveNetParams.Params, services, peers, p2pCfg, &serverLogger)
	if err != nil {
		serverLogger.Error().Msgf("Unable to start server: %v", err)
		return nil, err
	}
	return server, nil
}
