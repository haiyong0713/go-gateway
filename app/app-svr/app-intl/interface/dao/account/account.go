package account

import (
	"context"

	accwar "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

func (d *Dao) relations3(c context.Context, owners []int64, mid int64) (relations map[int64]*accwar.RelationReply, err error) {
	var (
		reply *accwar.RelationsReply
		ip    = metadata.String(c, metadata.RemoteIP)
	)
	arg := &accwar.RelationsReq{Owners: owners, Mid: mid, RealIp: ip}
	if reply, err = d.accClient.Relations3(c, arg); err != nil {
		log.Error("[AccClient] Relations3 Arg %v, Err %v", arg, err)
		return
	}
	relations = reply.Relations
	return
}

// Relations3 relatons
func (d *Dao) Relations3(c context.Context, owners []int64, mid int64) (follows map[int64]bool) {
	if len(owners) == 0 {
		return nil
	}
	follows = make(map[int64]bool, len(owners))
	for _, owner := range owners {
		follows[owner] = false
	}
	var (
		err       error
		relations map[int64]*accwar.RelationReply
	)
	if relations, err = d.relations3(c, owners, mid); err != nil {
		return
	}
	for i, a := range relations {
		if _, ok := follows[i]; ok {
			follows[i] = a.Following
		}
	}
	return
}

// IsAttention is
func (d *Dao) IsAttention(c context.Context, owners []int64, mid int64) (isAtten map[int64]int8) {
	if len(owners) == 0 || mid == 0 {
		return
	}
	var (
		err       error
		relations map[int64]*accwar.RelationReply
	)
	if relations, err = d.relations3(c, owners, mid); err != nil {
		return
	}
	isAtten = make(map[int64]int8, len(relations))
	for mid, rel := range relations {
		if rel.Following {
			isAtten[mid] = 1
		}
	}
	return
}

// Card3 get card info by mid
func (d *Dao) Card3(c context.Context, mid int64) (res *accwar.Card, err error) {
	var (
		arg   = &accwar.MidReq{Mid: mid}
		reply *accwar.CardReply
	)
	if reply, err = d.accClient.Card3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = reply.Card
	return
}

// Cards3 get cards info by mids
func (d *Dao) Cards3(c context.Context, mids []int64) (res map[int64]*accwar.Card, err error) {
	var (
		arg   = &accwar.MidsReq{Mids: mids}
		reply *accwar.CardsReply
	)
	if reply, err = d.accClient.Cards3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = reply.Cards
	return
}

// Following3 following.
func (d *Dao) Following3(c context.Context, mid, owner int64) (follow bool, err error) {
	var (
		ip  = metadata.String(c, metadata.RemoteIP)
		arg = &accwar.RelationReq{Mid: mid, Owner: owner, RealIp: ip}
	)
	rl, err := d.accClient.Relation3(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rl != nil {
		follow = rl.Following
	}
	return
}

// Infos3 rpc info get by mids .
func (d *Dao) Infos3(c context.Context, mids []int64) (res map[int64]*accwar.Info, err error) {
	var (
		arg   = &accwar.MidsReq{Mids: mids}
		reply *accwar.InfosReply
	)
	if reply, err = d.accClient.Infos3(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = reply.Infos
	return
}

// IsVip is
func (d *Dao) IsVip(c context.Context, mid int64) (isVip bool, err error) {
	vipReply, err := d.accClient.Vip3(c, &accwar.MidReq{Mid: mid})
	if err != nil {
		return
	}
	isVip = vipReply.IsValid()
	return
}
