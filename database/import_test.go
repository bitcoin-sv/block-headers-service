package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/bitcoin-sv/block-headers-service/internal/tests/assert"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/bitcoin-sv/block-headers-service/service"
)

type testCase struct {
	blockRecord        [][]string
	previousBlockHash  string
	blockHasher        service.BlockHasher
	cumulatedChainWork string
	rowIndex           int
	numberOfBlocks     int
	expected           dto.DbBlockHeader
	actual             dto.DbBlockHeader
}

var timeLayout = "2006-01-02 15:04:05-07:00"
var localTimezone, _ = time.LoadLocation("Local")

// TestPrepareRecordGenesisBlock tests the preparation (parsing and calculation of values) for genesis block,
// checking if the result is valid as expected.
func TestPrepareRecordGenesisBlock(t *testing.T) {

	testCSVRecord := [][]string{
		{"1", "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b", "2083236893", "486604799", "1231006505"},
	}

	blockTimestamp, _ := time.Parse(timeLayout, "2009-01-03 19:15:05+01:00")
	localTime := blockTimestamp.In(localTimezone)

	testOutputGenesisBlock := dto.DbBlockHeader{
		Height:        0,
		Hash:          "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		Version:       1,
		MerkleRoot:    "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		Timestamp:     localTime,
		Bits:          486604799,
		Nonce:         2083236893,
		State:         "LONGEST_CHAIN",
		Chainwork:     "4295032833",
		CumulatedWork: "4295032833",
		PreviousBlock: "0000000000000000000000000000000000000000000000000000000000000000",
	}

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000000000000000000000000000000000000000000000000000",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "0",
		rowIndex:           0,
		numberOfBlocks:     1,
		expected:           testOutputGenesisBlock,
	}
	var err error
	testCase.actual, err = PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)
	if err != nil {
		t.Errorf("Error while preparing record: %v", err)
	}
	assert.Equal[dto.DbBlockHeader](t, testCase.actual, testCase.expected)
}

// TestPrepareRecordTenBlocksBesideTheFork tests the preparation (parsing and calculation of values) for chain of ten blocks beside fork,
// checking if the result is valid as expected.
func TestPrepareRecordTenBlocksBesideTheFork(t *testing.T) {
	testCSVRecords := [][]string{
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
	}

	blockTimestamp, _ := time.Parse(timeLayout, "2018-11-15 19:44:57+01:00")
	localTime := blockTimestamp.In(localTimezone)
	testCaseOutputTenBlocks := dto.DbBlockHeader{
		Height:        556770,
		Hash:          "00000000000000000005569f09a80c66c8ebf514fdd1c03e803799c2420a4f5a",
		Version:       536870912,
		MerkleRoot:    "17e0aefc0154e0a3cdc4a837c66d9c0e0f0e4a44a703fd6e654e8cbc62c0b28f",
		Timestamp:     localTime,
		Bits:          402796026,
		Nonce:         4081063765,
		State:         "LONGEST_CHAIN",
		Chainwork:     "2166624730970898396303",
		CumulatedWork: "255349410425588691745638430",
		PreviousBlock: "0000000000000000005013e7cc2889ada8b01f24dfc325d1398be82197fc623b",
	}

	testCase := testCase{
		blockRecord:        testCSVRecords,
		previousBlockHash:  "000000000000000001f34f5eb45827af756e757498039f43ff6f7585c97f4d16",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "255327261802219463033558368",
		rowIndex:           556761,
		numberOfBlocks:     10,
		expected:           testCaseOutputTenBlocks,
	}

	var calculatedBlocks = []dto.DbBlockHeader{}

	for i := 0; i < testCase.numberOfBlocks; i++ {
		block, err := PrepareRecord(testCase.blockRecord[i], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)
		if err != nil {
			t.Errorf("Error while preparing record: %v", err)
		}
		calculatedBlocks = append(calculatedBlocks, block)
		testCase.previousBlockHash = block.Hash
		testCase.cumulatedChainWork = block.CumulatedWork
		testCase.rowIndex++
	}

	testCase.actual = calculatedBlocks[testCase.numberOfBlocks-1]
	assert.Equal[dto.DbBlockHeader](t, testCase.actual, testCase.expected)
}

