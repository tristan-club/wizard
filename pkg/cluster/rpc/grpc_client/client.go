package grpc_client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
)

func Start(clientAddr string) (*grpc.ClientConn, error) {
	var kacp = keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             3 * time.Second,  // wait 3 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}

	return grpc.Dial(clientAddr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))

}
