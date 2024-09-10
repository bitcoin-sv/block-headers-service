package service

import (
	"fmt"
	"testing"

	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
)

func TestMerkleRootConfirmations(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		request  []domains.MerkleRootConfirmationRequestItem
		expected []*domains.MerkleRootConfirmation
	}{
		{
			request: []domains.MerkleRootConfirmationRequestItem{
				{
					MerkleRoot:  "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight: 1,
				},
				{
					MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight: 2,
				},
			},
			expected: []*domains.MerkleRootConfirmation{
				{
					MerkleRoot:   "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight:  1,
					Hash:         "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
					Confirmation: domains.Confirmed,
				},
				{
					MerkleRoot:   "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
					BlockHeight:  2,
					Hash:         "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd",
					Confirmation: domains.Confirmed,
				},
			},
		},
		{
			request:  []domains.MerkleRootConfirmationRequestItem{},
			expected: []*domains.MerkleRootConfirmation{},
		},
		{
			request: []domains.MerkleRootConfirmationRequestItem{
				{
					MerkleRoot:  "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight: 1,
				},
				{
					MerkleRoot:  "invalid_merkle_root_abc123123",
					BlockHeight: 2,
				},
				{
					MerkleRoot:  "unable_to_verify_merkle_root_abc123",
					BlockHeight: 8, // Bigger than top height
				},
				{
					MerkleRoot:  "invalid_merkle_root_over_the_excess",
					BlockHeight: 100, // Bigger than top height + allowed excess
				},
			},
			expected: []*domains.MerkleRootConfirmation{
				{
					MerkleRoot:   "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
					BlockHeight:  1,
					Hash:         "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048",
					Confirmation: domains.Confirmed,
				},
				{
					MerkleRoot:   "invalid_merkle_root_abc123123",
					BlockHeight:  2,
					Hash:         "",
					Confirmation: domains.Invalid,
				},
				{
					MerkleRoot:   "unable_to_verify_merkle_root_abc123",
					BlockHeight:  8,
					Hash:         "",
					Confirmation: domains.UnableToVerify,
				},
				{
					MerkleRoot:   "invalid_merkle_root_over_the_excess",
					BlockHeight:  100,
					Hash:         "",
					Confirmation: domains.Invalid,
				},
			},
		},
	}

	for _, tt := range testCases {
		mrcfs, _ := tData.hs.Merkleroots.GetMerkleRootsConfirmations(tt.request)
		for i, mrcf := range mrcfs {
			assert.Equal(t, mrcf.Hash, tt.expected[i].Hash)
			assert.Equal(t, mrcf.BlockHeight, tt.expected[i].BlockHeight)
			assert.Equal(t, mrcf.Confirmation, tt.expected[i].Confirmation)
			assert.Equal(t, mrcf.MerkleRoot, tt.expected[i].MerkleRoot)
		}
	}
}

func TestMerkleRoots(t *testing.T) {
	tData := setUpServices()
	orderedByField := "BlockHeight"
	sortDirection := "ASC"

	testCases := []struct {
		RequestBatchSize        int
		RequestLastEvaluatedKey int
		Expected                domains.MerkleRootsESKPagedResponse
	}{
		{
			RequestBatchSize:        2,
			RequestLastEvaluatedKey: 1,
			Expected: domains.MerkleRootsESKPagedResponse{
				Content: []*domains.MerkleRootsResponse{
					{
						MerkleRoot:  "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
						BlockHeight: 2,
					},
					{
						MerkleRoot:  "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
						BlockHeight: 3,
					},
				},
				Page: domains.ExclusiveStartKeyPage[int]{
					OrderByField:     &orderedByField,
					SortDirection:    &sortDirection,
					TotalElements:    5,
					Size:             2,
					LastEvaluatedKey: 3,
				},
			},
		},
	}

	for _, tt := range testCases {
		merkleRoots, _ := tData.hs.Merkleroots.GetMerkleRoots(tt.RequestBatchSize, tt.RequestLastEvaluatedKey)
		assert.Equal(t, merkleRoots.Page.LastEvaluatedKey, tt.Expected.Page.LastEvaluatedKey)
		assert.Equal(t, merkleRoots.Page.OrderByField, tt.Expected.Page.OrderByField)
		assert.Equal(t, merkleRoots.Page.SortDirection, tt.Expected.Page.SortDirection)
		assert.Equal(t, merkleRoots.Page.TotalElements, tt.Expected.Page.TotalElements)
		assert.Equal(t, merkleRoots.Page.Size, tt.Expected.Page.Size)

		for i, mr := range merkleRoots.Content {
			assert.Equal(t, mr.BlockHeight, tt.Expected.Content[i].BlockHeight)
			assert.Equal(t, mr.MerkleRoot, tt.Expected.Content[i].MerkleRoot)
		}
		fmt.Print(&merkleRoots)
	}
}
