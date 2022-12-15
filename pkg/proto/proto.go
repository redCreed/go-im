package proto

import (
	"bufio"
	"encoding/binary"
	"errors"
	"go-im/api/protocol"
	"strconv"
)

const (
	// MaxBodySize 最大数据字节长度 4096
	MaxBodySize = int32(1 << 12)
	//PacketLen 包长度，在数据流传输过程中，先写入整个包的长度，方便整个包的数据读取。
	//HeaderLen 头长度，在处理数据时，会先解析头部，可以知道具体业务操作。
	//VersionLen 协议版本号，主要用于上行和下行数据包按版本号进行解析。
	//OperationLen 业务操作码，可以按操作码进行分发数据包到具体业务当中。
	//SequenceLen 序列号，数据包的唯一标记，可以做具体业务处理，或者数据包去重。
	PackageSize  = 4
	HeaderSize   = 2
	VersionSize  = 2
	OperateSize  = 4
	SequenceSize = 4

	HeartSize      = 4
	RawHeaderSize  = PackageSize + HeaderSize + VersionSize + OperateSize + SequenceSize
	MaxPackageSize = RawHeaderSize + MaxBodySize

	PackIndex     = 0
	HeaderIndex   = PackIndex + PackageSize
	VersionIndex  = HeaderIndex + HeaderSize
	OperateIndex  = VersionIndex + VersionSize
	SequenceIndex = OperateIndex + SequenceSize

	HeartIndex = SequenceIndex + SequenceSize
)

func ReadTcp(p *protocol.Proto, reader *bufio.Reader) error {
	var (
		buf        []byte
		err        error
		packageLen uint32
		headerLen  uint16
		bodyLen    uint32
	)

	if buf, err = reader.Peek(RawHeaderSize); err != nil {
		return err
	}

	packageLen = binary.BigEndian.Uint32(buf[PackIndex:HeaderIndex])
	headerLen = binary.BigEndian.Uint16(buf[HeaderIndex:VersionIndex])
	p.Ver = int32(binary.BigEndian.Uint16(buf[VersionIndex:OperateIndex]))
	p.Op = int32(binary.BigEndian.Uint32(buf[OperateIndex:SequenceIndex]))
	p.Seq = int32(binary.BigEndian.Uint32(buf[SequenceIndex:RawHeaderSize]))

	if packageLen > uint32(MaxBodySize) {
		return errors.New("package length error")
	}
	if headerLen != RawHeaderSize {
		return errors.New("header length error")
	}
	if bodyLen = packageLen - RawHeaderSize; bodyLen > 0 {
		pack := make([]byte, int(RawHeaderSize+bodyLen))
		_, err = reader.Read(pack)
		if err != nil {
			return err
		}
		p.Body = pack[RawHeaderSize:]
	} else {
		p.Body = nil
	}

	return err
}

func WriteTcp(p *protocol.Proto, writer *bufio.Writer) error {
	var err error
	packageSize := len(p.Body) + RawHeaderSize
	tmp := make([]byte, packageSize)
	// 封装头信息
	binary.BigEndian.PutUint32(tmp[PackIndex:HeaderIndex], uint32(packageSize))
	binary.BigEndian.PutUint16(tmp[HeaderIndex:VersionIndex], uint16(RawHeaderSize))
	binary.BigEndian.PutUint16(tmp[VersionIndex:OperateIndex], uint16(p.Ver))
	binary.BigEndian.PutUint32(tmp[OperateIndex:SequenceIndex], uint32(p.Op))
	binary.BigEndian.PutUint32(tmp[SequenceIndex:RawHeaderSize], uint32(p.Seq))
	copy(tmp[RawHeaderSize:], p.Body)

	//写入bufio中
	if _, err = writer.Write(tmp); err != nil {
		return err
	}
	//bufio有缓存，flush真正写数据
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

// WriteTCPHeart write TCP heartbeat with room online.
func WriteTCPHeart(p *protocol.Proto, wr *bufio.Writer, online int32) error {
	var err error
	packageSize := len(p.Body) + RawHeaderSize
	tmp := make([]byte, packageSize)
	// 封装头信息
	binary.BigEndian.PutUint32(tmp[PackIndex:HeaderIndex], uint32(packageSize))
	binary.BigEndian.PutUint16(tmp[HeaderIndex:VersionIndex], uint16(RawHeaderSize))
	binary.BigEndian.PutUint16(tmp[VersionIndex:OperateIndex], uint16(p.Ver))
	binary.BigEndian.PutUint32(tmp[OperateIndex:SequenceIndex], uint32(p.Op))
	binary.BigEndian.PutUint32(tmp[SequenceIndex:RawHeaderSize], uint32(p.Seq))
	p.Body = []byte(strconv.Itoa(int(online)))
	copy(tmp[RawHeaderSize:], p.Body)
	if _, err = wr.Write(tmp); err != nil {
		return err
	}
	return nil
}
