package dao

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/log"
)

const (
	activeSeasonList    = "esports:active:season:list:mc"
	activeSeasonListTtl = 300
	seasonContestIds    = "esports:season:contestIds:mc:id:%d"
	seasonContestIdsTtl = 86400
	seasonTeamsList     = "esports:season:teams:mc:id:%d"
	seasonTeamsListTtl  = 600
)

func (d *dao) PingMC(ctx context.Context) (err error) {
	if err = d.mc.Set(ctx, &memcache.Item{Key: "ping", Value: []byte("pong"), Expiration: 0}); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}
