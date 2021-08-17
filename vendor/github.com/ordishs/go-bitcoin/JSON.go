package bitcoin

type gbtParams struct {
	Mode         string   `json:"mode,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Rules        []string `json:"rules,omitempty"`
}

// BlockchainInfo comment
type BlockchainInfo struct {
	Chain                string  `json:"chain"`
	Blocks               int32   `json:"blocks"`
	Headers              int32   `json:"headers"`
	BestBlockHash        string  `json:"bestblockhash"`
	Difficulty           float64 `json:"difficulty"`
	MedianTime           int64   `json:"mediantime"`
	VerificationProgress float64 `json:"verificationprogress,omitempty"`
	Pruned               bool    `json:"pruned"`
	PruneHeight          int32   `json:"pruneheight,omitempty"`
	ChainWork            string  `json:"chainwork,omitempty"`
}

// GetInfo comment
type GetInfo struct {
	Version                      int32   `json:"version"`
	ProtocolVersion              int32   `json:"protocolversion"`
	WalletVersion                int32   `json:"walletversion"`
	Balance                      float64 `json:"balance"`
	Blocks                       int32   `json:"blocks"`
	TimeOffset                   int64   `json:"timeoffset"`
	Connections                  int32   `json:"connections"`
	Proxy                        string  `json:"proxy"`
	Difficulty                   float64 `json:"difficulty"`
	TestNet                      bool    `json:"testnet"`
	STN                          bool    `json:"stn"`
	KeyPoolOldest                int64   `json:"keypoololdest"`
	KeyPoolSize                  int32   `json:"keypoolsize"`
	PayTXFee                     float64 `json:"paytxfee"`
	RelayFee                     float64 `json:"relayfee"`
	Errors                       string  `json:"errors"`
	MaxBlockSize                 int64   `json:"maxblocksize"`
	MaxMinedBlockSize            int64   `json:"maxminedblocksize"`
	MaxStackMemoryUsagePolicy    uint64  `json:"maxstackmemoryusagepolicy"`
	MaxStackMemoryUsageConsensus uint64  `json:"maxstackmemoryusageconsensus"`
}

// Network comment
type Network struct {
	Name                       string `json:"name"`
	Limited                    bool   `json:"limited"`
	Reachable                  bool   `json:"reachable"`
	Proxy                      string `json:"proxy"`
	ProxyRandmomizeCredentials bool   `json:"proxy_randomize_credentials"`
}

// LocalAddress comment
type LocalAddress struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Score   int    `json:"score"`
}

// NetworkInfo comment
type NetworkInfo struct {
	Version           int            `json:"version"`
	SubVersion        string         `json:"subversion"`
	ProtocolVersion   int            `json:"protocolversion"`
	LocalServices     string         `json:"localservices"`
	LocalRelay        bool           `json:"localrelay"`
	TimeOffset        int            `json:"timeoffset"`
	TXPropagationFreq int            `json:"txnpropagationfreq"`
	TXPropagationLen  int            `json:"txnpropagationqlen"`
	NetworkActive     bool           `json:"networkactive"`
	Connections       int            `json:"connections"`
	AddressCount      int            `json:"addresscount"`
	Networks          []Network      `json:"networks"`
	RelayFee          float64        `json:"relayfee"`
	ExcessUTXOCharge  float64        `json:"excessutxocharge"`
	LocalAddresses    []LocalAddress `json:"localaddresses"`
	Warnings          string         `json:"warnings"`
}

// NetTotals comment
type NetTotals struct {
	TotalBytesRecv int `json:"totalbytesrecv"`
	TotalBytesSent int `json:"totalbytessent"`
	TimeMillis     int `json:"timemillis"`
	UploadTarget   struct {
		TimeFrame             int  `json:"timeframe"`
		Target                int  `json:"target"`
		TargetReached         bool `json:"target_reached"`
		ServeHistoricalBlocks bool `json:"serve_historical_blocks"`
		BytesLeftInCycle      int  `json:"bytes_left_in_cycle"`
		TimeLeftInCycle       int  `json:"time_left_in_cycle"`
	} `json:"uploadtarget"`
}

// MiningInfo comment
type MiningInfo struct {
	Blocks                int     `json:"blocks"`
	CurrentBlockSize      int     `json:"currentblocksize"`
	CurrentBlockTX        int     `json:"currentblocktx"`
	Difficulty            float64 `json:"difficulty"`
	BlocksPriorityPercent int     `json:"blockprioritypercentage"`
	Errors                string  `json:"errors"`
	NetworkHashPS         float64 `json:"networkhashps"`
	PooledTX              int     `json:"pooledtx"`
	Chain                 string  `json:"chain"`
}

// BytesData struct
type BytesData struct {
	Addr        int `json:"addr"`
	BlockTXN    int `json:"blocktxn"`
	CmpctBlock  int `json:"cmpctblock"`
	FeeFilter   int `json:"feefilter"`
	GetAddr     int `json:"getaddr"`
	GetData     int `json:"getdata"`
	GetHeaders  int `json:"getheaders"`
	Headers     int `json:"headers"`
	Inv         int `json:"inv"`
	NotFound    int `json:"notfound"`
	Ping        int `json:"ping"`
	Pong        int `json:"pong"`
	Reject      int `json:"reject"`
	SendCmpct   int `json:"sendcmpct"`
	SendHeaders int `json:"sendheaders"`
	TX          int `json:"tx"`
	VerAck      int `json:"verack"`
	Version     int `json:"version"`
}

// Peer struct
type Peer struct {
	ID             int     `json:"id"`
	Addr           string  `json:"addr"`
	AddrLocal      string  `json:"addrlocal"`
	Services       string  `json:"services"`
	RelayTXes      bool    `json:"relaytxes"`
	LastSend       int     `json:"lastsend"`
	LastRecv       int     `json:"lastrecv"`
	BytesSent      int     `json:"bytessent"`
	BytesRecv      int     `json:"bytesrecv"`
	ConnTime       int     `json:"conntime"`
	TimeOffset     int     `json:"timeoffset"`
	PingTime       float64 `json:"pingtime"`
	MinPing        float64 `json:"minping"`
	Version        int     `json:"version"`
	Subver         string  `json:"subver"`
	Inbound        bool    `json:"inbound"`
	AddNode        bool    `json:"addnode"`
	StartingHeight int     `json:"startingheight"`
	TXNInvSize     int     `json:"txninvsize"`
	Banscore       int     `json:"banscore"`
	SyncedHeaders  int     `json:"synced_headers"`
	SyncedBlocks   int     `json:"synced_blocks"`
	// "inflight": [],
	WhiteListed     bool      `json:"whitelisted"`
	BytesSendPerMsg BytesData `json:"bytessent_per_msg"`
	BytesRecvPerMsg BytesData `json:"bytesrecv_per_msg"`
}

// PeerInfo comment
type PeerInfo []Peer

// RawMemPool comment
type RawMemPool []string

// MempoolInfo comment
type MempoolInfo struct {
	Size           int     `json:"size"`
	Bytes          int     `json:"bytes"`
	Usage          int     `json:"usage"`
	MaxMemPool     int     `json:"maxmempool"`
	MemPoolMinFree float64 `json:"mempoolminfee"`
}

// ChainTXStats struct
type ChainTXStats struct {
	Time             int     `json:"time"`
	TXCount          int     `json:"txcount"`
	WindowBlockCount int     `json:"window_block_count"`
	WindowTXCount    int     `json:"window_tx_count"`
	WindowInterval   int     `json:"window_interval"`
	TXRate           float64 `json:"txrate"`
}

// Address comment
type Address struct {
	IsValid      bool   `json:"isvalid"`
	Address      string `json:"address"`
	ScriptPubKey string `json:"scriptPubKey"`
	IsMine       bool   `json:"ismine"`
	IsWatchOnly  bool   `json:"iswatchonly"`
	IsScript     bool   `json:"isscript"`
}

// Transaction comment
type Transaction struct {
	TXID string `json:"txid"`
	Hash string `json:"hash"`
	Data string `json:"data"`
}

// BlockTemplate comment
type BlockTemplate struct {
	Version                  uint32        `json:"version"`
	PreviousBlockHash        string        `json:"previousblockhash"`
	Target                   string        `json:"target"`
	Transactions             []Transaction `json:"transactions"`
	Bits                     string        `json:"bits"`
	CurTime                  uint64        `json:"curtime"`
	CoinbaseValue            uint64        `json:"coinbasevalue"`
	Height                   uint32        `json:"height"`
	MinTime                  uint64        `json:"mintime"`
	NonceRange               string        `json:"noncerange"`
	DefaultWitnessCommitment string        `json:"default_witness_commitment"`
	SizeLimit                uint64        `json:"sizelimit"`
	WeightLimit              uint64        `json:"weightlimit"`
	SigOpLimit               int64         `json:"sigoplimit"`
	VBRequired               int64         `json:"vbrequired"`
	// extra mining candidate fields
	IsMiningCandidate bool             `json:"-"`
	MiningCandidateID string           `json:"-"`
	MiningCandidate   *MiningCandidate `json:"-"`
	MerkleBranches    []string         `json:"-"`
}

// MiningCandidate comment
type MiningCandidate struct {
	ID                  string   `json:"id"`
	PreviousHash        string   `json:"prevhash"`
	CoinbaseValue       uint64   `json:"coinbaseValue"`
	Version             uint32   `json:"version"`
	Bits                string   `json:"nBits"`
	CurTime             uint64   `json:"time"`
	Height              uint32   `json:"height"`
	NumTx               uint32   `json:"num_tx"`
	SizeWithoutCoinbase uint64   `json:"sizeWithoutCoinbase"`
	MerkleProof         []string `json:"merkleProof"`
}

type submitMiningSolutionParams struct {
	ID       string `json:"id"`
	Nonce    uint32 `json:"nonce"`
	Coinbase string `json:"coinbase"`
	Time     uint32 `json:"time"`
	Version  uint32 `json:"version"`
}

// Block struct
type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	Size              uint64   `json:"size"`
	Height            uint64   `json:"height"`
	Version           int64    `json:"version"`
	VersionHex        string   `json:"versionHex"`
	MerkleRoot        string   `json:"merkleroot"`
	TxCount           uint64   `json:"txcount"`
	NTx               uint64   `json:"nTx"`
	NumTx             uint64   `json:"num_tx"`
	Tx                []string `json:"tx"`
	Time              uint64   `json:"time"`
	MedianTime        uint64   `json:"mediantime"`
	Nonce             uint64   `json:"nonce"`
	Bits              string   `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	Chainwork         string   `json:"chainwork"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     string   `json:"nextblockhash"`
	// extra properties
	CoinbaseTx *RawTransaction `json:"coinbaseTx"`
	TotalFees  float64         `json:"totalFees"`
	Miner      string          `json:"miner"`
	Pagination *BlockPage      `json:"pages"`
}

// Block2 struct
type Block2 struct {
	Hash              string   `json:"hash"`
	Size              int      `json:"size"`
	Height            int      `json:"height"`
	Version           uint32   `json:"version"`
	VersionHex        string   `json:"versionHex"`
	MerkleRoot        string   `json:"merkleroot"`
	TxCount           uint64   `json:"txcount"`
	NTx               uint64   `json:"nTx"`
	NumTx             uint64   `json:"num_tx"`
	Tx                []string `json:"tx"`
	Time              uint32   `json:"time"`
	MedianTime        uint32   `json:"mediantime"`
	Nonce             uint32   `json:"nonce"`
	Bits              string   `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	Chainwork         string   `json:"chainwork"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     string   `json:"nextblockhash"`
	BlockSubsidy      uint64   `json:"blockSubsidy"`
	BlockReward       uint64   `json:"blockReward"`
	USDPrice          float64  `json:"usdPrice"`
	Miner             string   `json:"miner"`
}

// BlockOverview struct
type BlockOverview struct {
	Hash          string `json:"hash"`
	Confirmations int64  `json:"confirmations"`
	Size          uint64 `json:"size"`
	Height        uint64 `json:"height"`
	Version       int64  `json:"version"`
	VersionHex    string `json:"versionHex"`
	MerkleRoot    string `json:"merkleroot"`
	// TxCount           uint64  `json:"txcount"`
	Time              uint64  `json:"time"`
	MedianTime        uint64  `json:"mediantime"`
	Nonce             uint64  `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	PreviousBlockHash string  `json:"previousblockhash"`
	NextBlockHash     string  `json:"nextblockhash"`
}

