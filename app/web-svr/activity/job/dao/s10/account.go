package s10

import (
	"context"

	user "git.bilibili.co/bapis/bapis-go/passport/service/user"

	"go-common/library/log"
)

func (d *Dao) MidByTel(ctx context.Context, tel string) (int64, error) {
	reply, err := d.userAccClient.MidByTel(ctx, &user.MidByTelReq{Tel: tel})
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.MidByTel(tel:%s) error:%v", tel, err)
		return 0, err
	}
	return reply.Mid, nil
}
