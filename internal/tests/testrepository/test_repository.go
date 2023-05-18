package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/tests/fixtures"
	"github.com/libsv/bitcoin-hc/repository"
)

// NewTestRepositories creates repository.Repositories for unit testing usage.
func NewTestRepositories(db *[]domains.BlockHeader) repository.Repositories {
	return repository.Repositories{
		Headers: NewHeadersTestRepository(db),
	}
}

// NewCleanTestRepositories creates repository.Repositories with minimal needed data (ex. with genesis block).
func NewCleanTestRepositories() repository.Repositories {
	db, _ := fixtures.StartingChain()
	var tokensTable []domains.Token

	return repository.Repositories{
		Headers:  NewHeadersTestRepository(&db),
		Tokens:   NewTokensTestRepository(&tokensTable),
		Webhooks: NewWebhooksTestRepository(&[]domains.Webhook{}),
	}
}
