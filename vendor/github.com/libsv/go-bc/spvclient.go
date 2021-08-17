package bc

import (
	"errors"
)

// An SPVClient is a struct used to specify interfaces
// used to complete Simple Payment Verification (SPV)
// in conjunction with a Merkle Proof.
//
// The implementation of BlockHeaderChain which is supplied will depend on the client
// you are using, some may return a HeaderJSON response others may return the blockhash.
type SPVClient struct {
	// BlockHeaderChain will be set when an implementation returning a bc.BlockHeader type is provided.
	bhc BlockHeaderChain
}

// SPVOpts can be implemented to provided functional options for an SPVClient.
type SPVOpts func(*SPVClient)

// WithBlockHeaderChain will inject the provided BlockHeaderChain into the SPVClient.
func WithBlockHeaderChain(bhc BlockHeaderChain) SPVOpts {
	return func(s *SPVClient) {
		s.bhc = bhc
	}
}

// NewSPVClient creates a new SPVClient based on the options provided.
// If no BlockHeaderChain implementation is provided, the setup will return an error.
func NewSPVClient(opts ...SPVOpts) (*SPVClient, error) {
	cli := &SPVClient{}
	for _, opt := range opts {
		opt(cli)
	}
	if cli.bhc == nil  {
		return nil, errors.New("at least one blockchain header implementation should be returned")
	}
	return cli, nil
}