package prechecker

import (
	"github.com/tristan-club/bot-wizard/handler/tghandler/flow/chain"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"github.com/tristan-club/bot-wizard/pkg/util"
)

func MustBeMessage(ctx *tcontext.Context, node *chain.Node) error {
	if ctx.U.Message == nil {
		log.Error().Msgf("got invalid user message %s", util.FastMarshal(ctx))
		return he.NewBusinessError(he.CodeInvalidUserState, "", nil)
	}
	return nil
}
