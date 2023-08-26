package http

import (
	"encoding/json"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	tusm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tus"
	"io/ioutil"

	"github.com/pkg/errors"
)

func TusList(c *bm.Context) {
	c.JSON(svc.BatchFetchTusInfos(c))
}

func TusDetail(c *bm.Context) {
	params := &struct {
		TusValue string `form:"tus_value" validate:"required"`
	}{}
	if err := c.Bind(params); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.FetchTusConfigDetail(c, params.TusValue))
}

func TusSave(c *bm.Context) {
	bs, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	_ = c.Request.Body.Close()
	params := &struct {
		Details []*tusm.Detail `json:"details"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if len(params.Details) == 0 {
		c.JSON(nil, errors.Wrapf(ecode.RequestErr, "No Details"))
		return
	}
	c.JSON(nil, svc.SaveTusConfigs(c, params.Details))
}
