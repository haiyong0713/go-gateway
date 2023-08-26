package resource

import (
	"context"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/conf"

	resourceapi "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

type Dao struct {
	resourceClient resourceapi.ResourceClient
	db             *xsql.DB
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		db: xsql.NewMySQL(c.MySQL.Show),
	}
	var err error
	if d.resourceClient, err = resourceapi.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) ResourceUse(c context.Context, arg *resourceapi.ResourceUseAsyncReq) error {
	_, err := d.resourceClient.ResourceUseAsync(c, arg)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) CodeOpen(c context.Context, mid int64, code string) error {
	arg := &resourceapi.CodeOpenReq{
		Mid:       mid,
		Code:      code,
		IP:        metadata.String(c, metadata.RemoteIP),
		Timestamp: time.Now().Unix(),
	}
	_, err := d.resourceClient.CodeOpen(c, arg)
	if err != nil {
		return err
	}
	return nil
}
