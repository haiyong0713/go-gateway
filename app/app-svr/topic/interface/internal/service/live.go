package service

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

func (s *Service) liveInfos(c context.Context, uids []int64, general *topiccardmodel.GeneralParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	lives := make(map[int64]*livexroom.Infos)
	playurls := make(map[int64]*livexroom.LivePlayUrlData)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) error {
			ls, ps, err := s.liveInfosSlice(ctx, partUids, general)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, l := range ls {
				lives[uid] = l
			}
			for uid, p := range ps {
				playurls[uid] = p
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("liveInfos uids(%+v) eg.wait(%+v)", uids, err)
		return nil, nil, err
	}
	return lives, playurls, nil
}

func (s *Service) liveInfosSlice(c context.Context, uids []int64, general *topiccardmodel.GeneralParam) (map[int64]*livexroom.Infos, map[int64]*livexroom.LivePlayUrlData, error) {
	resTmp, err := s.livexroomGRPC.GetMultipleByUids(c, &livexroom.UIDsReq{
		Uids:  uids,
		Attrs: []string{"show", "status", "area", "pendants"},
		Playurl: &livexroom.PlayURLParams{
			Switch:   1,
			ReqBiz:   "/bilibili.app.topic.v1.Topic/TopicDetailsAll",
			Uipstr:   metadata.String(c, metadata.RemoteIP),
			Uid:      general.Mid,
			Platform: general.GetPlatform(),
			Build:    general.GetBuild(),
		},
	})
	if err != nil {
		log.Error("s.livexroomGateGRPC.GetMultipleByUids error=%+v", err)
		return nil, nil, err
	}
	return resTmp.List, resTmp.PlayUrl, nil
}

func (s *Service) SessionInfo(c context.Context, liveAdditionals map[int64][]string, general *topiccardmodel.GeneralParam) (map[string]*livexroomgate.SessionInfos, error) {
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[string]*livexroomgate.SessionInfos)
	for mid, liveIds := range liveAdditionals {
		for _, liveId := range liveIds {
			upmid := mid
			lid := liveId
			uplives := map[int64]*livexroomgate.LiveIds{upmid: {LiveIds: []string{lid}}}
			g.Go(func(ctx context.Context) (err error) {
				req := &livexroomgate.SessionInfoBatchReq{
					UidLiveIds: uplives,
					EntryFrom:  []string{"dt_booking_dt"},
					Playurl: &livexroomgate.PlayUrlReq{
						ReqBiz:     "/bilibili.app.topic.v1.Topic/TopicDetailsAll",
						Uipstr:     metadata.String(c, metadata.RemoteIP),
						Uid:        general.Mid,
						Platform:   general.GetPlatform(),
						Build:      general.GetBuild(),
						DeviceName: general.GetDevice(),
						Network:    "other",
					},
				}
				reply, err := s.livexroomGateGRPC.SessionInfoBatch(ctx, req)
				if err != nil {
					log.Error("%+v", err)
					return err
				}
				mu.Lock()
				if item, ok := reply.List[upmid]; ok {
					res[lid] = item
				}
				mu.Unlock()
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
