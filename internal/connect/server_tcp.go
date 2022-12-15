package connect

import (
	"bufio"
	"context"
	"fmt"
	"go-im/api/protocol"
	"go-im/pkg/proto"
	"go.uber.org/zap"
	"io"
	"net"
	"runtime"
	"time"
)

const (
	maxInt = 1<<31 - 1
)

func InitTCP(s *Server, addrs []string) error {
	for _, addr := range addrs {
		var (
			err      error
			tcpAddr  *net.TCPAddr
			listener *net.TCPListener
		)
		if tcpAddr, err = net.ResolveTCPAddr("tcp", addr); err != nil {
			s.log.Error("resolve tcp addr err", zap.Error(err))
			return err
		}
		if listener, err = net.ListenTCP("tcp", tcpAddr); err != nil {
			s.log.Error("listen tcp err", zap.Error(err))
			return err
		}

		//默认最大go等于cpu核心数
		for i := 0; i < runtime.NumCPU(); i++ {
			//分割n核接收连接来提升性能
			go AcceptTCP(s, listener)
		}

	}
	return nil
}

func AcceptTCP(s *Server, listener *net.TCPListener) {
	var (
		err  error
		conn *net.TCPConn
		r    int
	)

	for {
		if conn, err = listener.AcceptTCP(); err != nil {
			s.log.Error("accept tcp err", zap.Error(err))
			return
		}

		if err = conn.SetKeepAlive(s.c.Tcp.KeepAlive); err != nil {
			s.log.Error("accept tcp err", zap.Error(err))
			return
		}

		if err = conn.SetReadBuffer(s.c.Tcp.ReceiveBuf); err != nil {
			s.log.Error("accept tcp err", zap.Error(err))
			return
		}
		if err = conn.SetWriteBuffer(s.c.Tcp.SendBuf); err != nil {
			s.log.Error("accept tcp err", zap.Error(err))
			return
		}

		go serverTCP(s, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serverTCP(s *Server, conn *net.TCPConn, r int) {
	var ch *Channel
	ch = NewChannel(0, s.c.Protocol.ProtoSize)
	ch.connTcp = conn
	s.ServeTCP(ch)
}

func (s *Server) closeTCP(ch *Channel, b *Bucket) {
	ch.connTcp.Close()
	//todo 处理
	if b != nil {
		b.Del(ch)
	}
	ch.Close()
}

func (s *Server) ServeTCP(ch *Channel) {
	var (
		err     error
		rid     string
		accepts []int32
		hb      time.Duration
		b       *Bucket
	)
	reader := bufio.NewReader(ch.connTcp)
	writer := bufio.NewWriter(ch.connTcp)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//远程连接的ip
	ch.IP, _, _ = net.SplitHostPort(ch.connTcp.RemoteAddr().String())
	p := new(protocol.Proto)
	//认证tcp连接
	if ch.Mid, ch.Key, rid, accepts, hb, err = s.authTCP(ctx, reader, writer, p); err != nil {
		s.log.Error("authTCP err:", zap.Error(err))
		s.closeTCP(ch, b)
		return
	}
	fmt.Println("heartBeat:", hb)
	ch.Watch(accepts...)
	//user key=>bucket=>room_id
	b = s.Bucket(ch.Key)
	if err = b.Put(rid, ch); err != nil {
		s.log.Error("put err:", zap.Error(err))
		s.closeTCP(ch, b)
		return
	}

	//读取消息并write数据到客户端
	go s.writeTCPData(ctx, ch)
	//读取前端发送过来的消息
	go s.readTCPData(ctx, ch, b)
}

func (s *Server) writeTCPData(ctx context.Context, ch *Channel) {
	var (
		p      *protocol.Proto
		finish bool
		online int32
		err    error
	)
	wr := bufio.NewWriter(ch.connTcp)
	for {
		//推送过来的消息
		p = ch.Ready()
		fmt.Println("read:", p.Op)
		switch p {
		case protocol.ProtoFinish:
			finish = true
			goto failed
		case protocol.ProtoReady:
			if p.Op == protocol.OpHeartbeatReply {
				if ch.Room != nil {
					online = ch.Room.OnlineNum()
				}
				//读到心跳将房间在线人数返回
				if err = proto.WriteTCPHeart(p, wr, online); err != nil {
					goto failed
				}
			} else {
				if err = proto.WriteTcp(p, wr); err != nil {
					goto failed
				}
			}
			p.Body = nil
		default:
			// server send 如果连接端口，写报错，直接关闭连接
			if err = proto.WriteTcp(p, wr); err != nil {
				goto failed
			}
		}
		if err = wr.Flush(); err != nil {
			break
		}
	}
failed:
	//todo 是否会重复关闭
	ch.connTcp.Close()
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = ch.Ready() == protocol.ProtoFinish
	}
}

func (s *Server) readTCPData(ctx context.Context, ch *Channel, b *Bucket) {
	var err error
	reader := bufio.NewReader(ch.connTcp)
	for {
		p := new(protocol.Proto)
		//消息解析
		err = proto.ReadTcp(p, reader)
		//todo 处理
		if err == io.EOF {
			s.log.Error("io.EOF err:", zap.Error(err))
		}
		if err != nil {
			s.log.Error("read data err:", zap.Error(err))
			break
		}
		fmt.Println("p:", p.Op, string(p.Body))
		if p.Op == protocol.OpHeartbeat {
			p.Op = protocol.OpHeartbeatReply
			p.Body = nil
		} else {
			if err = s.Operate(ctx, p, b, ch); err != nil {
				break
			}
		}
		//todo 敏感词过滤
		//channel长度不够会报错，等待数据被发出去
		if err = ch.Push(p); err != nil {
			s.log.Error(fmt.Sprintf("push proto err, key: %s mid: %d ", ch.Key, ch.Mid), zap.Error(err))
		}
	}

	s.closeTCP(ch, b)
	if err := s.Disconnect(ctx, ch.Mid, ch.Key); err != nil {
		s.log.Error(fmt.Sprintf("key: %s mid: %d operator do disconnect", ch.Key, ch.Mid), zap.Error(err))
	}
}

// auth for goim handshake with client, use rsa & aes.
//返回参数分别: 会员id 唯一key  房间id  用户切换 room, 也就在这里处理  心跳时间
func (s *Server) authTCP(ctx context.Context, rr *bufio.Reader, wr *bufio.Writer, p *protocol.Proto) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	for {
		if err = proto.ReadTcp(p, rr); err != nil {
			return 0, "", "", nil, 0, err
		}
		//判断是否是认证消息
		if p.Op == protocol.OpAuth {
			break
		} else {
			//todo 是否死循环
			s.log.Error(fmt.Sprintf("tcp request operation(%d) not auth", p.Op))
		}
	}
	if mid, key, rid, accepts, hb, err = s.Connect(ctx, p, ""); err != nil {
		s.log.Error("authTCP.Connect", zap.String("key", key), zap.Error(err))
		return
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil
	if err = proto.WriteTcp(p, wr); err != nil {
		s.log.Error("authTCP.WriteTCP", zap.String("key", key), zap.Error(err))
		return
	}
	//刷新数据到对端
	err = wr.Flush()
	return
}
