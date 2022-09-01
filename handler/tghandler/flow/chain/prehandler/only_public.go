package prehandler

import (
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/pconst"
)

func OnlyPublic(ctx *tcontext.Context) error {
	if ctx.U.FromChat().IsPrivate() {
		return he.NewBusinessError(pconst.CodeCmdNeedGroupChat, "", nil)
	}
	return nil
}
