package pconst

import (
	"fmt"
	"github.com/tristan-club/bot-wizard/config"
	"github.com/tristan-club/kit/chain_info"
)

const (
	ChainTypeBsc     = 1
	ChainTypeMetis   = 2
	ChainTypePolygon = 3
	ChainTypeKlaytn  = 4
)

var DebugChainTypeList = chain_info.GetSupportChainTypeList()
var ChainTypeList = chain_info.GetSupportChainTypeList()

func GetChainName(chainType uint32) string {
	return chain_info.GetChainInfo(chainType).Name
}

const (
	ExploreTypeTx = iota + 1
	ExploreTypeAddress
)

func GetExplore(chainType uint32, exploreType int32) string {
	netType := chain_info.NetworkTypeMainNet
	if config.IsTestNet() {
		netType = chain_info.NetworkTypeTestNet
	}
	var queryType string
	if exploreType == ExploreTypeTx {
		queryType = "tx/"
	} else {
		queryType = "address/"
	}

	for _, net := range chain_info.GetSupportNetList() {
		if net.ChainType == chainType && net.Type == uint8(netType) {
			return net.BlockExplorer + queryType
		}
	}

	return fmt.Sprintf("invalid explore for target chain[type:%d-%d]", chainType, exploreType)
}

func GetChainId(chainType uint32) uint64 {
	netType := chain_info.NetworkTypeMainNet
	if config.IsTestNet() {
		netType = chain_info.NetworkTypeTestNet
	}
	for _, net := range chain_info.GetSupportNetList() {
		if net.ChainType == chainType && net.Type == uint8(netType) {
			return net.ChainId
		}
	}
	return 0
}
