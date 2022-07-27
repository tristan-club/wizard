package dcontext

import (
	"context"
	"encoding/base64"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	he "github.com/tristan-club/wizard/pkg/error"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"time"
)

type WalletBot struct {
	WalletMgr controller_pb.ControllerServiceClient
	//Bot       *discordgo..Bot
}

func NewWalletBot(walletConn *grpc.ClientConn) (*WalletBot, error) {
	wb := &WalletBot{
		WalletMgr: controller_pb.NewControllerServiceClient(walletConn),
	}
	return wb, nil
}

type Context struct {
	CmdId         string
	Context       context.Context
	IC            *discordgo.InteractionCreate
	CM            controller_pb.ControllerServiceClient
	Session       *discordgo.Session
	BotName       string
	Requester     *controller_pb.Requester
	IsICResponded bool
}

func (ctx *Context) GetFromId() string {
	if ctx.IC == nil {
		return ""
	}
	if ctx.IC.Member != nil {
		return ctx.IC.Member.User.ID
	}
	return ctx.IC.User.ID
}

//func (ctx *Context) GetGroupId() string {
//	return ctx.IC.ChannelID
//}

func (ctx *Context) GetGroupChannelId() string {
	return ctx.IC.ChannelID
}

func (ctx *Context) IsPrivate() bool {
	return ctx.IC.User != nil
}

func (ctx *Context) GetChatId() string {
	if ctx.IsPrivate() {
		return ctx.GetFromId()
	}
	return ctx.IC.ChannelID
}

func (ctx *Context) GetUserName() string {
	user := ctx.IC.User
	if user == nil {
		user = ctx.IC.Member.User
	}
	return user.Username
}

func (ctx *Context) GetNickname() string {
	if ctx.IsPrivate() {
		return ""
	}
	return ctx.IC.Member.Nick
}

func (ctx *Context) GetAvailableName() string {
	if ctx.GetNickname() != "" {
		return "@" + ctx.GetUserName()
	}
	return ctx.GetUserName()
}

func (ctx *Context) GetNickNameMDV2() string {
	name := ctx.GetNickname()
	if name == "" {
		name = ctx.GetUserName()
	}
	return "@" + name
}

func (ctx *Context) CopyRequester() (context.Context, context.CancelFunc, he.Error) {
	c, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	b, err := proto.Marshal(ctx.Requester)
	if err != nil {
		cancel()
		return c, nil, he.NewServerError(he.ServerError, "marshal data error", err)
	}
	requestStr := base64.StdEncoding.EncodeToString(b)
	md := metadata.Pairs("requester", requestStr)
	c = metadata.NewOutgoingContext(c, md)
	return c, cancel, nil
}
