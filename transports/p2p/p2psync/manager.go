// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package p2psync

import (
	"fmt"
	"math"
	"math/big"
	"crypto/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libsv/bitcoin-hc/configs"
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/wire"
	"github.com/libsv/bitcoin-hc/service"
	"github.com/libsv/bitcoin-hc/transports/p2p/p2plog"
	peerpkg "github.com/libsv/bitcoin-hc/transports/p2p/peer"
)

const (

	// maxNetworkViolations is the max number of network violations a
	// sync peer can have before a new sync peer is found.
	maxNetworkViolations = 3

	// maxLastBlockTime is the longest time in seconds that we will
	// stay with a sync peer while below the current blockchain height.
	// Set to 3 minutes.
	maxLastBlockTime = 60 * 3 * time.Second

	// syncPeerTickerInterval is how often we check the current
	// syncPeer. Set to 30 seconds.
	syncPeerTickerInterval = 30 * time.Second
)

// zeroHash is the zero value hash (all zeros).  It is defined as a convenience.
var zeroHash chainhash.Hash

// newPeerMsg signifies a newly connected peer to the block handler.
type newPeerMsg struct {
	peer  *peerpkg.Peer
	reply chan struct{}
}

// invMsg packages a bitcoin inv message and the peer it came from together
// so the block handler has access to that information.
type invMsg struct {
	inv  *wire.MsgInv
	peer *peerpkg.Peer
}

// headersMsg packages a bitcoin headers message and the peer it came from
// together so the block handler has access to that information.
type headersMsg struct {
	headers *wire.MsgHeaders
	peer    *peerpkg.Peer
}

// donePeerMsg signifies a newly disconnected peer to the block handler.
type donePeerMsg struct {
	peer  *peerpkg.Peer
	reply chan struct{}
}

// getSyncPeerMsg is a message type to be sent across the message channel for
// retrieving the current sync peer.
type getSyncPeerMsg struct {
	reply chan int32
}

// isCurrentMsg is a message type to be sent across the message channel for
// requesting whether or not the sync manager believes it is synced with the
// currently connected peers.
type isCurrentMsg struct {
	reply chan bool
}

// pauseMsg is a message type to be sent across the message channel for
// pausing the sync manager.  This effectively provides the caller with
// exclusive access over the manager until a receive is performed on the
// unpause channel.
type pauseMsg struct {
	unpause <-chan struct{}
}

// syncPeerState stores additional info about the sync peer.
type syncPeerState struct {
	recvBytes         uint64
	recvBytesLastTick uint64
	lastBlockTime     time.Time
	violations        int
	ticks             uint64
}

// validNetworkSpeed checks if the peer is slow and
// returns an integer representing the number of network
// violations the sync peer has.
func (sps *syncPeerState) validNetworkSpeed(minSyncPeerNetworkSpeed uint64) int {
	// Fresh sync peer. We need another tick.
	if sps.ticks == 0 {
		return 0
	}

	// Number of bytes received in the last tick.
	recvDiff := sps.recvBytes - sps.recvBytesLastTick

	// If the peer was below the threshold, mark a violation and return.
	if recvDiff/uint64(syncPeerTickerInterval.Seconds()) < minSyncPeerNetworkSpeed {
		sps.violations++
		return sps.violations
	}

	// No violation found, reset the violation counter.
	sps.violations = 0

	return sps.violations
}

// updateNetwork updates the received bytes. Just tracks 2 ticks
// worth of network bandwidth.
func (sps *syncPeerState) updateNetwork(syncPeer *peerpkg.Peer) {
	sps.ticks++
	sps.recvBytesLastTick = sps.recvBytes
	sps.recvBytes = syncPeer.BytesReceived()
}

