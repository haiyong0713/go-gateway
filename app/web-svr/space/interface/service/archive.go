package service

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/model"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	uparcmdl "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

const (
	_checkTypeChannel = "channel"
)

var (
	_emptyArchiveReason = make([]*model.ArchiveReason, 0)
	_emptySearchVList   = make([]*model.SearchVList, 0)
)

// UpArcStat get up all article stat.
func (s *Service) UpArcStat(c context.Context, mid int64) (data *model.UpArcStat, err error) {
	addCache := true
	if data, err = s.dao.UpArcCache(c, mid); err != nil {
		addCache = false
	} else if data != nil {
		return
	}
	dt := time.Now().AddDate(0, 0, -1).Add(-12 * time.Hour).Format("20060102")
	if data, err = s.dao.UpArcStat(c, mid, dt); data != nil && addCache {
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.SetUpArcCache(c, mid, data)
		})
	}
	return
}

// TopArc get top archive.
// nolint:gomnd
func (s *Service) TopArc(c context.Context, mid, vmid int64) (res *model.ArchiveReason, err error) {
	var (
		topArc   *model.AidReason
		arcReply *arcmdl.ArcReply
	)
	if topArc, err = s.dao.TopArc(c, vmid); err != nil {
		return
	}
	if topArc == nil || topArc.Aid == 0 {
		err = ecode.SpaceNoTopArc
		return
	}
	if arcReply, err = s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: topArc.Aid}); err != nil {
		log.Error("TopArc s.arcClient.Arc(%d) error(%v)", topArc.Aid, err)
		return
	}
	arc := arcReply.Arc
	if arc.AttrValV2(arcapi.AttrBitV2OnlyFavView) == arcmdl.AttrYes {
		if err := s.cache.Do(c, func(ctx context.Context) {
			if err = s.DelTopArc(ctx, vmid); err != nil {
				log.Error("日志告警 置顶隐藏仅收藏可见稿件错误,mid:%v,error:%+v", vmid, err)
				return
			}
		}); err != nil {
			log.Error("%+v", err)
		}
		err = ecode.SpaceNoTopArc
		return
	}
	if mid != vmid && !arc.IsNormal() {
		err = ecode.SpaceNoTopArc
		return
	}
	if arc.Author.Mid != vmid {
		//证明访问的是合作人的空间 staff从名单中删除
		if !s.isCooperAuth(vmid, arc) {
			if mid != vmid {
				//客态不展示
				err = ecode.SpaceNoTopArc
				return
			}
			//主人态失效
			arc.State = -1
		}
	}
	if !arc.IsNormal() {
		//稿件状态不正常 则合作稿件标识不展示
		arc.Rights.IsCooperation = 0
	}
	if arc.Access >= 10000 {
		arc.Stat.View = -1
	}
	model.ClearAttrAndAccess(arc)
	res = &model.ArchiveReason{Arc: arc, Bvid: s.avToBv(arc.Aid), Reason: template.HTMLEscapeString(topArc.Reason), InterVideo: arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes}
	return
}

