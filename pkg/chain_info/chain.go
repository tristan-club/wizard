package chain_info

import (
	"fmt"
	"os"
)

const (
	ChainTypeBsc      = 1
	ChainTypeMetis    = 2
	ChainTypePolygon  = 3
	ChainTypeKlaytn   = 4
	ChainTypeOkc      = 5
	ChainTypeEthereum = 10
)

var supportChainTypeList = []uint32{ChainTypeBsc, ChainTypeMetis, ChainTypePolygon, ChainTypeOkc}

func GetSupportChainTypeList() []uint32 {
	return supportChainTypeList
}

const (
	NetworkTypeMainNet = iota + 1
	NetworkTypeTestNet
)

type Net struct {
	ChainType     uint32 `json:"chain_type"`
	ChainId       uint64 `json:"chain_id"`
	NetworkName   string `json:"network_name"`
	Symbol        string `json:"symbol"`
	Decimals      uint8  `json:"decimals"`
	RpcUrl        string `json:"rpc_url"`
	BlockExplorer string `json:"block_explorer"`
	Type          uint8  `json:"type"`
}

func (c *Net) IsAvailable() bool {
	return c.ChainId == 0
}

type Chain struct {
	ChainType    uint32 `json:"chain_type"`
	Name         string `json:"name"`
	Remark       string `json:"remark"`
	Icon         string `json:"icon"`
	Symbol       string `json:"coin_symbol"`
	CoinDecimals uint8  `json:"coin_decimals"`
	Type         uint8  `json:"type"`
}

var supportChainList = []*Chain{
	{
		ChainType:    ChainTypeBsc,
		Symbol:       "BNB",
		CoinDecimals: 18,
		Type:         NetworkTypeMainNet,
		Name:         "BNB Chain",
	},
	{
		ChainType:    ChainTypeMetis,
		Symbol:       "Metis",
		CoinDecimals: 18,
		Type:         NetworkTypeMainNet,
		Name:         "Metis",
	},
	{
		ChainType:    ChainTypePolygon,
		Symbol:       "Matic",
		CoinDecimals: 18,
		Type:         NetworkTypeMainNet,
		Name:         "Polygon",
	},
	{
		ChainType:    ChainTypeKlaytn,
		Symbol:       "KLAY",
		CoinDecimals: 18,
		Type:         NetworkTypeMainNet,
		Name:         "Klaytn",
	},
	{
		ChainType:    ChainTypeOkc,
		Symbol:       "OKC",
		CoinDecimals: 18,
		Type:         NetworkTypeMainNet,
		Name:         "OKC",
	},
}

func GetSupportChainList() []*Chain {
	return supportChainList
}

func GetChainInfo(chainType uint32) *Chain {
	for _, v := range supportChainList {
		if v.ChainType == chainType {
			return v
		}
	}

	return &Chain{
		ChainType:    0,
		Symbol:       "-1",
		CoinDecimals: 0,
		Type:         0,
	}
}

var supportChainNetList = []*Net{}