// SyncManager is used to communicate block related messages with peers. The
// SyncManager is started as by executing Start() in a goroutine. Once started,
// it selects peers to sync from and starts the initial block download. Once the
// chain is in sync, the SyncManager handles incoming block and header
// notifications and relays announcements of new blocks to peers.
type SyncManager struct {
	log            p2plog.Logger
	peerNotifier   PeerNotifier
	started        int32
	shutdown       int32
	chainParams    *chaincfg.Params
	progressLogger *blockProgressLogger
	msgChan        chan interface{}
	wg             sync.WaitGroup
	quit           chan struct{}
	syncPeer       *peerpkg.Peer
	syncPeerState  *syncPeerState
	peerStates     map[*peerpkg.Peer]*peerpkg.PeerSyncState

	// The following fields are used for headers-first mode.
	headersFirstMode bool
	startHeader      *domains.BlockHeader
	nextCheckpoint   *chaincfg.Checkpoint

	// minSyncPeerNetworkSpeed is the minimum speed allowed for
	// a sync peer.
	minSyncPeerNetworkSpeed uint64
	blocksToConfirmFork     int

	Services *service.Services
}

// findNextHeaderCheckpoint returns the next checkpoint after the passed height.
// It returns nil when there is not one either because the height is already
// later than the final checkpoint or some other reason such as disabled
// checkpoints.
// TODO: set next headers checkpoint.
func (sm *SyncManager) findNextHeaderCheckpoint(height int32) *chaincfg.Checkpoint {
	checkpoints := configs.Cfg.Checkpoints

	sm.log.Infof("[Headers] findNextHeaderCheckpoint count: %d, height: %d", len(checkpoints), height)

	if len(checkpoints) == 0 {
		return nil
	}

	// There is no next checkpoint if the height is already after the final
	// checkpoint.
	finalCheckpoint := &checkpoints[len(checkpoints)-1]
	if height >= finalCheckpoint.Height {
		return nil
	}

	sm.log.Infof("[Headers] height: %d, final checkpoint: %d", height, finalCheckpoint.Height)

	// Find the next checkpoint.
	nextCheckpoint := finalCheckpoint
	for i := len(checkpoints) - 2; i >= 0; i-- {
		if height >= checkpoints[i].Height {
			break
		}
		nextCheckpoint = &checkpoints[i]
	}
	return nextCheckpoint
}

func (sm *SyncManager) ignoreBlockHash(blockHash *chainhash.Hash) bool {
	for _, hash := range sm.chainParams.HeadersToIgnore {
		if blockHash.IsEqual(hash) {
			return true
		}
	}

	return false
}

