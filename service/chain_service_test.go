package service

import (
	"testing"
	"time"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	"github.com/libsv/bitcoin-hc/internal/tests/fixtures"
	testlog "github.com/libsv/bitcoin-hc/internal/tests/log"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/repository"
)

func TestRejectBlockHeader(t *testing.T) {
	//given
	r, longestChainTip := givenLongestChainInRepository()
	h, hash := givenIgnoredHeaderToAddNextTo(longestChainTip)

	cs := createChainsService(serviceSetup{Repositories: &r, IgnoredHash: hash})

	//when
	header, err := cs.Add(h)

	//then
	assert.IsError(t, err, BlockRejected.String())

	assertHeaderExist(t, header)

	assert.Equal(t, header.State, domains.Rejected)
}

func TestAddTheHeaderToLongestChain(t *testing.T) {
	//given
	r, longestChainTip := givenChainWithOnlyGenesisBlockInRepository()
	h := givenHeaderToAddNextTo(longestChainTip)

	cs := createChainsService(serviceSetup{Repositories: &r})

	//when
	header, addErr := cs.Add(h)

	//then
	assert.NoError(t, addErr)
	assertHeaderExist(t, header)
	assertHeaderInDb(t, r, header)

	if header.State != "LONGEST_CHAIN" {
		t.Errorf("Header should belong to the longest chain but is %s", header.State)
	}

	if !header.IsLongestChain() {
		t.Error("Header should be marked as longest chain but is not")
	}

	if header.Height != longestChainTip.Height+1 {
		t.Errorf("Expect header to be at height %d but it is at heigh %d", longestChainTip.Height+1, header.Height)
	}
}

func TestAddOrphanHeaderToChain(t *testing.T) {
	//given
	r, _ := givenLongestChainInRepository()
	h := givenOrphanedHeaderToAdd()

	cs := createChainsService(serviceSetup{Repositories: &r})

	//when
	header, addErr := cs.Add(h)

	//then
	assert.NoError(t, addErr)
	assertHeaderExist(t, header)
	assertHeaderInDb(t, r, header)

	if header.State != "ORPHAN" {
		t.Errorf("Header should belong to the orphan chain but is %s", header.State)
	}

	if !header.IsOrphan() {
		t.Error("Header should be marked as orphan but is not")
	}

	if header.Height != 1 {
		t.Errorf("Expect header to be at height 1 but it is at heigh %d", header.Height)
	}
}

func TestAddHeaderToOrphanChain(t *testing.T) {
	//given
	r, _ := givenLongestChainInRepository()
	tip := givenOrphanChainInRepository(&r)
	h := givenHeaderToAddNextTo(tip)

	cs := createChainsService(serviceSetup{Repositories: &r})

	//when
	header, addErr := cs.Add(h)

	//then
	assert.NoError(t, addErr)
	assertHeaderExist(t, header)
	assertHeaderInDb(t, r, header)

	if header.State != "ORPHAN" {
		t.Errorf("Header should belong to the orphan chain but is %s", header.State)
	}

	if !header.IsOrphan() {
		t.Error("Header should be marked as orphan but is not")
	}

	if header.Height != tip.Height+1 {
		t.Errorf("Expect header to be at height %d but it is at heigh %d", tip.Height+1, header.Height)
	}
}

func TestAddHeaderThatAlreadyExist(t *testing.T) {
	//given
	r, tip := givenLongestChainInRepository()
	h := fixtures.BlockHeaderSourceOf(tip)

	cs := createChainsService(serviceSetup{Repositories: &r})

	//when
	header, addErr := cs.Add(*h)

	//then
	assert.NoError(t, addErr)
	assertHeaderExist(t, header)
	assertHeaderInDb(t, r, header)

	if header.State != "LONGEST_CHAIN" {
		t.Errorf("Header should belong to the longest chain but is %s", header.State)
	}

	if !header.IsLongestChain() {
		t.Error("Header should be marked as longest chain but is not")
	}

	assertOnlyOneHeaderOnHeight(t, r, header)
}

