package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func rankDataInterventionSave(c *bm.Context) {
	p := new(model.S10RankDataInterventionReq)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(esSvc.RankDataInterventionSave(c, p))
}

func rankDataInterventionGet(c *bm.Context) {
	c.JSON(esSvc.RankDataInterventionGet(c))
}