// startSync will choose the best peer among the available candidate peers to
// download/sync the blockchain from.  When syncing is already running, it
// simply returns.  It also examines the candidates for any which are no longer
// candidates and removes them as needed.
func (sm *SyncManager) startSync() {
	sm.log.Info("[Manager] startSync")
	// Return now if we're already syncing.
	if sm.syncPeer != nil {
		return
	}

	best := sm.Services.Headers.GetTipHeight()
	bestPeers := []*peerpkg.Peer{}
	okPeers := []*peerpkg.Peer{}
	for peer, state := range sm.peerStates {
		if !state.SyncCandidate {
			continue
		}

		// Add any peers on the same block to okPeers. These should
		// only be used as a last resort.
		if peer.LastBlock() == best {
			okPeers = append(okPeers, peer)
			continue
		}

		// Remove sync candidate peers that are no longer candidates due
		// to passing their latest known block.
		if peer.LastBlock() < best {
			state.SyncCandidate = false
			continue
		}

		// Append each good peer to bestPeers for selection later.
		bestPeers = append(bestPeers, peer)
	}

	var bestPeer *peerpkg.Peer

	// Try to select a random peer that is at a higher block height,
	// if that is not available then use a random peer at the same
	// height and hope they find blocks.
	if len(bestPeers) > 0 {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(bestPeers))))

		if err == nil {
			bestPeer = bestPeers[int(randInt.Int64())]
		}
	} else if len(okPeers) > 0 {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(okPeers))))

		if err == nil {
			bestPeer = bestPeers[int(randInt.Int64())]
		}
	}

	// Start syncing from the best peer if one was selected.
	if bestPeer != nil {
		// Clear the requestedBlocks if the sync peer changes, otherwise
		// we may ignore blocks we need that the last sync peer failed
		// to send.
		// sm.requestedBlocks = make(map[chainhash.Hash]struct{})

		locator := sm.Services.Headers.LatestHeaderLocator()

		sm.log.Infof("Syncing to block height %d from peer %v",
			bestPeer.LastBlock(), bestPeer.Addr())

		// When the current height is less than a known checkpoint we
		// can use block headers to learn about which blocks comprise
		// the chain up to the checkpoint and perform less validation
		// for them.  This is possible since each header contains the
		// hash of the previous header and a merkle root.  Therefore if
		// we validate all of the received headers link together
		// properly and the checkpoint hashes match, we can be sure the
		// hashes for the blocks in between are accurate.  Further, once
		// the full blocks are downloaded, the merkle root is computed
		// and compared against the value in the header which proves the
		// full block hasn't been tampered with.
		//
		// Once we have passed the final checkpoint, or checkpoints are
		// disabled, use standard inv messages learn about the blocks
		// and fully validate them.  Finally, regression test mode does
		// not support the headers-first approach so do normal block
		// downloads when in regression test mode.
		if sm.nextCheckpoint != nil &&
			best < sm.nextCheckpoint.Height &&
			sm.chainParams != &chaincfg.RegressionNetParams {

			// TODO: request for next headers batch
			sm.log.Info("[Headers] startSync - Request for next headers batch")
			err := bestPeer.PushGetHeadersMsg(locator, sm.nextCheckpoint.Hash)
			if err != nil {
				sm.log.Info(err)
			}
			sm.headersFirstMode = true
			sm.log.Infof("Downloading headers for blocks %d to "+
				// "%d from peer %s", best.Height+1,
				"%d from peer %s", best+1,
				sm.nextCheckpoint.Height, bestPeer.Addr())
		} else {
			// TODO: initial request for headers
			sm.log.Info("[Headers] Initial request")
			err := bestPeer.PushGetBlocksMsg(locator, &zeroHash)
			if err != nil {
				sm.log.Info(err)
			}
		}

		bestPeer.SetSyncPeer(true)
		sm.syncPeer = bestPeer
		sm.syncPeerState = &syncPeerState{
			lastBlockTime:     time.Now(),
			recvBytes:         bestPeer.BytesReceived(),
			recvBytesLastTick: uint64(0),
		}
	} else {
		sm.log.Warnf("No sync peer candidates available")
	}
}

// SyncHeight returns latest known block being synced to.
func (sm *SyncManager) SyncHeight() uint64 {
	if sm.syncPeer == nil {
		return 0
	}

	return uint64(sm.topBlock())
}

// isSyncCandidate returns whether or not the peer is a candidate to consider
// syncing from.
func (sm *SyncManager) isSyncCandidate(peer *peerpkg.Peer) bool {
	// Typically a peer is not a candidate for sync if it's not a full node,
	// however regression test is special in that the regression tool is
	// not a full node and still needs to be considered a sync candidate.
	if sm.chainParams == &chaincfg.RegressionNetParams {
		// The peer is not a candidate if it's not coming from localhost
		// or the hostname can't be determined for some reason.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			return false
		}

		if host != "127.0.0.1" && host != "localhost" {
			return false
		}
	} else {
		// The peer is not a candidate for sync if it's not a full
		// node.
		nodeServices := peer.Services()
		if nodeServices&wire.SFNodeNetwork != wire.SFNodeNetwork {
			return false
		}
	}

	// Candidate if all checks passed.
	return true
}

