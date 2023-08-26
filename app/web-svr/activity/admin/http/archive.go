package http

import (
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func archives(c *bm.Context) {
	p := &model.ArchiveParam{}
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Bvids != "" {
		var aids []int64
		bvids := strings.Split(p.Bvids, ",")
		if len(bvids) == 0 || len(bvids) > 30 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		for _, bvid := range bvids {
			aid, err := bvArgCheck(0, bvid)
			if err != nil {
				c.JSON(nil, err)
				return
			}
			aids = append(aids, aid)
		}
		p.Aids = aids
	}
	if len(p.Aids) == 0 || len(p.Aids) > 30 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(actSrv.Archives(c, p.Aids))
}

func accounts(c *bm.Context) {
	p := new(struct {
		Mids []int64 `form:"mids,split" validate:"min=1,max=30,dive,gt=0"`
	})
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(actSrv.Accounts(c, p.Mids))
}
