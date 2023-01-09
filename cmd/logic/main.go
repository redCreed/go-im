package main

import (
	"flag"
	"go-im/internal/logic"
	"go-im/internal/logic/conf"
	"go-im/internal/logic/grpc"
	"go-im/internal/logic/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "configs/connect.yaml", "default config path.")
	flag.Parse()
	conf.Parse(confPath)
	//todo etcd
	l := logic.New(conf.Conf)
	rpcSrv := grpc.New(conf.Conf.RPCServer, l)
	httpSrv := http.New(conf.Conf.HTTPServer, l)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			//if cancel != nil {
			//	cancel()
			//}
			//srv.Close()
			httpSrv.Close()
			rpcSrv.GracefulStop()
			//log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
