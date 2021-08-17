# go-bitcoin
Go wrapper for bitcoin RPC

## RPC services
Start by creating a connection to a bitcoin node
```
  b, err := New("rcp host", rpc port, "rpc username", "rpc password", false)
  if err != nil {
    log.Fatal(err)
  }
```

Then make a call to bitcoin
```
  res, err := b.GetBlockchainInfo()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("%#v\n", res)
```

Available calls are:
```
GetConnectionCount()
GetBlockchainInfo()
GetNetworkInfo()
GetNetTotals()
GetMiningInfo()
Uptime()
GetMempoolInfo()
GetRawMempool(details bool)
GetChainTxStats(blockcount int)
ValidateAddress(address string)
GetHelp()
GetBestBlockHash()
GetBlockHash(blockHeight int)
SendRawTransaction(hex string)
GetBlock(blockHash string)
GetBlockOverview(blockHash string)
GetBlockHex(blockHash string)
GetRawTransaction(txID string)
GetRawTransactionHex(txID string)
GetBlockTemplate(includeSegwit bool)
GetMiningCandidate()
SubmitBlock(hexData string)
SubmitMiningSolution(candidateID string, nonce uint32,
                     coinbase string, time uint32, version uint32)
GetDifficulty()
DecodeRawTransaction(txHex string)
GetTxOut(txHex string, vout int, includeMempool bool)
ListUnspent(addresses []string)
```

## ZMQ
It is also possible to subscribe to a bitcoin node and be notified about new transactions and new blocks via the node's ZMQ interface.

First, create a ZMQ instance:
```
  zmq := bitcoin.NewZMQ("localhost", 28332)
```

Then create a buffered or unbuffered channel of strings and a goroutine to consume the channel:
```
	ch := make(chan string)

	go func() {
		for c := range ch {
			log.Println(c)
		}
	}()
```

Finally, subscribe to "hashblock" or "hashtx" topics passing in your channel:
```
	err := zmq.Subscribe("hashblock", ch)
	if err != nil {
		log.Fatalln(err)
	}
```