// handleNewPeerMsg deals with new peers that have signaled they may
// be considered as a sync peer (they have already successfully negotiated).  It
// also starts syncing if needed.  It is invoked from the syncHandler goroutine.
func (sm *SyncManager) handleNewPeerMsg(peer *peerpkg.Peer) {
	// Ignore if in the process of shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.log.Infof("New valid peer %s (%s)", peer, peer.UserAgent())

	// Initialize the peer state
	isSyncCandidate := sm.isSyncCandidate(peer)

	sm.peerStates[peer] = &peerpkg.PeerSyncState{
		SyncCandidate: isSyncCandidate,
	}

	// Start syncing by choosing the best candidate if needed.
	if isSyncCandidate && sm.syncPeer == nil {
		sm.startSync()
	}
}

// handleCheckSyncPeer selects a new sync peer.
func (sm *SyncManager) handleCheckSyncPeer() {
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	// If we don't have a sync peer, then there is nothing to do.
	if sm.syncPeer == nil {
		return
	}

	// Update network stats at the end of this tick.
	defer sm.syncPeerState.updateNetwork(sm.syncPeer)

	// Check network speed of the sync peer and its last block time. If we're currently
	// flushing the cache skip this round.
	if (sm.syncPeerState.validNetworkSpeed(sm.minSyncPeerNetworkSpeed) < maxNetworkViolations) &&
		(time.Since(sm.syncPeerState.lastBlockTime) <= maxLastBlockTime) {
		return
	}

	// Don't update sync peers if you have all the available
	// blocks.

	best := sm.Services.Headers.GetTip()

	if sm.topBlock() == best.Height {
		// Update the time and violations to prevent disconnects.
		sm.syncPeerState.lastBlockTime = time.Now()
		sm.syncPeerState.violations = 0
		return
	}

	_, exists := sm.peerStates[sm.syncPeer]
	if !exists {
		return
	}

	sm.updateSyncPeer()
}

// topBlock returns the best chains top block height.
func (sm *SyncManager) topBlock() int32 {

	if sm.syncPeer.LastBlock() > sm.syncPeer.StartingHeight() {
		return sm.syncPeer.LastBlock()
	}

	return sm.syncPeer.StartingHeight()
}

// handleDonePeerMsg deals with peers that have signaled they are done.  It
// removes the peer as a candidate for syncing and in the case where it was
// the current sync peer, attempts to select a new best peer to sync from.  It
// is invoked from the syncHandler goroutine.
func (sm *SyncManager) handleDonePeerMsg(peer *peerpkg.Peer) {
	_, exists := sm.peerStates[peer]
	if !exists {
		sm.log.Warnf("Received done peer message for unknown peer %s", peer)
		return
	}

	// Remove the peer from the list of candidate peers.
	delete(sm.peerStates, peer)

	sm.log.Infof("Lost peer %s", peer)

	// Fetch a new sync peer if this is the sync peer.
	if peer == sm.syncPeer {
		sm.updateSyncPeer()
	}
}

// updateSyncPeer picks a new peer to sync from.
func (sm *SyncManager) updateSyncPeer() {
	sm.log.Infof("Updating sync peer, last block: %v, violations: %v", sm.syncPeerState.lastBlockTime, sm.syncPeerState.violations)

	// Disconnect from the misbehaving peer.
	sm.syncPeer.Disconnect()

	// Attempt to find a new peer to sync from
	// Also, reset the headers-first state.
	sm.syncPeer.SetSyncPeer(false)
	sm.syncPeer = nil
	sm.syncPeerState = nil

	if sm.headersFirstMode {
		sm.log.Info("[Manager] updateSyncPeer, resetHeaderState")
		best := sm.Services.Headers.GetTip()
		sm.log.Infof("[Manager] BestSnapshot : %#v", best)
	}

	sm.startSync()
}

// current returns true if we believe we are synced with our peers, false if we
// still have blocks to check.
func (sm *SyncManager) current() bool {
	if !sm.Services.Headers.IsCurrent() {
		return false
	}

	// if blockChain thinks we are current and we have no syncPeer it
	// is probably right.
	if sm.syncPeer == nil {
		return true
	}

	// No matter what chain thinks, if we are below the block we are syncing
	// to we are not current.
	if sm.Services.Headers.GetTipHeight() < sm.syncPeer.LastBlock() {
		return false
	}
	return true
}

