package connect

import (
	"context"
	"go-im/api/logic"
	"go-im/internal/connect/conf"
	"go-im/pkg/cityhash"
	"go-im/pkg/log"
	"time"
)

type Server struct {
	c         *conf.Config
	serverID  string
	buckets   []*Bucket
	bucketIdx uint32
	rpcClient logic.LogicClient
	log       *log.Log
}

// NewServer returns a new Server.
func NewServer(c *conf.Config, serverId string) *Server {
	s := &Server{}
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	//生成uuid或者ip
	s.serverID = serverId
	s.log = log.NewLog("im", c.Mode.Debug)
	s.c = c
	//todo s.rpcClient

	//更新用户在线人数
	go s.onlineProc()
	return s
}

func (s *Server) onlineProc() {
	for {
		var (
			allRoomsCount map[string]int32
			err           error
		)
		roomCount := make(map[string]int32)
		for _, bucket := range s.buckets {
			for roomID, count := range bucket.RoomsCount() {
				roomCount[roomID] += count
			}
		}
		//防止出现死循环
		if allRoomsCount, err = s.RenewOnline(context.Background(), s.serverID, roomCount); err != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		for _, bucket := range s.buckets {
			bucket.UpdateRoomCount(allRoomsCount)
		}
		time.Sleep(time.Second * 10)
	}
}

func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

func (s *Server) Bucket(key string) *Bucket {
	idx := cityhash.CityHash32([]byte(key), uint32(len(key))) % uint32(len(s.buckets))
	return s.buckets[idx]
}
