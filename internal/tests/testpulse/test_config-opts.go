package testpulse

import (
	"github.com/libsv/bitcoin-hc/vconfig"
)

// WithApiAuthorization enable authorization with default config on API.
func WithApiAuthorization() ConfigOpt {
	return func(c *vconfig.Config) {
		c.WithAuthorization()
	}
}

// WithoutApiAuthorization disable API authorization.
func WithoutApiAuthorization() ConfigOpt {
	return func(c *vconfig.Config) {
		c.WithoutAuthorization()
	}
}
