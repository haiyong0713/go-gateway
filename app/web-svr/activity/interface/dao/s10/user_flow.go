package s10

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-gateway/app/web-svr/activity/interface/model/s10"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
)

const _userFlowSQL = "select id,source from act_s10_user_flow where mid=?;"

func (d *Dao) UserFlow(ctx context.Context, mid int64) (res, source int32, err error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessUserFlow) {
		return 0, 0, xecode.LimitExceed
	}
	row := component.S10GlobalDB.QueryRow(ctx, _userFlowSQL, mid)
	if err = row.Scan(&res, &source); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		log.Errorc(ctx, "s10 d.dao.ExchangeDataPackage(mid:%d) error:%v", mid, err)
		return
	}
	return
}
