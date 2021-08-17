package data

import (
	headers "github.com/libsv/bitcoin-hc"
)

type nodeHeaderFacade struct {
	nodeRdr headers.BlockheaderReader
	dbRdr   headers.BlockheaderReader
}

func NewNodeHeaderFacade(nodeRdr headers.BlockheaderReader, dbRdr headers.BlockheaderReader) *nodeHeaderFacade {
	return &nodeHeaderFacade{
		nodeRdr: nodeRdr,
		dbRdr:   dbRdr,
	}
}
