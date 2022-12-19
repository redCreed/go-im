package main

import (
	"flag"
	"go-im/internal/job"
	"go-im/internal/job/conf"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "configs/connect.yaml", "default config path.")
	flag.Parse()
	conf.Parse(confPath)
	s := job.NewServer(conf.Conf)
	s.Consume()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:

		default:
			return
		}
	}
}
