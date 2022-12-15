package connect

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"go-im/api/protocol"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strconv"
	"time"
)

func InitWebsocket(s *Server, addrs []string) error {
	for _, addr := range addrs {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			handleWs(s, w, r)
		})
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}
	return nil
}

func handleWs(s *Server, w http.ResponseWriter, r *http.Request) {
	var upGrader = websocket.Upgrader{
		ReadBufferSize:   4096,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		// cross origin domain
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		// 处理 Sec-WebSocket-Protocol Header
		Subprotocols:      []string{r.Header.Get("Sec-WebSocket-Protocol")},
		EnableCompression: true,
	}

	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Error("upgrade err", zap.Error(err))
		return
	}
	fmt.Println("ws connect")
	ch := NewChannel(0, s.c.Protocol.ProtoSize)
	ch.ws = conn
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//远程连接的ip
	ch.IP, _, _ = net.SplitHostPort(ch.ws.RemoteAddr().String())
	var (
		rid     string
		accepts []int32
		b       *Bucket
	)
	//认证tcp连接
	if ch.Mid, ch.Key, rid, accepts, _, err = s.authWebsocket(ctx, conn, r.Header.Get("Cookie")); err != nil {
		s.log.Error("authTCP err:", zap.Error(err))
		s.closeWs(ch, b)
		return
	}
	ch.Watch(accepts...)
	b = s.Bucket(ch.Key)
	if err = b.Put(rid, ch); err != nil {
		s.log.Error("put err:", zap.Error(err))
		s.closeWs(ch, b)
		return
	}

	//读取消息并write数据到客户端
	go s.writeWs(ctx, ch)
	//读取前端发送过来的消息
	go s.readWs(ctx, ch, b)
}

func (s *Server) readWs(ctx context.Context, ch *Channel, b *Bucket) {
	var (
		err error
		//mT    int
		msg []byte
		p   *protocol.Proto
	)
	for {
		p = new(protocol.Proto)
		if _, msg, err = ch.ws.ReadMessage(); err != nil {
			//检测到前端关闭
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
				break
			}
			s.log.Error("ws read data err", zap.Error(err))
			break
		}

		if err = jsoniter.Unmarshal(msg, p); err != nil {
			s.log.Error("Unmarshal data err", zap.String("data", string(msg)), zap.Error(err))
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

func (s *Server) writeWs(ctx context.Context, ch *Channel) {
	var (
		p      *protocol.Proto
		finish bool
		err    error
	)
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
					p.Body = []byte(strconv.Itoa(int(ch.Room.OnlineNum())))
				}
			} else {
				p.Body = nil
			}
			if err = s.write(ch.ws, p); err != nil {
				goto failed
			}
		default:
			if err = s.write(ch.ws, p); err != nil {
				goto failed
			}
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

func (s *Server) write(ws *websocket.Conn, p *protocol.Proto) error {
	var (
		err error
		msg []byte
	)

	if msg, err = jsoniter.Marshal(p); err != nil {
		return err
	}
	if err = ws.WriteMessage(websocket.TextMessage, msg); err != nil {
		return err
	}

	return nil
}

func (s *Server) read(ws *websocket.Conn, p *protocol.Proto) error {
	var (
		err error
		msg []byte
	)

	if _, msg, err = ws.ReadMessage(); err != nil {
		return nil
	}

	if err = jsoniter.Unmarshal(msg, p); err != nil {
		return nil
	}

	return nil
}

func (s *Server) closeWs(ch *Channel, b *Bucket) {
	ch.ws.Close()
	if b != nil {
		b.Del(ch)
	}
	ch.Close()
}

// auth for goim handshake with client, use rsa & aes.
func (s *Server) authWebsocket(ctx context.Context, ws *websocket.Conn, cookie string) (mid int64, key, rid string, accepts []int32, hb time.Duration, err error) {
	var (
		mT    int
		msg   []byte
		times = 0
	)
	p := new(protocol.Proto)
	for {
		times++
		if times >= 4 {
			s.log.Warn(fmt.Sprintf("超过3次认证失败, cookie:%s", cookie))
			err = errors.New("超过3次认证失败")
			return
		}
		if mT, msg, err = ws.ReadMessage(); err != nil {
			s.log.Error("ws read data err", zap.Error(err))
			return
		}

		if err = jsoniter.Unmarshal(msg, p); err != nil {
			s.log.Error("Unmarshal data err", zap.String("data", string(msg)), zap.Error(err))
			return
		}

		if p.Op == protocol.OpAuth && mT == websocket.TextMessage {
			break
		} else {
			s.log.Error("ws request operation not auth", zap.Error(err))
		}
	}
	if mid, key, rid, accepts, hb, err = s.Connect(ctx, p, cookie); err != nil {
		return
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil

	if err = s.write(ws, p); err != nil {
		s.log.Error("ws write auth reply ", zap.Error(err))
		return
	}
	return
}
