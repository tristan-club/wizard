package pconst

import (
	"github.com/tristan-club/kit/chain_info"
)

var DebugChainTypeList = chain_info.GetSupportChainTypeList()
var ChainTypeList = chain_info.GetSupportChainTypeList()

func GetChainName(chainType uint32) string {
	return chain_info.GetChainInfo(chainType).Name
}

func GetExplore(chainType uint32, txHash string, exploreType chain_info.ExplorerTargetType) string {
	return chain_info.GetExplorerTargetUrl(chain_info.GetNetByChainType(chainType).ChainId, txHash, exploreType)
}

func GetChainId(chainType uint32) uint64 {
	return chain_info.GetNetByChainType(chainType).ChainId
}
