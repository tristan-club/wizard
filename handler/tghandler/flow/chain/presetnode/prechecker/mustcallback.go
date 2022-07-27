package prechecker

import (
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/log"
	"github.com/tristan-club/wizard/pkg/util"
)

func MustBeCallback(ctx *tcontext.Context, node *chain.Node) error {
	if ctx.U.CallbackQuery == nil {
		log.Error().Msgf("got invalid user callback %s", util.FastMarshal(ctx))
		return he.NewBusinessError(he.CodeInvalidUserState, "", nil)
	}
	return nil
}
