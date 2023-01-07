package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	model "go-im/internal/logic/dto"
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
