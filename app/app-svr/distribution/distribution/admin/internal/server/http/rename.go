package http

import (
	"encoding/json"
	"io/ioutil"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model/rename"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

func Rename(c *bm.Context) {
	bs, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	_ = c.Request.Body.Close()
	params := &rename.Rename{}
	if err := json.Unmarshal(bs, params); err != nil {
		c.JSON(nil, err)
		return
	}
	if params.ID == "" {
		c.JSON(nil, errors.Wrap(ecode.RequestErr, "illegal rename id"))
		return
	}
	c.JSON(nil, svc.Rename(c, params))
}
