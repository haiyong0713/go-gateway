package native

import (
	"context"

	spaceGRPC "git.bilibili.co/bapis/bapis-go/space/service"
	"go-common/library/log"
)

func (d *Dao) UpActivityTab(c context.Context, mid int64, state int32, title string, pageID int64) (bool, error) {
	req := &spaceGRPC.UpActivityTabReq{
		Mid:     mid,
		State:   state,
		TabCont: pageID,
		TabName: title,
	}
	rly, err := d.spaceGRPC.UpActivityTab(c, req)
	if err != nil {
		log.Errorc(c, "Fail to handle upActivityTab, req=%+v error=%+v", req, err)
		return false, err
	}
	return rly.GetSuccess(), nil
}
