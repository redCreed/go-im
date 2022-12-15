package grpc

import (
	"context"
	pb "go-im/api/connect"
	"go-im/internal/connect"
	"go-im/internal/connect/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

// New comet grpc server.
func New(c *conf.RPCServer, s *connect.Server) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterCometServer(srv, &server{s})
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return srv
}

type server struct {
	srv *connect.Server
}

func (s server) PushMsg(ctx context.Context, req *pb.PushMsgReq) (*pb.PushMsgReply, error) {
	//TODO implement me
	panic("implement me")
}

func (s server) Broadcast(ctx context.Context, req *pb.BroadcastReq) (*pb.BroadcastReply, error) {
	//TODO implement me
	panic("implement me")
}

func (s server) BroadcastRoom(ctx context.Context, req *pb.BroadcastRoomReq) (*pb.BroadcastRoomReply, error) {
	//TODO implement me
	panic("implement me")
}

func (s server) Rooms(ctx context.Context, req *pb.RoomsReq) (*pb.RoomsReply, error) {
	//TODO implement me
	panic("implement me")
}

var _ pb.CometServer = &server{}
