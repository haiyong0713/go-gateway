package mission

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/mission"
)

func (s *Service) FetchVideoByOperId(ctx context.Context, operSourceDataId int64) (res []*mission.VideoInfo, err error) {
	operConfig, err := s.dao.GetVideoAIdsByOperSourceId(ctx, operSourceDataId)
	if err != nil {
		log.Errorc(ctx, "dao.GetRoomIdsByOperSourceId:%v, err: %v", operSourceDataId, err)
		return
	}
	return s.getVideoInfoByAids(ctx, operConfig)
}

func (s *Service) getVideoInfoByAids(c context.Context, aids []int64) (res []*mission.VideoInfo, err error) {
	res = make([]*mission.VideoInfo, 0, len(aids))
	archiveInfos, err := client.Archives(c, aids)
	if err != nil {
		return
	}
	for _, aid := range aids {
		tmp, ok := archiveInfos[aid]
		if !ok {
			continue
		}
		if !tmp.IsNormal() {
			log.Errorc(c, "getVideoInfoByAids skip aid %v because state=%v", tmp.Aid, tmp.State)
			continue
		}
		res = append(res, &mission.VideoInfo{
			Id:         tmp.Aid,
			Author:     tmp.Author,
			VideoCover: tmp.Pic,
			VideoTitle: tmp.Title,
			VideoUrl:   tmp.ShortLinkV2,
			Duration:   tmp.Duration,
			Stat:       tmp.Stat,
		})
	}
	return
}
