package database

import (
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
)

type testCase struct {
	testID               int
	name                 string
	data                 testCaseData
	expectedBlock        *dto.DbBlockHeader
	expectedErrorMessage string
}

type testCaseData struct {
	blockRecord        [][]string
	previousBlockHash  string
	cumulatedChainWork string
	rowIndex           int
	numberOfBlocks     int
}

var timeLayout = "2006-01-02 15:04:05-07:00"
var localTimezone, _ = time.LoadLocation("Local")

func TestPrepareRecordHappyPath(t *testing.T) {
	testCases := []testCase{
		{
			testID: 1,
			name:   "genesis block, should return valid block header",
			data: testCaseData{
				blockRecord: [][]string{
					{"1", "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", "2083236893", "486604799", "1231006505"},
				},
				previousBlockHash:  "0000000000000000000000000000000000000000000000000000000000000000",
				cumulatedChainWork: "0",
				rowIndex:           0,
				numberOfBlocks:     1,
			},
			expectedBlock: &dto.DbBlockHeader{
				Height:        0,
				Hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
				Version:       1,
				MerkleRoot:    "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
				Timestamp:     timestampInLocalTime("2009-01-03 19:15:05+01:00"),
				Bits:          486604799,
				Nonce:         2083236893,
				State:         "LONGEST_CHAIN",
				Chainwork:     "4295032833",
				CumulatedWork: "4295032833",
				PreviousBlock: "0000000000000000000000000000000000000000000000000000000000000000",
			},
			expectedErrorMessage: "",
		},

		{
			testID: 2,
			name:   "10 blocks beside the fork, should return valid block header",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "0fb7334d7fc33e3284cc47e6682ea19d478716576fc02431977ee856d8a11a7a", "4179160797", "402791861", "1542300873"},
					{"536870912", "334693eb277aa554d2c606041b97f03ba8689e8da373229375f5e5a1bbe5b5e3", "365081954", "402791587", "1542301036"},
					{"536870912", "559b5d64cb554ba90f33b4047941dd4b9797600ed46b63c8d1cd71888b8c6565", "483339668", "402792728", "1542301708"},
					{"536870912", "aa22501b3f7306aaeb7e222b3149becb874e11ff45d38c5e270b653cf3561ab4", "303622189", "402793092", "1542302315"},
					{"536870912", "8049cf9f0e767e1b96998e0b9dd68da9242b7e0e108c83cc02edab38f5854aaa", "379456570", "402791605", "1542303567"},
					{"536870912", "f093773ebea5c6a290d762d3d0b43eca7de94b5b25fd8a93c677f5fc63b3bf16", "1220874517", "402791706", "1542304321"},
					{"536870912", "da2b9eb7e8a3619734a17b55c47bdd6fd855b0afa9c7e14e3a164a279e51bba9", "1301274612", "402792411", "1542305817"},
					{"536870912", "b0ac8183e020907d399efb445e8fd6a90f611545c3f53ee9bab6130e1cf701a8", "522429575", "402792964", "1542306568"},
					{"536870912", "dc3f21e9e6cfbe895e9fc39b39538ed771a4ba8b0be0cb414b89fbcfaeda691f", "889633737", "402795325", "1542307349"},
					{"536870912", "17e0aefc0154e0a3cdc4a837c66d9c0e0f0e4a44a703fd6e654e8cbc62c0b28f", "4081063765", "402796026", "1542307497"},
				},
				previousBlockHash:  "000000000000000001f34f5eb45827af756e757498039f43ff6f7585c97f4d16",
				cumulatedChainWork: "255327261802219463033558368",
				rowIndex:           556761,
				numberOfBlocks:     10,
			},
			expectedBlock: &dto.DbBlockHeader{
				Height:        556770,
				Hash:          "00000000000000000005569f09a80c66c8ebf514fdd1c03e803799c2420a4f5a",
				Version:       536870912,
				MerkleRoot:    "17e0aefc0154e0a3cdc4a837c66d9c0e0f0e4a44a703fd6e654e8cbc62c0b28f",
				Timestamp:     timestampInLocalTime("2018-11-15 19:44:57+01:00"),
				Bits:          402796026,
				Nonce:         4081063765,
				State:         "LONGEST_CHAIN",
				Chainwork:     "2166624730970898396303",
				CumulatedWork: "255349410425588691745638430",
				PreviousBlock: "0000000000000000005013e7cc2889ada8b01f24dfc325d1398be82197fc623b",
			},
			expectedErrorMessage: "",
		},
		{
			testID: 1,
			name:   "newer block, should return valid block header",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock: &dto.DbBlockHeader{
				Height:        833233,
				Hash:          "00000000000000000676a9b9cdb44820a04c780ca152737124e36341b6c4cdd2",
				Version:       536870912,
				MerkleRoot:    "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885",
				Timestamp:     timestampInLocalTime("2024-02-26 14:33:43+01:00"),
				Bits:          403300437,
				Nonce:         3035389718,
				State:         "LONGEST_CHAIN",
				Chainwork:     "478151526252246136711",
				CumulatedWork: "409554917150373038158892043",
				PreviousBlock: "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
			},
			expectedErrorMessage: "",
		},
	}

	for _, tc := range testCases {
		var result = []dto.DbBlockHeader{}
		var err error
		for i := 0; i < tc.data.numberOfBlocks; i++ {
			block, err := PrepareRecord(tc.data.blockRecord[i], tc.data.previousBlockHash, tc.data.cumulatedChainWork, tc.data.rowIndex)
			if err != nil {
				t.Errorf("Error while preparing record: %v", err)
			}
			result = append(result, *block)
			tc.data.previousBlockHash = block.Hash
			tc.data.cumulatedChainWork = block.CumulatedWork
			tc.data.rowIndex++
		}
		assert.Equal[dto.DbBlockHeader](t, result[tc.data.numberOfBlocks-1], *tc.expectedBlock)
		assert.NoError(t, err)
	}
}