// BlockHeader comment
type BlockHeader struct {
	Hash              string  `json:"hash"`
	Confirmations     int64   `json:"confirmations"`
	Height            uint64  `json:"height"`
	Version           uint64  `json:"version"`
	VersionHex        string  `json:"versionHex"`
	MerkleRoot        string  `json:"merkleroot"`
	Time              uint64  `json:"time"`
	MedianTime        uint64  `json:"mediantime"`
	Nonce             uint64  `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	PreviousBlockHash string  `json:"previousblockhash"`
	NextBlockHash     string  `json:"nextblockhash"`
	TXCount           uint32  `json:"num_tx"`
}

// BlockHeaderAndCoinbase comment
type BlockHeaderAndCoinbase struct {
	Hash              string           `json:"hash"`
	Confirmations     int64            `json:"confirmations"`
	Height            uint64           `json:"height"`
	Version           uint64           `json:"version"`
	VersionHex        string           `json:"versionHex"`
	MerkleRoot        string           `json:"merkleroot"`
	Time              uint64           `json:"time"`
	MedianTime        uint64           `json:"mediantime"`
	Nonce             uint64           `json:"nonce"`
	Bits              string           `json:"bits"`
	Difficulty        float64          `json:"difficulty"`
	Chainwork         string           `json:"chainwork"`
	PreviousBlockHash string           `json:"previousblockhash"`
	NextBlockHash     string           `json:"nextblockhash"`
	Tx                []RawTransaction `json:"tx"`
}

// BlockPage to store links
type BlockPage struct {
	URI  []string `json:"uri"`
	Size uint64   `json:"size"`
}

// BlockTxid comment
type BlockTxid struct {
	BlockHash  string   `json:"blockhash"`
	Tx         []string `json:"tx"`
	StartIndex uint64   `json:"startIndex"`
	EndIndex   uint64   `json:"endIndex"`
	Count      uint64   `json:"count"`
}

// RawTransaction comment
type RawTransaction struct {
	Hex           string `json:"hex"`
	TxID          string `json:"txid"`
	Hash          string `json:"hash"`
	Version       int32  `json:"version"`
	Size          uint32 `json:"size"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint32 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
	BlockHeight   uint64 `json:"blockheight,omitempty"`
}

