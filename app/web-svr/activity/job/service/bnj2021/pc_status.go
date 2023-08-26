package bnj2021

import (
	"context"
	"time"

	bnjDao "go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/dao/like"
)

func ASyncResetPCConfiguration(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = resetWebViewData()
		case <-ctx.Done():
			return nil
		}
	}
}

func resetWebViewData() (err error) {
	if BnjRewardCfg.PCVID > 0 {
		var m map[string]interface{}
		m, err = like.FetchWebViewData(context.Background(), BnjRewardCfg.PCVID)
		if err != nil || len(m) == 0 {
			return
		}

		err = bnjDao.ResetWebViewData(m)
	}

	return
}
