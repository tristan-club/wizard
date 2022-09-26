package chain

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	he "github.com/tristan-club/kit/error"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext"
	"github.com/tristan-club/wizard/handler/tghandler/tcontext/expire_message"
	"github.com/tristan-club/wizard/handler/userstate"
	"github.com/tristan-club/wizard/pconst"
	"github.com/tristan-club/wizard/pkg/util"
)

type ChainHandler struct {
	Id          string
	PreHandlers []func(ctx *tcontext.Context) error
	StartState  string
	Nodes       []*Node
	SendHandler func(ctx *tcontext.Context) error
	Payload     map[*Node][]byte
	CmdParser   func(u *tgbotapi.Update) string
}

func NewChainHandler(id string, sendHandler func(ctx *tcontext.Context) error) *ChainHandler {
	return &ChainHandler{
		Id:          id,
		Nodes:       nil,
		SendHandler: sendHandler,
	}
}

func (c *ChainHandler) AddPreHandler(preHandler func(ctx *tcontext.Context) error) *ChainHandler {
	c.PreHandlers = append(c.PreHandlers, preHandler)
	return c
}

func (c *ChainHandler) AddNode(askFor, checkParma, enter func(ctx *tcontext.Context, node *Node) error) *ChainHandler {
	c.Nodes = append(c.Nodes, &Node{
		Id:                len(c.Nodes) + 1,
		AskForHandler:     askFor,
		CheckParamHandler: checkParma,
		EnterHandler:      enter,
	})
	return c
}

func (c *ChainHandler) AddPresetNode(node *Node, payload interface{}) *ChainHandler {
	if node == nil {
		log.Error().Msgf("nil node for handle id %s, payload %s", c.Id, util.FastMarshal(payload))
		return c
	}
	nodeCopy := &Node{}
	*nodeCopy = *node
	node = nodeCopy
	// 如果payload不为空，则代表要定制一个新的节点
	if !util.IsNil(payload) {
		herr := node.AddPayload(payload)
		if herr != nil {
			log.Error().Msgf("add payload error %s, node id %s, chandler id %s, payload %v", herr.Error(), node.Id, c.Id, payload)
		}
	}

	node.Id = len(c.Nodes) + 1
	c.Nodes = append(c.Nodes, node)
	return c
}

func (c *ChainHandler) GetCmdParser() func(u *tgbotapi.Update) string {
	return c.CmdParser
}

func (c *ChainHandler) AddCmdParser(parser func(u *tgbotapi.Update) string) *ChainHandler {
	c.CmdParser = parser
	return c
}

func (c *ChainHandler) Handle(ctx *tcontext.Context) error {
	log.Debug().Msgf("start chandler cmd %s, state %d, cmdId %s", ctx.CmdId, ctx.CurrentState, c.Id)
	if len(c.PreHandlers) != 0 && ctx.CurrentState == userstate.StateNone {
		for _, preHandler := range c.PreHandlers {
			if herr := preHandler(ctx); herr != nil {
				return herr
			}
		}
	}

	if ctx.CurrentState == userstate.StateNone && len(c.Nodes) > 0 {

		if askForHandler := c.Nodes[0].AskForHandler; askForHandler != nil {
			herr := c.Nodes[0].AskForHandler(ctx, c.Nodes[0])
			if herr != nil {
				return herr
			}
		}

		if herr := userstate.SetState(ctx.Requester.RequesterOpenId, c.Nodes[0].Id, "", "", nil); herr != nil {
			return herr
		} else {
			return nil
		}

	}

	if len(c.Nodes) != 0 {
		for k, node := range c.Nodes {
			if node.Id == ctx.CurrentState {
				if node.CheckParamHandler != nil {
					herr := node.CheckParamHandler(ctx, node)
					if herr != nil {
						return herr
					}
				}
				if node.EnterHandler != nil {
					herr := node.EnterHandler(ctx, node)
					if herr != nil {
						return herr
					}
					expire_message.ClearPreviousStepExpireMessage(ctx)
				}

				if k == len(c.Nodes)-1 {
					if herr := c.SendHandler(ctx); herr != nil {
						return herr
					}
					if herr := userstate.ResetState(ctx.OpenId()); herr != nil {
						return herr
					}
					return nil
				} else {
					herr := c.Nodes[k+1].AskForHandler(ctx, c.Nodes[k+1])
					if herr != nil {
						return herr
					} else {
						herr = userstate.SetState(ctx.Requester.RequesterOpenId, c.Nodes[k+1].Id, "", "", nil)
						if herr != nil {
							return herr
						} else {
							return nil
						}
					}
				}
			}
		}
		return he.NewBusinessError(pconst.CodeInvalidUserState, "", nil)
	} else {
		if herr := c.SendHandler(ctx); herr != nil {
			return herr
		}
		if herr := userstate.ResetState(ctx.OpenId()); herr != nil {
			return herr
		}
	}

	return nil
}
