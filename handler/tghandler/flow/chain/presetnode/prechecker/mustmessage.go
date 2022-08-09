package prechecker

import (
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/util"
)

func MustBeMessage(ctx *tcontext.Context, node *chain.Node) error {
	if ctx.U.Message == nil {
		log.Error().Msgf("got invalid user message %s", util.FastMarshal(ctx))
		return he.NewBusinessError(he.CodeInvalidUserState, "", nil)
	}
	return nil
}
