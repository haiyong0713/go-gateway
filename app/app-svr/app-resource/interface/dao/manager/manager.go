package manager

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/manager"
)

const (
	_splashListURL = "/x/admin/feed/open/splash/list"
)

type Dao struct {
	clientAsyn    *httpx.Client
	splashListURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		clientAsyn:    httpx.NewClient(c.HTTPClientAsyn),
		splashListURL: c.Host.Manager + _splashListURL,
	}
	return
}

func (d *Dao) SplashList(c context.Context) (*manager.SplashList, error) {
	var res struct {
		Code int                 `json:"code"`
		Data *manager.SplashList `json:"data"`
	}
	if err := d.clientAsyn.Get(c, d.splashListURL, "", nil, &res); err != nil {
		log.Error("日志报警 manager splash list url(%s) error(%v)", d.splashListURL, err)
		return nil, err
	}
	if res.Data == nil {
		log.Error("日志报警 manager splash list json data is null")
		return nil, ecode.NothingFound
	}
	if config := res.Data.DefaultConfig; config != nil {
		// 只有有配置的时候处理
		if err := config.SplashConfigChange(); err != nil {
			log.Error("日志报警 manager splash list defaultConfig change error(%v)", err)
			return nil, err
		}
	}
	if config := res.Data.SelectConfig; config != nil {
		// 只有有配置的时候处理
		if err := config.SplashConfigChange(); err != nil {
			log.Error("日志报警 manager splash list selectConfig change error(%v)", err)
			return nil, err
		}
	}
	for _, v := range res.Data.PrepareDefaultConfigs {
		if v == nil {
			continue
		}
		// 只有有配置的时候处理
		if err := v.SplashConfigChange(); err != nil {
			log.Error("日志报警 manager splash list preloadConfig change error(%v)", err)
			return nil, err
		}
	}
	if config := res.Data.BaseDefaultConfig; config != nil {
		// 只有有配置的时候处理
		if err := config.SplashConfigChange(); err != nil {
			log.Error("日志报警 manager splash list baseDefaultConfig change error(%v)", err)
			return nil, err
		}
	}
	return res.Data, nil
}
