package job

import (
	"context"
	"github.com/Shopify/sarama"
	"go-im/internal/job/conf"
	"go-im/pkg/log"
	"go.uber.org/zap"
	"sync"
)

type Server struct {
	log     *log.Log
	lock    sync.RWMutex
	c       *conf.Config
	k       *Kafka
	connect map[string]*ConnectServer
}

func NewServer(c *conf.Config) *Server {
	s := new(Server)
	s.c = c
	s.log = log.NewLog("im", c.Mode.Debug)
	//todo connect

	s.k = NewKafka(s)
	return s
}

func (s *Server) Consume() {
	config := newKafkaConfig()

	// 创建client
	newClient, err := sarama.NewClient(s.c.Kafka.Brokers, config)
	if err != nil {
		panic(err)
	}

	// 根据client创建consumerGroup
	client, err := sarama.NewConsumerGroupFromClient(s.c.Kafka.Group, newClient)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case err = <-client.Errors():
				s.log.Error("read channel err", zap.Error(err))
			default:
				if err := client.Consume(context.Background(), []string{s.c.Kafka.Topic}, s.k); err != nil {
					s.log.Error("new consumer group err", zap.Error(err))
				}
			}
		}
	}()
}
