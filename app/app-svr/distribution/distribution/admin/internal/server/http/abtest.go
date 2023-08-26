package http

import (
	"encoding/json"
	"io/ioutil"

	abm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/abtest"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

func ABTestList(c *bm.Context) {
	c.JSON(svc.BatchFetchABTestInfo(c))
}

func ABTestDetail(c *bm.Context) {
	params := &abm.DetailReq{}
	if err := c.Bind(params); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.FetchABTestConfigDetail(c, params))
}

func ABTestSave(c *bm.Context) {
	body, err := c.Request.GetBody()
	defer func() {
		_ = body.Close()
	}()
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	params := &struct {
		Details []*abm.Detail `json:"details"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if len(params.Details) == 0 {
		c.JSON(nil, errors.Wrapf(ecode.RequestErr, "No Details"))
		return
	}
	c.JSON(nil, svc.SaveABTestConfigs(c, params.Details))
}
