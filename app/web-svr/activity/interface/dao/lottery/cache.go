package lottery

import (
	"context"

	"go-gateway/app/web-svr/activity/interface/model/lottery"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -nullcache=&lottery.WxLotteryLog{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	WxLotteryLog(ctx context.Context, mid int64) (*lottery.WxLotteryLog, error)
	// bts: -nullcache=&lottery.WxLotteryHis{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	WxLotteryHisByBuvid(ctx context.Context, buvid string) (*lottery.WxLotteryHis, error)
}
