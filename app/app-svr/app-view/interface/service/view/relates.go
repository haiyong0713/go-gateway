package view

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	pageApi "git.bilibili.co/bapis/bapis-go/bilibili/pagination"

	"github.com/pkg/errors"

	"go-common/library/log"
	"go-common/library/pagination"

	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/relates"
	"go-gateway/app/app-svr/app-view/interface/model/view"
)

const (
	_noRelatesPage = 5
)

func (s *Service) PlayerRelatesGRPC(c context.Context, req *relates.RelatesFeedGRPCRequest) (*viewApi.PlayerRelatesReply, error) {
	err := s.checkArcStatus(c, req.Aid)
	if err != nil {
		return nil, err
	}

	params, err := s.buildAiRecommendParams(req, model.PlayerRelateCmd)
	if err != nil {
		return nil, err
	}

	aiResponse := s.getAIRecommend(c, params)

	filteredRelates := s.filterDupRelates(req.Aid, aiResponse.Relates)
	if i18n.PreferTraditionalChinese(c, req.Slocale, req.Clocale) {
		for _, rl := range filteredRelates {
			i18n.TranslateAsTCV2(&rl.Title)
		}
	}

	isMelloi := calcIsMelloi(c)
	if isMelloi == "" {
		// 相关推荐曝光上报
		s.RelateInfoc(req.Mid, req.Aid, int(req.Plat), strconv.FormatInt(req.Build, 10), req.Buvid, req.Ip,
			model.PathPlayerRelates, aiResponse.ReturnCode, aiResponse.UserFeature,
			req.From, "", filteredRelates, time.Now(), 1, 0,
			aiResponse.PlayParam, req.TrackId, model.PageTypeRelate, req.FromSpmid, req.Spmid,
			aiResponse.PvFeature, nil, isMelloi, aiResponse.RelateInfoc, req.PageIndex)
	}

	res := &viewApi.PlayerRelatesReply{
		List: relates.FromRelates(filteredRelates),
	}

	return res, nil
}

func (s *Service) RelatesFeedGRPC(c context.Context, req *relates.RelatesFeedGRPCRequest) (*viewApi.RelatesFeedReply, error) {
	err := s.checkArcStatus(c, req.Aid)
	if err != nil {
		return nil, err
	}

	params, err := s.buildAiRecommendParams(req, model.RelateCmd)
	if err != nil {
		return nil, err
	}

	aiResponse := s.getAIRecommend(c, params)

	filteredRelates := s.filterDupRelates(req.Aid, aiResponse.Relates)
	if i18n.PreferTraditionalChinese(c, req.Slocale, req.Clocale) {
		for _, rl := range filteredRelates {
			i18n.TranslateAsTCV2(&rl.Title)
		}
	}

	isMelloi := calcIsMelloi(c)
	if isMelloi == "" {
		// 相关推荐曝光上报
		s.RelateInfoc(req.Mid, req.Aid, int(req.Plat), strconv.FormatInt(req.Build, 10), req.Buvid, req.Ip,
			model.PathRelatesFeed, aiResponse.ReturnCode, aiResponse.UserFeature,
			req.From, "", filteredRelates, time.Now(), 1, 0,
			aiResponse.PlayParam, req.TrackId, model.PageTypeRelate, req.FromSpmid, req.Spmid,
			aiResponse.PvFeature, nil, isMelloi, aiResponse.RelateInfoc, params.PageIndex)
	}

	res := &viewApi.RelatesFeedReply{
		List:       relates.FromRelates(filteredRelates),
		HasNext:    len(filteredRelates) != 0 && aiResponse.ReturnCode != "3" && aiResponse.ReturnCode != "11",
		Pagination: &pageApi.PaginationReply{},
	}

	//处理ai返回的next标识
	if aiResponse.Next != "" {
		pi, err := strconv.ParseInt(aiResponse.Next, 10, 64)
		if err != nil {
			log.Error("日志告警 pagination aiResponse.Next parse error(%+v)", err)
		} else {
			res.Pagination.Next = pagination.TokenGeneratorWithSalt(model.PaginationTokenSalt).GetPageToken(pi)
		}
	} else {
		s.prom.Incr("相关推荐-新分页参数-ai空Next-relateFeed")
	}

	return res, nil
}

func (s *Service) checkArcStatus(c context.Context, aid int64) error {
	// 校验稿件审核屏蔽状态
	arcAddit, err := s.vuDao.ArcViewAddit(c, aid)
	if err != nil {
		log.Error("checkArcStatus aid(%d) ArcViewAddit err(%+v) or arcAddit=nil", aid, err)
	}
	// 有屏蔽推荐池属性的稿件下, 不出相关推荐任何信息
	if arcAddit != nil && arcAddit.ForbidReco != nil && arcAddit.ForbidReco.State == 1 {
		log.Warn("checkArcStatus no relates aid(%d) arcAddit.ForbidReco.State(%d)", aid, arcAddit.ForbidReco.State)
		return errors.Wrapf(err, "稿件审核屏蔽 error aid(%d) arcAddit.ForbidReco.State(%d)", aid, arcAddit.ForbidReco.State)
	}
	return nil
}