func init() {

	supportChainNetList = []*Net{
		{
			ChainType:   ChainTypeBsc,
			ChainId:     56,
			Symbol:      "BNB",
			Decimals:    18,
			Type:        NetworkTypeMainNet,
			NetworkName: "BSC Mainnet",
			RpcUrl:      "https://bsc-dataseed1.binance.org/",
			//RpcUrl:           fmt.Sprintf("https://bsc.getblock.io/mainnet/?api_key=%s", blockIOProvider),
			BlockExplorer: "",
		},
		{
			ChainType:   ChainTypeBsc,
			ChainId:     97,
			Symbol:      "BNB",
			Decimals:    18,
			Type:        NetworkTypeTestNet,
			NetworkName: "BSC Testnet",
			RpcUrl:      "https://data-seed-prebsc-1-s1.binance.org:8545/",
		},
		{
			ChainType:   ChainTypeMetis,
			ChainId:     1088,
			Symbol:      "Metis",
			Decimals:    18,
			Type:        NetworkTypeMainNet,
			NetworkName: "Metis Mainnet",
			RpcUrl:      "https://andromeda.metis.io/?owner=1088",
			//WssUrl:      "wss://andromeda-ws.metis.io",
			//BlockExplorer: "https://rinkeby.etherscan.io",

		},
		{
			ChainType:   ChainTypeMetis,
			ChainId:     588,
			Symbol:      "Metis",
			Decimals:    18,
			Type:        NetworkTypeTestNet,
			NetworkName: "Metis TestNet",
			RpcUrl:      "https://stardust.metis.io/?owner=588",
			//WssUrl:      "wss://stardust-ws.metis.io/",

		},
		{
			ChainType:   ChainTypePolygon,
			ChainId:     137,
			Symbol:      "Matic",
			Decimals:    18,
			Type:        NetworkTypeMainNet,
			NetworkName: "Polygon Mainnet",
			RpcUrl:      "https://polygon-rpc.com/",
			//RpcUrl:           fmt.Sprintf("https://matic.getblock.io/mainnet/?api_key=%s", blockIOProvider),

			//WssUrl:           "wss://rpc-mainnet.matic.network/",
		},
		{
			ChainType:   ChainTypePolygon,
			ChainId:     80001,
			Symbol:      "Matic",
			Decimals:    18,
			Type:        NetworkTypeTestNet,
			NetworkName: "Polygon TestNet",
			RpcUrl:      "https://matic-mumbai.chainstacklabs.com",
			//WssUrl:           "wss://rpc-mumbai.matic.today",
		},
		{
			ChainType:     ChainTypeKlaytn,
			ChainId:       8217,
			Symbol:        "KLAY",
			Decimals:      18,
			Type:          NetworkTypeMainNet,
			NetworkName:   "Klaytn Cypress",
			RpcUrl:        "https://public-node-api.klaytnapi.com/v1/cypress",
			BlockExplorer: "https://scope.klaytn.com/",
		},
		{
			ChainType:     ChainTypeKlaytn,
			ChainId:       1001,
			Symbol:        "KLAY",
			Decimals:      18,
			Type:          NetworkTypeTestNet,
			NetworkName:   "Klaytn Baobab",
			RpcUrl:        "https://api.baobab.klaytn.net:8651/",
			BlockExplorer: "https://baobab.scope.klaytn.com/",
		},
		{
			ChainType:     ChainTypeOkc,
			ChainId:       66,
			Symbol:        "OKT",
			Decimals:      18,
			Type:          NetworkTypeMainNet,
			NetworkName:   "OKC Mainnet",
			RpcUrl:        "https://exchainrpc.okex.org",
			BlockExplorer: "https://www.oklink.com/okc/",
		},
		{
			ChainType:     ChainTypeOkc,
			ChainId:       65,
			Symbol:        "OKT",
			Decimals:      18,
			Type:          NetworkTypeTestNet,
			NetworkName:   "OKC Testnet",
			RpcUrl:        "https://exchaintestrpc.okex.org",
			BlockExplorer: "https://www.oklink.com/okc-test/",
		},
	}

	blockIOProvider := os.Getenv("BLOCK_IO_PROVIDER")
	if blockIOProvider != "" {
		for k, _ := range supportChainNetList {
			if supportChainNetList[k].ChainId == 56 {
				supportChainNetList[k].RpcUrl = fmt.Sprintf("https://bsc.getblock.io/mainnet/?api_key=%s", blockIOProvider)
			}
			if supportChainNetList[k].ChainId == 137 {
				supportChainNetList[k].RpcUrl = fmt.Sprintf("https://matic.getblock.io/mainnet/?api_key=%s", blockIOProvider)
			}
		}
	}

}

func GetSupportNetList() []*Net {
	return supportChainNetList
}

func GetNetByChainId(chainId uint64) *Net {
	for _, v := range supportChainNetList {
		if v.ChainId == chainId {
			return v
		}
	}
	return nil
}
