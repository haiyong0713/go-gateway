package caldiff

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/appstatic/job/model"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

const (
	_preHeat = "/api/cache/preload"
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
