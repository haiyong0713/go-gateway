package account

import (
	"context"
	"fmt"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

// Dao is account dao.
type Dao struct {
	// account grpc
	accGRPC account.AccountClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accGRPC, err = account.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("account NewClient error (%+v)", err))
	}
	return
}

// Card3 get card info by mid
func (d *Dao) Card3(c context.Context, mid int64) (*account.Card, error) {
	arg := &account.MidReq{Mid: mid}
	reply, err := d.accGRPC.Card3(c, arg)
	if err != nil {
		return nil, err
	}
	card := reply.GetCard()
	if card == nil {
		return nil, ecode.NothingFound
	}
	return card, nil
}

// Infos3 is
func (d *Dao) Infos3(ctx context.Context, mids []int64) (map[int64]*account.Info, error) {
	arg := &account.MidsReq{Mids: mids}
	reply, err := d.accGRPC.Infos3(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.GetInfos(), nil
}

// Info3 user info
func (d *Dao) Info3(c context.Context, mid int64) (*account.Info, error) {
	arg := &account.MidReq{Mid: mid}
	reply, err := d.accGRPC.Info3(c, arg)
	if err != nil {
		return nil, err
	}
	info := reply.GetInfo()
	if info == nil {
		return nil, ecode.NothingFound
	}
	return info, nil
}

func (d *Dao) ProfileWithStat3(c context.Context, mid int64) (*account.ProfileStatReply, error) {
	arg := &account.MidReq{Mid: mid}
	return d.accGRPC.ProfileWithStat3(c, arg)
}
