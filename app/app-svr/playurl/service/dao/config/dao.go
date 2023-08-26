package config

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/playurl/service/conf"

	appConf "git.bilibili.co/bapis/bapis-go/community/service/appconfig"
)

type Dao struct {
	confClient appConf.AppConfigClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.confClient, err = appConf.NewClient(c.AppConfClient)
	if err != nil {
		panic(fmt.Sprintf("appconf NewClient error(%v)", err))
	}
	return
}

// SubtitleExist .
func (d *Dao) ShakeConfig(c context.Context, aid, cid int64) (string, error) {
	res, err := d.confClient.ShakeConfig(c, &appConf.ShakeConfigReq{Aid: aid, Cid: cid})
	if err != nil {
		log.Error("ShakeConfig err(%+v) aid(%d) cid(%d)", err, aid, cid)
		return "", err
	}
	return res.GetConfig().GetUrl(), nil
}
