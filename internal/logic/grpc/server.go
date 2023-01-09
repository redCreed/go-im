package grpc

import (
	"context"
	pb "go-im/api/logic"
	"go-im/internal/logic"
	"go-im/internal/logic/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
)

// New logic grpc server
func New(c *conf.RPCServer, l *logic.Logic) *grpc.Server {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     c.IdleTimeout,
		MaxConnectionAgeGrace: c.ForceCloseWait,
		Time:                  c.KeepAliveInterval,
		Timeout:               c.KeepAliveTimeout,
		MaxConnectionAge:      c.MaxLifeTime,
	})
	srv := grpc.NewServer(keepParams)
	pb.RegisterLogicServer(srv, &server{l})
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
	logic *logic.Logic
}

func (s server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.ConnectReply, error) {
	mid, key, room, accepts, hb, err := s.logic.Connect(ctx, req.Server, req.Cookie, req.Token)
	if err != nil {
		return &pb.ConnectReply{}, err
	}
	return &pb.ConnectReply{Mid: mid, Key: key, RoomID: room, Accepts: accepts, Heartbeat: hb}, nil
}

func (s server) Disconnect(ctx context.Context, req *pb.DisconnectReq) (*pb.DisconnectReply, error) {
	has, err := s.logic.Disconnect(ctx, req.Mid, req.Key, req.Server)
	if err != nil {
		return &pb.DisconnectReply{}, err
	}
	return &pb.DisconnectReply{Has: has}, nil
}

func (s server) Heartbeat(ctx context.Context, req *pb.HeartbeatReq) (*pb.HeartbeatReply, error) {
	//TODO implement me
	panic("implement me")
}

func (s server) RenewOnline(ctx context.Context, req *pb.OnlineReq) (*pb.OnlineReply, error) {
	allRoomCount, err := s.logic.RenewOnline(ctx, req.Server, req.RoomCount)
	if err != nil {
		return &pb.OnlineReply{}, err
	}
	return &pb.OnlineReply{AllRoomCount: allRoomCount}, nil
}

func (s server) Receive(ctx context.Context, req *pb.ReceiveReq) (*pb.ReceiveReply, error) {
	//TODO implement me
	panic("implement me")
}

func (s server) Nodes(ctx context.Context, req *pb.NodesReq) (*pb.NodesReply, error) {
	//TODO implement me
	panic("implement me")
}

var _ pb.LogicServer = &server{}
