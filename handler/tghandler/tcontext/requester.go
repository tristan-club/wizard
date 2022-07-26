package tcontext

import (
	"fmt"
	"github.com/tristan-club/bot-wizard/entity/entity_pb/controller_pb"
	he "github.com/tristan-club/bot-wizard/pkg/error"
)

func RespToError(in interface{}) he.Error {
	switch in.(type) {
	case *controller_pb.ControllerCommonResponse:
		resp := in.(*controller_pb.ControllerCommonResponse)
		if resp.Code == he.BusinessError {
			return he.NewBusinessError(int(resp.Code), resp.Message, nil)
		} else {
			return he.NewServerError(int(resp.Code), "", fmt.Errorf(resp.Message))
		}
	default:
		return he.NewServerError(he.ServerError, "", fmt.Errorf("invalid response parse format"))
	}

}
