package conf

var Conf *Config

// New 解析配置
func New(path string) {

}

type Config struct {
	Bucket   *Bucket
	Tcp      *TCP
	Mode     *Mode
	Protocol *Protocol
}

type Bucket struct {
	Size          int //bucket数量
	Channel       int //每个房间的channel
	Room          int //每个bucket中的房价数量
	RoutineAmount int //广播房间的channel数量
	RoutineSize   int //每个channel长度
}

// TCP is tcp config.
type TCP struct {
	Host         []string
	SendBuf      int
	ReceiveBuf   int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

type Mode struct {
	Debug bool
}

type Protocol struct {
	ProtoSize int
}
