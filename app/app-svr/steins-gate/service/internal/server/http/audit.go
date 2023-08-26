package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func graphAudit(c *bm.Context) {
	v := new(model.AuditParam)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.State == model.GraphStateRepulse && (v.RejectReason == "") {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.State != model.GraphStateRepulse && v.State != model.GraphStatePass {
		log.Error("Invalid State param %+v", v)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, svc.GraphAudit(c, v))

}
