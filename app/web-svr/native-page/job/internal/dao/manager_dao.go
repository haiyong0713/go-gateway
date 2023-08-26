package dao

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
)

const (
	uriTsOnline     = "/x/admin/native_page/native/ts/online"
	spaceOfflineURI = "/x/admin/native_page/native/ts/space/offline"
)

type managerCfg struct {
	HTTPClient *httpx.ClientConfig
	Host       string
}

type managerDao struct {
	http            *httpx.Client
	tsOnlineURL     string
	spaceOfflineURL string
}

func newManagerDao(cfg *managerCfg) *managerDao {
	return &managerDao{
		http:            httpx.NewClient(cfg.HTTPClient),
		tsOnlineURL:     cfg.Host + uriTsOnline,
		spaceOfflineURL: cfg.Host + spaceOfflineURI,
	}
}

func (d *managerDao) TsOnline(c context.Context, tsID, pid, auditTime int64) error {
	params := url.Values{}
	params.Set("oid", strconv.FormatInt(tsID, 10))
	params.Set("pid", strconv.FormatInt(pid, 10))
	params.Set("audit_time", strconv.FormatInt(auditTime, 10))
	var rly struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.http.Post(c, d.tsOnlineURL, "", params, &rly); err != nil {
		log.Errorc(c, "Fail to post tsOnline request, params=%+v error=%+v", params, err)
		return err
	}
	if rly.Code != ecode.OK.Code() {
		log.Errorc(c, "Fail to online upActivity, params=%+v rly=%+v", params, rly)
		return errors.Wrap(ecode.Int(rly.Code), d.tsOnlineURL+" msg="+rly.Msg)
	}
	return nil
}

func (d *managerDao) SpaceOffline(c context.Context, mid, pageID int64, tabType string) error {
	var err error
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	params.Set("tab_type", tabType)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.http.Post(c, d.spaceOfflineURL, "", params, &res); err != nil {
		log.Error("Fail to offline space, url=%+v error=%+v", d.spaceOfflineURL+"?"+params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
	}
	return err
}
