package tcontext

import (
	"context"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/mdparse"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/pconst"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"strconv"
)

type Context struct {
	CmdId        string
	CurrentState int
	Context      context.Context
	U            *tgbotapi.Update
	CM           controller_pb.ControllerServiceClient
	BotApi       *tgbotapi.BotAPI
	BotName      string
	BotId        int64
	Requester    *controller_pb.Requester
	Payload      interface{}
	CmdParam     []string
	Msg          *tgbotapi.Message

	Param  interface{}
	Result interface{}
}

func DefaultContext(u *tgbotapi.Update, api *tgbotapi.BotAPI) *Context {
	return &Context{U: u, BotApi: api}
}

func (ctx *Context) OpenId() string {
	return strconv.FormatInt(ctx.U.SentFrom().ID, 10)
}

func (ctx *Context) GetUserName() string {
	return ctx.U.SentFrom().UserName
}

func (ctx *Context) GetNickname() string {
	return ctx.U.SentFrom().FirstName + " " + ctx.U.SentFrom().LastName
}

func (ctx *Context) GetAvailableName() string {
	if ctx.GetUserName() != "" {
		return "@" + ctx.GetUserName()
	}
	return ctx.GetNickname()
}

func (ctx *Context) GetMentionName() string {
	name := ctx.U.SentFrom().UserName
	if name == "" {
		name = ctx.U.SentFrom().FirstName + " " + ctx.U.SentFrom().LastName
	}
	return "@" + name
}

func (ctx *Context) GenerateDeepLink(cid *customid.CustomId) string {
	return fmt.Sprintf("https://t.me/%s?start=%s", ctx.BotName, cid.String())
}

func (ctx *Context) GetNickNameMDV2() string {
	nicknameAt := fmt.Sprintf("[@%s](tg://user?id=%s)", mdparse.ParseV2(ctx.GetNickname()), ctx.OpenId())
	return nicknameAt
}

func (ctx *Context) GenerateNickName(label string, openId string) string {
	return fmt.Sprintf("[@%s](tg://user?id=%s)", label, openId)
}

func (ctx *Context) CopyRequester() (context.Context, he.Error) {
	c := context.Background()
	b, err := proto.Marshal(ctx.Requester)
	if err != nil {
		return c, he.NewServerError(pconst.CodeMarshalError, "", err)
	}
	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	c = metadata.NewOutgoingContext(c, md)
	return c, nil
}
