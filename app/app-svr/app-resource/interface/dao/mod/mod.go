package mod

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/mod"

	"github.com/pkg/errors"
)

const (
	_appKeyList = "/x/admin/fawkes/business/mod/appkey/list"
	_fileList   = "/x/admin/fawkes/business/mod/appkey/file/list"
)

type Dao struct {
	client     *bm.Client
	appKeyList string
	fileList   string
	redis      credis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     bm.NewClient(c.HTTPClientAsyn),
		appKeyList: c.Host.Fawkes + _appKeyList,
		fileList:   c.Host.Fawkes + _fileList,
		redis:      credis.NewRedis(c.Redis.Fawkes.Config),
	}
	return
}

func (d *Dao) Close() {
}

func (d *Dao) AppKeyList(ctx context.Context) ([]string, error) {
	var res struct {
		Code int      `json:"code"`
		Data []string `json:"data"`
	}
	if err := d.client.Get(ctx, d.appKeyList, "", nil, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.appKeyList)
	}
	return res.Data, nil
}

func (d *Dao) FileList(ctx context.Context, appKey string, env mod.Env, md5Val string) (map[string]map[string][]*mod.File, string, error) {
	params := url.Values{}
	params.Set("app_key", appKey)
	params.Set("env", string(env))
	params.Set("md5", md5Val)
	var res struct {
		Code int `json:"code"`
		Data struct {
			File map[string]map[string][]*mod.File `json:"file"`
			Md5  string                            `json:"md5"`
		} `json:"data"`
	}
	if err := d.client.Get(ctx, d.fileList, "", params, &res); err != nil {
		return nil, "", err
	}
	if res.Code != ecode.OK.Code() {
		return nil, "", errors.Wrap(ecode.Int(res.Code), d.fileList)
	}
	return res.Data.File, res.Data.Md5, nil
}

func (d *Dao) WhitelistData(c context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return d.client.Raw(c, req)
}

func (d *Dao) ModuleDisableList(ctx context.Context, appKey string, env mod.Env) (map[string]int64, error) {
	key := fmt.Sprintf("mod_disable_%s_%s", appKey, env)
	return redis.Int64Map(d.redis.Do(ctx, "HGETALL", key))
}
