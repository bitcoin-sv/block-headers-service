package testpulse

import "github.com/bitcoin-sv/pulse/domains"

// When exposes functions to easy testing operations that can happen in pulse.
type When struct {
	*TestPulse
}

// NewHeaderReceived simulates sending new header to application.
func (w *When) NewHeaderReceived(bs domains.BlockHeaderSource) error {
	_, err := w.services.Chains.Add(bs)
	return err
}