// handleHeadersMsg handles block header messages from all peers.  Headers are
// requested when performing a headers-first sync.
func (sm *SyncManager) handleHeadersMsg(hmsg *headersMsg) {
	peer := hmsg.peer

	_, exists := sm.peerStates[peer]
	if !exists {
		sm.log.Warnf("Received headers message from unknown peer %s", peer)
		return
	}

	// The remote peer is misbehaving if we didn't request headers.
	msg := hmsg.headers
	numHeaders := len(msg.Headers)
	sm.log.Infof("[Headers] received headers count: %d", numHeaders)

	if !sm.headersFirstMode {
		sm.log.Warnf("Got %d unrequested headers from %s -- "+
			"disconnecting", numHeaders, peer.Addr())
		peer.Disconnect()
		return
	}

	// Nothing to do for an empty headers message.
	if numHeaders == 0 {
		return
	}

	// Process all of the received headers ensuring each one connects to the
	// previous and that checkpoints match.
	receivedCheckpoint := false
	var finalHash *chainhash.Hash
	for i, blockHeader := range msg.Headers {
		h, addErr := sm.Services.Chains.Add(service.BlockHeaderSource(*blockHeader))

		if service.BlockRejected.Is(addErr) {
			sm.peerNotifier.BanPeer(peer)
			peer.Disconnect()
			return
		}

		if service.BlockSaveFail.Is(addErr) {
			sm.log.Errorf("Couldn't save header %v in database, because of %+v", h, addErr)
		}

		sm.logSyncState(i, *h)

		// Verify the header at the next checkpoint height matches.
		var err error
		receivedCheckpoint, err = verifyCheckpointHeight(sm, *h, receivedCheckpoint, peer)
		if err != nil {
			sm.log.Warnf(err.Error())
			return
		}

		if h.IsLongestChain() {
			finalHash = &h.Hash
		}
		if sm.startHeader == nil {
			sm.startHeader = h
		}
	}

	// When this header is a checkpoint, switch to fetching the blocks for
	// all the headers since the last checkpoint.
	if receivedCheckpoint {
		// Since the first entry of the list is always the final block
		// that is already in the database and is only used to ensure
		// the next header links properly, it must be removed before
		// fetching the blocks.
		// sm.headerList.Remove(sm.headerList.Front())
		// sm.log.Infof("Received %v block headers: Fetching blocks",
		// sm.headerList.Len())

		sm.progressLogger.SetLastLogTime(time.Now())
		sm.log.Infof("Received checkpoint headers: %v - Fetching next headers", sm.Services.Headers.CountHeaders())
		prevHeight := sm.nextCheckpoint.Height
		prevHash := sm.nextCheckpoint.Hash
		sm.nextCheckpoint = sm.findNextHeaderCheckpoint(prevHeight)
		if sm.nextCheckpoint != nil {
			sm.requestForNextHeaderBatch(prevHash, peer, prevHeight)
			return
		}
	}

	if sm.nextCheckpoint == nil {
		// This is headers-first mode, the block is a checkpoint, and there are
		// no more checkpoints, so switch to normal mode by requesting blocks
		// from the block after this one up to the end of the chain (zero hash).
		sm.log.Infof("Reached the final checkpoint -- switching to normal mode")
		sm.log.Infof("Reached the final checkpoint -- lastHash: %#v", finalHash.String())
		sm.sendGetHeadersWithPassedParams([]*chainhash.Hash{finalHash}, &zeroHash, peer)
		return
	}

	// This header is not a checkpoint, so request the next batch of
	// headers starting from the latest known header and ending with the
	// next checkpoint.
	sm.sendGetHeadersWithPassedParams([]*chainhash.Hash{finalHash}, sm.nextCheckpoint.Hash, peer)

}

