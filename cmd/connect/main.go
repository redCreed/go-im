package main

import (
	"flag"
	"go-im/internal/connect"
	"go-im/internal/connect/conf"
	"go-im/internal/connect/grpc"
	"go-im/pkg/etcd"
	"os"
	"os/signal"
	"syscall"
)

var (
	key = "discovery:"
)

func main() {
	var (
		confPath string
		serverId string
	)

	flag.StringVar(&confPath, "conf", "configs/connect.yaml", "default config path.")
	flag.StringVar(&serverId, "server", "127.0.0.1", "ip")
	flag.Parse()
	conf.Parse(confPath)
	s := connect.NewServer(conf.Conf, serverId)
	if err := connect.InitTCP(s, conf.Conf.Tcp.Host); err != nil {
		panic(err)
	}
	if err := connect.InitWebsocket(s, conf.Conf.Websocket.Host); err != nil {
		panic(err)
	}

	ser, err := etcd.NewRegister([]string{conf.Conf.Discovery.Host}, key+serverId, serverId, int64(conf.Conf.Discovery.Lease))
	if err != nil {
		panic(err)
	}

	// new grpc server
	rpcSrv := grpc.New(conf.Conf.RPCServer, s)
	// signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			ser.Close()
			return
		case syscall.SIGHUP:

		default:
			return
		}
	}
}
