package http

import (
	"strconv"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/feature/service/api"

	bm "go-common/library/net/http/blademaster"
)

func buildLimit(c *bm.Context) {
	var (
		params = c.Request.Form
		treeID int64
		err    error
	)
	if treeID, err = strconv.ParseInt(params.Get("tree_id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(featureSvr.BuildLimit(c, &api.BuildLimitReq{TreeId: treeID}))
}