// TestPrepareRecordNewerBlock tests the preparation (parsing and calculation of values) for newer (833233) block,
// checking if the result is valid as expected.
func TestPrepareRecordNewerBlock(t *testing.T) {

	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423"},
	}

	blockTimestamp, _ := time.Parse(timeLayout, "2024-02-26 14:33:43+01:00")
	localTime := blockTimestamp.In(localTimezone)
	testOutputGenesisBlock := dto.DbBlockHeader{
		Height:        833233,
		Hash:          "00000000000000000676a9b9cdb44820a04c780ca152737124e36341b6c4cdd2",
		Version:       536870912,
		MerkleRoot:    "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885",
		Timestamp:     localTime,
		Bits:          403300437,
		Nonce:         3035389718,
		State:         "LONGEST_CHAIN",
		Chainwork:     "478151526252246136711",
		CumulatedWork: "409554917150373038158892043",
		PreviousBlock: "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
	}

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
		expected:           testOutputGenesisBlock,
	}
	var err error
	testCase.actual, err = PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)
	if err != nil {
		t.Errorf("Error while preparing record: %v", err)
	}
	assert.Equal[dto.DbBlockHeader](t, testCase.actual, testCase.expected)
}

func TestPrepareRecordLongMerkleRootError(t *testing.T) {

	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885a", "3035389718", "403300437", "1708954423"},
	}
	expectedErrorMessage := "max hash string length is 64 bytes"

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}
	var err error
	testCase.actual, err = PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}

func TestPrepareCharInVersionError(t *testing.T) {

	version := "536870912a"
	testCSVRecord := [][]string{
		{version, "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", "1708954423"},
	}
	expectedErrorMessage := fmt.Sprintf("strconv.Atoi: parsing \"%s\": invalid syntax", version)

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}
	var err error
	testCase.actual, err = PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}

func TestPrepareCharInNonceError(t *testing.T) {

	nonce := "3035389718a"
	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", nonce, "403300437", "1708954423"},
	}
	expectedErrorMessage := fmt.Sprintf("strconv.Atoi: parsing \"%s\": invalid syntax", nonce)

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}

	_, err := PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}

func TestPrepareCharInBitsError(t *testing.T) {

	bits := "403300437a"
	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", bits, "1708954423"},
	}
	expectedErrorMessage := fmt.Sprintf("strconv.Atoi: parsing \"%s\": invalid syntax", bits)

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}

	_, err := PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}

func TestPrepareCharInTimestampError(t *testing.T) {

	timestamp := "1708954423a"
	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437", timestamp},
	}
	expectedErrorMessage := fmt.Sprintf("strconv.ParseInt: parsing \"%s\": invalid syntax", timestamp)

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}

	_, err := PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}

func TestPrepareWrongArgumentCountError(t *testing.T) {

	testCSVRecord := [][]string{
		{"536870912", "e9446d4ebeb301aeb5a2f375ac062bf3581269d783362cf066f08bbe6040a885", "3035389718", "403300437"},
	}
	expectedErrorMessage := fmt.Sprintf("invalid record length: expected 5 elements, got %d", len(testCSVRecord[0]))

	testCase := testCase{
		blockRecord:        testCSVRecord,
		previousBlockHash:  "0000000000000000031817e0b646350cac1b8770d6cba60717e86185cadb15cc",
		blockHasher:        service.DefaultBlockHasher(),
		cumulatedChainWork: "409554438998846785912755332",
		rowIndex:           833233,
		numberOfBlocks:     1,
	}

	_, err := PrepareRecord(testCase.blockRecord[0], testCase.previousBlockHash, testCase.blockHasher, testCase.cumulatedChainWork, testCase.rowIndex)

	assert.IsError(t, err, expectedErrorMessage)
}
