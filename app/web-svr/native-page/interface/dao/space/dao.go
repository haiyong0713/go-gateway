package space

import (
	"context"

	spaceGRPC "git.bilibili.co/bapis/bapis-go/space/service"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

type Dao struct {
	spaceClient spaceGRPC.SpaceClient
}

func New(c *conf.Config) *Dao {
	spaceClient, err := spaceGRPC.NewClient(c.SpaceClient)
	if err != nil {
		panic(err)
	}
	return &Dao{spaceClient: spaceClient}
}

func (d *Dao) UserTab(c context.Context, mid int64) (*spaceGRPC.UserTabReply, error) {
	rly, err := d.spaceClient.UserTab(c, &spaceGRPC.UserTabReq{Mid: mid})
	if err != nil {
		log.Errorc(c, "Fail to get userTab, mid=%+v error=%+v", mid, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) UpActivityTab(c context.Context, mid int64, state int32, title string, pageID int64) (bool, error) {
	req := &spaceGRPC.UpActivityTabReq{
		Mid:     mid,
		State:   state,
		TabCont: pageID,
		TabName: title,
	}
	rly, err := d.spaceClient.UpActivityTab(c, req)
	if err != nil {
		log.Errorc(c, "Fail to get upActivityTab, req=%+v error=%+v", req, err)
		return false, err
	}
	return rly.GetSuccess(), nil
}
