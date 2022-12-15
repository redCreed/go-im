package connect

import (
	"go-im/api/connect"
	"go-im/internal/connect/conf"
	"sync"
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
	//for {
	//	arg := <-c
	//	if room := b.Room(arg.RoomID); room != nil {
	//		room.Push(arg.Proto)
	//	}
	//}
}
