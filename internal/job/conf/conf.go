package conf

import (
	"github.com/spf13/viper"
	"strings"
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
	v.SetConfigName("job")
	v.SetConfigType("yaml")
	v.AddConfigPath(p)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := v.Unmarshal(Conf); err != nil {
		panic(err)
	}
}

type Config struct {
	Mode      *Mode
	Discovery *Discovery
	Kafka     *Kafka
}

type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

type Discovery struct {
	driver  string
	host    string
	timeout int
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
