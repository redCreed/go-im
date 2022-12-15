package connect

import (
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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
	c.signal = make(chan *protocol.Proto, 1024)
	c.watchOps = make(map[int32]struct{})
	return c
}

func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	c.mutex.Unlock()
}

//Close 发送关闭信号 关闭这个channel
func (c *Channel) Close() {
	c.signal <- protocol.ProtoFinish
}

// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- protocol.ProtoReady
}
func (c *Channel) Ready() *protocol.Proto {
	return <-c.signal
}

func (c *Channel) Push(p *protocol.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
		err = errors.New("signal channel not enough")
	}
	return
}
