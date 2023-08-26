package popups

import (
	"fmt"
	crowd "git.bilibili.co/bapis/bapis-go/platform/service/bgroup"
	"go-common/library/conf/env"
	"go-common/library/database/sql"
	"go-common/library/database/taishan"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/model"
)

type tableConfig struct {
	Table string
	Token string
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type Dao struct {
	db       *sql.DB
	c        *conf.Config
	Taishan  *Taishan
	popCache []*model.PopUps
	// crowd gprc
	CrowdGRPC crowd.BGroupServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.DB.Manager),
	}
	d.popCache = make([]*model.PopUps, 0)
	zone := env.Zone
	t, err := taishan.NewClient(&warden.ClientConfig{Zone: zone})
	if err != nil {
		panic(fmt.Sprintf("taishan.NewClient err(%v)", err))
	}
	if d.CrowdGRPC, err = crowd.NewClient(c.CrowdGRPC); err != nil {
		log.Error("PopUps NewCrowdGRPC error(%+v)", err)
		panic(err)
	}
	d.Taishan = &Taishan{
		client: t,
		tableCfg: tableConfig{
			Table: c.Taishan.Popups.Table,
			Token: c.Taishan.Popups.Token,
		},
	}
	d.FlushPopUpsCache()
	return
}

// Close close db resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
