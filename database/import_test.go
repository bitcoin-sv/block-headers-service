package database

import (
	"github.com/rs/zerolog"
	"os"
	"testing"

	"github.com/bitcoin-sv/pulse/internal/tests/assert"
)

func TestImportHeadersFromFile(t *testing.T) {
	// given
	// Create a temporary CSV file for testing
	content := []byte("hash,version,merkleroot,nonce,bits,chainwork,timestamp,cumulatedWork\n000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f,1,4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b,2083236893,486604799,4295032833,1231006505,4295032833\n000000009b7262315dbf071787ad3656097b892abffd1f95a1a022f896f533fc,1,63522845d294ee9b0188ae5cac91bf389a0c3723f084ca1025e7d9cdfe481ce1,2011431709,486604799,4295032833,1231471428,25770196998\n000000003031a0e73735690c5a1ff2a4be82553b2a12b776fbd3a215dc8f778d,1,20251a76e64e920e58291a30d4b212939aae976baca40e70818ceaa596fb9d37,2538380312,486604799,4295032833,1231471789,30065229831")
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up after closing the file

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	repo := MockHeaders{db: nil}
	log := zerolog.Nop()

	// when
	_, err = importHeadersFromFile(&repo, tmpfile, &log)

	// then
	assert.NoError(t, err)
}
