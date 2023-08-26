package v1

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"

	"github.com/pkg/errors"
)

func (d *dao) Relations3(c context.Context, owners []int64, mid int64) (follows map[int64]bool) {
	if len(owners) == 0 {
		return nil
	}
	follows = make(map[int64]bool, len(owners))
	var (
		am        *accgrpc.RelationsReply
		err       error
		ip        = metadata.String(c, metadata.RemoteIP)
		ownersMap = make(map[int64]struct{}, len(owners))
		ownerLeft []int64
	)
	for _, owner := range owners {
		if _, ok := ownersMap[owner]; ok {
			continue
		}
		follows[owner] = false
		ownersMap[owner] = struct{}{}
		ownerLeft = append(ownerLeft, owner)
	}
	arg := &accgrpc.RelationsReq{Owners: ownerLeft, Mid: mid, RealIp: ip}
	if am, err = d.accountClient.Relations3(c, arg); err != nil {
		log.Error("d.accRPC.Relations2(%v) error(%v)", arg, err)
		return
	}
	for i, a := range am.Relations {
		if _, ok := follows[i]; ok {
			follows[i] = a.Following
		}
	}
	return
}

func (d *dao) ProfilesWithoutPrivacy3(c context.Context, mids []int64) (map[int64]*accgrpc.ProfileWithoutPrivacy, error) {
	arg := &accgrpc.MidsReq{Mids: mids}
	card, err := d.accountClient.ProfilesWithoutPrivacy3(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	return card.ProfilesWithoutPrivacy, nil
}

func (d *dao) Cards3(c context.Context, mids []int64) (res map[int64]*accgrpc.Card, err error) {
	var cardTmp *accgrpc.CardsReply
	arg := &accgrpc.MidsReq{Mids: mids}
	if cardTmp, err = d.accountClient.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = cardTmp.Cards
	return
}

func (d *dao) NFTBatchInfo(ctx context.Context, in *memberAPI.NFTBatchInfoReq) (*memberAPI.NFTBatchInfoReply, error) {
	reply, err := d.memberClient.NFTBatchInfo(ctx, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (d *dao) Infos3(c context.Context, mids []int64) (res map[int64]*accgrpc.Info, err error) {
	var resTmp *accgrpc.InfosReply
	arg := &accgrpc.MidsReq{Mids: mids}
	if resTmp, err = d.accountClient.Infos3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = resTmp.Infos
	return
}

func (d *dao) CheckRegTime(ctx context.Context, req *accgrpc.CheckRegTimeReq) bool {
	res, err := d.accountClient.CheckRegTime(ctx, req)
	if err != nil {
		log.Error("d.accGRPC.CheckRegTime req=%+v", req)
		return false
	}
	return res.GetHit()
}
