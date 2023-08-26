package dao

import (
	"context"
	"sync"

	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	ugcSeasonSvc "git.bilibili.co/bapis/bapis-go/ugc-season/service"
	"go-common/library/sync/errgroup.v2"
)

const _ugcBatchSize = 50

func (d *dao) UgcSeasonsInfo(ctx context.Context, ss []int64) (ret map[int64]*ugcSeasonSvc.Season, err error) {
	eg := errgroup.WithContext(ctx)
	ret = make(map[int64]*ugcSeasonSvc.Season)
	mu := sync.Mutex{}
	for i := 0; i < len(ss); i += _ugcBatchSize {
		var partSS []int64
		if i+_ugcBatchSize > len(ss) {
			partSS = ss[i:]
		} else {
			partSS = ss[i : i+_ugcBatchSize]
		}
		eg.Go(func(c context.Context) error {
			req := &ugcSeasonSvc.SeasonsRequest{SeasonIds: partSS}
			resp, err := d.ugcSeasonGRPC.Seasons(ctx, req)
			if err != nil {
				return wrapDaoError(err, "ugcSeasonGRPC.Seasons", req)
			}
			mu.Lock()
			for k := range resp.GetSeasons() {
				ret[k] = resp.GetSeasons()[k]
			}
			mu.Unlock()
			return nil
		})
	}
	err = eg.Wait()
	return ret, err
}

func (d *dao) UgcSeasonDetail(ctx context.Context, ss int64) (model.UgcSeasonDetail, error) {
	resp, err := d.ugcSeasonGRPC.View(ctx, &ugcSeasonSvc.ViewRequest{SeasonID: ss})
	if err != nil {
		return model.UgcSeasonDetail{}, wrapDaoError(err, "ugcSeasonGRPC.View", ss)
	}
	return model.UgcSeasonDetail{
		Season:   resp.GetView().GetSeason(),
		Sections: resp.GetView().GetSections(),
	}, nil
}
