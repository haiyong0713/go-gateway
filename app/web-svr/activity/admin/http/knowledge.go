package http

import (
	"fmt"

	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

const _maxRecords = 500

func historyUpdate(c *bm.Context) {
	param := new(model.ParamKnowledge)
	err := c.Bind(param)
	if err != nil {
		return
	}
	if len(param.UpdateMids) > _maxRecords {
		err = xecode.Errorf(xecode.RequestErr, fmt.Sprintf("不能超过%d个用户", _maxRecords))
		c.JSON(nil, err)
	}
	c.JSON(nil, actSrv.UpdateKnowledgeHistoryBatch(c, param))
}
