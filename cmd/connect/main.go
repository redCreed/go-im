package main

import (
	"flag"
	"go-im/internal/connect"
	"go-im/internal/connect/conf"
	"go-im/internal/connect/grpc"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "configs/connect.yaml", "default config path.")
	flag.Parse()
	conf.Parse(confPath)
	s := connect.NewServer(conf.Conf)
	if err := connect.InitTCP(s, conf.Conf.Tcp.Host); err != nil {
		panic(err)
	}
	//todo websocket
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
			return
		case syscall.SIGHUP:

		default:
			return
		}
	}
}
