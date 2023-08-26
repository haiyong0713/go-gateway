package dao

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	spaceGRPC "git.bilibili.co/bapis/bapis-go/space/service"
)

type spaceCfg struct {
	Client *warden.ClientConfig
}

type spaceDao struct {
	client spaceGRPC.SpaceClient
}

func newSpaceDao(cfg *spaceCfg) *spaceDao {
	d := &spaceDao{}
	var err error
	if d.client, err = spaceGRPC.NewClient(cfg.Client); err != nil {
		panic(fmt.Sprintf("Fail to new activityClient, config=%+v error=%+v", cfg.Client, err))
	}
	return d
}

func (d *spaceDao) UpActivityTab(c context.Context, mid int64, state int32, title string, pageID int64) (bool, error) {
	req := &spaceGRPC.UpActivityTabReq{
		Mid:     mid,
		State:   state,
		TabCont: pageID,
		TabName: title,
	}
	rly, err := d.client.UpActivityTab(c, req)
	if err != nil {
		log.Errorc(c, "Fail to handle upActivityTab, req=%+v error=%+v", req, err)
		return false, err
	}
	return rly.GetSuccess(), nil
}
