package dao

import (
	"context"

	copyrightSvc "git.bilibili.co/bapis/bapis-go/copyright-manage/interface"
)

type CopyrightBansOpt struct {
	Aids []int64
}

func (d *dao) CopyrightBans(ctx context.Context, opt CopyrightBansOpt) (map[int64]bool, error) {
	req := &copyrightSvc.AidsReq{Aids: opt.Aids, Option: copyrightSvc.BanOption_BanListen}
	resp, err := d.copyrightGRPC.GetArcsBanPlay(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "copyrightGRPC.GetArcsBanPlay", req)
	}
	ret := make(map[int64]bool)
	for k, v := range resp.GetBanPlay() {
		ret[k] = v == copyrightSvc.BanPlayEnum_IsBan
	}
	return ret, nil
}

const (
	copyRightBanListen = "ban_listen"
	copyRightBan       = 1
)

func (d *dao) CopyrightBan(ctx context.Context, aid int64) (bool, error) {
	req := &copyrightSvc.AidReq{Aid: aid}
	resp, err := d.copyrightGRPC.GetArcBanPlay(ctx, req)
	if err != nil {
		return false, wrapDaoError(err, "copyrightGRPC.GetArcBanPlay", req)
	}
	for _, v := range resp.GetBanPlay() {
		if v.GetKey() == copyRightBanListen && v.GetValue() == copyRightBan {
			return true, nil
		}
	}
	return false, nil
}