func TestPrepareRecordErrorPath(t *testing.T) {

	testCases := []testCase{
		{
			testID: 1,
			name:   "version out of int32 range, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"2147483648", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseInt: parsing \"2147483648\": value out of range",
		},
		{
			testID: 2,
			name:   "nonce out of uint32 range, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"2147483646", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "4294967296", "403300437", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseUint: parsing \"4294967296\": value out of range",
		},
		{
			testID: 3,
			name:   "bits out of uint32 range, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"2147483646", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "4294967296", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseUint: parsing \"4294967296\": value out of range",
		},
		{
			testID: 4,
			name:   "too little values in row, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "invalid record length: expected 5 elements, got 4",
		},
		{
			testID: 5,
			name:   "too much values in row, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "403300437", "403300437"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "invalid record length: expected 5 elements, got 6",
		},
		{
			testID: 6,
			name:   "wrong character in version, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912a", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseInt: parsing \"536870912a\": invalid syntax",
		},
		{
			testID: 7,
			name:   "wrong character in nonce, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718a", "403300437", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseUint: parsing \"3035389718a\": invalid syntax",
		},
		{
			testID: 8,
			name:   "wrong character in bits, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437a", "1708954423"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseUint: parsing \"403300437a\": invalid syntax",
		},
		{
			testID: 9,
			name:   "wrong character in timestamp, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423a"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "strconv.ParseInt: parsing \"1708954423a\": invalid syntax",
		},
		{
			testID: 10,
			name:   "too long markleroot, should return error",
			data: testCaseData{
				blockRecord: [][]string{
					{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885a", "3035389718", "403300437", "1708954423a"},
				},
				previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
				cumulatedChainWork: "409554438998846785912755332",
				rowIndex:           833233,
				numberOfBlocks:     1,
			},
			expectedBlock:        nil,
			expectedErrorMessage: "max hash string length is 64 bytes",
		},
	}

	for _, tc := range testCases {
		result, err := PrepareRecord(tc.data.blockRecord[0], tc.data.previousBlockHash, tc.data.cumulatedChainWork, tc.data.rowIndex)
		assert.Equal[*dto.DbBlockHeader](t, result, tc.expectedBlock)
		assert.IsError(t, err, tc.expectedErrorMessage)
	}
}

func timestampInLocalTime(timestamp string) time.Time {
	blockTimestamp, _ := time.Parse(timeLayout, timestamp)
	localTime := blockTimestamp.In(localTimezone)
	return localTime
}
