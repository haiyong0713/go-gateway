package feature

import (
	"fmt"

	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"

	"github.com/jinzhu/gorm"
)

const (
	_auth          = "/v1/auth"
	_fetchRoleTree = "/v1/node/role/app"
)

type Dao struct {
	authURL     string
	roleTreeURL string
	http        *bm.Client
	db          *gorm.DB
	plats       []*feature.Plat
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		authURL:     fmt.Sprintf("%s/%s", c.Host.Easyst, _auth),
		roleTreeURL: fmt.Sprintf("%s/%s", c.Host.Easyst, _fetchRoleTree),
		http:        bm.NewClient(c.HTTPClient.Read),
		db:          orm.NewMySQL(c.ORMFeature),
		plats:       c.Plats,
	}

	return d
}
