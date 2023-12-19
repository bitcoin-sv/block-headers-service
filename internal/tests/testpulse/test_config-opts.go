package testpulse

import (
	"github.com/bitcoin-sv/pulse/internal/tests/testrepository"
)

// WithLongestChain fills the initialized header test repository with 4 additional blocks.
func WithLongestChain() RepoOpt {
	return func(r *testrepository.TestRepositories) {
		r.Headers.FillWithLongestChain()
	}
}
