package main

import (
	"flag"
	"go-im/internal/logic"
	"go-im/internal/logic/conf"
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
	srv := logic.New(conf.Conf)
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
			//httpSrv.Close()
			//rpcSrv.GracefulStop()
			//log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
