package http

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

func dynamicNew(c *bm.Context) {
	// params := c.Request.Form
	// data, err := externalSvc.DynamicNew(c, params.Encode())
	// dynamicResult(c, data, err)
	dynamicResult(c, nil, ecode.NothingFound)
}

func dynamicCount(c *bm.Context) {
	params := c.Request.Form
	data, err := externalSvc.DynamicCount(c, params.Encode())
	dynamicResult(c, data, err)
}

func dynamicHistory(c *bm.Context) {
	// params := c.Request.Form
	// data, err := externalSvc.DynamicHistory(c, params.Encode())
	// dynamicResult(c, data, err)
	dynamicResult(c, nil, ecode.NothingFound)
}

func dynamicResult(c *bm.Context, data json.RawMessage, err error) {
	if err != nil {
		code := ecode.Int(-22)
		c.JSON(nil, code)
	} else {
		if !bytes.Contains(data, []byte(`"code":0`)) {
			var res struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(data, &res); err != nil {
				log.Error("Failed to unmarshal: %+v", errors.WithStack(err))
			}
		}
		c.Bytes(http.StatusOK, "text/json; charset=utf-8", data)
	}
}
