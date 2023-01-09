package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gomodule/redigo/redis"
	pb "go-im/api/logic"
	model "go-im/internal/logic/dto"
	"go-im/pkg/cityhash"
	"google.golang.org/protobuf/proto"
	"log"
	"strconv"
)

const (
	_prefixMidServer    = "mid_%d" // mid -> key:server
	_prefixKeyServer    = "key_%s" // key -> server
	_prefixServerOnline = "ol_%s"  // server -> online
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

// AddMapping add a mapping.
// Mapping:
//	mid -> key_server
//	key -> server
func (d *Dao) AddMapping(c context.Context, mid int64, key, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if _, err = conn.Do("HSET", keyMidServer(mid), key, server); err != nil {
		log.Fatalf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	if _, err = conn.Do("EXPIRE", keyMidServer(mid), d.redisExpire); err != nil {
		log.Fatalf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if _, err = conn.Do("SET", keyKeyServer(key), server); err != nil {
		log.Fatalf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	if _, err = conn.Do("EXPIRE", keyKeyServer(key), d.redisExpire); err != nil {
		log.Fatalf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	return
}

// DelMapping del a mapping.
func (d *Dao) DelMapping(c context.Context, mid int64, key, server string) (has bool, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	if _, err = conn.Do("HDEL", keyMidServer(mid), key); err != nil {
		log.Fatalf("conn.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}

	if _, err = conn.Do("DEL", keyKeyServer(key)); err != nil {
		log.Fatalf("conn.Send(HDEL %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}

	return
}

// AddServerOnline add a server online.
func (d *Dao) AddServerOnline(c context.Context, server string, online *model.Online) (err error) {
	roomsMap := map[uint32]map[string]int32{}
	for room, count := range online.RoomCount {
		rMap := roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64]
		if rMap == nil {
			rMap = make(map[string]int32)
			roomsMap[cityhash.CityHash32([]byte(room), uint32(len(room)))%64] = rMap
		}
		rMap[room] = count
	}
	key := keyServerOnline(server)
	//map[13:map[xxtttdig:15] 19:map[ss3indig:23] 27:map[3idd2ndig:2 xx3indig:22]]
	for hashKey, value := range roomsMap {
		err = d.addServerOnline(c, key, strconv.FormatInt(int64(hashKey), 10), &model.Online{RoomCount: value, Server: online.Server, Updated: online.Updated})
		if err != nil {
			return
		}
	}
	return
}

func (d *Dao) addServerOnline(c context.Context, key string, hashKey string, online *model.Online) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, _ := json.Marshal(online)
	//redis管道 可以通过send方法一次性向服务器发送一个命令或者多个命令，命令发送完毕后
	//用flush方法将缓冲区的命令一次性的发送到服务端，客户端用receive的方法读取命令的结果
	if err = conn.Send("HSET", key, hashKey, b); err != nil {
		log.Fatalf("conn.Send(SET %s,%s) error(%v)", key, hashKey, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisExpire); err != nil {
		log.Fatalf("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Fatalf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Fatalf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// ServerOnline get a server online.
func (d *Dao) ServerOnline(c context.Context, server string) (online *model.Online, err error) {
	online = &model.Online{RoomCount: map[string]int32{}}
	key := keyServerOnline(server)
	for i := 0; i < 64; i++ {
		ol, err := d.serverOnline(c, key, strconv.FormatInt(int64(i), 10))
		if err == nil && ol != nil {
			online.Server = ol.Server
			if ol.Updated > online.Updated {
				online.Updated = ol.Updated
			}
			for room, count := range ol.RoomCount {
				online.RoomCount[room] = count
			}
		}
	}
	return
}

func (d *Dao) serverOnline(c context.Context, key string, hashKey string) (online *model.Online, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("HGET", key, hashKey))
	if err != nil {
		if err != redis.ErrNil {
			d.log.Error(fmt.Sprintf("conn.Do(HGET %s %s) error(%v)", key, hashKey, err))
		}
		return
	}
	online = new(model.Online)
	if err = json.Unmarshal(b, online); err != nil {
		d.log.Error(fmt.Sprintf("serverOnline json.Unmarshal(%s) error(%v)", b, err))
		return
	}
	return
}

// DelServerOnline del a server online.
func (d *Dao) DelServerOnline(c context.Context, server string) (err error) {
	conn := d.redis.Get()
	defer conn.Close()
	key := keyServerOnline(server)
	if _, err = conn.Do("DEL", key); err != nil {
		d.log.Error(fmt.Sprintf("conn.Do(DEL %s) error(%v)", key, err))
	}
	return
}

// ServersByKeys get a server by key.
func (d *Dao) ServersByKeys(c context.Context, keys []string) (res []string, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	var args []interface{}
	for _, key := range keys {
		args = append(args, keyKeyServer(key))
	}
	if res, err = redis.Strings(conn.Do("MGET", args...)); err != nil {
		log.Fatalf("conn.Do(MGET %v) error(%v)", args, err)
	}
	return
}

// PushMsg push a message to databus.
func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(keys[0]), //todo
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Fatalf("PushMsg.send(push pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// KeysByMids get a key server by mid.
func (d *Dao) KeysByMids(c context.Context, mids []int64) (ress map[string]string, olMids []int64, err error) {
	conn := d.redis.Get()
	defer conn.Close()
	ress = make(map[string]string)
	for _, mid := range mids { //mid key server
		if err = conn.Send("HGETALL", keyMidServer(mid)); err != nil {
			log.Fatalf("conn.Do(HGETALL %d) error(%v)", mid, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Fatalf("conn.Flush() error(%v)", err)
		return
	}
	for idx := 0; idx < len(mids); idx++ {
		var (
			res map[string]string
		)
		if res, err = redis.StringMap(conn.Receive()); err != nil {
			log.Fatalf("conn.Receive() error(%v)", err)
			return
		}
		if len(res) > 0 {
			olMids = append(olMids, mids[idx])
		}
		for k, v := range res {
			ress[k] = v
		}
	}
	return
}

func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, room string, msg []byte) error {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Room:      room,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}

	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(room),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Fatalf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return nil
}

func (d *Dao) BroadcastMsg(c context.Context, op, speed int32, msg []byte) error {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Speed:     speed,
		Msg:       msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}

	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(op), 10)),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Fatalf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return nil
}
