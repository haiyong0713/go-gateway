package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/dynamic/service/conf"
)

// init region archive redis .
func initRegionArc(c *bm.Context) {
	var (
		err error
		v   = new(struct {
			Rid  int32  `form:"rid"`
			Ower string `form:"ower"`
		})
	)
	if err = c.Bind(v); err != nil {
		return
	}
	if !_judgePerm(v.Ower) {
		c.JSON(nil, ecode.AccessDenied)
		return
	}
	c.JSON(nil, dySvc.RegionArcInit(c, v.Rid))
}

func _judgePerm(ow string) (f bool) {
	var sc = conf.Conf.Rule.PermInit
	for _, v := range sc {
		if v == ow {
			f = true
			return
		}
	}
	return
}
