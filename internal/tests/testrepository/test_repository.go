package testrepository

import (
	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/tests/fixtures"
	"github.com/libsv/bitcoin-hc/notification"
	"github.com/libsv/bitcoin-hc/repository"
)

// TestRepositories is a struct used for testing pulse repositories.
type TestRepositories struct {
	Headers  *HeaderTestRepository
	Tokens   *TokensTestRepository
	Webhooks *WebhooksTestRepository
}

// NewTestRepositories creates repository.Repositories for unit testing usage.
func NewTestRepositories(db *[]domains.BlockHeader) repository.Repositories {
	return repository.Repositories{
		Headers: NewHeadersTestRepository(db),
	}
}

// NewCleanTestRepositories creates TestRepositories with minimal needed data (ex. with genesis block).
func NewCleanTestRepositories() TestRepositories {
	db, _ := fixtures.StartingChain()
	var tokensTable []domains.Token

	return TestRepositories{
		Headers:  NewHeadersTestRepository(&db),
		Tokens:   NewTokensTestRepository(&tokensTable),
		Webhooks: NewWebhooksTestRepository(&[]notification.Webhook{}),
	}
}

// ToDomainRepo creates a domain repository.Repositories struct to comply with pulse structs.
func (t *TestRepositories) ToDomainRepo() *repository.Repositories {
	return &repository.Repositories{
		Headers: t.Headers,
		Tokens: t.Tokens,
		Webhooks: t.Webhooks,
	}
}
