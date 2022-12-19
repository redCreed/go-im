package job

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-im/api/connect"
	"go-im/internal/job/conf"
	"go-im/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"sync/atomic"
	"time"
)

var (
	// grpc options
	grpcKeepAliveTime    = time.Duration(10) * time.Second
	grpcKeepAliveTimeout = time.Duration(3) * time.Second
	grpcBackoffMaxDelay  = time.Duration(3) * time.Second
	grpcMaxSendMsgSize   = 1 << 24
	grpcMaxCallMsgSize   = 1 << 24
)

const (
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
)

type ConnectServer struct {
	serverId      string
	client        connect.CometClient
	pushCh        []chan *connect.PushMsgReq
	roomCh        []chan *connect.BroadcastRoomReq
	broadcastChan chan *connect.BroadcastReq

	pushChanNum uint64
	roomChanNum uint64
	routineSize uint64

	log *log.Log
}

func newConnectServer(c *conf.Config) (*ConnectServer, error) {
	s := new(ConnectServer)
	s.serverId = "connect_server_1"
	//
	client, err := newCometClient("127.0.0.1:5566")
	if err != nil {
		return nil, err
	}
	s.client = client
	s.pushCh = make([]chan *connect.PushMsgReq, s.routineSize)
	s.roomCh = make([]chan *connect.BroadcastRoomReq, s.routineSize)
	s.broadcastChan = make(chan *connect.BroadcastReq, s.routineSize)
	for i := 0; i < int(s.routineSize); i++ {
		s.pushCh[i] = make(chan *connect.PushMsgReq, s.pushChanNum)
		s.roomCh[i] = make(chan *connect.BroadcastRoomReq, s.roomChanNum)
		go s.process(s.pushCh[i], s.roomCh[i], s.broadcastChan)
	}

	return s, nil
}

func newCometClient(addr string) (connect.CometClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
		}...,
	)
	if err != nil {
		return nil, err
	}
	return connect.NewCometClient(conn), err
}

func (c *ConnectServer) process(pushCh chan *connect.PushMsgReq, roomCh chan *connect.BroadcastRoomReq, broadcastChan chan *connect.BroadcastReq) {
	for {
		select {
		case pushData := <-pushCh:
			_, err := c.client.PushMsg(context.Background(), pushData)
			if err != nil {
				c.log.Error("推送指定key错误", zap.Error(err))
			}
		case roomData := <-roomCh:
			_, err := c.client.BroadcastRoom(context.Background(), roomData)
			if err != nil {
				c.log.Error("推送指定房间错误", zap.Error(err))
			}
		case broadcastData := <-broadcastChan:
			_, err := c.client.Broadcast(context.Background(), broadcastData)
			if err != nil {
				c.log.Error("广播错误", zap.Error(err))
			}
		}
	}
}

func (c *ConnectServer) PushKey(msg *connect.PushMsgReq) error {
	index := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	select {
	case c.pushCh[index] <- msg:
	default:
		return errors.New(fmt.Sprintf("serverId:%s index:%d pushKeyChan not enough", c.serverId, index))
	}
	return nil
}

func (c *ConnectServer) PushRoom(msg *connect.BroadcastRoomReq) error {
	index := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	select {
	case c.roomCh[index] <- msg:
	default:
		return errors.New(fmt.Sprintf("serverId:%s index:%d broadcastRoomChan not enough", c.serverId, index))
	}
	return nil
}

func (c *ConnectServer) Broadcast(msg *connect.BroadcastReq) error {
	c.broadcastChan <- msg
	select {
	case c.broadcastChan <- msg:
	default:
		return errors.New(fmt.Sprintf("serverId:%s broadcastChan not enough", c.serverId))
	}
	return nil
}
