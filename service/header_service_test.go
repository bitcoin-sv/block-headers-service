package service

import (
	"math/big"
	"testing"

	"github.com/libsv/bitcoin-hc/domains"
	"github.com/libsv/bitcoin-hc/internal/chaincfg/chainhash"
	"github.com/libsv/bitcoin-hc/internal/tests/assert"
	"github.com/libsv/bitcoin-hc/internal/tests/testrepository"
	"github.com/libsv/bitcoin-hc/repository"
)

type testData struct {
	db *[]domains.BlockHeader
	hs *Services
}

func TestGetHeadersByHeight(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		height        int
		count         int
		expectedError bool
		expectedCount int
	}{
		{
			height:        1,
			count:         1,
			expectedError: false,
			expectedCount: 1,
		},
		{
			height:        2,
			count:         2,
			expectedError: false,
			expectedCount: 2,
		},
		{
			height:        100,
			count:         1,
			expectedError: true,
			expectedCount: 0,
		},
		{
			height:        1,
			count:         100,
			expectedError: false,
			expectedCount: 4,
		},
	}

	for _, tt := range testCases {
		headers, err := tData.hs.Headers.GetHeadersByHeight(tt.height, tt.count)

		assert.Equal(t, err != nil, tt.expectedError)
		assert.Equal(t, len(headers), tt.expectedCount)
	}
}

func TestGetHeadersByHash(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		hash          string
		expectedError bool
	}{
		{
			// Genesis
			hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
			expectedError: false,
		},
		{
			hash:          "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd",
			expectedError: false,
		},
		{
			// Height: 100
			hash:          "000000007bc154e0fa7ea32218a72fe2c1bb9f86cf8c9ebf9a715ed27fdb229a",
			expectedError: true,
		},
	}

	for _, tt := range testCases {
		header, err := tData.hs.Headers.GetHeaderByHash(tt.hash)

		assert.Equal(t, err != nil, tt.expectedError)
		assert.Equal(t, header == nil, tt.expectedError)
		if !tt.expectedError {
			assert.Equal(t, header.Hash.String(), tt.hash)
		}
	}
}

func TestCountHeaders(t *testing.T) {
	tData := setUpServices()

	count := tData.hs.Headers.CountHeaders()
	assert.Equal(t, count, 5)

	fifthHeader := createHeader(5, *testrepository.FifthHash, *testrepository.FourthHash)
	*tData.db = append(*tData.db, fifthHeader)

	count = tData.hs.Headers.CountHeaders()
	assert.Equal(t, count, 6)
}

func TestGetTipAndGetTipHeight(t *testing.T) {
	tData := setUpServices()

	tip := tData.hs.Headers.GetTip()
	tipHeight := tData.hs.Headers.GetTipHeight()
	assert.Equal(t, tip != nil, true)
	assert.Equal(t, tip.Height, 4)
	assert.Equal(t, tip.Height, tipHeight)

	fifthHeader := createHeader(5, *testrepository.FifthHash, *testrepository.FourthHash)
	*tData.db = append(*tData.db, fifthHeader)

	tip = tData.hs.Headers.GetTip()
	tipHeight = tData.hs.Headers.GetTipHeight()
	assert.Equal(t, tip != nil, true)
	assert.Equal(t, tip.Height, 5)
	assert.Equal(t, tip.Height, tipHeight)
}

func TestGetHeaderAncestorsByHash(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		hash          string
		ancestorHash  string
		expectedCount int
		expectedError bool
	}{
		{
			hash:          "0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449", // Height = 3
			ancestorHash:  "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048", //Height = 1
			expectedCount: 3,
			expectedError: false,
		},
		{
			hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f", // Height = 0
			ancestorHash:  "000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485", //Height = 4
			expectedCount: 0,
			expectedError: true,
		},
		{
			hash:          "000000007bc154e0fa7ea32218a72fe2c1bb9f86cf8c9ebf9a715ed27fdb229a", // Height = 100
			ancestorHash:  "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", //Height = 2
			expectedCount: 0,
			expectedError: true,
		},
		{
			hash:          "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
			ancestorHash:  "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", //Height = 2
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tt := range testCases {
		headers, err := tData.hs.Headers.GetHeaderAncestorsByHash(tt.hash, tt.ancestorHash)

		assert.Equal(t, err != nil, tt.expectedError)
		assert.Equal(t, len(headers), tt.expectedCount)
	}
}

func TestGetCommonAncestor(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		hashes         []string
		ancestorHeight int32
		expectedError  bool
	}{
		{
			hashes: []string{
				"000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485", // Height = 4
				"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
			},
			ancestorHeight: 1,
			expectedError:  false,
		},
		{
			hashes: []string{
				"000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485", // Height = 4
				"0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449", // Height = 4 Fork
				"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
				"00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048", // Height = 1
			},
			ancestorHeight: 0,
			expectedError:  false,
		},
		{
			hashes: []string{
				"000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485", // Height = 4
				"000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
				"00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048", // Height = 1
				"000000007bc154e0fa7ea32218a72fe2c1bb9f86cf8c9ebf9a715ed27fdb229a", // Height = 100
			},
			ancestorHeight: 0,
			expectedError:  true,
		},
	}

	for _, tt := range testCases {
		header, err := tData.hs.Headers.GetCommonAncestors(tt.hashes)

		assert.Equal(t, err != nil, tt.expectedError)
		if header != nil {
			assert.Equal(t, header.Height, tt.ancestorHeight)
		}
	}
}

func TestLatestHeaderLocator(t *testing.T) {
	tData := setUpServices()

	locator := tData.hs.Headers.LatestHeaderLocator()
	headers, _ := tData.hs.Headers.GetHeadersByHeight(0, 5)

	for _, header := range headers {
		check := false
		for _, hash := range locator {
			if *hash == header.Hash {
				check = true
			}
		}
		assert.Equal(t, check, true)
	}
}

func TestGetAllTips(t *testing.T) {
	tData := setUpServices()

	//Check tips without fork
	tips, _ := tData.hs.Headers.GetTips()
	assert.Equal(t, len(tips), 1)

	//Check tip with fork
	forkHeader := createHeader(4, *testrepository.SixthHash, *testrepository.FifthHash)
	*tData.db = append(*tData.db, forkHeader)
	tips, _ = tData.hs.Headers.GetTips()
	assert.Equal(t, len(tips), 2)
}

func setUpServices() *testData {
	db, _ := testrepository.LongestChain()
	var array []domains.BlockHeader = db
	repo := &repository.Repositories{
		Headers: testrepository.NewHeadersTestRepository(&array),
	}

	hs := NewServices(Dept{
		Repositories: repo,
		Peers:        nil,
	})

	return &testData{
		db: &array,
		hs: hs,
	}
}

func createHeader(height int32, hash chainhash.Hash, prevBlock chainhash.Hash) domains.BlockHeader {
	return domains.BlockHeader{
		Height:        height,
		Hash:          hash,
		PreviousBlock: prevBlock,
		Chainwork:     big.NewInt(4295032833),
	}
}
