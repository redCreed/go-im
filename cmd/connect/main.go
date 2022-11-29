package main

import "flag"

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "configs/connect.yaml", "default config path.")
	flag.Parse()

}
