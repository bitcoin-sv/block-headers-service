package bitcoin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Utxo bitindex comment
type Utxo struct {
	TxID   string `json:"txid"`
	Vout   uint32 `json:"vout"`
	Height uint32 `json:"height"`
	Value  uint64 `json:"value"`
}

// UtxoResponse comment
type UtxoResponse struct {
	Address string `json:"address"`
	Utxos   []Utxo `json:"utxos"`
	Balance uint64 `json:"balance"`
}

type bitIndexResponseData struct {
	Data UtxoResponse `json:"data"`
}

// BitIndex comment
type BitIndex struct {
	BaseURL string
}

// NewBitIndexClient returns a new bitIndex client for the given url
func NewBitIndexClient(url string) (*BitIndex, error) {
	return &BitIndex{
		BaseURL: url,
	}, nil
}

// GetUtxos comment
func (b *BitIndex) GetUtxos(addr string) (*UtxoResponse, error) {
	bitindexURL := fmt.Sprintf("%s/%s/%s", b.BaseURL, "utxos", addr)

	resp, err := http.Get(bitindexURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bir := bitIndexResponseData{}
	err = json.Unmarshal(body, &bir)
	if err != nil {
		fmt.Printf("error unarshalling body %+v", err)
		return nil, err
	}

	return &bir.Data, nil
}
