package testapp

import "github.com/bitcoin-sv/block-headers-service/domains"

// When exposes functions to easy testing operations that can happen in block headers service.
type When struct {
	*TestBlockHeaderService
}

// NewHeaderReceived simulates sending new header to application.
func (w *When) NewHeaderReceived(bs domains.BlockHeaderSource) error {
	_, err := w.services.Chains.Add(bs)
	return err
}
