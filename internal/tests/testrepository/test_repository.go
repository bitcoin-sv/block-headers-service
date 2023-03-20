package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/repository"
)

func NewTestRepositories(db []domains.BlockHeader) repository.Repositories {
	return repository.Repositories{
		Headers: NewHeadersTestRepository(db),
		Tips:    NewTestTipRepository(db),
	}
}
