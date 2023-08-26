package http

import (
	"encoding/json"
	"io/ioutil"

	vcm "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func tusMultipleVersionAdd(c *bm.Context) {
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
		FieldName  string            `json:"field_name"`
		BuildLimit []*vcm.BuildLimit `json:"build_limit"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(svc.AddVersion(c, params.FieldName, params.BuildLimit))
}

func tusMultipleUpdateBuildLimit(c *bm.Context) {
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
		FieldName   string           `json:"field_name"`
		VersionInfo *vcm.VersionInfo `json:"version_info"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if params.VersionInfo.ConfigVersion == "v1.0" {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "初始版本不要设置版本号"))
		return
	}
	c.JSON(nil, svc.UpdateBuildLimit(c, params.FieldName, params.VersionInfo))
}

func tusMultipleVersion(c *bm.Context) {
	params := struct {
		FieldName string `form:"field_name"`
	}{}
	if err := c.Bind(&params); err != nil {
		return
	}
	c.JSON(svc.FetchConfigVersionManagerByField(c, params.FieldName))
}

func tusMultipleVersionDelete(c *bm.Context) {
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
		FieldName     string `json:"field_name"`
		ConfigVersion string `json:"config_version"`
	}{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if params.ConfigVersion == vcm.FirstVersion {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "初始版本不能删除"))
		return
	}
	c.JSON(nil, svc.DeleteConfigVersion(c, params.FieldName, params.ConfigVersion))
}
