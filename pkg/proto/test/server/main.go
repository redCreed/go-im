package main

import (
	"bufio"
	"fmt"
	"go-im/api/protocol"
	"go-im/pkg/proto"
	"io"
	"net"
)

func main() {
	listen, _ := net.Listen("tcp", "127.0.0.1:9000")
	defer listen.Close()
	fmt.Println("start listen 9000")
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go handler(conn)
	}
	select {}
}

func handler(conn net.Conn) {
	p := new(protocol.Proto)
	r := bufio.NewReader(conn)
	for {
		err := proto.ReadTcp(p, r)
		if err == io.EOF {
			break
		}
		if err != io.EOF && err != nil {
			fmt.Println("server err:", err)
		}

		fmt.Println(p.Ver, p.Op, p.Seq, string(p.Body))
	}
}
