package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
)

// NewTestRepositories creates repository.Repositories for unit testing usage.
func NewTestRepositories(db []domains.BlockHeader) repository.Repositories {
	return repository.Repositories{
		Headers: NewHeadersTestRepository(db),
	}
}
