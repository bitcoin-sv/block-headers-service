package bitcoin

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	cache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

// A Bitcoind represents a Bitcoind client
type Bitcoind struct {
	client  *rpcClient
	Storage *cache.Cache
	group   singleflight.Group
}

// New return a new bitcoind
func New(host string, port int, user, passwd string, useSSL bool) (*Bitcoind, error) {
	rpcClient, err := newClient(host, port, user, passwd, useSSL)
	if err != nil {
		return nil, err
	}

	defaultExpiration := 5 * time.Second
	cleanupInterval := 10 * time.Second

	return &Bitcoind{
		client:  rpcClient,
		Storage: cache.New(defaultExpiration, cleanupInterval),
		group:   singleflight.Group{},
	}, nil
}

func (b *Bitcoind) call(method string, params []interface{}) (rpcResponse, error) {
	key := fmt.Sprintf("%s|%v", method, params)
	// Check cache
	value, found := b.Storage.Get(key)
	if found {
		// fmt.Printf("CACHED: ")
		return value.(rpcResponse), nil
	}

	// Combine memoized function with a cache store
	value, err, _ := b.group.Do(key, func() (interface{}, error) {
		// fmt.Printf("EXECED: ")
		data, innerErr := b.client.call(method, params)

		if innerErr == nil {
			b.Storage.Set(key, data, cache.DefaultExpiration)
		}

		return data, innerErr
	})
	return value.(rpcResponse), err
}

// GetConnectionCount returns the number of connections to other nodes.
func (b *Bitcoind) GetConnectionCount() (count uint64, err error) {
	r, err := b.call("getconnectioncount", nil)
	if err != nil {
		return 0, err
	}
	count, err = strconv.ParseUint(string(r.Result), 10, 64)
	return
}

