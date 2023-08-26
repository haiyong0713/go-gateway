package service

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/up_reserve"
	"strconv"
	"strings"
)

func (s *Service) UpReserveList(ctx context.Context, arg *up_reserve.ParamList) (rly *up_reserve.UpReserveListReply, err error) {
	if arg.Sid == 0 && arg.Mid == 0 {
		return
	}
	rly, err = s.dao.GetUpReserve(ctx, arg.Sid, arg.Mid, arg.Pn, arg.Ps)
	if err != nil {
		log.Errorc(ctx, "s.dao.UpReserveList err, arg:(%+v), rly:(%+v), err:(%+v)", arg, rly, err)
	}
	return
}

func (s *Service) UpReserveHang(ctx context.Context, arg *up_reserve.ParamHang) (err error) {
	mids := strings.Split(arg.Mid, ",")
	if len(mids) == 0 {
		err = ecode.Error(ecode.RequestErr, "mid格式非法,请用因为逗号分隔并且数量要大于0")
		return
	}
	if len(mids) > 100 {
		err = ecode.Error(ecode.RequestErr, "批量添加每次mid不得超过100个")
		return
	}

	for _, item := range mids {
		mid, _ := strconv.ParseInt(item, 10, 64)
		if mid <= 0 {
			err = ecode.Error(ecode.RequestErr, "存在非法的mid")
			return
		}
	}
	if err = s.dao.BatchInsertUpReserveHangAndLog(ctx, arg.Sid, mids, arg.Operator); err != nil {
		err = errors.Wrap(err, "s.dao.BatchInsertUpReserveHangAndLog err")
		log.Errorc(ctx, err.Error())
		err = ecode.Error(ecode.RequestErr, err.Error())
		return
	}

	return
}

func (s *Service) UpReserveHangLogList(ctx context.Context, arg *up_reserve.HangLogListParams) (res *up_reserve.HangLogListReply, err error) {
	res, err = s.dao.GetUpReserveHangLogList(ctx, arg.Sid, arg.Pn, arg.Ps)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUpReserveHangLogList err")
		log.Errorc(ctx, err.Error())
		err = ecode.Error(ecode.RequestErr, err.Error())
		return
	}
	return
}
