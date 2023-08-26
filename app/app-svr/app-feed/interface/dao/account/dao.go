package account

import (
	"context"
	"fmt"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	// grpc
	accGRPC accountgrpc.AccountClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accGRPC, err = accountgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("accountgrpc NewClientt error (%+v)", err))
	}
	return
}

// Relations3GRPC relations grpc
func (d *Dao) Relations3GRPC(c context.Context, owners []int64, mid int64) (follows map[int64]bool) {
	var (
		am  *accountgrpc.RelationsReply
		err error
	)
	if len(owners) == 0 {
		return nil
	}
	follows = make(map[int64]bool, len(owners))
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.RelationsReq{Mid: mid, Owners: owners, RealIp: ip}
	if am, err = d.accGRPC.Relations3(c, arg); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, o := range owners {
		if a, ok := am.Relations[o]; ok {
			follows[o] = a.Following
		} else {
			follows[o] = false
		}
	}
	return
}

// IsAttentionGRPC is attention grpc
func (d *Dao) IsAttentionGRPC(c context.Context, owners []int64, mid int64) (isAtten map[int64]int8) {
	var (
		am  *accountgrpc.RelationsReply
		err error
	)
	if len(owners) == 0 || mid == 0 {
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.RelationsReq{Mid: mid, Owners: owners, RealIp: ip}
	if am, err = d.accGRPC.Relations3(c, arg); err != nil {
		log.Error("%+v", err)
		return
	}
	isAtten = make(map[int64]int8, len(am.Relations))
	for mid, rel := range am.Relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return
}

// Cards3GRPC card grpc
func (d *Dao) Cards3GRPC(c context.Context, mids []int64) (res map[int64]*accountgrpc.Card, err error) {
	var cardsReply *accountgrpc.CardsReply
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.MidsReq{Mids: mids, RealIp: ip}
	if cardsReply, err = d.accGRPC.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = cardsReply.Cards
	return
}

func (d *Dao) CheckRegTime(ctx context.Context, req *accountgrpc.CheckRegTimeReq) bool {
	res, err := d.accGRPC.CheckRegTime(ctx, req)
	if err != nil {
		log.Error("d.accGRPC.CheckRegTime req=%+v", req)
		return false
	}
	return res.GetHit()
}
