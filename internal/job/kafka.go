package job

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	pb "go-im/api/logic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"time"
)

type Kafka struct {
	brokers           []string
	topics            []string
	startOffset       int64
	version           string
	ready             chan bool
	group             string
	channelBufferSize int
	assignor          string
	s                 *Server
}

func NewKafka(s *Server) *Kafka {
	return &Kafka{s: s}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (k *Kafka) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (k *Kafka) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (k *Kafka) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// 具体消费消息
	for message := range claim.Messages() {
		//log.Infof("[topic:%s] [partiton:%d] [offset:%d] [value:%s] [time:%v]",
		//	message.Topic, message.Partition, message.Offset, string(message.Value), message.Timestamp)
		pushMsg := new(pb.PushMsg)
		if err := proto.Unmarshal(message.Value, pushMsg); err != nil {
			k.s.log.Error(fmt.Sprintf("proto.Unmarshal(%v)  ", message), zap.Error(err))
			continue
		}
		if err := k.s.push(context.Background(), pushMsg); err != nil {
			k.s.log.Error("", zap.Error(err))
		}
		//手动更新位移同时开启幂等
		session.MarkMessage(message, "")
	}
	return nil
}

func newKafkaConfig() *sarama.Config {
	config := sarama.NewConfig()
	//config.ClientID = "sarama_demo" //
	//config.Version = sarama.V0_11_0_1                // kafka server的版本号
	config.Producer.Return.Successes = true                       // sync必须设置这个
	config.Producer.RequiredAcks = sarama.WaitForLocal            // WaitForAll也就是等待foolower同步，才会返回
	config.Producer.Return.Errors = true                          //投递失败返回错误
	config.Producer.Partitioner = sarama.NewRoundRobinPartitioner //轮询分区投放
	//config.Producer.Retry.Max=1  //默认重试3次
	//config.Producer.Retry.Backoff=1  //默认重试间隔100ms
	config.Producer.Compression = sarama.CompressionZSTD              //保证生产与消费的压缩类型一致
	config.Producer.CompressionLevel = sarama.CompressionLevelDefault //压缩类型

	config.ChannelBufferSize = 1024 * 1024
	config.Metadata.Full = false // 不用拉取全部的信息

	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false                                                       // 自动提交偏移量，默认开启
	config.Consumer.Offsets.AutoCommit.Interval = time.Second                                               // 这个看业务需求，commit提交频率，不然容易down机后造成重复消费。默认1s
	config.Consumer.Offsets.Initial = sarama.OffsetOldest                                                   // 从最开始的地方消费，业务中看有没有需求，新业务重跑topic。
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange} // rb策略，默认就是range
	// 分区分配策略
	//switch assignor {
	//case "sticky":
	//	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	//case "roundrobin":
	//	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	//case "range":
	//	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	//default:
	//	log.Panicf("Unrecognized consumer group partition assignor: %s", assignor)
	//}
	return config
}