func TestAddConcurrentChainBlock(t *testing.T) {
	var blockFromLongestChain = fixtures.HashHeight1
	var blockFromStaleChain = fixtures.StaleHashHeight2
	const bitsExceedingCumulatedChainWork uint32 = 0x180f0dc7

	testCases := map[string]struct {
		previous             *chainhash.Hash
		bits                 uint32
		newBlockChainState   domains.HeaderState
		oldLongestChainState domains.HeaderState
	}{
		"header with less chain work then concurrent block should be stale": {
			previous:             blockFromLongestChain,
			bits:                 fixtures.DefaultBits - 1,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header with the same chain work as concurrent block should be stale": {
			previous:             blockFromLongestChain,
			bits:                 fixtures.DefaultBits,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header with more chain work then concurrent block, but less then tip cumulated work should be stale": {
			previous:             blockFromLongestChain,
			bits:                 fixtures.DefaultBits + 1,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header next to other stale block with less chain work then concurrent block should be stale": {
			previous:             blockFromStaleChain,
			bits:                 fixtures.DefaultBits - 1,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header next to other stale block with the same chain work as concurrent block should be stale": {
			previous:             blockFromStaleChain,
			bits:                 fixtures.DefaultBits,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header next to other stale block with more chain work then concurrent block, but less then tip cumulated work should be stale": {
			previous:             blockFromStaleChain,
			bits:                 fixtures.DefaultBits + 1,
			newBlockChainState:   domains.Stale,
			oldLongestChainState: domains.LongestChain,
		},
		"header with the greatest chainwork next to the middle of stale chain should become longest chain tip": {
			previous:             fixtures.StaleHashHeight2,
			bits:                 bitsExceedingCumulatedChainWork,
			newBlockChainState:   domains.LongestChain,
			oldLongestChainState: domains.Stale,
		},
		"header with the greatest chainwork next to the tip of stale chain become longest chain tip": {
			previous:             fixtures.StaleHashHeight4,
			bits:                 bitsExceedingCumulatedChainWork,
			newBlockChainState:   domains.LongestChain,
			oldLongestChainState: domains.Stale,
		},
	}

	for name, params := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := givenLongestChainInRepository()
			givenStaleChainInRepository(&r)

			prev, _ := r.Headers.GetHeaderByHash(params.previous.String())
			h := givenHeaderToAddNextTo(prev)
			h.Bits = params.bits

			cs := createChainsService(serviceSetup{Repositories: &r})

			//when
			header, addErr := cs.Add(h)

			//then
			assert.NoError(t, addErr)
			assertHeaderExist(t, header)

			assertHeaderInDb(t, r, header)

			if header.Height != prev.Height+1 {
				t.Errorf("Expect header to be at height %d but it is at heigh %d", prev.Height+1, header.Height)
			}

			assertHeaderInState(t, header, params.newBlockChainState)

			tc := getHeadersFromThisChainUpTo(t, r, prev.Height)
			for _, ch := range tc {
				assertHeaderInState(t, ch, params.newBlockChainState)
			}

			cc := getHeadersFromConcurrentChain(t, r)
			for _, ch := range cc {
				assertHeaderInState(t, &ch, params.oldLongestChainState)
			}
		})
	}
}

func givenStaleChainInRepository(r *repository.Repositories) {
	sc, _ := fixtures.StaleChain()
	for _, h := range sc {
		if h.Hash != chaincfg.GenesisHash {
			r.Headers.AddHeaderToDatabase(h)
		}
	}
}

func assertHeaderInDb(t *testing.T, r repository.Repositories, header *domains.BlockHeader) {
	_, err := r.Headers.GetHeaderByHash(header.Hash.String())
	assert.NoError(t, err)
}

func assertOnlyOneHeaderOnHeight(t *testing.T, r repository.Repositories, header *domains.BlockHeader) {
	headers, err := r.Headers.GetLongestChainHeadersFromHeight(header.Height)
	assert.NoError(t, err)
	assert.Equal(t, len(headers), 1)
}

