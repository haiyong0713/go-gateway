package dm

import (
	"context"
	"fmt"

	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	"go-gateway/app/app-svr/playurl/service/conf"
)

type Dao struct {
	dmClient dmgrpc.DMClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.dmClient, err = dmgrpc.NewClient(c.DmClient)
	if err != nil {
		panic(fmt.Sprintf("dmgrpc NewClient error(%v)", err))
	}
	return
}

// SubtitleExist .
func (d *Dao) SubtitleExist(c context.Context, cid int64) (*dmgrpc.SubtitleExistReply, error) {
	return d.dmClient.SubtitleExist(c, &dmgrpc.SubtitleExistReq{Type: 1, Oid: cid})
}
