package account

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

// Dao is account dao.
type Dao struct {
	// grpc
	accGRPC accountgrpc.AccountClient
}

// New new a account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accGRPC, err = accountgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("accountgrpc NewClient error (%+v)", err))
	}
	return
}

// Profile3 get profile
func (d *Dao) Profile3(c context.Context, mid int64) (*accountgrpc.Profile, error) {
	arg := &accountgrpc.MidReq{Mid: mid}
	card, err := d.accGRPC.ProfileWithStat3(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return card.GetProfile(), nil
}

func (d *Dao) Cards3(c context.Context, uids []int64) (map[int64]*accountgrpc.Card, error) {
	cardReply, err := d.accGRPC.Cards3(c, &accountgrpc.MidsReq{Mids: uids})
	if err != nil || cardReply == nil {
		log.Error("Failed to call Cards3(). uids: %+v. error: %+v", uids, errors.WithStack(err))
		return nil, err
	}
	return cardReply.GetCards(), nil
}

func (d *Dao) Cards3All(c context.Context, uids []int64) (map[int64]*accountgrpc.Card, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*accountgrpc.Card)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			cs, err := d.Cards3(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, c := range cs {
				res[uid] = c
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Cards3 uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}
