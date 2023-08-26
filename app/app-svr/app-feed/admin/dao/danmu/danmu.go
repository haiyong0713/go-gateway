package danmu

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	danmu "git.bilibili.co/bapis/bapis-go/community/interface/dm-admin"
	"go-common/library/sync/errgroup.v2"
)

// Dao is danmu dao.
type Dao struct {
	// danmu grpc
	dmGRPC danmu.DMClient
	dmCfg  *conf.Danmu
}

// New danmu dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.dmGRPC, err = danmu.NewClientDM(c.DanmuGRPC); err != nil {
		panic(fmt.Sprintf("damnu NewClient error (%+v)", err))
	}
	d.dmCfg = c.WeeklySelected.Danmu
	return
}

func (d *Dao) AddWeekViewDanmuV1(ctx context.Context, cids []int64) (err error) {
	req := &danmu.BatchAddActivityCidReq{
		Oids:    cids,
		BizType: danmu.BatchAddActivityBizType_WeekView,
	}
	_, err = d.dmGRPC.BatchAddActivityCid(ctx, req)
	if err != nil {
		log.Error("damnu AddWeekViewDanmuV1 err(%v)", err)
	}

	return
}

func (d *Dao) AddWeekViewDanmuV2(ctx context.Context, cids []int64) (err error) {
	resouceMeta := &selected.ResourceMeta{
		Icon:    d.dmCfg.Icon,
		KeyWord: "每周必看",
	}
	resouceMetaBs, err := json.Marshal(resouceMeta)
	if err != nil {
		log.Error("damnu AddWeekViewDanmuV2 Marshal resouceMeta(%+v) err(%v)", resouceMeta, err)
		return
	}
	purifyExtra := &selected.PurifyExtra{
		PurifyEffective:    d.dmCfg.PurifyExtra.PurifyEffective,
		EffectivePeriod:    d.dmCfg.PurifyExtra.EffectivePeriod,
		EffectiveMax:       d.dmCfg.PurifyExtra.EffectiveMax,
		PurifyNonEffective: d.dmCfg.PurifyExtra.PurifyNonEffective,
		NonEffectivePeriod: d.dmCfg.PurifyExtra.NonEffectivePeriod,
		NonEffectiveMax:    d.dmCfg.PurifyExtra.NonEffectiveMax,
	}
	purifyExtraBs, err := json.Marshal(purifyExtra)
	if err != nil {
		log.Error("damnu AddWeekViewDanmuV2 Marshal purifyExtra(%+v) err(%v)", purifyExtra, err)
		return
	}
	var cidSection []*danmu.CidSection
	for _, v := range d.dmCfg.CidSection {
		cidSection = append(cidSection, &danmu.CidSection{
			Start: v.StartTime,
			End:   v.EndTime,
		})
	}
	e := errgroup.WithContext(ctx)
	for i := 0; i <= int(len(cids)/selected.DanmuBatchCids); i++ {
		left := i * selected.DanmuBatchCids
		right := left + selected.DanmuBatchCids
		if left >= len(cids) {
			break
		}
		if right > len(cids) {
			right = len(cids)
		}
		e.Go(func(ctx context.Context) (err error) {
			var oidItem []*danmu.OidItem
			for j := left; j < right; j++ {
				oidItem = append(oidItem, &danmu.OidItem{
					Oid: cids[j],
					ExtraSet: &danmu.ExtraSet{
						PurifyExtra: string(purifyExtraBs),
						CidSection:  cidSection,
					},
				})
			}
			req := &danmu.EditActivityByBizReq{
				BizType:      danmu.EditActivityBizType_WeekView,
				OidType:      selected.CidType, //视频分段
				ResourceMeta: string(resouceMetaBs),
				OidItem:      oidItem,
				AutoDeploy:   true, //自动上架
			}
			_, err = d.dmGRPC.EditActivityByBiz(ctx, req)
			if err != nil {
				log.Error("damnu AddWeekViewDanmuV2 err(%v)", err)
			}
			return
		})
	}
	if err = e.Wait(); err != nil {
		return
	}
	log.Warn("damnu AddWeekViewDanmuV2 success!")
	return
}
