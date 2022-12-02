package connect

import (
	"bufio"
	"go.uber.org/zap"
	"net"
	"runtime"
)

const (
	maxInt = 1<<31 - 1
)

func InitTcp(s *Server, addrs []string) error {
	var (
		err      error
		tcpAddr  *net.TCPAddr
		listener *net.TCPListener
	)
	for _, addr := range addrs {
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
			go AcceptTcp(s, listener)
		}

	}
	return nil
}

func AcceptTcp(s *Server, listener *net.TCPListener) {
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

		go serverTcp(s, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serverTcp(s *Server, conn *net.TCPConn, r int) {
	var ch *Channel
	ch = NewChannel(0, s.c.Protocol.ProtoSize)
	ch.connTcp = conn
	s.ServeTCP(ch)
}

func (s *Server) ServeTCP(ch *Channel) {
	bufio.NewReader(ch.connTcp)
}
