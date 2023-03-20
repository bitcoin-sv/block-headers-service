package service

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	testlog "github.com/libsv/bitcoin-hc/internal/tests/log"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/repository"
	"testing"
	"time"
)

func TestRejectBlockHeader(t *testing.T) {
	//given
	r, longestChainTip := givenLongestChainInRepository()
	h, hash := givenIgnoredHeaderToAddNextTo(longestChainTip)

	cs := createChainsService(serviceSetup{Repositories: &r, IgnoredHash: hash})

	//when
	header, err := cs.Add(h)

	//then
	if err == nil {
		t.Error("Expect to receive error BlockRejected")
	} else {
		assert.Equal(t, err.Error(), BlockRejected.String())
	}

	assert.Equal(t, header.State, domains.Rejected)
}

func TestAddTheHeaderToLongestChain(t *testing.T) {
	//given
	r, longestChainTip := givenChainWithOnlyGenesisBlockInRepository()
	h := givenHeaderToAddNextTo(longestChainTip)

	cs := createChainsService(serviceSetup{Repositories: &r})

	//when adding header
	header, addErr := cs.Add(h)

	//then
	if addErr != nil {
		t.Errorf("Doesn't expect to receive error, but get one %v", addErr)
	}

	hash := header.Hash
	_, err := r.Headers.GetHeaderByHash(hash.String())
	if err != nil {
		t.Errorf("Could not find header by hash %s because %s", hash, err)
	}

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

	//when adding header
	header, addErr := cs.Add(h)

	//then
	if addErr != nil {
		t.Errorf("Doesn't expect to receive error, but get one %v", addErr)
	}

	hash := header.Hash
	_, err := r.Headers.GetHeaderByHash(hash.String())
	if err != nil {
		t.Errorf("Could not find header by hash %s because %s", hash, err)
	}

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

	//when adding header
	header, addErr := cs.Add(h)

	//then
	if addErr != nil {
		t.Errorf("Doesn't expect to receive error, but get one %v", addErr)
	}

	hash := header.Hash
	_, err := r.Headers.GetHeaderByHash(hash.String())
	if err != nil {
		t.Errorf("Could not find header by hash %s because %s", hash, err)
	}

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

func givenChainWithOnlyGenesisBlockInRepository() (repository.Repositories, *domains.BlockHeader) {
	db, tip := testrepository.StartingChain()
	return testrepository.NewTestRepositories(db), tip
}

func givenLongestChainInRepository() (repository.Repositories, *domains.BlockHeader) {
	db, tip := testrepository.LongestChain()
	return testrepository.NewTestRepositories(db), tip
}

func givenOrphanChainInRepository(r *repository.Repositories) *domains.BlockHeader {
	_, orphanTip := testrepository.OrphanChain()
	r.Headers.AddHeaderToDatabase(*orphanTip)
	return orphanTip
}

func givenHeaderToAddNextTo(prev *domains.BlockHeader) BlockHeaderSource {
	return createHeaderSource(prev.Hash)
}

func createHeaderSource(ph chainhash.Hash) BlockHeaderSource {
	t, _ := time.Parse("yyyy-MM-dd hh:mm:ss", "2009-01-09 04:23:48")
	return BlockHeaderSource{
		Version:    1,
		PrevBlock:  ph,
		MerkleRoot: *testrepository.HashOf("63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1"),
		Timestamp:  t,
		Bits:       486604799,
		Nonce:      2011431709,
	}
}

func givenIgnoredHeaderToAddNextTo(prev *domains.BlockHeader) (BlockHeaderSource, domains.BlockHash) {
	h := createHeaderSource(prev.Hash)
	return h, DefaultBlockHasher().BlockHash(&h)
}

func givenOrphanedHeaderToAdd() BlockHeaderSource {
	return createHeaderSource(*testrepository.HashOf("0000000000000000000000000000000000000000000000000000000000001ce1"))
}

func createChainsService(s serviceSetup) Chains {
	return NewChainsService(ChainServiceDependencies{
		Repositories: s.Repositories,
		Params:       s.Params(),
		Logger:       testlog.InitializeMockLogger(),
		BlockHasher:  DefaultBlockHasher(),
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
