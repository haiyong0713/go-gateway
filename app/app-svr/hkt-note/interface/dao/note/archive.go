package note

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"

	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	"github.com/pkg/errors"
)

const (
	_needAll = 1
)

func (d *Dao) arc(c context.Context, aid int64) (*arcapi.Arc, error) {
	var arg = &arcapi.ArcRequest{Aid: aid}
	reply, err := d.arcClient.Arc(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "Arc aid(%d)", aid)
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	return reply.Arc, nil
}

func (d *Dao) cheeseSeason(c context.Context, sid int32) (*cssngrpc.SeasonCard, error) {
	arg := &cssngrpc.SeasonCardsReq{Ids: []int32{sid}, NeedAll: _needAll}
	reply, err := d.chSsnClient.Cards(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "CheeseSeason sid(%d)", sid)
	}
	if reply == nil || reply.Cards == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "CheeseSeason sid(%d)", sid)
	}
	if ssn, ok := reply.Cards[sid]; !ok || ssn == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "CheeseSeason sid(%d)", sid)
	}
	return reply.Cards[sid], nil
}

// 获取稿件详情
func (d *Dao) ToArcCore(c context.Context, oid int64, oidType int) *note.ArcCore {
	res := &note.ArcCore{Oid: oid}
	switch oidType {
	case note.OidTypeUgc:
		arc, err := d.arc(c, oid)
		if err != nil {
			log.Warn("noteWarn toArcCore err(%+v)", err)
			res.Status = note.ArcStatusWrong
			return res
		}
		res.FromUGC(arc)
	case note.OidTypeCheese:
		ssn, err := d.cheeseSeason(c, int32(oid))
		if err != nil {
			log.Warn("noteWarn toArcCore err(%+v)", err)
			res.Status = note.ArcStatusWrong
			return res
		}
		res.FromCheese(ssn)
	default:
		res.Status = note.ArcStatusWrong
		log.Warn("noteInfo noteInfo oid(%d) oidType(%d) invalid", oid, oidType)
	}
	return res
}
