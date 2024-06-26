//go:build regression
// +build regression

package regressiontests

type merkleRootFixtures struct {
	MerkleRoot string `json:"merkleRoot"`
	Height     int32  `json:"blockHeight"`
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
