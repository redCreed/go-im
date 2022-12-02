package proto

import (
	"encoding/binary"
	"fmt"
	"go-im/api/protocol"
	"io"
)

const (
	// MaxBodySize 最大数据字节长度 4096
	MaxBodySize    = int32(1 << 12)
	PackageSize    = 4
	HeaderSize     = 2
	VersionSize    = 2
	OperateSize    = 4
	SequenceSize   = 4
	HeartSize      = 4
	RawHeaderSize  = PackageSize + HeaderSize + VersionSize + OperateSize + SequenceSize
	MaxPackageSize = RawHeaderSize + MaxBodySize

	PackIndex     = 0
	HeaderIndex   = PackIndex + PackageSize
	VersionIndex  = HeaderIndex + HeaderSize
	OperateIndex  = VersionIndex + VersionSize
	SequenceIndex = OperateIndex + SequenceSize
	HeartIndex    = SequenceIndex + SequenceSize
)

func ReadTcp(p *protocol.Proto, reader io.Reader) {
	//binary.Read(reader, binary.BigEndian, p)
	err := binary.Read(reader, binary.BigEndian, &p.Ver)
	if err != nil {
		fmt.Println("err:", err)
	}
	binary.Read(reader, binary.BigEndian, &p.Op)
	binary.Read(reader, binary.BigEndian, &p.Seq)
	binary.Read(reader, binary.BigEndian, &p.Body)
}

func WriteTcp(p *protocol.Proto, writer io.Writer) {
	binary.Write(writer, binary.BigEndian, &p.Ver)
	binary.Write(writer, binary.BigEndian, &p.Op)
	binary.Write(writer, binary.BigEndian, &p.Seq)
	binary.Write(writer, binary.BigEndian, &p.Body)
}
