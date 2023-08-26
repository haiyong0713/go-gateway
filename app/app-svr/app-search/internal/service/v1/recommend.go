package v1

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	historyGrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

func (s *Service) RecommendTags(ctx context.Context, mid int64, req *search.RecommendTagsReq) (*search.RecommendTagsRsp, error) {
	if s.c.SearchRcmdTagsConfig.CloseRcmdTagsSwitch {
		return nil, nil
	}
	dev, _ := device.FromContext(ctx)
	historyReply, err := s.dao.GetHistoryFrequent(ctx, &historyGrpc.HistoryFrequentReq{
		Mid:        mid,
		Business:   "archive",
		Businesses: []string{"archive"},
		Ip:         metadata.String(ctx, metadata.RemoteIP),
		Buvid:      dev.Buvid,
		StartTs:    req.StartTs,
		EndTs:      req.EndTs,
	})
	if err != nil {
		log.Error("s.historyDao.GetHistoryFrequent req=%+v, err=%+v", req, err)
	}
	num := getNumNot1stInHistoryFrequent(historyReply)
	reply, err := s.dao.GetAiRecommendTags(ctx, req.Style, num, req.DisableRcmd, req.Gt, req.Id1st, i18n.PreferTraditionalChinese(ctx, req.SLocale, req.CLocale))
	if err != nil {
		log.Error("s.aiDao.GetAiRecommendTags req=%+v, err=%+v", req, err)
		return nil, err
	}
	if reply == nil || len(reply.Tags) == 0 {
		// 避免前端无效展示
		return nil, nil
	}
	return reply, nil
}

func getNumNot1stInHistoryFrequent(reply *historyGrpc.HistoryFrequentReply) int64 {
	if reply == nil || len(reply.Res) == 0 {
		return 0
	}
	return int64(len(reply.Res)) - 1
}