func (sm *SyncManager) requestForNextHeaderBatch(prevHash *chainhash.Hash, peer *peerpkg.Peer, prevHeight int32) {
	sm.log.Infof("[Manager] receivedCheckpoint    : %d", sm.nextCheckpoint.Height)
	sm.log.Infof("[Manager] nextCheckpoint.Height : %d", sm.nextCheckpoint.Height)
	sm.log.Infof("[Manager] nextCheckpoint.Hash   : %v", sm.nextCheckpoint.Hash)

	sm.sendGetHeadersWithPassedParams([]*chainhash.Hash{prevHash}, sm.nextCheckpoint.Hash, peer)
	if sm.syncPeer != nil {
		sm.log.Infof("Downloading headers for blocks %d to %d from "+
			"peer %s", prevHeight+1, sm.nextCheckpoint.Height,
			sm.syncPeer.Addr())
	}
}

// TODO: Consider removing this method after finishing devleopment.
func (sm *SyncManager) logSyncState(i int, h domains.BlockHeader) {
	length := sm.Services.Headers.CountHeaders()
	if math.Mod(float64(length), 1000) == 0 || length > 760000 {
		sm.log.Infof("[Manager][%d] sm.headerList.Len()    : %#v", i, length)
	}
	if length > 762595 {
		sm.log.Infof("[Manager][%d] -------------------", i)
		sm.log.Infof("[Manager][%d] sm.headerList.Len()    : %#v", i, length)
		sm.log.Infof("[Manager][%d] node.hash              : %#v", i, h.Hash)
		sm.log.Infof("[Manager][%d] node.height            : %#v", i, h.Height)
		sm.log.Infof("[Manager][%d] prevNode.hash          : %#v", i, h.PreviousBlock)
	}
}

func verifyCheckpointHeight(sm *SyncManager, h domains.BlockHeader, receivedCheckpoint bool, peer *peerpkg.Peer) (bool, error) {
	if sm.nextCheckpoint != nil && h.Height == sm.nextCheckpoint.Height {
		if h.Hash == *sm.nextCheckpoint.Hash {
			receivedCheckpoint = true
			sm.log.Infof("Verified downloaded block "+
				"header against checkpoint at height "+
				"%d/hash %s", h.Height, h.Hash)
		} else {
			sm.log.Warnf("Block header at height %d/hash "+
				"%s from peer %s does NOT match "+
				"expected checkpoint hash of %s -- "+
				"disconnecting", h.Height,
				h.Hash, peer.Addr(),
				sm.nextCheckpoint.Hash)
			peer.Disconnect()
			return false, fmt.Errorf("corresponding checkpoint height does not match got: %v, exp: %v", h.Height, sm.nextCheckpoint.Height)
		}
	}
	return receivedCheckpoint, nil
}

func (sm *SyncManager) sendGetHeadersWithPassedParams(chainHash []*chainhash.Hash, stopHash *chainhash.Hash, peer *peerpkg.Peer) {
	locator := domains.BlockLocator(chainHash)
	err := peer.PushGetHeadersMsg(locator, stopHash)
	if err != nil {
		sm.log.Warnf("Failed to send getheaders message to "+
			"peer %s: %v", peer.Addr(), err)
	}
}

