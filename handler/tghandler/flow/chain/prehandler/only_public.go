package prehandler

import (
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/wizard/pkg/error"
)

func OnlyPublic(ctx *tcontext.Context) error {
	if ctx.U.FromChat().IsPrivate() {
		return he.NewBusinessError(he.CodeCmdNeedGroupChat, "", nil)
	}
	return nil
}
