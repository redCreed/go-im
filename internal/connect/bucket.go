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

// roomproc
func (b *Bucket) roomProc(c chan *connect.BroadcastRoomReq) {
	//for {
	//	arg := <-c
	//	if room := b.Room(arg.RoomID); room != nil {
	//		room.Push(arg.Proto)
	//	}
	//}
}
