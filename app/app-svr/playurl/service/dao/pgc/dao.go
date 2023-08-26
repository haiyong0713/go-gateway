package pgc

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/playurl/service/conf"

	pgcApi "git.bilibili.co/bapis/bapis-go/pgc/service/player"
)

// Dao is ugcpay dao.
type Dao struct {
	// rpc
	pgcRPC pgcApi.PlayerClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.pgcRPC, err = pgcApi.NewClient(c.PGCPlayerClient)
	if err != nil {
		panic(fmt.Sprintf("pgcplayer NewClient error(%v)", err))
	}
	return
}

// AssetRelation is
func (d *Dao) PGCCanPlay(c context.Context, mid, cid int64, platform, device, mobiApp string) (canPlay bool, err error) {
	var (
		plat int32
		req  *pgcApi.VerifyPlayReq
		res  *pgcApi.VerifyPlayReply
		ip   = metadata.String(c, metadata.RemoteIP)
	)
	if platform != "pc" && platform != "html5" && mobiApp == "" {
		mobiApp = platform
		platform = ""
	}
	req = &pgcApi.VerifyPlayReq{
		User: &pgcApi.UserProto{
			Mid:      mid,
			Ip:       ip,
			MobiApp:  mobiApp,
			Platform: platform,
			Device:   device,
		},
		Cids: []uint32{uint32(cid)},
	}
	if res, err = d.pgcRPC.VerifyPlay(c, req); err != nil {
		log.Error("d.pgcRPC.VerifyPlay cid(%d) mid(%d) ip(%s) plat(%d) error(%+v)", cid, mid, ip, plat, err)
		return
	}
	if res != nil && res.CidMap != nil {
		if cplay, ok := res.CidMap[uint32(cid)]; ok {
			canPlay = cplay.AllowPlay
		}
	}
	return
}
