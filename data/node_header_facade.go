package data

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos"

	headers "github.com/libsv/bitcoin-hc"
)

// nodeHeaderFacade is a layer on top of the node and db
// data stores. This will attempt to lookup headers by block hash, first from the
// db and then if not found, from the node itself.
type nodeHeaderFacade struct {
	nodeRdr headers.BlockheaderReader
	dbRdr   headers.BlockheaderReader
}

// NewNodeHeaderFacade will return a new NodeHeaderFacade used to lookup headers from the db then fallback to the node.
//
// Unlike a true cache, this will not then store the result in the db as validation needs completed.
func NewNodeHeaderFacade(nodeRdr headers.BlockheaderReader, dbRdr headers.BlockheaderReader) *nodeHeaderFacade {
	return &nodeHeaderFacade{
		nodeRdr: nodeRdr,
		dbRdr:   dbRdr,
	}
}

// Header will return a header firstly by attempting a lookup from the db and if not present, then
// falling back to the node to locate it.
func (n *nodeHeaderFacade) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	hdr, err := n.dbRdr.Header(ctx, args)
	// if we have an error & it is an unknown error, return it.
	if err != nil && !lathos.IsNotFound(err) {
		return nil, errors.Wrapf(err, "failed to read header '%s' from db", args.Blockhash)
	}
	if hdr != nil {
		return hdr, nil
	}
	// fallback to a node lookup
	hdr, err = n.nodeRdr.Header(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read header '%s' from node", args.Blockhash)
	}
	return hdr, nil
}

// Height will return the current block height cached.
func (n *nodeHeaderFacade) Height(ctx context.Context) (int, error) {
	return n.dbRdr.Height(ctx)
}
