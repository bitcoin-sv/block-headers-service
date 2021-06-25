package service

import (
	"context"

	"github.com/pkg/errors"

	"github.com/libsv/headers-client"
)

type headersService struct {
	hRdr  headers.BlockheaderReader
	hWrtr headers.BlockheaderWriter
}

// NewHeadersService will setup and return a new headers service.
func NewHeadersService(hRdr headers.BlockheaderReader, hWrtr headers.BlockheaderWriter) *headersService {
	return &headersService{hRdr: hRdr, hWrtr: hWrtr}
}

// Header will return a single header by block hash.
func (h *headersService) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	resp, err := h.hRdr.Header(ctx, args)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Create will add a new header.
func (h *headersService) Create(ctx context.Context, req headers.BlockHeader) error {
	// TODO - validate header
	if err := h.hWrtr.Create(ctx, req); err != nil {
		return errors.Wrapf(err, "failed to create blockheader with hash %s", req.Hash)
	}
	return nil
}

// Height will return the current block height stored in the service data store.
func (h *headersService) Height(ctx context.Context) (*headers.Height,error){
	height, err := h.hRdr.Height(ctx)
	if err != nil{
		return nil, errors.Wrap(err, "failed to get current stored height")
	}
	return &headers.Height{Height: height}, nil
}
