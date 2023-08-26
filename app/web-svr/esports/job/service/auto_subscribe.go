package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/dao"
	"go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/tool"

	favpb "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

const (
	cacheKey4AutoSubscribeList = "auto_subscribe"

	bizListLabel4AutoSub = "auto_subscribe"
)

func (s *Service) AsyncAutoSubscribe(ctx context.Context) {
	conn := component.GlobalAutoSubCache.Get(ctx)
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			if l, err := redis.Int64(conn.Do("LLEN", cacheKey4AutoSubscribeList)); err == nil {
				tool.Metric4BizListLen.WithLabelValues([]string{bizListLabel4AutoSub}...).Set(float64(l))
			}

			if bs, err := redis.Bytes(conn.Do("RPOP", cacheKey4AutoSubscribeList)); err == nil {
				detail := model.AutoSubscribeDetail{}
				if err := json.Unmarshal(bs, &detail); err == nil {
					_ = s.autoSubscribeByDetail(ctx, detail)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) autoSubscribeByDetail(ctx context.Context, detail model.AutoSubscribeDetail) error {
	var startID int64
	seasonIDStr := strconv.FormatInt(detail.SeasonID, 10)
	teamIDStr := strconv.FormatInt(detail.TeamId, 10)

	for {
		midList, lastID, err := dao.AutoSubMids(ctx, detail, startID)
		if err != nil {
			return err
		}

		for _, v := range midList {
			arg := &favpb.AddFavReq{Tp: int32(favmdl.TypeEsports), Mid: v, Oid: detail.ContestID, Fid: 0}
			_, err = component.FavClient.AddFav(ctx, arg)
			if err != nil {
				tool.Metric4AutoSub.WithLabelValues([]string{seasonIDStr, teamIDStr, "error"}...).Inc()
			} else {
				tool.Metric4AutoSub.WithLabelValues([]string{seasonIDStr, teamIDStr, "succeed"}...).Inc()
				if s.isNewBGroup(detail.ContestID) {
					s.dao.AsyncSendBGroupDatabus(ctx, v, detail.ContestID) // 新小卡.
				} else {
					s.dao.AsyncSendTunnelDatabus(ctx, _platform, v, detail.ContestID) // 老天马卡.
				}
			}
		}

		if len(midList) < dao.Limit4EveryQuery {
			return nil
		}

		startID = lastID
	}
}

func (s *Service) isNewBGroup(contestID int64) bool {
	if s.c.TunnelBGroup.SendNew == 1 {
		return true
	}
	for _, grayID := range s.c.TunnelBGroup.NewContests {
		if grayID == contestID {
			return true
		}
	}
	return false
}
