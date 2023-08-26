package account

import (
	"context"
	"fmt"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	acvalidate "go-gateway/app/app-svr/app-card/interface/model/thirdValidate"
	"go-gateway/app/app-svr/app-show/interface/conf"

	"github.com/pkg/errors"
)

// Dao is rpc dao.
type Dao struct {
	// grpc
	accGRPC accountgrpc.AccountClient
}

// New new a account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accGRPC, err = accountgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(fmt.Sprintf("accountgrpc NewClientt error (%+v)", err))
	}
	return
}

// Cards3GRPC card grpc
func (d *Dao) Cards3GRPC(c context.Context, rawMids []int64) (map[int64]*accountgrpc.Card, error) {
	var (
		cardsReply *accountgrpc.CardsReply
		err        error
	)
	mids := midFilter(rawMids)
	if len(mids) == 0 {
		return nil, ecode.NothingFound
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.MidsReq{Mids: mids, RealIp: ip}
	if cardsReply, err = d.accGRPC.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	acValidate := &acvalidate.AccountCardValidator{AccountCards: cardsReply}
	// compare
	acValidate.CompareLength(mids)
	return cardsReply.Cards, nil
}

// Relations3GRPC relations grpc
func (d *Dao) Relations3GRPC(c context.Context, mid int64, owners []int64) (res map[int64]*accountgrpc.RelationReply, err error) {
	var (
		am *accountgrpc.RelationsReply
	)
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &accountgrpc.RelationsReq{Mid: mid, Owners: owners, RealIp: ip}
	if am, err = d.accGRPC.Relations3(c, arg); err != nil {
		log.Error("%+v", err)
		return
	}
	res = am.Relations
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

func (d *Dao) Info3GRPC(c context.Context, mid int64) (res *accountgrpc.InfoReply, err error) {
	if mid < 1 {
		return
	}
	arg := &accountgrpc.MidReq{Mid: mid}
	if res, err = d.accGRPC.Info3(c, arg); err != nil {
		log.Error("%+v", err)
	}
	return
}

// Infos3GRPC is
func (d *Dao) Infos3GRPC(ctx context.Context, rawMids []int64) (map[int64]*accountgrpc.Info, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	//剔除rawMids中为0的值
	var mids []int64
	for _, mid := range rawMids {
		if mid > 0 {
			mids = append(mids, mid)
		}
	}
	if len(mids) == 0 {
		return nil, ecode.NothingFound
	}
	arg := &accountgrpc.MidsReq{
		Mids:   mids,
		RealIp: ip,
	}
	reply, err := d.accGRPC.Infos3(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Infos, nil
}

func midFilter(rawMids []int64) []int64 {
	//去重+剔除为0的值
	midsSet := make(map[int64]struct{})
	for _, v := range rawMids {
		midsSet[v] = struct{}{}
	}
	var mids []int64
	for mid := range midsSet {
		if mid > 0 {
			mids = append(mids, mid)
		}
	}
	return mids
}
