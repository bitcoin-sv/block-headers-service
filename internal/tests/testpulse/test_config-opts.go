package testpulse

import (
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/vconfig"
)

// WithApiAuthorization enable authorization with default config on API.
func WithApiAuthorization() ConfigOpt {
	return func(c *vconfig.Config) {
		c.WithAuthorization()
	}
}

// WithLongestChain fills the initialized header test repository with 4 additional blocks.
func WithLongestChain() RepoOpt {
	return func(r *testrepository.TestRepositories) {
		r.Headers.FillWithLongestChain()
	}
}
