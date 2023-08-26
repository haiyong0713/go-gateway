package dynamic

import (
	"context"
	"net/url"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
)

const (
	decoCardsURL = "/x/internal/garb/user/card/multi"
)

func (d *Dao) Cards3(c context.Context, uids []int64) (*accountgrpc.CardsReply, error) {
	cardReply, err := d.accountGRPC.Cards3(c, &accountgrpc.MidsReq{Mids: uids})
	if err != nil || cardReply == nil {
		log.Errorc(c, "Dao.Cards3(uids: %+v) failed. error(%+v)", uids, err)
		return nil, err
	}
	return cardReply, nil
}

func (d *Dao) Infos(c context.Context, uids []int64) (*accountgrpc.InfosReply, error) {
	infos, err := d.accountGRPC.Infos3(c, &accountgrpc.MidsReq{Mids: uids})
	if err != nil || infos == nil {
		log.Errorc(c, "Dao.Infos(uids: %+v) failed. error(%+v)", uids, err)
		return nil, err
	}
	return infos, nil
}

func (d *Dao) DecorateCards(c context.Context, uids []int64) (map[int64]*dynmdl.DecoCards, error) {
	params := url.Values{}
	params.Set("mids", xstr.JoinInts(uids))
	decoCard := d.decoCard
	var ret struct {
		Code int                         `json:"code"`
		Msg  string                      `json:"message"`
		Data map[int64]*dynmdl.DecoCards `json:"data"`
	}
	if err := d.client.Get(c, decoCard, "", params, &ret); err != nil {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), error(%+v)", decoCard, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Errorc(c, "PGCBatch http GET(%s) failed, params:(%s), code: %v, msg: %v", decoCard, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "PGCBatch url(%v) code(%v) msg(%v)", decoCard, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}

func (d *Dao) Followings(c context.Context, uid int64) (*relagrpc.FollowingsReply, error) {
	following, err := d.relaGRPC.Followings(c, &relagrpc.MidReq{Mid: uid})
	if err != nil || following == nil {
		log.Errorc(c, "Dao.Followings(uid: %v) failed. error(%+v)", uid, err)
		return nil, err
	}
	return following, nil
}

func (d *Dao) Stats(c context.Context, uids []int64) (*relagrpc.StatsReply, error) {
	stats, err := d.relaGRPC.Stats(c, &relagrpc.MidsReq{Mids: uids})
	if err != nil || stats == nil {
		log.Errorc(c, "Dao.Stats(uids:%+v) failed. error(%+v)", uids, err)
		return nil, err
	}
	return stats, nil
}
