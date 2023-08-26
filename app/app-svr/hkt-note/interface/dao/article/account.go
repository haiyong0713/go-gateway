package article

import (
	"context"

	"go-common/library/ecode"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	upgrpc "git.bilibili.co/bapis/bapis-go/archive/service/up"
	"github.com/pkg/errors"
)

const (
	_fromNote = 10
	_showNote = 0
)

func (d *Dao) AccCards(c context.Context, mids []int64) (map[int64]*accgrpc.Card, error) {
	cardsReply, err := d.accClient.Cards3(c, &accgrpc.MidsReq{Mids: mids})
	if err != nil {
		return nil, errors.Wrapf(err, "AccCards mids(%v)", mids)
	}
	if cardsReply == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "AccCards mids(%v)", mids)
	}
	if cardsReply.Cards == nil {
		return make(map[int64]*accgrpc.Card), nil
	}
	return cardsReply.Cards, nil
}

func (d *Dao) UpSwitch(c context.Context, mid int64) (bool, error) {
	res, err := d.upClient.UpSwitch(c, &upgrpc.UpSwitchReq{Mid: mid, From: _fromNote})
	if err != nil {
		return false, errors.Wrapf(err, "UpSwitch mid(%d)", mid)
	}
	if res == nil {
		return false, errors.Wrapf(ecode.NothingFound, "UpSwitch mid(%d)", mid)
	}
	// state 开关状态 0-打开 1-关闭 (和文档相反)
	return res.State == _showNote, nil
}
