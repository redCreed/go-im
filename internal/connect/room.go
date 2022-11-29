package connect

import "sync"

type Room struct {
	Id     string
	lock   sync.Mutex
	next   *Channel
	drop   bool
	Online int
}

func NewRoom(roomId string) *Room {
	return &Room{
		Id:     roomId,
		next:   nil,
		drop:   false,
		Online: 0,
	}
}
