package p2p

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/addrmgr"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/connmgr"
	"github.com/bitcoin-sv/block-headers-service/transports/p2p/peer"
	"github.com/rs/zerolog"
)

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
		"disconnecting", sp, msg.NumberOfFields)
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
		sp.log.Trace().Msgf("Added address %s from host %s", na.IP.String(), sp.NA().IP.String())
	}

	// Add addresses to server address manager.  The address manager handles
	// the details of things such as preventing duplicate addresses, max
	// addresses, and last seen updates.
	// XXX bitcoind gives a 2 hour time penalty here, do we want to do the
	// same?
	sp.server.addrManager.AddAddresses(msg.AddrList, sp.NA())
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
		UserAgentName:     sp.server.p2pConfig.UserAgentName,
		UserAgentVersion:  sp.server.p2pConfig.UserAgentVersion,
		UserAgentComments: initUserAgentComments(config.ExcessiveBlockSize),
		ChainParams:       sp.server.chainParams,
		Services:          sp.server.wireServices,
		ProtocolVersion:   uint32(70013),
		TrickleInterval:   config.TrickleInterval,
	}
}
