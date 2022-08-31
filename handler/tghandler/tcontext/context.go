package tcontext

import (
	"context"
	"encoding/base64"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	he "github.com/tristan-club/wizard/pkg/error"
	"github.com/tristan-club/wizard/pkg/mdparse"
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
	Requester    *controller_pb.Requester
	Payload      interface{}
}

func DefaultContext(u *tgbotapi.Update, api *tgbotapi.BotAPI) *Context {
	return &Context{U: u, BotApi: api}
}

func (ctx *Context) OpenId() string {
	var openId string
	if openId = ctx.Requester.RequesterOpenId; openId == "" {
		openId = strconv.FormatInt(ctx.U.SentFrom().ID, 10)
	}
	return openId
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

func (ctx *Context) GetNickNameMDV2() string {
	nicknameAt := fmt.Sprintf("[@%s](tg://user?id=%s)", mdparse.ParseV2(ctx.GetNickname()), ctx.OpenId())
	return nicknameAt
}

func (ctx *Context) CopyRequester() (context.Context, he.Error) {
	c := context.Background()
	b, err := proto.Marshal(ctx.Requester)
	if err != nil {
		return c, he.NewServerError(he.CodeMarshalError, "", err)
	}
	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	c = metadata.NewOutgoingContext(c, md)
	return c, nil
}
