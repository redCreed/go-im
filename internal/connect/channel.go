package connect

import (
	"github.com/gorilla/websocket"
	"go-im/api/protocol"
	"net"
	"sync"
)

type Channel struct {
	Room     *Room
	Next     *Channel
	Prev     *Channel
	signal   chan *protocol.Proto
	Mid      int64  //memberID
	Key      string //相等于sessionId
	IP       string
	watchOps map[int32]struct{} //int32 是房间号 map 多个房间号 一个 goim 终端能够接收多个房间发送来的 im 消息
	mutex    sync.RWMutex
	conn     *websocket.Conn
	connTcp  *net.TCPConn
}

// NewChannel new a channel.
func NewChannel(cli, svr int) *Channel {
	c := new(Channel)
	//c.CliProto.Init(cli)
	c.signal = make(chan *protocol.Proto, svr)
	c.watchOps = make(map[int32]struct{})
	return c
}