// GetBlockchainInfo returns the number of connections to other nodes.
func (b *Bitcoind) GetBlockchainInfo() (info BlockchainInfo, err error) {
	r, err := b.call("getblockchaininfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetInfo returns the number of connections to other nodes.
func (b *Bitcoind) GetInfo() (info GetInfo, err error) {
	r, err := b.call("getinfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetNetworkInfo returns the number of connections to other nodes.
func (b *Bitcoind) GetNetworkInfo() (info NetworkInfo, err error) {
	r, err := b.call("getnetworkinfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetNetTotals returns the number of connections to other nodes.
func (b *Bitcoind) GetNetTotals() (totals NetTotals, err error) {
	r, err := b.call("getnettotals", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &totals)
	return
}

// GetMiningInfo comment
func (b *Bitcoind) GetMiningInfo() (info MiningInfo, err error) {
	r, err := b.call("getmininginfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// Uptime returns the number of connections to other nodes.
func (b *Bitcoind) Uptime() (uptime uint64, err error) {
	r, err := b.call("uptime", nil)
	if err != nil {
		return 0, err
	}
	uptime, err = strconv.ParseUint(string(r.Result), 10, 64)
	return
}

// GetPeerInfo returns the number of connections to other nodes.
func (b *Bitcoind) GetPeerInfo() (info PeerInfo, err error) {
	r, err := b.call("getpeerinfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetMempoolInfo comment
func (b *Bitcoind) GetMempoolInfo() (info MempoolInfo, err error) {
	r, err := b.call("getmempoolinfo", nil)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &info)
	return
}

// GetRawMempool returns the number of connections to other nodes.
func (b *Bitcoind) GetRawMempool(details bool) (raw []byte, err error) {
	p := []interface{}{details}
	r, err := b.call("getrawmempool", p)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	//err = json.Unmarshal(r.Result, &raw)
	raw, err = json.Marshal(r.Result)
	return
}

// GetChainTxStats returns the number of connections to other nodes.
func (b *Bitcoind) GetChainTxStats(blockcount int) (stats ChainTXStats, err error) {
	p := []interface{}{blockcount}
	r, err := b.call("getchaintxstats", p)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &stats)
	return
}

// ValidateAddress returns the number of connections to other nodes.
func (b *Bitcoind) ValidateAddress(address string) (addr Address, err error) {
	p := []interface{}{address}
	r, err := b.call("validateaddress", p)
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &addr)
	return
}

// GetHelp returns the number of connections to other nodes.
func (b *Bitcoind) GetHelp() (j []byte, err error) {
	r, err := b.call("help", nil)
	if err != nil {
		return
	}
	j, err = json.Marshal(r.Result)

	return
}

// GetBestBlockHash comment
func (b *Bitcoind) GetBestBlockHash() (hash string, err error) {
	r, err := b.call("getbestblockhash", nil)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(r.Result, &hash); err != nil {
		return "", err
	}
	return
}

// GetBlockHash comment
func (b *Bitcoind) GetBlockHash(blockHeight int) (blockHash string, err error) {
	p := []interface{}{blockHeight}
	r, err := b.call("getblockhash", p)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(r.Result, &blockHash); err != nil {
		return "", err
	}

	return
}

// SendRawTransaction comment
func (b *Bitcoind) SendRawTransaction(hex string) (txid string, err error) {
	r, err := b.call("sendrawtransaction", []interface{}{hex, true})
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(r.Result, &txid); err != nil {
		return "", err
	}

	return
}

// SendRawTransactionWithoutFeeCheck comment
func (b *Bitcoind) SendRawTransactionWithoutFeeCheck(hex string) (txid string, err error) {
	r, err := b.call("sendrawtransaction", []interface{}{hex, false, true})
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(r.Result, &txid); err != nil {
		return "", err
	}

	return
}

// SignRawTransaction comment
func (b *Bitcoind) SignRawTransaction(hex string) (sr *SignRawTransactionResponse, err error) {
	r, err := b.call("signrawtransaction", []interface{}{hex})
	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &sr)
	return
}

// GetBlock returns information about the block with the given hash.
func (b *Bitcoind) GetBlock(blockHash string) (block *Block, err error) {
	r, err := b.call("getblock", []interface{}{blockHash})

	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &block)
	return
}

// GetRawBlock returns the raw bytes of the block with the given hash.
func (b *Bitcoind) GetRawBlock(blockHash string) ([]byte, error) {
	r, err := b.call("getblock", []interface{}{blockHash, 0})
	if err != nil {
		return nil, err
	}

	var rawHex string
	err = json.Unmarshal(r.Result, &rawHex)
	if err != nil {
		return nil, err
	}

	res, err := hex.DecodeString(rawHex)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetBlockOverview returns basic information about the block with the given hash.
func (b *Bitcoind) GetBlockOverview(blockHash string) (block *BlockOverview, err error) {
	r, err := b.call("getblock", []interface{}{blockHash})

	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &block)
	return
}

// GetBlockHeaderHex returns the block header hex for the given hash.
func (b *Bitcoind) GetBlockHeaderHex(blockHash string) (blockHeader *string, err error) {
	r, err := b.call("getblockheader", []interface{}{blockHash, false})

	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &blockHeader)
	return
}

// GetBlockHeader returns the block header for the given hash.
func (b *Bitcoind) GetBlockHeader(blockHash string) (blockHeader *BlockHeader, err error) {
	r, err := b.call("getblockheader", []interface{}{blockHash})

	if err != nil {
		return
	}

	if r.Err != nil {
		rr := r.Err.(map[string]interface{})
		err = fmt.Errorf("ERROR %s: %s", rr["code"], rr["message"])
		return
	}

	err = json.Unmarshal(r.Result, &blockHeader)
	return
}

// GetBlockHex returns information about the block with the given hash.
func (b *Bitcoind) GetBlockHex(blockHash string) (raw *string, err error) {
	r, err := b.call("getblock", []interface{}{blockHash, 0})
	if err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &raw)
	return
}

// GetBlockHeaderAndCoinbase returns information about the block with the given hash.
func (b *Bitcoind) GetBlockHeaderAndCoinbase(blockHash string) (blockHeaderAndCoinbase *BlockHeaderAndCoinbase, err error) {
	r, err := b.call("getblock", []interface{}{blockHash, 3})
	if err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &blockHeaderAndCoinbase)
	return
}

// GetRawTransaction returns raw transaction representation for given transaction id.
func (b *Bitcoind) GetRawTransaction(txID string) (rawTx *RawTransaction, err error) {
	if txID == "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b" {
		// This is the genesis coinbase transaction and cannot be retrieved in this way.
		return &RawTransaction{
			Hex:      "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000",
			TxID:     "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			Hash:     "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
			Version:  1,
			Size:     204,
			LockTime: 0,
			Vin: []Vin{
				{
					Coinbase: "04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73",
					Sequence: 4294967295,
				},
			},
			Vout: []Vout{
				{
					Value: 50.00000000,
					N:     0,
					ScriptPubKey: ScriptPubKey{
						ASM:       "04678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5f OP_CHECKSIG",
						Hex:       "4104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac",
						ReqSigs:   1,
						Type:      "pubkey",
						Addresses: []string{"1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"},
					},
				},
			},
			BlockHash: "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
			Time:      1231006505,
			Blocktime: 1231006505,
		}, nil
	}

	r, err := b.call("getrawtransaction", []interface{}{txID, 1})
	if err != nil {
		return
	}
	err = json.Unmarshal(r.Result, &rawTx)
	return
}

// GetRawTransactionHex returns raw transaction representation for given transaction id.
func (b *Bitcoind) GetRawTransactionHex(txID string) (rawTx *string, err error) {
	if txID == "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b" {
		// This is the genesis coinbase transaction and cannot be retrieved in this way.
		genesisHex := "01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"
		return &genesisHex, nil
	}

	r, err := b.call("getrawtransaction", []interface{}{txID, 0})
	if err != nil {
		return
	}

	err = json.Unmarshal(r.Result, &rawTx)
	return
}

// GetBlockTemplate comment
func (b *Bitcoind) GetBlockTemplate(includeSegwit bool) (template *BlockTemplate, err error) {
	params := gbtParams{}
	if includeSegwit {
		params = gbtParams{
			Mode:         "",
			Capabilities: []string{},
			Rules:        []string{"segwit"},
		}
	}

	r, err := b.call("getblocktemplate", []interface{}{params})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(r.Result, &template); err != nil {
		return nil, err
	}

	return
}

// GetMiningCandidate comment
func (b *Bitcoind) GetMiningCandidate() (template *MiningCandidate, err error) {

	r, err := b.call("getminingcandidate", nil)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(r.Result, &template); err != nil {
		return nil, err
	}

	return
}

// SubmitBlock comment
func (b *Bitcoind) SubmitBlock(hexData string) (result string, err error) {
	r, err := b.client.call("submitblock", []interface{}{hexData})
	if err != nil || r.Err != nil || string(r.Result) != "null" {
		msg := fmt.Sprintf("******* BLOCK SUBMIT FAILED with error: %+v and result: %s\n", err, string(r.Result))
		return "", errors.New(msg)
	}

	return string(r.Result), nil
}

// SubmitMiningSolution comment
func (b *Bitcoind) SubmitMiningSolution(miningCandidateID string, nonce uint32, coinbase string, time uint32, version uint32) (result string, err error) {
	params := submitMiningSolutionParams{
		ID:       miningCandidateID,
		Nonce:    nonce,
		Coinbase: coinbase,
		Time:     time,
		Version:  version,
	}

	r, err := b.client.call("submitminingsolution", []interface{}{params})
	if (err != nil && err.Error() != "") || r.Err != nil || (string(r.Result) != "null" && string(r.Result) != "true") {
		msg := fmt.Sprintf("******* BLOCK SUBMIT FAILED with error: %+v and result: %s\n", err, string(r.Result))
		return "", errors.New(msg)
	}

	return string(r.Result), nil
}

// GetDifficulty comment
func (b *Bitcoind) GetDifficulty() (difficulty float64, err error) {
	r, err := b.call("getdifficulty", nil)
	if err != nil {
		return 0.0, err
	}

	difficulty, err = strconv.ParseFloat(string(r.Result), 64)
	return
}

// DecodeRawTransaction comment
func (b *Bitcoind) DecodeRawTransaction(txHex string) (string, error) {
	r, err := b.call("decoderawtransaction", []interface{}{txHex})
	if err != nil {
		return "", err
	}

	return string(r.Result), nil
}

// GetTxOut comment
func (b *Bitcoind) GetTxOut(txHex string, vout int, includeMempool bool) (string, error) {
	r, err := b.call("gettxout", []interface{}{txHex, vout, includeMempool})
	if err != nil {
		return "", err
	}

	return string(r.Result), nil
}

// ListUnspent comment
func (b *Bitcoind) ListUnspent(addresses []string) (res []*UnspentTransaction, err error) {
	var minConf uint32 = 0
	var maxConf uint32 = 9999999

	r, err := b.call("listunspent", []interface{}{minConf, maxConf, addresses})
	if err != nil {
		return
	}

	json.Unmarshal(r.Result, &res)

	for _, utxo := range res {
		if utxo.Amount > 0 && utxo.Satoshis == 0 {
			utxo.Satoshis = uint64(utxo.Amount * 100000000)
		}
	}

	return
}

// SendToAddress comment
func (b *Bitcoind) SendToAddress(address string, amount float64) (string, error) {
	r, err := b.call("sendtoaddress", []interface{}{address, amount})
	if err != nil {
		return "", err
	}

	var txid string
	json.Unmarshal(r.Result, &txid)

	return txid, nil
}

// Generate for regtest
func (b *Bitcoind) Generate(amount float64) ([]string, error) {
	r, err := b.call("generate", []interface{}{amount})
	if err != nil {
		return nil, err
	}

	var hashes []string
	json.Unmarshal(r.Result, &hashes)

	return hashes, nil
}

// GenerateToAddress for regtest
func (b *Bitcoind) GenerateToAddress(amount float64, address string) ([]string, error) {
	r, err := b.call("generatetoaddress", []interface{}{amount, address})
	if err != nil {
		return nil, err
	}

	var hashes []string
	json.Unmarshal(r.Result, &hashes)

	return hashes, nil
}
