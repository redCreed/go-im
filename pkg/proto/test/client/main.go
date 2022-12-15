package main

import (
	"bufio"
	"fmt"
	"go-im/api/protocol"
	"go-im/pkg/proto"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:3101")
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(conn)
	p := &protocol.Proto{
		Ver:  1,
		Op:   17,
		Seq:  3,
		Body: nil,
	}
	if err := proto.WriteTcp(p, w); err != nil {
		fmt.Println("client, err:", err)
	}

	r := bufio.NewReader(conn)
	rp := new(protocol.Proto)
	if err := proto.ReadTcp(rp, r); err != nil {
		fmt.Println("client read:", err)
	}

	fmt.Println("read:", rp.Op)
	go func() {
		rand.Seed(time.Now().UnixNano())
		w := bufio.NewWriter(conn)
		for i := 0; i < 20; i++ {
			p := &protocol.Proto{
				Ver:  1,
				Op:   11,
				Seq:  3,
				Body: nil,
			}
			data := fmt.Sprintf("[这是世界的一部分是不是]：%d", i)
			p.Body = []byte(data)
			if err := proto.WriteTcp(p, w); err != nil {
				fmt.Println("client, err:", err)
			}
		}

	}()

	go func() {
		r := bufio.NewReader(conn)
		for {
			p := new(protocol.Proto)
			if err := proto.ReadTcp(p, r); err != nil {
				fmt.Println("client read:", err)
			}
			fmt.Println("read:", p.Op)
		}
	}()
	s := make(chan os.Signal, 1)
	<-s
}