// SetTopArc set top archive.
func (s *Service) SetTopArc(c context.Context, mid, aid int64, reason string) (err error) {
	eg := errgroup.Group{}
	eg.Go(func(ctx context.Context) error {
		arcReply, e := s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if e != nil || arcReply.Arc == nil {
			log.Error("SetTopArc s.arcClient.Arc mid(%d) aid(%d) error(%v)", mid, aid, e)
			return e
		}
		arc := arcReply.Arc
		if !arc.IsNormal() {
			return ecode.SpaceFakeAid
		}
		if arc.Author.Mid != mid && !s.isCooperAuth(mid, arc) {
			return ecode.SpaceNotAuthor
		}
		if arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrYes {
			return ecode.SpacePayUGV
		}
		if arc.AttrValV2(arcapi.AttrBitV2OnlyFavView) == arcapi.AttrYes {
			return ecode.ForbitTop
		}
		return nil
	})
	if reason != "" {
		eg.Go(func(ctx context.Context) error {
			e := s.Filter(c, []string{reason})
			if e != nil {
				return e
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		topArc, e := s.dao.TopArc(c, mid)
		if e != nil {
			return e
		}
		if topArc != nil && aid == topArc.Aid && reason == topArc.Reason {
			return xecode.NotModified
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if err = s.dao.AddTopArc(c, mid, aid, reason); err == nil {
		if err := s.dao.AddCacheTopArc(c, mid, &model.AidReason{Aid: aid, Reason: reason}); err != nil {
			log.Error("%+v", err)
		}
	}
	return
}

// DelTopArc delete top archive.
func (s *Service) DelTopArc(c context.Context, mid int64) (err error) {
	var topArc *model.AidReason
	if topArc, err = s.dao.TopArc(c, mid); err != nil {
		return
	}
	if topArc == nil {
		err = xecode.RequestErr
		return
	}
	if err = s.dao.DelTopArc(c, mid); err == nil {
		if err := s.dao.AddCacheTopArc(c, mid, &model.AidReason{Aid: -1}); err != nil {
			log.Error("%+v", err)
		}
	}
	return
}

// Masterpiece get masterpiece.
// nolint:gomnd
func (s *Service) Masterpiece(c context.Context, mid, vmid int64) (res []*model.ArchiveReason, err error) {
	var (
		mps       *model.AidReasons
		arcsReply *arcmdl.ArcsReply
		aids      []int64
	)
	if mps, err = s.dao.Masterpiece(c, vmid); err != nil {
		return
	}
	if mps == nil || len(mps.List) == 0 {
		res = _emptyArchiveReason
		return
	}
	for _, v := range mps.List {
		aids = append(aids, v.Aid)
	}
	if arcsReply, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("Masterpiece s.arcClient.Arcs(%v) error(%v)", aids, err)
		return
	}
	for _, v := range mps.List {
		if arc, ok := arcsReply.Arcs[v.Aid]; ok && arc != nil {
			if arc.AttrValV2(arcapi.AttrBitV2OnlyFavView) == arcapi.AttrYes {
				tmpAid := v.Aid
				if err := s.cache.Do(c, func(ctx context.Context) {
					if err = s.CancelMasterpiece(ctx, vmid, tmpAid); err != nil {
						log.Error("日志告警 代表作隐藏仅收藏可见稿件错误,mid:%v,aid:%v,error:%+v", vmid, tmpAid, err)
					}
				}); err != nil {
					log.Error("%+v", err)
				}
				continue
			}
			if !arc.IsNormal() && mid != vmid {
				continue
			}
			if arc.Access >= 10000 {
				arc.Stat.View = -1
			}
			//staff从名单中删除
			if arc.Author.Mid != vmid {
				if !s.isCooperAuth(vmid, arc) {
					if mid != vmid {
						//客态不展示
						continue
					}
					//主人态失效
					arc.State = -1
				}
			}
			if !arc.IsNormal() {
				//稿件状态不正常 则合作稿件标识不展示
				arc.Rights.IsCooperation = 0
			}
			model.ClearAttrAndAccess(arc)
			res = append(res, &model.ArchiveReason{Arc: arc, Bvid: s.avToBv(arc.Aid), Reason: template.HTMLEscapeString(v.Reason), InterVideo: arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes})
		}
	}
	if len(res) == 0 {
		res = _emptyArchiveReason
	}
	return
}

// AddMasterpiece add masterpiece.
func (s *Service) AddMasterpiece(c context.Context, mid, aid int64, reason string) (err error) {
	var (
		mps *model.AidReasons
	)
	eg := errgroup.Group{}
	eg.Go(func(ctx context.Context) error {
		mpsTmp, e := s.dao.Masterpiece(c, mid)
		if e != nil {
			return e
		}
		if mpsTmp == nil {
			mpsTmp = &model.AidReasons{}
		}
		mpLen := len(mpsTmp.List)
		if mpLen >= s.c.Rule.MaxMpLimit {
			return ecode.SpaceMpMaxCount
		}
		for _, v := range mpsTmp.List {
			if v.Aid == aid {
				return ecode.SpaceMpExist
			}
		}
		mps = mpsTmp
		return nil
	})
	if reason != "" {
		eg.Go(func(ctx context.Context) error {
			e := s.Filter(c, []string{reason})
			if e != nil {
				return e
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		arcReply, e := s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if e != nil || arcReply.Arc == nil {
			log.Error("AddMasterpiece s.arcClient.Arc(%d) error(%v)", aid, e)
			return e
		}
		arc := arcReply.Arc
		if !arc.IsNormal() {
			return ecode.SpaceFakeAid
		}
		if arc.Author.Mid != mid && !s.isCooperAuth(mid, arc) {
			return ecode.SpaceNotAuthor
		}
		if arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrYes {
			return ecode.SpacePayUGV
		}
		if arc.AttrValV2(arcapi.AttrBitV2OnlyFavView) == arcapi.AttrYes {
			return ecode.ForbitMasterpiece
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if err = s.dao.AddMasterpiece(c, mid, aid, reason); err == nil {
		mps.List = append(mps.List, &model.AidReason{Aid: aid, Reason: reason})
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.AddCacheMasterpiece(c, mid, mps)
		})
	}
	return
}

func (s *Service) isCooperAuth(mid int64, a *arcmdl.Arc) bool {
	if a.AttrVal(arcmdl.AttrBitIsCooperation) != arcmdl.AttrYes {
		return false
	}
	for _, v := range a.StaffInfo {
		if v.Mid == mid {
			return true
		}
	}
	return false
}

// EditMasterpiece edit masterpiece.
func (s *Service) EditMasterpiece(c context.Context, mid, preAid, aid int64, reason string) (err error) {
	var (
		preCheck bool
		mps      *model.AidReasons
	)
	eg := errgroup.Group{}
	eg.Go(func(ctx context.Context) error {
		mpsTmp, mpsErr := s.dao.Masterpiece(c, mid)
		if mpsErr != nil {
			return mpsErr
		}
		if mpsTmp == nil || len(mpsTmp.List) == 0 {
			return ecode.SpaceMpNoArc
		}
		for _, v := range mpsTmp.List {
			if v.Aid == preAid {
				preCheck = true
			}
			if v.Aid == aid {
				return ecode.SpaceMpExist
			}
		}
		if !preCheck {
			return ecode.SpaceMpNoArc
		}
		mps = mpsTmp
		return nil
	})
	if reason != "" {
		eg.Go(func(ctx context.Context) (err error) {
			if err = s.Filter(c, []string{reason}); err != nil {
				return err
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		arcReply, arcErr := s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if arcErr != nil || arcReply.Arc == nil {
			log.Error("AddMasterpiece s.arcClient.Arc(%d) error(%v)", aid, err)
			return arcErr
		}
		arc := arcReply.Arc
		if !arc.IsNormal() {
			return ecode.SpaceFakeAid
		}
		if arc.Author.Mid != mid && !s.isCooperAuth(mid, arc) {
			return ecode.SpaceNotAuthor
		}
		if arc.AttrVal(arcapi.AttrBitIsPUGVPay) == arcapi.AttrYes {
			return ecode.SpacePayUGV
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	if err = s.dao.EditMasterpiece(c, mid, aid, preAid, reason); err == nil {
		newAidReasons := &model.AidReasons{}
		for _, v := range mps.List {
			if v.Aid == preAid {
				newAidReasons.List = append(newAidReasons.List, &model.AidReason{Aid: aid, Reason: reason})
			} else {
				newAidReasons.List = append(newAidReasons.List, v)
			}
		}
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.AddCacheMasterpiece(c, mid, newAidReasons)
		})
	}
	return
}

// CancelMasterpiece delete masterpiece.
func (s *Service) CancelMasterpiece(c context.Context, mid, aid int64) (err error) {
	var (
		mps        *model.AidReasons
		existCheck bool
	)
	if mps, err = s.dao.Masterpiece(c, mid); err != nil {
		return
	}
	if mps == nil || len(mps.List) == 0 {
		err = ecode.SpaceMpNoArc
		return
	}
	for _, v := range mps.List {
		if v.Aid == aid {
			existCheck = true
			break
		}
	}
	if !existCheck {
		err = ecode.SpaceMpNoArc
		return
	}
	if err = s.dao.DelMasterpiece(c, mid, aid); err == nil {
		newAidReasons := &model.AidReasons{}
		for _, v := range mps.List {
			if v.Aid == aid {
				continue
			}
			newAidReasons.List = append(newAidReasons.List, v)
		}
		if len(newAidReasons.List) == 0 {
			newAidReasons.List = append(newAidReasons.List, &model.AidReason{Aid: -1})
		}
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.AddCacheMasterpiece(c, mid, newAidReasons)
		})
	}
	return
}

// UpArcs get upload archive .
func (s *Service) UpArcs(c context.Context, mid int64, pn, ps int32) (*model.UpArc, error) {
	res := &model.UpArc{List: []*model.ArcItem{}}
	reply, err := s.upArcClient.ArcPassed(c, &uparcmdl.ArcPassedReq{Mid: mid, Pn: int64(pn), Ps: int64(ps)})
	if err != nil {
		log.Error("UpArcs upArcClient.ArcPassed mid(%d) error(%v)", mid, err)
		return res, nil
	}
	if reply == nil {
		log.Warn("UpArcs upArcClient.ArcPassed mid:%d reply nil", mid)
		return res, nil
	}
	res.Count = reply.Total
	for _, v := range reply.Archives {
		if v == nil {
			continue
		}
		item := &model.ArcItem{}
		item.FromUpArc(v)
		item.Bvid = s.avToBv(v.Aid)
		res.List = append(res.List, item)
	}
	return res, nil
}

// ArcSearch get archive from search.
// nolint:gomnd
func (s *Service) ArcSearch(ctx context.Context, mid int64, arg *model.SearchArg, riskParams *model.RiskManagement) (*model.ArcSearchRes, error) {
	data := &model.ArcSearchRes{
		IsRisk:      false,
		GaiaResType: model.GaiaResponseType_Default,
	}
	riskResult := s.RiskVerifyAndManager(ctx, riskParams)
	if riskResult != nil {
		data.GaiaResType = riskResult.GaiaResType
		data.IsRisk = riskResult.IsRisk
		data.GaiaData = riskResult.GaiaData
		return data, nil
	}
	reply, total, err := s.arcSearch(ctx, arg)
	if err != nil {
		return nil, err
	}
	if reply == nil || len(reply.VList) == 0 {
		return &model.ArcSearchRes{
			List: &model.SearchRes{VList: _emptySearchVList},
			Page: &model.SearchPage{Pn: arg.Pn, Ps: arg.Ps, Count: total}}, nil
	}
	checkAids := make(map[int64]int64)
	if arg.CheckType == _checkTypeChannel {
		if mid == 0 {
			return nil, xecode.RequestErr
		}
		chArcs, err := s.dao.ChannelVideos(ctx, mid, arg.CheckID, false)
		if err != nil {
			log.Error("%+v", err)
		}
		for _, chArc := range chArcs {
			checkAids[chArc.Aid] = chArc.Aid
		}
	}
	var vList []*model.SearchVList
	for _, v := range reply.VList {
		if v.HideClick {
			v.Play = "--"
		}
		if _, ok := checkAids[v.Aid]; !ok {
			v.Bvid = s.avToBv(v.Aid)
			switch created := v.Created.(type) {
			case string:
				if ts, err := time.ParseInLocation("2006-01-02 15:04:05", created, time.Local); err == nil {
					v.Created = ts.Unix()
				}
			}
			vList = append(vList, v)
		}
		lengthStr := strings.Split(v.Length, ":")
		if len(lengthStr) == 2 {
			min := fmt.Sprintf("%02s", lengthStr[0])
			sec := fmt.Sprintf("%02s", lengthStr[1])
			v.Length = min + ":" + sec
		}
	}
	data.List = &model.SearchRes{
		TList: reply.TList,
		VList: vList,
	}
	data.Page = &model.SearchPage{
		Pn:    arg.Pn,
		Ps:    arg.Ps,
		Count: total,
	}
	// 稿件数量大于1时才出按钮
	if len(vList) > 1 && !s.forbidEpisodicButton(arg.Mid) {
		data.EpisodicButton = &model.ArcListButton{
			Text: s.c.PlayButton.Text,
			URI:  fmt.Sprintf(s.c.PlayButton.BaseURI, arg.Mid),
		}
	}
	return data, nil
}

// nolint:gomnd
func (s *Service) arcSearch(ctx context.Context, arg *model.SearchArg) (*model.SearchRes, int64, error) {
	var (
		arcs  []*uparcmdl.Arc
		total int64
		tList map[string]*model.SearchTList
	)
	if err := func() error {
		var order uparcmdl.SearchOrder
		switch arg.Order {
		case "pubdate":
			order = uparcmdl.SearchOrder_pubtime
		case "click":
			order = uparcmdl.SearchOrder_click
		case "stow":
			order = uparcmdl.SearchOrder_fav
		}
		without := []uparcmdl.Without{uparcmdl.Without_no_space}
		if arg.Index == 1 || len(arg.Keyword) == 0 {
			setting := s.privacy(ctx, arg.Mid)
			if setting[model.LivePlayback] == 0 {
				without = append(without, uparcmdl.Without_live_playback)
			}
		}
		if arg.Index == 1 { // 空间tab稿件列表
			req := &uparcmdl.ArcPassedReq{
				Mid:     arg.Mid,
				Pn:      int64(arg.Pn),
				Ps:      int64(arg.Ps),
				Order:   order,
				Without: without,
			}
			reply, err := s.upArcClient.ArcPassed(ctx, req)
			if err != nil {
				return err
			}
			arcs = reply.Archives
			total = reply.Total
			return nil
		}
		req := &uparcmdl.ArcPassedSearchReq{
			Mid:     arg.Mid,
			Tid:     arg.Tid,
			Keyword: arg.Keyword,
			Pn:      int64(arg.Pn),
			Ps:      int64(arg.Ps),
			Order:   order,
			HasTags: true,
			Without: without,
		}
		reply, err := s.upArcClient.ArcPassedSearch(ctx, req)
		if err != nil {
			return err
		}
		arcs = reply.Archives
		total = reply.Total
		tList = make(map[string]*model.SearchTList, len(reply.Tags))
		for _, v := range reply.Tags {
			tList[strconv.FormatInt(v.Tid, 10)] = &model.SearchTList{Tid: v.Tid, Count: v.Count, Name: v.Name}
		}
		return nil
	}(); err != nil {
		log.Error("%+v", err)
		return nil, 0, err
	}
	var vList []*model.SearchVList
	for _, v := range arcs {
		if v == nil {
			continue
		}
		var isPay, isUnionVideo, isSteinGate, isLivePlayback int
		if model.AttrVal(v, arcapi.AttrBitUGCPay) == arcapi.AttrYes {
			isPay = 1
		}
		if model.AttrVal(v, arcapi.AttrBitIsCooperation) == arcapi.AttrYes {
			isUnionVideo = 1
		}
		if model.AttrVal(v, arcapi.AttrBitSteinsGate) == arcapi.AttrYes {
			isSteinGate = 1
		}
		for _, val := range s.c.LivePlayback.UpFrom {
			if v.UpFrom == val {
				isLivePlayback = 1
				break
			}
		}
		lengthH := v.Duration / 60
		lengthM := v.Duration - (lengthH * 60)
		vList = append(vList, &model.SearchVList{
			Comment:        int64(v.Stat.Reply),
			TypeID:         int64(v.TypeID),
			Play:           v.Stat.View,
			Pic:            v.Pic,
			Description:    v.Desc,
			Copyright:      strconv.FormatInt(int64(v.Copyright), 10),
			Title:          v.Title,
			Author:         v.Author.Name,
			Mid:            v.Author.Mid,
			Created:        v.PubDate.Time().Format("2006-01-02 15:04:05"),
			Length:         fmt.Sprintf("%d:%d", lengthH, lengthM),
			VideoReview:    int64(v.Stat.Danmaku),
			Aid:            v.Aid,
			Bvid:           s.avToBv(v.Aid),
			HideClick:      v.Access >= 10000,
			IsPay:          isPay,
			IsUnionVideo:   isUnionVideo,
			IsSteinsGate:   isSteinGate,
			IsLivePlayback: isLivePlayback,
		})
	}
	return &model.SearchRes{VList: vList, TList: tList}, total, nil
}

func (s *Service) forbidEpisodicButton(mid int64) bool {
	if !s.c.PlayButton.Open {
		return true
	}
	for _, v := range s.c.PlayButton.ForbidMids {
		if mid == v {
			return true
		}
	}
	return false
}
