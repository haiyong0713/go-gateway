package service

import (
	"context"

	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
	acpAPI "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	"go-common/library/sync/pipeline/fanout"
	arcclient "go-gateway/app/app-svr/archive/service/api"
	actclient "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/dao"
	espclient "go-gateway/app/web-svr/esports/interface/api/v1"

	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"
	tunnelapi "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

// Service biz service def.
type Service struct {
	c            *conf.Config
	dao          *dao.Dao
	arcClient    arcclient.ArchiveClient
	accClient    accwarden.AccountClient
	actClient    actclient.ActivityClient
	ACPClient    acpAPI.AccountControlPlaneClient
	espClient    espclient.EsportsClient
	tunnelClient tunnelapi.TunnelClient
	cache        *fanout.Fanout
}

const (
	_notDeleted    = 0
	_deleted       = 1
	_checkPass     = 4
	_checkNopass   = 3
	_online        = 1
	_downLine      = 0
	_statusOn      = 0
	_statusAll     = -1
	_eventAlready  = 108009
	_noAddEvent    = 108007
	_cardNotExists = 108019
	_bgroupExits   = 145202
	_cardStatusErr = 108014
)

var (
	liveRoomClient liveRoom.RoomClient
)

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   dao.New(c),
		cache: fanout.New("cache"),
	}
	var err error
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.accClient, err = accwarden.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.actClient, err = actclient.NewClient(c.ActClient); err != nil {
		panic(err)
	}
	if s.ACPClient, err = acpAPI.NewClient(s.c.ACPRPC); err != nil {
		panic(err)
	}
	if s.espClient, err = espclient.NewClient(s.c.EspClient); err != nil {
		panic(err)
	}
	if liveRoomClient, err = liveRoom.NewClient(c.RoomGRPC); err != nil {
		panic(err)
	}
	if s.tunnelClient, err = tunnelapi.NewClient(c.TunnelClient); err != nil {
		panic(err)
	}
	return s
}

// Ping .
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

func unique(ids []int64) (outs []int64) {
	idMap := make(map[int64]int64, len(ids))
	for _, v := range ids {
		if _, ok := idMap[v]; ok {
			continue
		} else {
			idMap[v] = v
		}
		outs = append(outs, v)
	}
	return
}
