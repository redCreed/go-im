package job

import (
	"context"
	"fmt"
	"go-im/api/connect"
	pb "go-im/api/logic"
	"go-im/api/protocol"
	"go.uber.org/zap"
)

func (s *Server) push(ctx context.Context, pushMsg *pb.PushMsg) (err error) {
	switch pushMsg.Type {
	case pb.PushMsg_PUSH:
		err = s.pushKeys(pushMsg.Operation, pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	case pb.PushMsg_ROOM:
		err = s.pushRoom(pushMsg.Operation, pushMsg.Server, pushMsg.Room, pushMsg.Msg)
	case pb.PushMsg_BROADCAST:
		err = s.broadcast(pushMsg.Operation, pushMsg.Msg, pushMsg.Speed)
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return
}

func (s *Server) broadcast(operation int32, body []byte, speed int32) error {
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}
	speed /= int32(len(s.connect))
	var args = connect.BroadcastReq{
		ProtoOp: operation,
		Proto:   p,
		Speed:   speed,
	}
	var err error
	for serverID, c := range s.connect {
		if err = c.Broadcast(&args); err != nil {
			s.log.Error(fmt.Sprintf("c.Broadcast(%v) serverID:%s  ", args, serverID), zap.Error(err))
		}
	}
	return nil
}

func (s *Server) pushRoom(operation int32, serverId string, roomId string, body []byte) error {
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}

	msg := &connect.BroadcastRoomReq{
		RoomID: roomId,
		Proto:  p,
	}
	var err error
	if c, ok := s.connect[serverId]; ok {
		if err = c.PushRoom(msg); err != nil {
			s.log.Error("", zap.Error(err))
		}
	}
	return nil
}

//个推
func (s *Server) pushKeys(operation int32, serverId string, keys []string, body []byte) error {
	p := &protocol.Proto{
		Ver:  1,
		Op:   operation,
		Body: body,
	}

	msg := &connect.PushMsgReq{
		Keys:    keys,
		ProtoOp: operation,
		Proto:   p,
	}
	var err error
	if c, ok := s.connect[serverId]; ok {
		if err = c.PushKey(msg); err != nil {
			s.log.Error("", zap.Error(err))
		}
	}
	return nil
}
