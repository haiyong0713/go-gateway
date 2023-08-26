package account

import (
	accwar "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/app-svr/app-intl/interface/conf"
)

// Dao is archive dao.
type Dao struct {
	accClient accwar.AccountClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accClient, err = accwar.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	return
}