// Vout represent an OUT value
type Vout struct {
	Value        float64      `json:"value"`
	N            int          `json:"n"`
	ScriptPubKey ScriptPubKey `json:"scriptPubKey"`
}

// Vin represent an IN value
type Vin struct {
	Coinbase  string    `json:"coinbase"`
	Txid      string    `json:"txid"`
	Vout      uint64    `json:"vout"`
	ScriptSig ScriptSig `json:"scriptSig"`
	Sequence  uint32    `json:"sequence"`
}

// OpReturn comment
type OpReturn struct {
	Type   string   `json:"type"`
	Action string   `json:"action"`
	Text   string   `json:"text"`
	Parts  []string `json:"parts"`
}

// ScriptPubKey Comment
type ScriptPubKey struct {
	ASM         string    `json:"asm"`
	Hex         string    `json:"hex"`
	ReqSigs     int64     `json:"reqSigs,omitempty"`
	Type        string    `json:"type"`
	Addresses   []string  `json:"addresses,omitempty"`
	OpReturn    *OpReturn `json:"opReturn"`
	IsTruncated bool      `json:"isTruncated"`
}

// A ScriptSig represents a scriptsyg
type ScriptSig struct {
	ASM string `json:"asm"`
	Hex string `json:"hex"`
}

// UnspentTransaction type
type UnspentTransaction struct {
	TXID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Address       string  `json:"address"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Satoshis      uint64  `json:"satoshis"`
	Confirmations uint32  `json:"confirmations"`
}

// SignRawTransactionResponse struct
type SignRawTransactionResponse struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

// Error comment
type Error struct {
	Code    float64
	Message string
}
