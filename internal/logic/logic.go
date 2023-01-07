package logic

import (
	"context"
	"go-im/internal/logic/conf"
	"go-im/internal/logic/dao"
	model "go-im/internal/logic/dto"
	"log"
	"time"
)

const (
	_onlineTick     = time.Second * 10
	_onlineDeadline = time.Minute * 5
)

type Logic struct {
	c *conf.Config
	// online
	totalIPs   int64
	totalConns int64
	roomCount  map[string]int32
	regions    map[string]string // province -> region
	dao        *dao.Dao
	HostName   string
}

func New(c *conf.Config) *Logic {
	s := new(Logic)
	s.c = c
	s.HostName = "tt"
	s.regions = make(map[string]string)
	s.initRegions()
	//todo etcd get all node info

	s.dao = dao.New(c)
	s.initRegions()

	_ = s.loadOnline()
	go s.onlineproc()
	return s
}

func (l *Logic) initRegions() {
	for region, ps := range l.c.Regions {
		for _, province := range ps {
			l.regions[province] = region
		}
	}
}
func (l *Logic) onlineproc() {
	for {
		time.Sleep(_onlineTick)
		if err := l.loadOnline(); err != nil {
			log.Fatalf("onlineproc error(%v)", err)
		}
	}
}

func (l *Logic) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	//for _, server := range s.nodes {
	var online *model.Online
	online, err = l.dao.ServerOnline(context.Background(), l.HostName)
	if err != nil {
		return
	}
	if time.Since(time.Unix(online.Updated, 0)) > _onlineDeadline {
		_ = l.dao.DelServerOnline(context.Background(), l.HostName)
		return
	}
	for roomID, count := range online.RoomCount {
		roomCount[roomID] += count
	}
	//}
	l.roomCount = roomCount
	return
}
