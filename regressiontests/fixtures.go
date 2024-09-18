//go:build regression
// +build regression

package regressiontests

import "net/http"

type merkleRootFixtures struct {
	MerkleRoot string `json:"merkleRoot"`
	Height     int32  `json:"blockHeight"`
}

type merkleRootBatchSample struct {
	queryParams  string
	expectedBody string
	expectedCode int
}

var fixtures = []merkleRootFixtures{
	{"2d05f0c9c3e1c226e63b5fac240137687544cf631cd616fd34fd188fc9020866", 100},
	{"5032c865ec9dffc052c0ae492d0b42c05d5904e7d540db1f8ca18118a2e561b3", 100100},
	{"f0caff9e93bc0ffe943538e242acdcbcdfb1759f59b0ef06da2ca834187cbb18", 200100},
	{"3e017719a1920bb4174fce5e165383ea0d389e50cd07afd7dd2f96762dfc8632", 300100},
	{"9d74e0ada54c8b1ac145803ed8d27b620888dbd6af9c0dd6ef1e9ff90c60c4cc", 400100},
	{"c1cfca75eefe6b5421354e13b2ac6f873e164aa2c8a82c7bba9eca8deca0047a", 500100},
	{"3286caac01f7e30b1a61285cf4adcf0a3f3f4afa2b548c22ea433f6abd128e02", 600100},
	// Height before BSV hard fork
	{"0931f7d640a770f1cc25af171baa2844a385473a84f3ee327480072bc802fdd3", 620537},
	// BSV hard fork
	{"880ae112f20e84077945f74156ec3f557681e818a80ba7070292b21c6e81ab54", 620538},
	{"33c22bcde3e974dd1409cf44d11765730376e526421192b48afc642bddd3715a", 700100},
	{"5577e5499957946166154297b6811f80a2c1c342d9743d05a63fe8ea9defed4c", 800100},
}

// czy bez evaluation zwroci 5 pierwszych headerow
// czy podajac 10 merkleroot ze zwroci kolejne 11-15 cos takiego
var merkleRootsTestSamples = []merkleRootBatchSample{
	{
		queryParams:  "?batchSize=5",
		expectedCode: http.StatusOK,
		expectedBody: `{
		                     "content": [
		                        {
		                          "merkleRoot": "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
		                          "blockHeight": 0
		                        },
		                        {
		                          "merkleRoot": "0e3e2357e806b6cdb1f70b54c3a3a17b6714ee1f0e68bebb44a74b1efd512098",
		                          "blockHeight": 1
		                        },
		                        {
		                          "merkleRoot": "9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5",
		                          "blockHeight": 2
		                        },
		                        {
		                          "merkleRoot": "999e1c837c76a1b7fbb7e57baf87b309960f5ffefbf2a9b95dd890602272f644",
		                          "blockHeight": 3
		                        },
		                        {
		                          "merkleRoot": "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a",
		                          "blockHeight": 4
		                        }
		                      ],
		                      "page": {
		                        "size": 5,
		                        "lastEvaluatedKey": "df2b060fa2e5e9c8ed5eaf6a45c13753ec8c63282b2688322eba40cd98ea067a"
		                     }
		                  }`,
	}, {
		queryParams:  "?batchSize=5&lastEvaluatedKey=ae9d8d06f1859e27f22a4454fbb82aca954ac9f4b5d553ff5760a3393a2818b1",
		expectedCode: http.StatusOK,
		expectedBody: `{
		                     "content": [
		                        {
		                          "merkleRoot": "8e2762752ef06265fd8266777338a14d2a89e3e45e83bbcb3d31877d5f2bcfdf",
		                          "blockHeight": 849838
		                        },
		                        {
		                          "merkleRoot": "abe45c884b116744a6e73da09fc58c02aa18b99073bd465d4880d16fdcbcfd48",
		                          "blockHeight": 849839 
		                        },
		                        {
		                          "merkleRoot": "1e06d817ce765e6186e25632506fb03d631890d115d5cd00ca96a91896a1c5ad",
		                          "blockHeight": 849840 
		                        },
		                        {
		                          "merkleRoot": "5c1d84ed999dd5df94bca96c6a5ed5e1450be6dcd6effa5359f3ebf36795df01",
		                          "blockHeight": 849841 
		                        },
		                        {
		                          "merkleRoot": "dcbb3d71d870805aa0657452efc8a12410f0281069c3f75974e7d9a2221e5c16",
		                          "blockHeight": 849842
		                        }
		                      ],
		                      "page": {
		                        "size": 5,
		                        "lastEvaluatedKey": "dcbb3d71d870805aa0657452efc8a12410f0281069c3f75974e7d9a2221e5c16"
		                     }
		                  }`,
	},
}
