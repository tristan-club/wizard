package pconst

import (
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	"github.com/tristan-club/wizard/config"
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

func GetExplore(chainType uint32, txHash string, exploreType chain_info.ExplorerTargetType) string {

	return chain_info.GetExplorerTargetUrl(chain_info.GetNetByChainType(chainType).ChainId, txHash, exploreType)

	netType := chain_info.NetworkTypeMainNet
	if config.IsTestNet() {
		netType = chain_info.NetworkTypeTestNet
	}
	var queryType string
	if exploreType == chain_info.ExplorerTargetTransaction {
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
	return chain_info.GetNetByChainType(chainType).ChainId
}
