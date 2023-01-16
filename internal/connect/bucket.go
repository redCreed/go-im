package connect

import (
	"go-im/api/connect"
	pb "go-im/api/connect"
	"go-im/api/protocol"
	"go-im/internal/connect/conf"
	"sync"
	"sync/atomic"
)

type Bucket struct {
	cLock       sync.RWMutex
	rooms       map[string]*Room
	chs         map[string]*Channel
	routines    []chan *connect.BroadcastRoomReq
	routinesNum uint64
	ipCnts      map[string]int32
}

func NewBucket(bucket *conf.Bucket) (b *Bucket) {
	b = new(Bucket)
	b.ipCnts = make(map[string]int32)
	b.chs = make(map[string]*Channel, bucket.Channel)
	b.rooms = make(map[string]*Room, bucket.Room)
	b.routines = make([]chan *connect.BroadcastRoomReq, bucket.RoutineAmount)
	for i := 0; i < bucket.RoutineAmount; i++ {
		c := make(chan *connect.BroadcastRoomReq, bucket.RoutineSize)
		b.routines[i] = c
		go b.roomProc(c)
	}
	return
}

func (b *Bucket) Put(roomId string, ch *Channel) error {
	var (
		room *Room
		ok   bool
		err  error
	)

	b.cLock.Lock()
	defer b.cLock.Unlock()
	//close old channel
	if oldCh := b.chs[ch.Key]; oldCh != nil {
		oldCh.Close()
	}
	b.chs[ch.Key] = ch
	if roomId != "" {
		//房间不存在则创建新房间
		if room, ok = b.rooms[roomId]; !ok {
			room = NewRoom(roomId)
			b.rooms[roomId] = room
		}
		ch.Room = room
	}
	b.ipCnts[ch.IP]++
	if room != nil {
		err = room.Put(ch)
	}
	return err
}

//Del 删除bucket和room的channel的信息
func (b *Bucket) Del(ch *Channel) {
	room := ch.Room
	b.cLock.Lock()
	if oldCh, ok := b.chs[ch.Key]; ok {
		if oldCh == ch {
			delete(b.rooms, ch.Key)
		}

		//ip记录
		if b.ipCnts[ch.IP] > 1 {
			b.ipCnts[ch.IP]--
		} else {
			delete(b.ipCnts, ch.IP)
		}
	}
	b.cLock.Unlock()
	if room != nil && room.Del(ch) {
		// if empty room, must delete from bucket
		b.DelRoom(room)
	}
	return
}

// DelRoom delete a room by roomid.
func (b *Bucket) DelRoom(room *Room) {
	b.cLock.Lock()
	delete(b.rooms, room.Id)
	b.cLock.Unlock()
	room.Close()
}

func (b *Bucket) ChangeRoom(roomId string, ch *Channel) error {
	var (
		oldRoom = ch.Room
		room    *Room
		ok      bool
	)
	//离开房间,则roomId为空
	if roomId == "" {
		//最后一个离开房间的
		if oldRoom != nil && oldRoom.Del(ch) {
			b.DelRoom(oldRoom)
		}
		//当前所在的房间也为空
		ch.Room = nil
		return nil
	}
	b.cLock.Lock()
	//判断房间是否已经创建
	if room, ok = b.rooms[roomId]; !ok {
		room = NewRoom(roomId)
		b.rooms[roomId] = room
	}
	b.cLock.Unlock()
	if oldRoom != nil && oldRoom.Del(ch) {
		b.DelRoom(oldRoom)
	}
	if err := room.Put(ch); err != nil {
		return err
	}
	ch.Room = room
	return nil
}

// roomproc
func (b *Bucket) roomProc(c chan *connect.BroadcastRoomReq) {
	for {
		arg := <-c
		if room := b.Room(arg.RoomID); room != nil {
			room.Push(arg.Proto)
		}
	}
}

func (b *Bucket) Room(roomId string) *Room {
	b.cLock.RLock()
	defer b.cLock.RUnlock()
	if room, ok := b.rooms[roomId]; ok {
		return room
	}

	return nil
}

func (b *Bucket) RoomsCount() map[string]int32 {
	b.cLock.RLock()
	defer b.cLock.RUnlock()
	resp := make(map[string]int32)
	for roomId, v := range b.rooms {
		resp[roomId] = v.Online
	}

	return resp
}

func (b *Bucket) UpdateRoomCount(roomCount map[string]int32) {
	b.cLock.RLock()
	defer b.cLock.RUnlock()
	for roomId, room := range b.rooms {
		room.OnlineCount = roomCount[roomId]
	}
}

// Rooms get all room id where online number > 0.
func (b *Bucket) Rooms() (res map[string]struct{}) {
	var (
		roomID string
		room   *Room
	)
	res = make(map[string]struct{})
	b.cLock.RLock()
	for roomID, room = range b.rooms {
		if room.Online > 0 {
			res[roomID] = struct{}{}
		}
	}
	b.cLock.RUnlock()
	return
}

// BroadcastRoom broadcast a message to specified room
func (b *Bucket) BroadcastRoom(req *pb.BroadcastRoomReq) {
	num := atomic.AddUint64(&b.routinesNum, 1) % uint64(len(b.routines))
	b.routines[num] <- req
}

// Broadcast push msgs to all channels in the bucket.
func (b *Bucket) Broadcast(p *protocol.Proto, op int32) {
	var ch *Channel
	b.cLock.RLock()
	for _, ch = range b.chs {
		if !ch.NeedPush(op) {
			continue
		}
		_ = ch.Push(p)
	}
	b.cLock.RUnlock()
}

// NeedPush verify if in watch.
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, ok := c.watchOps[op]; ok {
		return true
	}
	return false
}

// ChannelCount channel count in the bucket
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// RoomCount room count in the bucket
func (b *Bucket) RoomCount() int {
	return len(b.rooms)
}

// Channel get a channel by sub key.
func (b *Bucket) Channel(key string) (ch *Channel) {
	b.cLock.RLock()
	ch = b.chs[key]
	b.cLock.RUnlock()
	return
}
