package connect

import (
	"github.com/pkg/errors"
	"sync"
)

type Room struct {
	Id          string
	lock        sync.RWMutex
	next        *Channel //channel链表中的头,新加入的channel统一放头部
	drop        bool     // make room is live  true表示房间关闭 false表示房间直播中
	Online      int
	OnlineCount int // room online user count
}

func NewRoom(roomId string) *Room {
	return &Room{
		Id:     roomId,
		next:   nil,
		drop:   false,
		Online: 0,
	}
}

func (r *Room) Put(ch *Channel) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if !r.drop {
		//新的ch会在链表头部
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		r.next = ch
		ch.Prev = nil
		r.Online++
	} else {
		return errors.Errorf("room drop id:%s", r.Id)
	}

	return nil
}
func (r *Room) Del(ch *Channel) bool {
	r.lock.Lock()
	if ch.Next != nil {
		// if not footer
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// if not header
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	ch.Next = nil
	ch.Prev = nil
	r.Online--
	//online等于0 drop为true
	r.drop = r.Online == 0
	r.lock.Unlock()
	return r.drop
}

func (r *Room) Close() {
	r.lock.RLock()
	defer r.lock.RUnlock()
	//变量链表 发送停止信号
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
}

// OnlineNum the room all online.
func (r *Room) OnlineNum() int32 {
	if r.OnlineCount > 0 {
		return int32(r.OnlineCount)
	}
	return int32(r.Online)
}
