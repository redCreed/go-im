package connect

import (
	"context"
	"fmt"
	"go-im/api/logic"
	"go-im/api/protocol"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"strconv"
	"strings"
	"time"
)

// Connect connected a connection.
func (s *Server) Connect(c context.Context, p *protocol.Proto, cookie string) (mid int64, key, rid string, accepts []int32, heartbeat time.Duration, err error) {
	return 11022, "inson", "ooodi", []int32{15314}, time.Duration(3), nil

	reply, err := s.rpcClient.Connect(c, &logic.ConnectReq{
		Server: s.serverID,
		Cookie: cookie,
		Token:  p.Body,
	})
	if err != nil {
		return
	}
	return reply.Mid, reply.Key, reply.RoomID, reply.Accepts, time.Duration(reply.Heartbeat), nil
}

// Disconnect disconnected a connection.
func (s *Server) Disconnect(c context.Context, mid int64, key string) (err error) {
	return nil
	_, err = s.rpcClient.Disconnect(context.Background(), &logic.DisconnectReq{
		Server: s.serverID,
		Mid:    mid,
		Key:    key,
	})
	return
}

func (s *Server) Operate(ctx context.Context, p *protocol.Proto, b *Bucket, ch *Channel) error {
	switch p.Op {
	case protocol.OpChangeRoom:
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			s.log.Error("change room err",
				zap.String("room_id", string(p.Body)),
				zap.String("ch key", ch.Key),
				zap.Error(err))
		}

		p.Op = protocol.OpChangeRoomReply
	case protocol.OpSub:
		if ops, err := SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = protocol.OpSubReply
	case protocol.OpUnsub:
		if ops, err := SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = protocol.OpUnsubReply
	default: //发送到logic(真正发送消息是http请求)  默认为发送消息
		if err := s.Receive(ctx, ch.Mid, p); err != nil {
			s.log.Error(fmt.Sprintf("s.Report(%d) op:%d", ch.Mid, p.Op), zap.Error(err))
		}
		p.Body = nil
		p.Op = protocol.OpUnsubReply
	}
	return nil
}

// Receive receive a message.
func (s *Server) Receive(ctx context.Context, mid int64, p *protocol.Proto) (err error) {
	_, err = s.rpcClient.Receive(ctx, &logic.ReceiveReq{Mid: mid, Proto: p})
	return
}

// SplitInt32s split string into int32 slice.
func SplitInt32s(s, p string) ([]int32, error) {
	if s == "" {
		return nil, nil
	}
	sArr := strings.Split(s, p)
	res := make([]int32, 0, len(sArr))
	for _, sc := range sArr {
		i, err := strconv.ParseInt(sc, 10, 32)
		if err != nil {
			return nil, err
		}
		res = append(res, int32(i))
	}
	return res, nil
}

// RenewOnline renew room online.
func (s *Server) RenewOnline(ctx context.Context, serverID string, roomCount map[string]int32) (allRoom map[string]int32, err error) {
	reply, err := s.rpcClient.RenewOnline(ctx, &logic.OnlineReq{
		Server:    s.serverID,
		RoomCount: roomCount,
	}, grpc.UseCompressor(gzip.Name))
	if err != nil {
		return
	}
	return reply.AllRoomCount, nil
}
