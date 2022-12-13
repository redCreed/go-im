package main

import (
	"bufio"
	"fmt"
	"go-im/api/protocol"
	"go-im/pkg/proto"
	"math/rand"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	w := bufio.NewWriter(conn)
	for i := 0; i < 20; i++ {
		p := &protocol.Proto{
			Ver:  1,
			Op:   2,
			Seq:  3,
			Body: nil,
		}
		data := fmt.Sprintf("[这是世界的一部分是不是]：%d", i)
		p.Body = []byte(data)
		if err := proto.WriteTcp(p, w); err != nil {
			fmt.Println("client, err:", err)
		}
	}
}
