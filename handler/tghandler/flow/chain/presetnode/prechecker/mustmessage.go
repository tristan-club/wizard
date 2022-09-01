package prechecker

import (
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
)

func MustBeMessage(ctx *tcontext.Context, node *chain.Node) error {
	if ctx.U.Message == nil {
		log.Error().Msgf("got invalid user message %s", util.FastMarshal(ctx))
		return he.NewBusinessError(pconst.CodeInvalidUserState, "", nil)
	}
	return nil
}
