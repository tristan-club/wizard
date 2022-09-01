package prechecker

import (
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
)

func MustBeCallback(ctx *tcontext.Context, node *chain.Node) error {
	if ctx.U.CallbackQuery == nil {
		log.Error().Msgf("got invalid user callback %s", util.FastMarshal(ctx))
		return he.NewBusinessError(pconst.CodeInvalidUserState, "", nil)
	}
	return nil
}
