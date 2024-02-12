package testbhs

import (
	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testrepository"
)

// WithApiAuthorizationDisabled allows to not use authorization in Block Headers Service.
func WithApiAuthorizationDisabled() ConfigOpt {
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
