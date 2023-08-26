package history

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/archive/service/conf"

	his "git.bilibili.co/bapis/bapis-go/community/interface/history"
	"github.com/pkg/errors"
)

// Dao dao
type Dao struct {
	c         *conf.Config
	hisClient his.HistoryClient
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c: c,
	}
	var err error
	if dao.hisClient, err = his.NewClient(c.HisClient); err != nil {
		panic(fmt.Sprintf("history newClient panic(%+v)", err))
	}
	return
}

// Progress .
func (d *Dao) Progress(c context.Context, aids []int64, mid int64, buvid string) (map[int64]*his.ModelHistory, error) {
	req := &his.ProgressReq{Mid: mid, Aids: aids, Buvid: buvid}
	res, err := d.hisClient.Progress(c, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("history Progress is nil")
	}
	return res.Res, nil
}
