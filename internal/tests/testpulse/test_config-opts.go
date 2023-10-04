package testpulse

import (
	"github.com/libsv/bitcoin-hc/config"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
)

// WithApiAuthorization enable authorization with default config on API.
func WithApiAuthorization() ConfigOpt {
	return func(c *config.Config) {
		c.WithAuthorization()
	}
}

// WithLongestChain fills the initialized header test repository with 4 additional blocks.
func WithLongestChain() RepoOpt {
	return func(r *testrepository.TestRepositories) {
		r.Headers.FillWithLongestChain()
	}
}
