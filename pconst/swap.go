package pconst

import (
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/wizard/config"
)

const (
	SwapAssetBnb   = "BNB"
	SwapAssetMetis = "METIS"
	SwapAssetBusd  = "BUSD"

	TOKEN_BSC_BUSD  = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	TOKEN_BSC_WBNB  = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c"
	TOKEN_BSC_METIS = "0xe552Fb52a4F19e44ef5A967632DBc320B0820639"

	PANCAKE_BSC = "0x10ed43c718714eb63d5aa57b78b54704e256024e"
)

type SwapAsset struct {
	AssetSymbol   string `json:"asset_symbol"`
	AssetAddress  string `json:"asset_address"`
	SwapAmountMin string `json:"swap_amount_min"`
	SwapAmountMax string `json:"swap_amount_max"`
	Decimals      int32  `json:"decimals"`
}

var swapAssetAddress = map[string][]string{
	SwapAssetBnb:   {AssetAddressCoin, AssetAddressCoin},
	SwapAssetMetis: {"0xe552Fb52a4F19e44ef5A967632DBc320B0820639", "0xdB2D16f61e3e1bD54Ba8068E04F5552e718d5b7c"},
	SwapAssetBusd:  {TOKEN_BSC_BUSD, "0xe552Fb52a4F19e44ef5A967632DBc320B0820639"},
}

func GetSwapAssetAddress(assetSymbol string) string {
	index := 0
	if config.IsTestNet() {
		index = 1
	}
	if asset, ok := swapAssetAddress[assetSymbol]; ok && len(asset) == 2 {
		return asset[index]
	}

	return ""
}

var swapAsset = map[uint32][]SwapAsset{
	chain_info.ChainTypeBsc: []SwapAsset{{
		AssetAddress:  AssetAddressCoin,
		AssetSymbol:   SwapAssetBnb,
		Decimals:      18,
		SwapAmountMin: "0.001",
		SwapAmountMax: "0.5",
		//}, {
		//	AssetAddress:  GetSwapAssetAddress(SwapAssetMetis),
		//	AssetSymbol:   SwapAssetMetis,
		//	Decimals:      18,
		//	SwapAmountMin: "0.01",
		//	SwapAmountMax: "2",
	}, {
		AssetAddress:  GetSwapAssetAddress(SwapAssetBusd),
		AssetSymbol:   SwapAssetBusd,
		Decimals:      18,
		SwapAmountMin: "1",
		SwapAmountMax: "200",
	},
	},
}

func GetSwapAssetList(chainId uint32) []SwapAsset {
	if resp, ok := swapAsset[chainId]; ok {
		return resp
	}

	return []SwapAsset{}
}

func GetSwapAsset(chainId uint32, assetSymbol string) SwapAsset {
	if swapAssetList := GetSwapAssetList(chainId); len(swapAssetList) != 0 {
		for _, v := range swapAssetList {
			if v.AssetSymbol == assetSymbol {
				return v
			}
		}
	}
	return SwapAsset{}
}

var bridgeAsset = map[uint32][]SwapAsset{
	chain_info.ChainTypeMetis: []SwapAsset{{
		AssetAddress:  GetSwapAssetAddress(SwapAssetMetis),
		AssetSymbol:   SwapAssetMetis,
		Decimals:      18,
		SwapAmountMin: "0.01",
		SwapAmountMax: "2",
	},
	},
}

func GetBridgeAssetList(chainId uint32) []SwapAsset {
	if resp, ok := bridgeAsset[chainId]; ok {
		return resp
	}

	return []SwapAsset{}
}

func GetBridgeAsset(chainId uint32, assetSymbol string) SwapAsset {
	if swapAssetList := GetBridgeAssetList(chainId); len(swapAssetList) != 0 {
		for _, v := range swapAssetList {
			if v.AssetSymbol == assetSymbol {
				return v
			}
		}
	}
	return SwapAsset{}
}
