package account

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/playurl/service/conf"

	accrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

// Dao is account dao.
type Dao struct {
	// rpc
	accRPC accrpc.AccountClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.accRPC, err = accrpc.NewClient(c.AccountClient)
	if err != nil {
		panic(fmt.Sprintf("account NewClient error(%v)", err))
	}
	return
}

// IsVip check mid is vip
func (d *Dao) IsVip(c context.Context, mid int64) (isVip int, err error) {
	vipReply, err := d.accRPC.Vip3(c, &accrpc.MidReq{Mid: mid})
	if err != nil {
		return
	}
	if vipReply.IsValid() {
		isVip = 1
	}
	return
}