// handleInvMsg handles inv messages from all peers.
// We examine the inventory advertised by the remote peer and act accordingly.
func (sm *SyncManager) handleInvMsg(imsg *invMsg) {
	sm.log.Infof("[Headers] handleInvMsg, peer.ID: %d", imsg.peer.ID())

	lastHeader, _ := sm.Services.Headers.BackElement()
	sm.log.Infof("[Manager] handleInvMsg lastHeaderNode.height : %d", lastHeader.Height)

	peer := imsg.peer
	_, exists := sm.peerStates[peer]
	if !exists {
		sm.log.Warnf("Received inv message from unknown peer %s", peer)
		return
	}

	// Attempt to find the final block in the inventory list.  There may
	// not be one.
	invVects := imsg.inv.InvList
	lastBlock := searchForFinalBlock(invVects)

	// If this inv contains a block announcement, and this isn't coming from
	// our current sync peer or we're current, then update the last
	// announced block for this peer. We'll use this information later to
	// update the heights of peers based on blocks we've accepted that they
	// previously announced.
	if lastBlock != -1 && (peer != sm.syncPeer || sm.current()) {
		peer.UpdateLastAnnouncedBlock(&invVects[lastBlock].Hash)
	}

	// Ignore invs from peers that aren't the sync if we are not current.
	// Helps prevent fetching a mass of orphans.
	if peer != sm.syncPeer && !sm.current() {
		return
	}

	// If our chain is current and a peer announces a block we already
	// know of, then update their current block height.
	if lastBlock != -1 && sm.current() {
		blkHeight, err := sm.Services.Headers.BlockHeightByHash(&invVects[lastBlock].Hash)
		if err == nil {
			peer.UpdateLastBlockHeight(blkHeight)
		}
	}

	if lastBlock != -1 {
		lastHeader, _ := sm.Services.Headers.BackElement()
		sm.log.Infof("[Manager] handleInvMsg  lastConfirmedHeaderNode.hash  : %s", lastHeader.Hash)
		sm.log.Infof("[Manager] handleInvMsg lastConfirmedHeaderNode.height : %d", lastHeader.Height)
		sm.log.Infof("[Manager] handleInvMsg &invVects[lastBlock].Hash  : %v", &invVects[lastBlock].Hash)

		sm.sendGetHeadersWithPassedParams([]*chainhash.Hash{&lastHeader.Hash}, &invVects[lastBlock].Hash, peer)
	}
}

func searchForFinalBlock(invVects []*wire.InvVect) int {
	lastBlock := -1
	for i := len(invVects) - 1; i >= 0; i-- {
		if invVects[i].Type == wire.InvTypeBlock {
			lastBlock = i
			break
		}
	}
	return lastBlock
}

// blockHandler is the main handler for the sync manager.  It must be run as a
// goroutine.  It processes block and inv messages in a separate goroutine
// from the peer handlers so the block (MsgBlock) messages are handled by a
// single thread without needing to lock memory data structures.  This is
// important because the sync manager controls which blocks are needed and how
// the fetching should proceed.
func (sm *SyncManager) blockHandler() {
	ticker := time.NewTicker(syncPeerTickerInterval)
	defer ticker.Stop()

out:
	for {
		select {
		case <-ticker.C:
			sm.log.Infof("[Event] handleCheckSyncPeer")
			sm.handleCheckSyncPeer()
		case m := <-sm.msgChan:
			switch msg := m.(type) {
			case *newPeerMsg:
				sm.log.Infof("[Event] newPeerMsg")
				sm.handleNewPeerMsg(msg.peer)
				if msg.reply != nil {
					msg.reply <- struct{}{}
				}

			case *invMsg:
				sm.handleInvMsg(msg)

			case *headersMsg:
				sm.handleHeadersMsg(msg)

			case *donePeerMsg:
				sm.log.Infof("[Event] donePeerMsg")
				sm.handleDonePeerMsg(msg.peer)
				if msg.reply != nil {
					msg.reply <- struct{}{}
				}

			case getSyncPeerMsg:
				sm.log.Infof("[Event] getSyncPeerMsg")
				var peerID int32

				if sm.syncPeer != nil {
					peerID = sm.syncPeer.ID()
				}
				msg.reply <- peerID

			case isCurrentMsg:
				sm.log.Infof("[Event] isCurrentMsg")
				msg.reply <- sm.current()

			case pauseMsg:
				sm.log.Infof("[Event] pauseMsg")
				// Wait until the sender unpauses the manager.
				<-msg.unpause

			default:
				sm.log.Warnf("Invalid message type in block "+
					"handler: %T", msg)
			}

		case <-sm.quit:
			break out
		}
	}
	sm.log.Debug("Block handler shutting down")

	sm.wg.Done()
	sm.log.Trace("Block handler done")
}

