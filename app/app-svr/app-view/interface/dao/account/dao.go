package account

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"
)

// Dao is account dao.
type Dao struct {
	accApi api.AccountClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	client, err := api.NewClient(c.AccClient)
	if err != nil {
		panic(fmt.Sprintf("accountGRPC error(%v)", err))
	}
	d.accApi = client
	return
}

// Card3 get card info by mid
func (d *Dao) Card3(c context.Context, mid int64) (res *api.Card, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidReq{Mid: mid, RealIp: ip}
	var card *api.CardReply
	if card, err = d.accApi.Card3(c, arg); err != nil || card.Card == nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = card.Card
	return
}

func (d *Dao) GetInfos(c context.Context, mids []int64) (*api.InfosReply, error) {
	if len(mids) == 0 {
		return nil, ecode.RequestErr
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidsReq{
		Mids:   mids,
		RealIp: ip,
	}
	infos, err := d.accApi.Infos3(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	if infos == nil {
		return nil, ecode.NothingFound
	}
	return infos, nil
}

// Cards3 get cards info by mids
func (d *Dao) Cards3(c context.Context, mids []int64) (res map[int64]*api.Card, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidsReq{Mids: mids, RealIp: ip}
	var cards *api.CardsReply
	if cards, err = d.accApi.Cards3(c, arg); err != nil || cards.Cards == nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = cards.Cards
	return
}

// Following3 following.
func (d *Dao) Following3(c context.Context, mid, owner int64) (follow bool, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.RelationReq{Mid: mid, Owner: owner, RealIp: ip}
	rl, err := d.accApi.Relation3(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rl != nil {
		follow = rl.Following
	}
	return
}

// IsAttention is attention
func (d *Dao) IsAttention(c context.Context, owners []int64, mid int64) (isAtten map[int64]int8) {
	if len(owners) == 0 || mid == 0 {
		return
	}
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.RelationsReq{Owners: owners, Mid: mid, RealIp: ip}
	res, err := d.accApi.Relations3(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	isAtten = make(map[int64]int8, len(res.Relations))
	for mid, rel := range res.Relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return
}

func (d *Dao) IsBlueV(c context.Context, mid int64) bool {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidReq{Mid: mid, RealIp: ip}
	rl, err := d.accApi.Profile3(c, arg)
	if err != nil {
		log.Error("d.accApi.Profile3 arg(%+v) err(%v)", arg, err)
		return false
	}

	if rl != nil && rl.Profile != nil {
		return rl.Profile.Official.Type == 1
	}

	return false
}

// ContractRelation3 契约关系
func (d *Dao) ContractRelation3(c context.Context, mid, owner int64) (*api.ContractRelationReply, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.RelationReq{Mid: mid, Owner: owner, RealIp: ip}
	infos, err := d.accApi.ContractRelation(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}

	if infos == nil {
		return nil, ecode.NothingFound
	}
	return infos, nil
}

func (d *Dao) GetInfo(c context.Context, mid int64) (res *api.Info, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &api.MidReq{Mid: mid, RealIp: ip}
	info, err := d.accApi.Info3(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if info != nil && info.Info != nil {
		res = info.Info
	}
	return
}

// 先不删了，说不准以后的实验还要用.....
func (d *Dao) IsNewDevice(c context.Context, buvid, periods string) bool {
	res, err := d.accApi.CheckRegTime(c, &api.CheckRegTimeReq{Buvid: buvid, Periods: periods})
	if err != nil {
		log.Error("d.accApi.CheckRegTime(%s) error(%v)", buvid, err)
		return false
	}
	return res.GetHit()
}
