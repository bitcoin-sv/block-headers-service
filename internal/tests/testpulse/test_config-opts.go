package testpulse

import (
	"github.com/bitcoin-sv/pulse/config"
	"github.com/bitcoin-sv/pulse/internal/tests/testrepository"
)

// WithApiAuthorization allows to not use authorization in Pulse.
func WithoutApiAuthorization() ConfigOpt {
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
