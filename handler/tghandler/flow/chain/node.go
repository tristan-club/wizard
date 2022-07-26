package chain

import (
	"encoding/json"
	"fmt"
	"github.com/tristan-club/bot-wizard/handler/tghandler/tcontext"
	he "github.com/tristan-club/bot-wizard/pkg/error"
	"github.com/tristan-club/bot-wizard/pkg/log"
	"github.com/tristan-club/bot-wizard/pkg/util"
)

type Node struct {
	Id                int
	AskForHandler     func(ctx *tcontext.Context, node *Node) error
	CheckParamHandler func(ctx *tcontext.Context, node *Node) error
	EnterHandler      func(ctx *tcontext.Context, node *Node) error
	Payload           []byte
}

func NewNode(askFor, checkParma, enter func(ctx *tcontext.Context, node *Node) error) *Node {
	return &Node{
		AskForHandler:     askFor,
		CheckParamHandler: checkParma,
		EnterHandler:      enter,
	}
}

func (n *Node) IsPayloadNil() bool {
	return len(n.Payload) == 0
}

func (n *Node) AddPayload(input interface{}) error {
	if !util.IsNil(input) {
		b, err := json.Marshal(input)
		if err != nil {
			log.Error().Msgf("got invalid input %v, node %s", input, util.FastMarshal(n))
			return he.NewServerError(he.CodeMarshalError, "", err)
		}
		n.Payload = b
	}

	return nil
}

func (n *Node) TryGetPayload(out interface{}) error {
	if util.IsNil(n.Payload) {
		return nil
	}
	if util.IsNil(out) {
		log.Error().Fields(map[string]interface{}{"action": "Get payload invalid", "node id": n.Id, "payload": string(n.Payload)}).Send()
		//log.Error().Msgf("got invalid payload, node %s", util.FastMarshal(n.Id))
		return he.NewServerError(he.CodeInvalidPayload, "", fmt.Errorf("empty payload"))
	}

	if err := json.Unmarshal(n.Payload, &out); err != nil {
		log.Error().Msgf("unmarshal payload error %s, node %s", err.Error(), util.FastMarshal(n))
		return he.NewServerError(he.CodeMarshalError, "", err)
	}
	return nil

}
