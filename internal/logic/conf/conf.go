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

// Config config.
type Config struct {
	Env        *Env
	Discovery  *Discovery
	RPCClient  *RPCClient
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Node       *Node
	Backoff    *Backoff
	Regions    map[string][]string
}

type Discovery struct {
	driver  string
	host    string
	timeout int
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
}

// Node node config.
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int
	HeartbeatMax  int
	Heartbeat     time.Duration
	RegionWeight  float64
}

type Backoff struct {
	MaxDelay  int32
	BaseDelay int32
	Factor    float32
	Jitter    float32
}

// Redis .
type Redis struct {
	Network      string
	Addr         string
	Auth         string
	Active       int
	Idle         int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Expire       time.Duration
}

// Kafka .
type Kafka struct {
	Topic   string
	Brokers []string
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    time.Duration
	Timeout time.Duration
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

// HTTPServer is http server config.
type HTTPServer struct {
	Network      string
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
