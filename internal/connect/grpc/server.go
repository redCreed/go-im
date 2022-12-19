package grpc

import (
	"context"
	"errors"
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
		MaxConnectionIdle:     c.IdleTimeout,
		MaxConnectionAgeGrace: c.ForceCloseWait,
		Time:                  c.KeepAliveInterval,
		Timeout:               c.KeepAliveTimeout,
		MaxConnectionAge:      c.MaxLifeTime,
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

var _ pb.CometServer = &server{}

func (s server) PushMsg(ctx context.Context, req *pb.PushMsgReq) (*pb.PushMsgReply, error) {
	if len(req.Keys) == 0 || req.Proto == nil {
		return nil, errors.New("参数非法")
	}
	var err error
	for _, key := range req.Keys {
		b := s.srv.Bucket(key)
		if b == nil {
			continue
		}
		if channel := b.Channel(key); channel != nil {
			if !channel.NeedPush(req.ProtoOp) {
				continue
			}
			if err = channel.Push(req.Proto); err != nil {
				return &pb.PushMsgReply{}, err
			}
		}
	}
	return &pb.PushMsgReply{}, nil
}

func (s server) Broadcast(ctx context.Context, req *pb.BroadcastReq) (*pb.BroadcastReply, error) {
	if req.Proto == nil {
		return &pb.BroadcastReply{}, errors.New("参数错误")
	}
	go func() {
		for _, bucket := range s.srv.Buckets() {
			bucket.Broadcast(req.Proto, req.ProtoOp)
			if req.Speed > 0 {
				t := bucket.ChannelCount() / int(req.Speed)
				time.Sleep(time.Duration(t) * time.Second)
			}
		}
	}()
	return &pb.BroadcastReply{}, nil
}

func (s server) BroadcastRoom(ctx context.Context, req *pb.BroadcastRoomReq) (*pb.BroadcastRoomReply, error) {
	if req.Proto == nil || req.RoomID == "" {
		return nil, errors.New("参数错误")
	}
	for _, bucket := range s.srv.Buckets() {
		bucket.BroadcastRoom(req)
	}

	return &pb.BroadcastRoomReply{}, nil
}

func (s server) Rooms(ctx context.Context, req *pb.RoomsReq) (*pb.RoomsReply, error) {
	var (
		roomIds = make(map[string]bool)
	)
	for _, bucket := range s.srv.Buckets() {
		for roomID := range bucket.Rooms() {
			roomIds[roomID] = true
		}
	}
	return &pb.RoomsReply{Rooms: roomIds}, nil
}
