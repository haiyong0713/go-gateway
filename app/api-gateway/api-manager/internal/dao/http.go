package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

func (d *dao) GWAddPath(c context.Context, pathInfo *model.DynpathParam, discovery string) (err error) {
	if pathInfo == nil {
		return
	}
	var info = &struct {
		AppID    string `json:"app_id"`
		Endpoint string `json:"endpoint"`
	}{
		AppID:    discovery,
		Endpoint: fmt.Sprintf("discovery://%s", discovery),
	}
	var bs []byte
	bs, _ = json.Marshal(info)
	params := url.Values{}
	params.Set("node", pathInfo.Node)
	params.Set("gateway", pathInfo.Gateway)
	params.Set("pattern", pathInfo.Pattern)
	params.Set("client_info", string(bs))
	params.Set("enable", strconv.Itoa(pathInfo.Enable))
	params.Set("client_timeout", strconv.Itoa(pathInfo.ClientTimeout))
	paramStr := params.Encode()
	var (
		buffer bytes.Buffer
		query  string
	)
	buffer.WriteString(paramStr)
	query = buffer.String()
	req, err := http.NewRequest("POST", "http://manager.bilibili.co/x/admin/app-gw/dynpath/add", strings.NewReader(query))
	if err != nil {
		log.Errorc(c, "http.NewRequest error:%+v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.httpCli.Do(c, req, &res); err != nil {
		log.Errorc(c, "d.httpCli.Do error:%+v", err)
		return
	}
	if res.Code != 0 {
		err = ecode.ServerErr
		log.Errorc(c, "res.Code error:%+v", res.Code)
	}
	return
}
