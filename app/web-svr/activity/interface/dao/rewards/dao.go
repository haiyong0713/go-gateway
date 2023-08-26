package rewards

import (
	"fmt"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const _userSub = 100

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) *Dao {
	d := &Dao{db: component.GlobalRewardsDB}
	return d
}

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%_userSub)
}