// NewPeer informs the sync manager of a newly active peer.
func (sm *SyncManager) NewPeer(peer *peerpkg.Peer, done chan struct{}) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		done <- struct{}{}
		return
	}
	sm.msgChan <- &newPeerMsg{peer: peer, reply: done}
}

// QueueInv adds the passed inv message and peer to the block handling queue.
func (sm *SyncManager) QueueInv(inv *wire.MsgInv, peer *peerpkg.Peer) {
	// No channel handling here because peers do not need to block on inv
	// messages.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &invMsg{inv: inv, peer: peer}
}

// QueueHeaders adds the passed headers message and peer to the block handling
// queue.
func (sm *SyncManager) QueueHeaders(headers *wire.MsgHeaders, peer *peerpkg.Peer) {
	// No channel handling here because peers do not need to block on
	// headers messages.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		return
	}

	sm.msgChan <- &headersMsg{headers: headers, peer: peer}
}

// DonePeer informs the blockmanager that a peer has disconnected.
func (sm *SyncManager) DonePeer(peer *peerpkg.Peer, done chan struct{}) {
	// Ignore if we are shutting down.
	if atomic.LoadInt32(&sm.shutdown) != 0 {
		done <- struct{}{}
		return
	}

	sm.msgChan <- &donePeerMsg{peer: peer, reply: done}
}

// Start begins the core block handler which processes block and inv messages.
func (sm *SyncManager) Start() {
	// Already started?
	if atomic.AddInt32(&sm.started, 1) != 1 {
		return
	}

	sm.log.Trace("Starting sync manager")
	sm.wg.Add(1)
	go sm.blockHandler()
}

// Stop gracefully shuts down the sync manager by stopping all asynchronous
// handlers and waiting for them to finish.
func (sm *SyncManager) Stop() {
	if atomic.AddInt32(&sm.shutdown, 1) != 1 {
		sm.log.Warnf("Sync manager is already in the process of " +
			"shutting down")
	}

	sm.log.Infof("Sync manager shutting down")
	close(sm.quit)
	sm.wg.Wait()
	sm.log.Infof("Sync manager stopped")
}

// IsCurrent returns whether or not the sync manager believes it is synced with
// the connected peers.
func (sm *SyncManager) IsCurrent() bool {
	reply := make(chan bool)
	sm.msgChan <- isCurrentMsg{reply: reply}
	return <-reply
}

// New constructs a new SyncManager. Use Start to begin processing asynchronous
// block, tx, and inv updates.
func New(config *Config, peers map[*peerpkg.Peer]*peerpkg.PeerSyncState) (*SyncManager, error) {
	sm := SyncManager{
		log:                     config.Logger,
		peerNotifier:            config.PeerNotifier,
		chainParams:             config.ChainParams,
		peerStates:              peers,
		progressLogger:          newBlockProgressLogger("Processed", config.Logger),
		msgChan:                 make(chan interface{}, config.MaxPeers*3),
		quit:                    make(chan struct{}),
		minSyncPeerNetworkSpeed: config.MinSyncPeerNetworkSpeed,
		blocksToConfirmFork:     config.BlocksForForkConfirmation,
		Services:                config.Services,
	}

	if !config.DisableCheckpoints {
		// Initialize the next checkpoint based on the current height.
		height := config.Services.Headers.GetTipHeight()
		sm.nextCheckpoint = sm.findNextHeaderCheckpoint(height)
		if sm.nextCheckpoint != nil {
			sm.headersFirstMode = false
			sm.startHeader = nil
		}
	} else {
		sm.log.Info("Checkpoints are disabled")
	}

	return &sm, nil
}
