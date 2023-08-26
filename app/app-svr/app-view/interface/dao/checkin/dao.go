package checkin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/community/interface/checkin"
)

type Dao struct {
	checkClient api.CheckinClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.checkClient, err = api.NewClientCheckin(c.CheckinClient); err != nil {
		panic(fmt.Sprintf("NewClientCheckin not found err(%v)", err))
	}
	return
}

// 获取活动内容
func (d *Dao) CheckinActivity(ctx context.Context, req *api.ActivityReq) (*api.ActivityReply, error) {
	reply, err := d.checkClient.Activity(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "req:%+v", req)
	}
	return reply, nil
}

// 打卡活动
func (d *Dao) CheckinAddUserActivityRecord(ctx context.Context, req *api.AddUserActivityRecordReq) (*api.ActivityReply, error) {
	reply, err := d.checkClient.AddUserActivityRecord(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "req:%+v", req)
	}
	return reply, nil
}

// 组件曝光
func (d *Dao) CheckinWidgetExpose(ctx context.Context, req *api.WidgetExposeReq) error {
	_, err := d.checkClient.WidgetExpose(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "req:%+v", req)
	}
	return nil
}
