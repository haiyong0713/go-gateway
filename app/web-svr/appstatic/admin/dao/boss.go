package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const (
	_preHeat      = "/api/cache/preload"
	_preHeatQuery = "/api/cache/query_progress"
)

// UploadBoss .
func (d *Dao) UploadBoss(c context.Context, path string, payload []byte) (res *s3.PutObjectOutput, err error) {
	return d.boss.PutObject(c, model.BossBucket, path, payload)
}

// CdnDoPreload .
func (d *Dao) CdnDoPreload(c context.Context, reqParam []string) error {
	var msgBytes []byte
	params := map[string]interface{}{
		"Urls":   reqParam,
		"Action": "refresh",
	}
	msgBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}
	url := d.host.Cdn + _preHeat
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(msgBytes)))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	res := &struct {
		ErrCode int    `json:"ErrCode"`
		ErrMsg  string `json:"ErrMsg"`
	}{}
	err = d.client.Do(c, req, &res)
	if err != nil {
		return errors.Wrapf(err, "CdnDoPreload req(%v) res(%v)", reqParam, res)
	}
	if res.ErrCode != ecode.OK.Code() {
		return errors.Wrapf(err, "CdnDoPreload req(%v) res(%v)", reqParam, res)
	}
	return nil
}

// CdnPreloadQuery .
func (d *Dao) CdnPreloadQuery(c context.Context, reqParam []string) (map[string]*model.CdnKsyun, error) {
	var msgBytes []byte
	params := map[string]interface{}{
		"Urls":   reqParam,
		"Action": "preload",
	}
	msgBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := d.host.Cdn + _preHeatQuery
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(msgBytes)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	res := &model.CdnPreloadResult{}
	err = d.client.Do(c, req, res)
	if err != nil {
		return nil, err
	}
	if len(res.Ksyun) == 0 {
		return nil, nil
	}
	mapRes := make(map[string]*model.CdnKsyun)
	for _, v := range res.Ksyun {
		mapRes[v.URL] = v
	}
	return mapRes, nil
}
