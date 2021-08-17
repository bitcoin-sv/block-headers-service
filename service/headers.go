package service

import (
	"context"

	"github.com/pkg/errors"

	headers "github.com/libsv/bitcoin-hc"
)

type headersService struct {
	hRdr          headers.BlockheaderReader
	hWrtr         headers.BlockheaderWriter
	networkHeight headers.HeightReader
}

// NewHeadersService will setup and return a new headers service.
func NewHeadersService(hRdr headers.BlockheaderReader, hWrtr headers.BlockheaderWriter, networkHeight headers.HeightReader) *headersService {
	return &headersService{hRdr: hRdr, hWrtr: hWrtr, networkHeight: networkHeight}
}

// Header will return a single header by block hash.
func (h *headersService) Header(ctx context.Context, args headers.HeaderArgs) (*headers.BlockHeader, error) {
	if err := args.Validate(); err != nil {
		return nil, err
	}
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

// CreateBatch will batch insert headers.
func (h *headersService) CreateBatch(ctx context.Context, req []*headers.BlockHeader) error {
	return errors.Wrap(h.hWrtr.CreateBatch(ctx, req), "failed to insert blockheaders")
}

// Height will return the current block height stored in the service data store.
func (h *headersService) Height(ctx context.Context) (*headers.Height, error) {
	height, err := h.hRdr.Height(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current stored height")
	}
	networkHeight, err := h.networkHeight.Height(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current network height")
	}
	return &headers.Height{Height: height, NetworkHeight: networkHeight, Synced: height == networkHeight}, nil
}
