package service

import (
	"math/big"
	"testing"

	"github.com/rs/zerolog"

	"github.com/bitcoin-sv/block-headers-service/config"
	"github.com/bitcoin-sv/block-headers-service/domains"
	"github.com/bitcoin-sv/block-headers-service/internal/chaincfg/chainhash"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/fixtures"
	"github.com/bitcoin-sv/block-headers-service/internal/tests/testrepository"
	"github.com/bitcoin-sv/block-headers-service/internal/wire"
	"github.com/bitcoin-sv/block-headers-service/repository"
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

	fifthHeader := createHeader(5, *fixtures.HashHeight5, *fixtures.HashHeight4)
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

	fifthHeader := createHeader(5, *fixtures.HashHeight5, *fixtures.HashHeight4)
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
			ancestorHash:  "00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048", // Height = 1
			expectedCount: 3,
			expectedError: false,
		},
		{
			hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f", // Height = 0
			ancestorHash:  "000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485", // Height = 4
			expectedCount: 0,
			expectedError: true,
		},
		{
			hash:          "000000007bc154e0fa7ea32218a72fe2c1bb9f86cf8c9ebf9a715ed27fdb229a", // Height = 100
			ancestorHash:  "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
			expectedCount: 0,
			expectedError: true,
		},
		{
			hash:          "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
			ancestorHash:  "000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd", // Height = 2
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
		header, err := tData.hs.Headers.GetCommonAncestor(tt.hashes)

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

	// Check tips without fork
	tips, _ := tData.hs.Headers.GetTips()
	assert.Equal(t, len(tips), 1)

	// Check tip with fork
	forkHeader := createHeader(4, *fixtures.HashHeight6, *fixtures.HashHeight5)
	*tData.db = append(*tData.db, forkHeader)
	tips, _ = tData.hs.Headers.GetTips()
	assert.Equal(t, len(tips), 2)
}

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
		mrcfs, _ := tData.hs.Headers.GetMerkleRootsConfirmations(tt.request)

		for i, mrcf := range mrcfs {
			assert.Equal(t, mrcf.Hash, tt.expected[i].Hash)
			assert.Equal(t, mrcf.BlockHeight, tt.expected[i].BlockHeight)
			assert.Equal(t, mrcf.Confirmation, tt.expected[i].Confirmation)
			assert.Equal(t, mrcf.MerkleRoot, tt.expected[i].MerkleRoot)
		}
	}
}

