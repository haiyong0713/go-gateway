package http

import (
	"encoding/json"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	tusmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"
	"io/ioutil"

	"github.com/pkg/errors"
)

func MultipleTusFields(c *bm.Context) {
	c.JSON(svc.GetMultipleTusFields(c))
}

func MultipleTusDetail(c *bm.Context) {
	params := &struct {
		Field         string `form:"field_name" validate:"required"`
		ConfigVersion string `form:"config_version" default:"v1.0"`
	}{}
	if err := c.Bind(params); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.FetchMultipleTusDetail(c, params.Field, params.ConfigVersion))
}

func MultipleTusSave(c *bm.Context) {
	body, err := c.Request.GetBody()
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer func() {
		_ = body.Close()
	}()
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	params := &struct {
		FieldName     string          `json:"field_name"`
		Details       []*tusmm.Detail `json:"details"`
		ConfigVersion string          `json:"config_version"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if len(params.Details) == 0 {
		c.JSON(nil, errors.Wrapf(ecode.RequestErr, "No Details"))
		return
	}
	if params.ConfigVersion == "" {
		c.JSON(nil, errors.Wrapf(ecode.RequestErr, "No ConfigVersion"))
		return
	}
	c.JSON(nil, svc.SaveMultipleTusConfig(c, params.Details, params.FieldName, params.ConfigVersion))
}
