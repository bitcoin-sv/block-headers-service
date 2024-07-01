package testapp

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testrepository"
)

// WithAPIAuthorizationDisabled allows to not use authorization in Block Headers Service.
func WithAPIAuthorizationDisabled() ConfigOpt {
	return func(c *config.AppConfig) {
		c.WithoutAuthorization()
	}
}

// WithLongestChain fills the initialized header test repository with 4 additional blocks.
func WithLongestChain() RepoOpt {
	return func(r *testrepository.TestRepositories) {
		r.Headers.FillWithLongestChain()
	}
}