func TestLocateHeadersGetHeaders(t *testing.T) {
	tData := setUpServices()

	testCases := []struct {
		name                 string
		locator              []*chainhash.Hash
		hashstop             *chainhash.Hash
		result               []*wire.BlockHeader
		expectedErrorMessage string
	}{
		{
			name: "happy path, all block locator hashes are valid (hashStart height equals 3), hashstop is valid (height equals 4), returns block with height 4",
			locator: []*chainhash.Hash{
				toChainhashPtr("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
				toChainhashPtr("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd"),
				toChainhashPtr("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"),
				toChainhashPtr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},

			hashstop: toChainhashPtr("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485"),
			result: []*wire.BlockHeader{
				{
					Version:    1,
					PrevBlock:  toChainhash("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
					MerkleRoot: toChainhash("df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"),
					Bits:       486604799,
					Nonce:      2850094635,
				},
			},
			expectedErrorMessage: "",
		},
		{
			name: "happy path, all block locator hashes are invalid, hashstop is valid (height equals 3), returns blocks from height 1 to 3",
			locator: []*chainhash.Hash{
				toChainhashPtr("bad000008d9dc510f23c2657fc4f67bea30078cc05a90eb89e84cc475c080805"),
				toChainhashPtr("bad000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd"),
				toChainhashPtr("bad00000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"),
				toChainhashPtr("bad000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},

			hashstop: toChainhashPtr("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
			result: []*wire.BlockHeader{
				{
					Version:    1,
					PrevBlock:  toChainhash("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
					MerkleRoot: toChainhash("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"),
					Bits:       486604799,
					Nonce:      2573394689,
				},
				{
					Version:    1,
					PrevBlock:  toChainhash("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"),
					MerkleRoot: toChainhash("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"),
					Bits:       486604799,
					Nonce:      1639830024,
				},
				{
					Version:    1,
					PrevBlock:  toChainhash("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd"),
					MerkleRoot: toChainhash("999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"),
					Bits:       486604799,
					Nonce:      1844305925,
				},
			},
			expectedErrorMessage: "",
		},
		{
			name: "happy path, all block locator hashes are invalid, hashstop is equal to 0, returns blocks from height 1 to 3",
			locator: []*chainhash.Hash{
				toChainhashPtr("bad000008d9dc510f23c2657fc4f67bea30078cc05a90eb89e84cc475c080805"),
				toChainhashPtr("bad000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd"),
				toChainhashPtr("bad00000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"),
				toChainhashPtr("bad000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},
			result: []*wire.BlockHeader{
				{
					Version:    1,
					PrevBlock:  toChainhash("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
					MerkleRoot: toChainhash("0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098"),
					Bits:       486604799,
					Nonce:      2573394689,
				},
				{
					Version:    1,
					PrevBlock:  toChainhash("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048"),
					MerkleRoot: toChainhash("9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5"),
					Bits:       486604799,
					Nonce:      1639830024,
				},
				{
					Version:    1,
					PrevBlock:  toChainhash("000000006a625f06636b8bb6ac7b960a8d03705d1ace08b1a19da3fdcc99ddbd"),
					MerkleRoot: toChainhash("999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644"),
					Bits:       486604799,
					Nonce:      1844305925,
				},
				{
					Version:    1,
					PrevBlock:  toChainhash("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
					MerkleRoot: toChainhash("df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"),
					Bits:       486604799,
					Nonce:      2850094635,
				},
			},
			hashstop:             toChainhashPtr("0"),
			expectedErrorMessage: "",
		},
		{
			name: "unhappy path, valid locator hash height equals 4, hashstop is valid (height equals 3), returns error",
			locator: []*chainhash.Hash{
				toChainhashPtr("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485"),
			},

			hashstop:             toChainhashPtr("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
			result:               []*wire.BlockHeader{},
			expectedErrorMessage: "hashStop is lower than first valid height",
		},
		{
			name: "unhappy path, no locators, hashstop is valid (height equals 3), returns error",
			locator: []*chainhash.Hash{
				toChainhashPtr("000000004ebadb55ee9096c9a2f8880e09da59c0d68b1c228da88e48844a1485"),
			},

			hashstop:             toChainhashPtr("0000000082b5015589a3fdf2d4baff403e6f0be035a5d9742c1cae6295464449"),
			result:               []*wire.BlockHeader{},
			expectedErrorMessage: "hashStop is lower than first valid height",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tData.hs.Headers.LocateHeadersGetHeaders(tc.locator, tc.hashstop)
			if err != nil {
				assert.Equal(t, err.Error(), tc.expectedErrorMessage)
			} else {
				for j, r := range result {
					assert.Equal(t, r.Version, tc.result[j].Version)
					assert.Equal(t, r.PrevBlock, tc.result[j].PrevBlock)
					assert.Equal(t, r.MerkleRoot, tc.result[j].MerkleRoot)
					assert.Equal(t, r.Bits, tc.result[j].Bits)
					assert.Equal(t, r.Nonce, tc.result[j].Nonce)
				}
			}
		})
	}
}

func setUpServices() *testData {
	log := zerolog.Nop()
	db, _ := fixtures.LongestChain()
	var array []domains.BlockHeader = db
	repo := &repository.Repositories{
		Headers: testrepository.NewHeadersTestRepository(&array),
	}

	p2pcfg := config.GetDefaultAppConfig().P2P
	mrconfig := config.MerkleRootConfig{
		MaxBlockHeightExcess: 6,
	}
	cfg := config.AppConfig{
		P2P:        p2pcfg,
		MerkleRoot: &mrconfig,
	}
	hs := NewServices(Dept{
		Repositories: repo,
		Peers:        nil,
		Logger:       &log,
		Config:       &cfg,
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

func toChainhash(hash string) chainhash.Hash {
	chainhash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		panic("Invalid hash string")
	}

	return *chainhash
}

func toChainhashPtr(hash string) *chainhash.Hash {
	chainhash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		panic("Invalid hash string")
	}

	return chainhash
}
