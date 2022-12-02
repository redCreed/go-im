package proto

import (
	"bufio"
	"bytes"
	"fmt"
	"go-im/api/protocol"
	"testing"
)

func Test_Proto(t *testing.T) {
	p := &protocol.Proto{
		Ver:  1,
		Op:   1,
		Seq:  2,
		Body: []byte("hello world"),
	}
	buf := make([]byte, 4096)
	buffer := bytes.NewBuffer(buf)
	w := bufio.NewWriter(buffer)
	WriteTcp(p, w)

	resp := new(protocol.Proto)
	ReadTcp(resp, bufio.NewReader(buffer))

	fmt.Println("resp:", resp)
}
