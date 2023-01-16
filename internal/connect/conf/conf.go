package conf

import (
	"github.com/spf13/viper"
	"strings"
	"time"
)

var Conf *Config

// Parse 解析配置
func Parse(path string) {
	if path == "" {
		panic("cannot find config file ")
	}
	//读取指定配置文件
	Conf = new(Config)
	v := viper.New()
	ps := strings.Split(path, "/")
	//文件路径前缀
	p := strings.Join(ps[:len(ps)-1], "/")
	v.SetConfigName("connect")
	v.SetConfigType("yaml")
	v.AddConfigPath(p)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(Conf); err != nil {
		panic(err)
	}

	Conf.RPCServer = &RPCServer{
		Network:           "tcp",
		Addr:              ":3109",
		Timeout:           time.Second * 3,
		IdleTimeout:       time.Second * 60,
		MaxLifeTime:       time.Hour * 2,
		ForceCloseWait:    time.Second * 20,
		KeepAliveInterval: time.Second * 60,
		KeepAliveTimeout:  time.Second * 20,
	}
}

type Config struct {
	Discovery *Discovery
	Bucket    *Bucket
	Tcp       *TCP
	Mode      *Mode
	Protocol  *Protocol
	RPCServer *RPCServer
	Websocket *Websocket
}

type Discovery struct {
	Driver string
	Host   string
	Lease  int
}

type Websocket struct {
	Host        []string
	TlsOpen     bool
	CertFile    string
	PrivateFile string
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

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           time.Duration
	IdleTimeout       time.Duration
	MaxLifeTime       time.Duration
	ForceCloseWait    time.Duration
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
}
