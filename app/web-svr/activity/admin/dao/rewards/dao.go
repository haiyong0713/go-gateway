package rewards

import (
	"fmt"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/admin/conf"
)

const _userSub = 100

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) *Dao {
	d := &Dao{db: sql.NewMySQL(c.RewardsMySQL)}
	return d
}

func userHit(mid int64) string {
	return fmt.Sprintf("%02d", mid%_userSub)
}
