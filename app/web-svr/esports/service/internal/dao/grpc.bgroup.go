package dao

import (
	"context"
	"fmt"
	"strconv"

	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) InitBGroup(ctx context.Context, contest *model.ContestModel) (err error) {
	// 创建人群包 https://info.bilibili.co/pages/viewpage.action?pageId=184996626#id-%E4%BA%BA%E7%BE%A4%E5%8C%85service%E6%9C%8D%E5%8A%A1%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3-%E4%BA%BA%E7%BE%A4%E5%8C%85%E5%88%9B%E5%BB%BA
	req := &bGroup.AddBGroupReq{
		Type:       3,
		Name:       strconv.FormatInt(contest.ID, 10),
		AppName:    "pink",
		Business:   d.conf.TunnelBGroup.NewBusiness,
		Creator:    d.conf.TunnelBGroup.NewBusiness,
		Definition: "{\"oid\":" + strconv.FormatInt(contest.ID, 10) + "}",
		Dimension:  1,
	}
	_, err = d.bGroupClient.AddBGroup(ctx, req)
	if xecode.Cause(err).Code() == model.BGroupExits { //人群包已经存在
		err = nil
		return
	}
	if err != nil {
		log.Errorc(ctx, "InitBGroup d.bGroupClient.AddBGroup() contestID(%d) error(%+v)", contest.ID, err)
		err = fmt.Errorf("创建人群包出错(%+v)", err)
		return
	}
	return
}
