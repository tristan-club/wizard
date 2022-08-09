package tstore

import (
	"context"
	"errors"
	"github.com/tristan-club/kit/log"
	"github.com/tristan-club/wizard/entity/entity_pb/tstore_pb"
	"github.com/tristan-club/wizard/pkg/cluster/rpc/grpc_client"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"time"
)

var conn *grpc.ClientConn

func InitTStore(svc string) (err error) {
	conn, err = grpc_client.Start(svc)
	return err
}

func PBSave(uid string, path string, value proto.Message) error {

	_, err := Save(&tstore_pb.SaveParam{
		Uid:    uid,
		Path:   path,
		IValue: tstore_pb.NewIValue(value),
	})
	return err
}

func PBSaveString(uid string, path string, value string) error {

	_, err := Save(&tstore_pb.SaveParam{
		Uid:    uid,
		Path:   path,
		IValue: tstore_pb.NewIValue(value),
	})
	return err
}

func PBGetStr(uid string, path string) (string, error) {
	v, err := Fetch(uid, path)
	if err != nil {
		log.Error().Msgf("fetch tstore error,%s", err.Error())
		return "", err
	}
	if v.Code == 404 {
		return "", nil
	}
	if v.IValue.Itype != tstore_pb.IValue_str || v.Code != CodeSuccess {
		log.Error().Msgf("fetch tstore error,%s", err)
		return "", errors.New("pb get error")
	}
	return v.IValue.StrValue, nil
}

func PBGet(uid string, path string) ([]byte, error) {
	v, err := Fetch(uid, path)
	log.Info().Fields(map[string]interface{}{
		"pb get value": v,
	}).Send()
	if err != nil {
		log.Error().Msgf("fetch tstore error,%s", err.Error())
		return nil, err
	}
	if v.Code == 404 {
		return nil, nil
	}

	if v.IValue.Itype != tstore_pb.IValue_any || v.Code != CodeSuccess {
		log.Error().Msgf("fetch tstore error,%s", err)
		return []byte{}, errors.New("pb get error")
	}

	return v.IValue.AnyValue.Value, nil
}

func Save(v *tstore_pb.SaveParam) (*tstore_pb.SaveResp, error) {
	// 设定请求超时时间 3s
	cli := tstore_pb.NewTStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return cli.Save(ctx, v)
}

func Fetch(uid string, path string) (*tstore_pb.FetchResp, error) {
	cli := tstore_pb.NewTStoreServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return cli.Fetch(ctx, &tstore_pb.FetchParam{
		Uid:  uid,
		Path: path,
	})
}