func (s *Service) getAIRecommend(c context.Context, req *view.RecommendReq) *relates.AIRecommendResponse {
	var (
		rls                        []*view.Relate
		pkCode                     int
		advNum, adCode, returnCode string
		advertNew                  *advo.SunspotAdReplyForView
		userFeature                string
		pvFeature                  json.RawMessage
		playParam                  int
		err                        error
		relateConf                 *view.RelateConf
	)
	relatesInfoc := &view.RelatesInfoc{}
	relatesInfoc.SetAdCode("NULL")
	if req.Mid > 0 || req.Buvid != "" {
		rls, advNum, playParam, pkCode, userFeature, returnCode, pvFeature, relateConf, advertNew, _, adCode, err = s.newRcmdRelateV2(c, req)
		if err != nil {
			log.Error("getAIRecommend s.newRcmdRelateV2 error, req(%+v) error(%+v)", req, err)
		}
		//设置pk_code
		pkDesc, ok := view.PkCode[pkCode] //pk_code描述，用于prom
		if !ok {
			log.Error("getAIRecommend pk_code desc is not exist(%d),(%+v)", pkCode, view.PkCode)
		}
		relatesInfoc.SetPKCode(pkDesc)
		relatesInfoc.SetAdCode(adCode)
		relatesInfoc.SetAdNum(advNum)
	}

	//记录无相关推荐数据发生的页数
	s.prom.Incr("relates_return_code_" + returnCode)
	if len(rls) == 0 {
		if req.PageIndex < _noRelatesPage {
			s.prom.Incr("没有任何相关推荐-" + strconv.FormatInt(req.PageIndex, 10))
		}
		relatesInfoc.SetPKCode(view.AdFirstForRelate0)
	}

	//如果出现广告
	if advertNew != nil {
		log.Error("日志告警 相关推荐无限下拉出现广告 req(%+v), advertNew(%+v)", req, advertNew)
	}

	res := &relates.AIRecommendResponse{
		Relates:     rls,
		RelateInfoc: relatesInfoc,
		ReturnCode:  returnCode,
		UserFeature: userFeature,
		PvFeature:   pvFeature,
		PlayParam:   playParam,
	}
	if relateConf != nil && relateConf.Next != "" {
		res.Next = relateConf.Next
	}
	return res
}

// 兜底过滤相关推荐内重复aid且不和当前播放页aid重复 注意番剧游戏等无aid
func (s *Service) filterDupRelates(aid int64, relates []*view.Relate) []*view.Relate {
	if len(relates) == 0 {
		return nil
	}
	var (
		newRelates []*view.Relate
		aids       = make(map[int64]struct{}, len(relates))
	)
	aids[aid] = struct{}{}
	for _, r := range relates {
		//无aid的一般是运营商业卡，不过滤
		if r.Aid <= 0 {
			newRelates = append(newRelates, r)
			continue
		}
		if _, ok := aids[r.Aid]; ok {
			s.prom.Incr("RelatesFeed:重复的aid")
			log.Warn("filterDupRelates relate duplicated aid:%d, relate aid:%d", aid, r.Aid)
			continue
		}
		aids[r.Aid] = struct{}{}
		newRelates = append(newRelates, r)
	}
	return newRelates
}

func (s *Service) buildAiRecommendParams(req *relates.RelatesFeedGRPCRequest, relateCmd string) (*view.RecommendReq, error) {
	aiRecommendReq := &view.RecommendReq{
		Aid:         req.Aid,
		Mid:         req.Mid,
		Build:       int(req.Build),
		Buvid:       req.Buvid,
		SourcePage:  req.From,
		TrackId:     req.TrackId,
		Cmd:         relateCmd,
		Plat:        req.Plat,
		MobileApp:   req.MobileApp,
		Network:     req.Network,
		AdExp:       1,
		Device:      req.Device,
		RequestType: "wise",
		DisableRcmd: req.DisableRcmd,
		SessionId:   req.SessionId,
		RecStyle:    1,
		FromSpmid:   req.FromSpmid,
		Spmid:       req.Spmid,
		RefreshNum:  req.RefreshNum,
	}
	//无限下滑默认是2，表示相关推荐列表点击
	if relateCmd == model.RelateCmd {
		aiRecommendReq.SourcePage = "2"
	}

	//获取页码
	pi, err := s.getRelatePageIndex(req.PageIndex, req.Pagination)
	if err != nil {
		return nil, err
	}
	aiRecommendReq.PageIndex = pi

	return aiRecommendReq, nil
}

// getPageIndex 处理分页, 旧方式使用RelatesPage, 标准化使用Pagination分页参数
func (s *Service) getRelatePageIndex(relatesPage int64, page *pageApi.Pagination) (int64, error) {
	if page == nil {
		return relatesPage, nil
	}
	s.prom.Incr("相关推荐-新分页参数")
	pi, err := pagination.TokenGeneratorWithSalt(model.PaginationTokenSalt).ParsePageToken(page.Next)
	if err != nil {
		log.Error("日志告警 pagination ParsePageToken error(%+v)", err)
		return 0, err
	}
	//相关推荐页面从1开始
	if pi == 0 {
		pi = pi + 1
	}
	return pi, nil
}
