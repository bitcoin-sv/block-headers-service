package service

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/repository"
	"github.com/rs/zerolog"
)

// MerklerootsService represents Merkleroots service and provide access to repositories.
type MerklerootsService struct {
	repo      *repository.Repositories
	merkleCfg *config.MerkleRootConfig
	log       *zerolog.Logger
}

// NewMerklerootsService creates and returns MerklerootsService instance.
func NewMerklerootsService(repo *repository.Repositories, merkleCfg *config.MerkleRootConfig, log *zerolog.Logger) *MerklerootsService {
	merklerootsLogger := log.With().Str("service", "merkleroots").Logger()
	return &MerklerootsService{
		repo:      repo,
		merkleCfg: merkleCfg,
		log:       &merklerootsLogger,
	}
}

// GetMerkleRootsConfirmations returns a confirmation of merkle roots inclusion in the longest chain
// with hash of the block in which the merkle root is included.
func (ms *MerklerootsService) GetMerkleRootsConfirmations(
	request []domains.MerkleRootConfirmationRequestItem,
) ([]*domains.MerkleRootConfirmation, error) {
	// correct where domains.merkelerootconfreqitem is declared
	return ms.repo.Headers.GetMerkleRootsConfirmations(request, ms.merkleCfg.MaxBlockHeightExcess)
}

// GetMerkleRoots returns ExclusiveStartKey pagination with merkle roots from lastEvaluatedKey which
// is the last height of the block that a client has processed
func (ms *MerklerootsService) GetMerkleRoots(batchSize int, lastEvaluatedKey int) (*domains.MerkleRootsESKPagedResponse, error) {
	return ms.repo.Headers.GetMerkleRoots(batchSize, lastEvaluatedKey)
}
