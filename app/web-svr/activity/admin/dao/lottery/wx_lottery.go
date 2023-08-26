package lottery

import (
	"context"

	"go-gateway/app/web-svr/activity/admin/model/lottery"

	"github.com/pkg/errors"
)

func (d *Dao) WxLotteryLog(_ context.Context, mid, giftType, pn, ps int64) ([]*lottery.WxLotteryLog, int64, error) {
	wxLog := &lottery.WxLotteryLog{Mid: mid}
	source := d.orm.Table(wxLog.TableName())
	if mid > 0 {
		source = source.Where("mid=?", mid)
	}
	if giftType > 0 {
		source = source.Where("gift_type=?", giftType)
	}
	var count int64
	if err := source.Count(&count).Error; err != nil {
		err = errors.Wrapf(err, "WxLotteryLog Count mid:%d type:%d", mid, giftType)
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, nil
	}
	var list []*lottery.WxLotteryLog
	if err := source.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		err = errors.Wrapf(err, "WxLotteryLog Find mid:%d type:%d", mid, giftType)
		return nil, 0, err
	}
	return list, count, nil
}
