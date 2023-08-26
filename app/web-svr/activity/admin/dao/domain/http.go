package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	//"github.com/go-errors/errors"
	"github.com/pkg/errors"
	"go-common/library/log"
	mdomain "go-gateway/app/web-svr/activity/admin/model/domain"
)

type item struct {
	Description string `json:"description"`
	Group       string `json:"group"`
	Key         string `json:"key"`
	Value       string `json:"value"`
}

type Response struct {
	ErrCode int    `json:"code"`
	ErrMsg  string `json:"message"`
}

// GetDomainList ...
func (d *Dao) GetDomainList(ctx context.Context, pageNo, pageSize int) (records []mdomain.Record, err error) {
	var (
		params = url.Values{}
		res    struct {
			ErrCode int    `json:"code"`
			ErrMsg  string `json:"message"`
			Data    struct {
				PageNo   int              `json:"page_no"`
				PageSize int              `json:"page_size"`
				List     []mdomain.Record `json:"list"`
			} `json:"data"`
		}
	)
	params.Set("page_no", strconv.Itoa(pageNo))
	params.Set("page_size", strconv.Itoa(pageSize))

	if err = d.httpClient.Get(ctx, d.listUrl, "", params, &res); err != nil {
		err = errors.Errorf("GetDomainList d.httpClient.Get(%v) failed. error(%v)", d.listUrl, err)
		return
	}

	if res.ErrCode != 0 {
		err = errors.Errorf("GetDomainList: errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
		return
	}
	records = res.Data.List
	return
}

// GetFawkesConfig ...
func (d *Dao) GetFawkesConfig(ctx context.Context, appkey string) (records []mdomain.Record, err error) {
	var (
		params = url.Values{}
		res    struct {
			Response
			Data []*struct {
				AppKey      string `json:"app_key"`
				Env         string `json:"env"`
				Cvid        int    `json:"cvid"`
				Group       string `json:"group"`
				Key         string `json:"key"`
				Value       string `json:"value"`
				Type        int    `json:"type"`
				Operator    string `json:"operator"`
				Description string `json:"description"`
				Mtime       int64  `json:"mtime"`
			} `json:"data"`
		}
	)

	params.Set("app_key", appkey)
	params.Set("env", d.c.ActDomainConf.FawkesConf.Env)
	params.Set("business", d.c.ActDomainConf.FawkesConf.Business)

	if err = d.httpClient.Get(ctx, d.fawkesGetUrl, "", params, &res); err != nil {
		err = errors.Wrapf(err, "GetFawkesConfig d.httpClient.Get(%v) failed. error(%v)", d.fawkesGetUrl, err)
		return
	}

	if res.ErrCode != 0 {
		err = errors.Wrapf(err, "GetFawkesConfig errcode:%d, errmsg:%s", res.ErrCode, res.ErrMsg)
		return
	}
	for _, v := range res.Data {
		if v != nil && v.Key == d.c.ActDomainConf.FawkesConf.ItemKey {
			err = json.Unmarshal([]byte(v.Value), &records)
			if err != nil {
				log.Errorc(ctx, "GetFawkesConfig json_Unmarshal err:%v , %v", v.Value, err)
				return
			}
		}
	}
	return
}

// AddFawkesConfig ...
func (d *Dao) AddFawkesConfig(ctx context.Context, records []mdomain.Record, appKeys []string) (err error) {
	if appKeys == nil && len(appKeys) <= 0 {
		log.Errorc(ctx, "empty app_key:%v", appKeys)
		return
	}
	var (
		res Response
	)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	if err = jsonEncoder.Encode(records); err != nil {
		return
	}

	params := struct {
		AppKey      string `json:"app_key"`
		Env         string `json:"env"`
		Operator    string `json:"operator"`
		Description string `json:"description"`
		Business    string `json:"business"`
		Items       []item `json:"items"`
	}{
		AppKey:      strings.Join(appKeys, ","),
		Env:         d.c.ActDomainConf.FawkesConf.Env,
		Operator:    d.c.ActDomainConf.FawkesConf.Operator,
		Description: d.c.ActDomainConf.FawkesConf.Description,
		Business:    d.c.ActDomainConf.FawkesConf.Business,
		Items: []item{
			{
				Description: d.c.ActDomainConf.FawkesConf.ItemDescription +
					"version:" + strconv.FormatInt(time.Now().Unix(), 10),
				Group: d.c.ActDomainConf.FawkesConf.ItemGroupName,
				Key:   d.c.ActDomainConf.FawkesConf.ItemKey,
				Value: strings.TrimSpace(bf.String()),
			},
		},
	}
	buf, _ := json.Marshal(params)
	var req *http.Request
	log.Infoc(ctx, "SyncFawkes start : %v , %s", d.fawkesAddUrl, buf)
	if req, err = http.NewRequest(http.MethodPost, d.fawkesAddUrl, bytes.NewReader(buf)); err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.httpClient.Do(ctx, req, &res); err != nil {
		return errors.Wrapf(err, "SyncFawkes d.client.Do(%s) , err:%v", d.fawkesAddUrl, err)
	}

	if res.ErrCode != 0 {
		err = errors.Wrapf(err, "GetDomainList: errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
	}
	return
}
