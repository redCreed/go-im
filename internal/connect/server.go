package connect

import (
	"github.com/google/uuid"
	"go-im/api/logic"
	"go-im/internal/connect/conf"
	"go-im/pkg/cityhash"
	"go-im/pkg/log"
	"strconv"
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
func NewServer(c *conf.Config) *Server {
	s := &Server{}
	s.buckets = make([]*Bucket, c.Bucket.Size)
	s.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		s.buckets[i] = NewBucket(c.Bucket)
	}
	//生成uuid或者ip
	s.serverID = uuid.New().String()
	s.log = log.NewLog("im", c.Mode.Debug)
	s.c = c
	//s.rpcClient
	go s.onlineProc()
	return s
}

func (s *Server) onlineProc() {

}

func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

func (s *Server) Bucket(userId int) *Bucket {
	str := strconv.Itoa(userId)
	idx := cityhash.CityHash32([]byte(str), uint32(len(str))) % uint32(len(s.buckets))
	return s.buckets[idx]
}
