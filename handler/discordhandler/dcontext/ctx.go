package dcontext

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/tristan-club/kit/customid"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/wizard/entity/entity_pb/controller_pb"
	"github.com/tristan-club/wizard/pconst"
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
	CmdId     string
	Cid       *customid.CustomId
	Context   context.Context
	IC        *discordgo.InteractionCreate
	CM        controller_pb.ControllerServiceClient
	Session   *discordgo.Session
	BotName   string
	Requester *controller_pb.Requester
}

func DefaultContext(s *discordgo.Session, i *discordgo.InteractionCreate) *Context {
	return &Context{Session: s, IC: i}
}

// GetContext todo not support user_no and default_address
func GetContext(cc controller_pb.ControllerServiceClient, s *discordgo.Session, i *discordgo.InteractionCreate) *Context {
	ctx := &Context{
		//CmdId:     ,
		Context: context.Background(),
		IC:      i,
		CM:      cc,
		Session: s,
		BotName: "",
	}
	ctx.Requester = &controller_pb.Requester{
		//RequesterUserNo: ,
		RequesterOpenId:         ctx.GetFromId(),
		RequesterOpenType:       int32(pconst.PlatformDiscord),
		RequesterOpenNickname:   ctx.GetNickname(),
		RequesterOpenUserName:   ctx.GetUserName(),
		RequesterChannelId:      ctx.GetGroupChannelId(),
		RequesterDefaultAddress: "",
	}
	return ctx
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

	return fmt.Sprintf("<@%s>", ctx.GetFromId())
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
