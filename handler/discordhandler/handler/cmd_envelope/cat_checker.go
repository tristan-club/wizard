package cmd_envelope

import (
	"context"
	"fmt"
	"github.com/tristan-club/kit/chain_info"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/handler/discordhandler/dcontext"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
)

func catChecker(ctx *dcontext.Context) error {

	assetListResp, err := ctx.CM.AssetList(context.Background(), &controller_pb.AssetListReq{
		AppId:      ctx.Requester.RequesterAppId,
		PresetType: controller_pb.AssetPresetType_PRESET_TYPE_CAT,
		TokenType:  pconst.TokenTypeERC721,
	})
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request controller svc error", "error": err.Error(), "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, "", err)
	} else if assetListResp.CommonResponse.Code != he.Success {
		log.Error().Fields(map[string]interface{}{"action": "get cat error", "error": assetListResp, "ctx": ctx}).Send()
		return he.NewServerError(he.ServerError, assetListResp.CommonResponse.Message, fmt.Errorf(assetListResp.CommonResponse.Inner))
	} else if assetListResp.Data.Count == 0 {
		log.Info().Fields(map[string]interface{}{"action": "current app not configure cat", "ctx": ctx}).Send()
		return he.NewBusinessError(pconst.CodeCATNotConfigure, "", nil)
	}

	var chainIdList []uint32
	for _, v := range assetListResp.Data.List {
		chainIdList = append(chainIdList, chain_info.GetNetByChainId(v.ChainId).ChainType)
	}

	userstate.SetParam(ctx.GetFromId(), "envelope_option", uint32(controller_pb.ENVELOPE_OPTION_HAS_CAT))
	userstate.SetParam(ctx.GetFromId(), "chain_type_list", chainIdList)

	return nil
}