func getHeadersFromThisChainUpTo(t *testing.T, r repository.Repositories, height int32) []*domains.BlockHeader {
	c, _ := fixtures.StaleChain()
	hs := make([]*domains.BlockHeader, 0)
	for _, h := range c {
		if h.Hash != chaincfg.GenesisHash && h.Height <= height {
			dbh, err := r.Headers.GetHeaderByHash(h.Hash.String())
			assert.NoError(t, err)
			hs = append(hs, dbh)
		}
	}
	return hs
}

func getHeadersFromConcurrentChain(t *testing.T, r repository.Repositories) []domains.BlockHeader {
	o := make([]domains.BlockHeader, 0)

	h, err := r.Headers.GetHeaderByHash(fixtures.HashHeight1.String())
	assert.NoError(t, err)
	o = append(o, *h)

	h, err = r.Headers.GetHeaderByHash(fixtures.HashHeight2.String())
	assert.NoError(t, err)
	o = append(o, *h)

	h, err = r.Headers.GetHeaderByHash(fixtures.HashHeight3.String())
	assert.NoError(t, err)
	o = append(o, *h)

	h, err = r.Headers.GetHeaderByHash(fixtures.HashHeight4.String())
	assert.NoError(t, err)
	o = append(o, *h)

	return o
}

func assertHeaderExist(t *testing.T, h *domains.BlockHeader) {
	t.Helper()
	if h == nil {
		t.Fatal("Expect to receive header, but doesn't get one")
	}
}

func assertHeaderInState(t *testing.T, h *domains.BlockHeader, s domains.HeaderState) {
	t.Helper()
	if h.State != s {
		t.Errorf("Header %s should be in a %s state but is %s. \n Details: %+v", h.Hash, s, h.State, h)
	}
}

func givenChainWithOnlyGenesisBlockInRepository() (repository.Repositories, *domains.BlockHeader) {
	db, tip := fixtures.StartingChain()
	return testrepository.NewTestRepositories(&db), tip
}

func givenLongestChainInRepository() (repository.Repositories, *domains.BlockHeader) {
	db, tip := fixtures.LongestChain()

	var array []domains.BlockHeader = db
	return testrepository.NewTestRepositories(&array), tip
}

func givenOrphanChainInRepository(r *repository.Repositories) *domains.BlockHeader {
	_, orphanTip := fixtures.OrphanChain()
	r.Headers.AddHeaderToDatabase(*orphanTip)
	return orphanTip
}

func givenHeaderToAddNextTo(prev *domains.BlockHeader) domains.BlockHeaderSource {
	return createHeaderSource(prev.Hash)
}

func createHeaderSource(ph chainhash.Hash) domains.BlockHeaderSource {
	t, _ := time.Parse("yyyy-MM-dd hh:mm:ss", "2009-01-09 04:23:48")
	return domains.BlockHeaderSource{
		Version:    1,
		PrevBlock:  ph,
		MerkleRoot: *fixtures.HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"),
		Timestamp:  t,
		Bits:       486604799,
		Nonce:      2011431709,
	}
}

func givenIgnoredHeaderToAddNextTo(prev *domains.BlockHeader) (domains.BlockHeaderSource, domains.BlockHash) {
	h := createHeaderSource(prev.Hash)
	return h, DefaultBlockHasher().BlockHash(&h)
}

func givenOrphanedHeaderToAdd() domains.BlockHeaderSource {
	return createHeaderSource(*fixtures.HashOf("0000000000000000000000000000000000000000000000000000000000001ce1"))
}

func createChainsService(s serviceSetup) Chains {
	return NewChainsService(ChainServiceDependencies{
		Repositories:  s.Repositories,
		Params:        s.Params(),
		LoggerFactory: testlog.NewTestLoggerFactory(),
		BlockHasher:   DefaultBlockHasher(),
	})
}

type serviceSetup struct {
	*repository.Repositories
	IgnoredHash domains.BlockHash
}

func (s *serviceSetup) Params() *chaincfg.Params {
	ign := chainhash.Hash(s.IgnoredHash)

	return &chaincfg.Params{
		HeadersToIgnore: []*chainhash.Hash{&ign},
	}
}
