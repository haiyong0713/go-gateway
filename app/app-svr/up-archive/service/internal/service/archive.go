package service

import (
	"context"
	"sync"
	"time"
	"unicode/utf8"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/up-archive/ecode"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/model"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
)

// nolint:gocognit,staticcheck
func (s *Service) ArcPassed(ctx context.Context, req *api.ArcPassedReq) (*api.ArcPassedReply, error) {
	var ok bool
	req.Without, ok = s.convertWithout(req.Without, req.WithoutStaff)
	if ok && req.Order == api.SearchOrder_pubtime {
		total, err := s.arcPassedTotal(ctx, req.Mid, req.Without[0])
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("ArcPassed mid:%d error:%+v", req.Mid, err)
			}
			return nil, err
		}
		start := (req.Pn - 1) * req.Ps
		end := start + req.Ps - 1
		isAsc := req.Sort == "asc"
		aids, err := s.dao.CacheArcPassed(ctx, req.Mid, start, end, isAsc, req.Without[0])
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("ArcPassed CacheArcPassed mid:%d start:%d end:%d isAsc:%v error:%+v", req.Mid, start, end, isAsc, err)
			}
			return nil, err
		}
		res := &api.ArcPassedReply{Total: total}
		if len(aids) == 0 {
			return res, nil
		}
		archives, err := s.archiveGRPC.Arcs(ctx, &arcapi.ArcsRequest{Aids: aids})
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("ArcPassed req:%+v archiveGRPC.Arcs aids:%v error:%+v", req, aids, err)
			}
			return nil, err
		}
		for _, aid := range aids {
			arc, ok := archives.GetArcs()[aid]
			if !ok || arc == nil || !arc.IsNormal() {
				log.Error("日志告警 ArcPassed 稿件数据不对 mid:%d aid:%d", req.Mid, aid)
				continue
			}
			res.Archives = append(res.Archives, model.CopyFromArc(arc))
		}
		nowTs := time.Now().Unix()
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			needReset := func() bool {
				for i, v := range res.Archives {
					if i == 0 {
						continue
					}
					now := v
					pre := res.Archives[i-1]
					if now != nil && pre != nil {
						if (!isAsc && now.PubDate.Time().Unix() > pre.PubDate.Time().Unix()) || (isAsc && now.PubDate.Time().Unix() < pre.PubDate.Time().Unix()) {
							log.Error("日志告警 ArcPassed 顺序不对 mid:%d pn:%d ps:%d total:%d isAsc:%v index:%d aid:%d pubdate:%d preAid:%d prePubdate:%d", req.Mid, req.Pn, req.Ps, res.Total, isAsc, i, now.Aid, now.PubDate.Time().Unix(), pre.Aid, pre.PubDate.Time().Unix())
							return true
						}
					}
				}
				return false
			}()
			if needReset {
				if pubErr := s.dao.SendBuildCacheMsg(ctx, req.Mid, nowTs); pubErr != nil {
					log.Error("ArcPassed upArcPub.Send mid:%d error:%+v", req.Mid, pubErr)
				}
			}
		})
		return res, nil
	}
	searchReply, err := s.arcPassedSearch(ctx, req.Mid, 0, "", nil, false, int(req.Pn), int(req.Ps), req.Order, req.Without, convertSort(req.Sort))
	if err != nil {
		log.Error("ArcPassed req:%+v error:%+v", req, err)
		return nil, err
	}
	reply := &api.ArcPassedReply{}
	var aids []int64
	for _, val := range searchReply.Result {
		aids = append(aids, val.Aid)
	}
	if len(aids) == 0 {
		return reply, nil
	}
	archives, err := s.multiArcs(ctx, aids)
	if err != nil {
		return nil, err
	}
	for _, val := range searchReply.Result {
		arc, ok := archives[val.Aid]
		if !ok || arc == nil || !arc.IsNormal() {
			log.Error("日志告警 ArcPassedSearch 稿件数据不对 mid:%d aid:%d", req.Mid, val.Aid)
			continue
		}
		a := model.CopyFromArc(arc)
		reply.Archives = append(reply.Archives, a)
	}
	if searchReply != nil && searchReply.Page != nil {
		reply.Total = searchReply.Page.Total
	}
	return reply, nil
}

func (s *Service) ArcsPassed(ctx context.Context, req *api.ArcsPassedReq) (*api.ArcsPassedReply, error) {
	var without api.Without
	if req.WithoutStaff {
		without = api.Without_staff
	}
	totalm, err := s.arcsPassedTotal(ctx, req.Mids, without)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			log.Error("ArcsPassed mid:%+d error:%+v", req.Mids, err)
		}
		return nil, err
	}
	start := (req.Pn - 1) * req.Ps
	end := start + req.Ps - 1
	isAsc := req.Sort == "asc"
	var tmpMids []int64
	for mid, total := range totalm {
		if total > 0 {
			tmpMids = append(tmpMids, mid)
		}
	}
	aidsm, err := s.dao.CacheArcsPassed(ctx, tmpMids, start, end, isAsc, without)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			log.Error("ArcsPassed CacheArcsPassed mid:%+d start:%d end:%d isAsc:%v error:%+v", tmpMids, start, end, isAsc, err)
		}
		return nil, err
	}
	var tmpAids []int64
	for _, aids := range aidsm {
		tmpAids = append(tmpAids, aids...)
	}
	archives, err := s.multiArcs(ctx, tmpAids)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			log.Error("ArcsPassed req:%+v multiArcs aids:%v error:%+v", req, tmpMids, err)
		}
		return nil, err
	}
	res := &api.ArcsPassedReply{
		Archives: map[int64]*api.ArcPassedReply{},
	}
	for mid, total := range totalm {
		res.Archives[mid] = &api.ArcPassedReply{Total: total}
	}
	for mid, aids := range aidsm {
		for _, aid := range aids {
			arc, ok := archives[aid]
			if !ok || arc == nil || !arc.IsNormal() {
				log.Error("日志告警 ArcsPassed 稿件数据不对 mid:%d aid:%d", mid, aid)
				continue
			}
			res.Archives[mid].Archives = append(res.Archives[mid].Archives, model.CopyFromArc(arc))
		}
	}
	return res, nil
}

func (s *Service) ArcPassedTotal(ctx context.Context, req *api.ArcPassedTotalReq) (*api.ArcPassedTotalReply, error) {
	var ok bool
	req.Without, ok = s.convertWithout(req.Without, req.WithoutStaff)
	if ok && req.Tid == 0 {
		total, err := s.arcPassedTotal(ctx, req.Mid, req.Without[0])
		if err != nil {
			if ecode.EqualError(ecode.NothingFound, err) {
				return &api.ArcPassedTotalReply{Total: 0}, nil
			}
			log.Error("ArcPassedTotal mid:%d error:%+v", req.Mid, err)
			return nil, err
		}
		return &api.ArcPassedTotalReply{Total: total}, nil
	}
	searchReply, err := s.arcPassedSearch(ctx, req.Mid, req.Tid, "", nil, false, 1, 1, api.SearchOrder_pubtime, req.Without, api.Sort_desc)
	if err != nil {
		log.Error("ArcPassedTotal req:%+v error:%+v", req, err)
		return nil, err
	}
	reply := &api.ArcPassedTotalReply{}
	if searchReply != nil && searchReply.Page != nil {
		reply.Total = searchReply.Page.Total
	}
	return reply, nil
}

func (s *Service) ArcsPassedTotal(ctx context.Context, req *api.ArcsPassedTotalReq) (*api.ArcsPassedTotalReply, error) {
	var without api.Without
	if req.WithoutStaff {
		without = api.Without_staff
	}
	total, err := s.arcsPassedTotal(ctx, req.Mids, without)
	if err != nil {
		if !ecode.EqualError(ecode.NothingFound, err) {
			log.Error("ArcsPassedTotal mid:%+v error:%+v", req.Mids, err)
		}
		return nil, err
	}
	return &api.ArcsPassedTotalReply{Total: total}, nil
}

func (s *Service) arcPassedTotal(ctx context.Context, mid int64, without api.Without) (int64, error) {
	total, err := s.dao.CacheArcPassedTotal(ctx, mid, without)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		// 检查key是否存在，发回源消息
		nowTs := time.Now().Unix()
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			exist, existErr := s.dao.CacheArcPassedExists(ctx, mid, without)
			if existErr != nil {
				log.Error("arcPassedTotal CacheArcPassedExists mid:%d error:%+v", mid, existErr)
				return
			}
			if !exist {
				if pubErr := s.dao.SendBuildCacheMsg(ctx, mid, nowTs); pubErr != nil {
					log.Error("arcPassedTotal upArcPub.Send mid:%d error:%+v", mid, pubErr)
				}
			}
		})
		return 0, ecode.NothingFound
	}
	if total == 1 {
		aids, err := s.dao.CacheArcPassed(ctx, mid, 0, 1, false, without)
		if err != nil {
			return 0, err
		}
		if len(aids) == 1 && aids[0] == -1 {
			// expire empty cache
			_ = s.cache.Do(ctx, func(ctx context.Context) {
				if expireErr := s.dao.ExpireEmptyArcPassed(ctx, mid, without); expireErr != nil {
					log.Error("arcPassedTotal ExpireEmptyArcPassed mid:%d error:%+v", mid, expireErr)
				}
			})
			return 0, ecode.NothingFound
		}
	}
	return total, nil
}

func (s *Service) arcsPassedTotal(ctx context.Context, mids []int64, without api.Without) (map[int64]int64, error) {
	totalm, err := s.dao.CacheArcsPassedTotal(ctx, mids, without)
	if err != nil {
		return nil, err
	}
	var tmpMids []int64
	for _mid, total := range totalm {
		mid := _mid
		if total == 0 {
			// 检查key是否存在，发回源消息
			nowTs := time.Now().Unix()
			_ = s.cache.Do(ctx, func(ctx context.Context) {
				exist, existErr := s.dao.CacheArcPassedExists(ctx, mid, without)
				if existErr != nil {
					log.Error("arcsPassedTotal CacheArcPassedExists mid:%d error:%+v", mid, existErr)
					return
				}
				if !exist {
					if pubErr := s.dao.SendBuildCacheMsg(ctx, mid, nowTs); pubErr != nil {
						log.Error("arcsPassedTotal upArcPub.Send mid:%d error:%+v", mid, pubErr)
					}
				}
			})
			continue
		}
		if total == 1 {
			tmpMids = append(tmpMids, mid)
		}
	}
	aidsm, err := s.dao.CacheArcsPassed(ctx, tmpMids, 0, 1, false, without)
	if err != nil {
		return nil, err
	}
	for _mid, aids := range aidsm {
		mid := _mid
		if len(aids) == 1 && aids[0] == -1 {
			// expire empty cache
			_ = s.cache.Do(ctx, func(ctx context.Context) {
				if expireErr := s.dao.ExpireEmptyArcPassed(ctx, mid, without); expireErr != nil {
					log.Error("arcsPassedTotal ExpireEmptyArcPassed mid:%d error:%+v", mid, expireErr)
				}
			})
			totalm[mid] = 0
		}
	}
	return totalm, nil
}

func (s *Service) ArcPassedCursor(ctx context.Context, req *api.ArcPassedCursorReq) (*api.ArcPassedCursorReply, error) {
	var ok bool
	req.Without, ok = s.convertWithout(req.Without, req.WithoutStaff)
	if ok {
		_, err := s.arcPassedTotal(ctx, req.Mid, req.Without[0])
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("ArcPassedCursor mid:%d error:%+v", req.Mid, err)
			}
			return nil, err
		}
		isAsc := req.Sort == "asc"
		if !isAsc && req.Score == 0 {
			req.Score = model.MaxScore()
		}
		list, err := s.dao.CacheArcPassedCursor(ctx, req.Mid, req.Score, req.Ps, isAsc, false, req.Without[0])
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("ArcPassedCursor req:%+v error:%+v", req, err)
			}
			return nil, err
		}
		return &api.ArcPassedCursorReply{List: list}, nil
	}
	if req.Score > time.Now().Unix() {
		return nil, ecode.RequestErr
	}
	reply, err := s.dao.ArcSearchCursor(ctx, req.Mid, req.Score, false, int(req.Ps), req.Without, convertSort(req.Sort))
	if err != nil {
		log.Error("ArcPassedCursor mid:%d error:%+v", req.Mid, err)
		return nil, err
	}
	var list []*api.ArcPassed
	for _, val := range reply.Result {
		pubtime, err := time.ParseInLocation("2006-01-02 15:04:05", val.Pubtime, time.Local)
		if err != nil {
			log.Error("日志告警 转换投稿时间错误 data:%+v,error:%+v", val, err)
			continue
		}
		list = append(list, &api.ArcPassed{Aid: val.Aid, Score: pubtime.Unix()})
	}
	return &api.ArcPassedCursorReply{List: list}, nil
}

func (s *Service) ArcPassedByAid(ctx context.Context, req *api.ArcPassedByAidReq) (*api.ArcPassedByAidReply, error) {
	var ok bool
	req.Without, ok = s.convertWithout(req.Without, req.WithoutStaff)
	if ok && req.Tid == 0 && req.Order == api.SearchOrder_pubtime {
		return s.cacheArcPassedByAid(ctx, req.Mid, req.Aid, req.Ps, convertSort(req.Sort), req.Without[0])
	}
	var (
		reply       []*model.ArcsCursorAidResult
		left, total int64
		backupReply *api.ArcPassedByAidReply
		backupErr   error
	)
	degrade := s.ac.Search.Degrade
	res := &api.ArcPassedByAidReply{Cursor: &api.CursorReply{}}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		var err error
		reply, left, err = s.arcPassedByAid(ctx, req)
		return err
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.ArcSearch(ctx, req.Mid, req.Tid, "", nil, false, 1, 1, api.SearchOrder_pubtime, req.Without, api.Sort_desc)
		if err != nil {
			return err
		}
		if reply == nil || reply.Page == nil {
			return nil
		}
		total = reply.Page.Total
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if !degrade {
			return nil
		}
		var err error
		if backupReply, backupErr = s.cacheArcPassedByAid(ctx, req.Mid, req.Aid, req.Ps, convertSort(req.Sort), api.Without_no_space); err != nil {
			log.Error("ArcPassedByAid req:%+v error:%+v", req, err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("ArcPassedByAid req:%+v error:%+v", req, err)
		if degrade {
			log.Warn("ArcPassedByAid 降级成功,req:%+v error:%+v", req, backupErr)
			return backupReply, backupErr
		}
		return nil, err
	}
	res.Total = total
	if len(reply) == 0 {
		return res, nil
	}
	headRank := total - left
	var tailRank int64
	for i, val := range reply {
		tailRank = headRank + int64(i)
		res.Archives = append(res.Archives, &api.ArcPassedWithIndex{
			Archive: &api.ArcPassed{Aid: val.Aid, Score: val.Score},
			Rank:    tailRank,
		})
	}
	// 尾不等于total-1
	if tailRank != total-1 {
		res.Cursor.HasMore = true
	}
	return res, nil
}

func (s *Service) cacheArcPassedByAid(ctx context.Context, mid, aid, ps int64, sort api.Sort, without api.Without) (*api.ArcPassedByAidReply, error) {
	total, err := s.arcPassedTotal(ctx, mid, without)
	if err != nil {
		log.Error("cacheArcPassedByAid mid:%d,aid:%d,ps:%d,sort:%v,without:%v error:%+v", mid, aid, ps, sort, without, err)
		return nil, err
	}
	isAsc := sort == api.Sort_asc
	var score, headRank int64
	if aid != 0 {
		if score, headRank, err = s.dao.CacheArcPassedScoreRank(ctx, mid, aid, isAsc, without); err != nil {
			log.Error("cacheArcPassedByAid mid:%d,aid:%d,ps:%d,sort:%v,without:%v error:%+v", mid, aid, ps, sort, without, err)
			if err == redis.ErrNil {
				return nil, xecode.InvalidAid
			}
			return nil, err
		}
	}
	if !isAsc && score == 0 {
		score = model.MaxScore()
	}
	// 包含score的稿件，这个逻辑来着于迁移接口的逻辑
	reply, err := s.dao.CacheArcPassedCursor(ctx, mid, score, ps, isAsc, true, without)
	if err != nil {
		log.Error("cacheArcPassedByAid mid:%d,aid:%d,ps:%d,sort:%v,without:%v error:%+v", mid, aid, ps, sort, without, err)
		return nil, err
	}
	var (
		tailRank int64
		arcs     []*api.ArcPassedWithIndex
	)
	for i, arc := range reply {
		tailRank = headRank + int64(i)
		res := &api.ArcPassedWithIndex{
			Archive: arc,
			Rank:    tailRank,
		}
		arcs = append(arcs, res)
	}
	cursor := &api.CursorReply{}
	// 尾不等于total-1
	if tailRank != total-1 {
		cursor.HasMore = true
	}
	return &api.ArcPassedByAidReply{
		Archives: arcs,
		Total:    total,
		Cursor:   cursor,
	}, nil
}

func (s *Service) arcPassedByAid(ctx context.Context, req *api.ArcPassedByAidReq) ([]*model.ArcsCursorAidResult, int64, error) {
	sort := convertSort(req.Sort)
	if req.Aid == 0 {
		reply, err := s.dao.ArcSearchCursorAid(ctx, req.Mid, 0, false, 0, req.Tid, int(req.Ps), req.Order, req.Without, sort)
		if err != nil {
			log.Error("arcPassedByAid req:%+v error:%+v", req, err)
			return nil, 0, err
		}
		var left int64
		if reply != nil && reply.Page != nil {
			left = reply.Page.Total
		}
		return reply.Result, left, nil
	}
	scoreReply, err := s.dao.ArcSearchScore(ctx, req.Mid, req.Aid, req.Tid, req.Order, req.Without)
	if err != nil {
		log.Error("arcPassedByAid req:%+v error:%+v", req, err)
		return nil, 0, err
	}
	if scoreReply == nil {
		log.Error("arcPassedByAid req:%+v scoreReply is nil", req)
		return nil, 0, xecode.InvalidAid
	}
	score := scoreReply.Score
	var (
		equalReply, rangeReply []*model.ArcsCursorAidResult
		equalLeft, rangeLeft   int64
	)
	eg := errgroup.WithCancel(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.ArcSearchCursorAid(ctx, req.Mid, score, true, req.Aid, req.Tid, int(req.Ps), req.Order, req.Without, sort)
		if err != nil {
			log.Error("arcPassedByAid req:%+v error:%+v", req, err)
			return err
		}
		if reply != nil {
			equalReply = reply.Result
			if reply.Page != nil {
				equalLeft = reply.Page.Total
			}
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.ArcSearchCursorAid(ctx, req.Mid, score, false, req.Aid, req.Tid, int(req.Ps), req.Order, req.Without, sort)
		if err != nil {
			log.Error("arcPassedByAid req:%+v error:%+v", req, err)
			return err
		}
		if reply != nil {
			rangeReply = reply.Result
			if reply.Page != nil {
				rangeLeft = reply.Page.Total
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("arcPassedByAid req:%+v error:%+v", req, err)
		return nil, 0, err
	}
	reply := append(equalReply, rangeReply...)
	if len(reply) > int(req.Ps) {
		reply = reply[:int(req.Ps)]
	}
	left := equalLeft + rangeLeft
	return reply, left, nil
}

func (s *Service) UpsPassed(ctx context.Context, req *api.UpsArcsReq) (*api.UpsAidPubTimeReply, error) {
	var without api.Without
	if req.WithoutStaff {
		without = api.Without_staff
	}
	totalm, err := s.arcsPassedTotal(ctx, req.Mids, without)
	if err != nil {
		log.Error("UpsPassed mid:%+d error:%+v", req.Mids, err)
		return nil, err
	}
	start := (req.Pn - 1) * req.Ps
	end := start + req.Ps - 1
	var tmpMids []int64
	for mid, total := range totalm {
		if total > 0 {
			tmpMids = append(tmpMids, mid)
		}
	}
	upsm, err := s.dao.CacheUpsPassed(ctx, tmpMids, start, end, false, without)
	if err != nil {
		log.Error("UpsPassed CacheUpsPassed mid:%+d start:%d end:%d isAsc:%v error:%+v", tmpMids, start, end, false, err)
		return nil, err
	}
	arcm := map[int64]*api.UpAidPubTimeReply{}
	for mid, ups := range upsm {
		arcs := &api.UpAidPubTimeReply{
			Archives: ups,
		}
		arcm[mid] = arcs
	}
	return &api.UpsAidPubTimeReply{
		Archives: arcm,
	}, nil
}

func (s *Service) ArcPassedSearch(ctx context.Context, req *api.ArcPassedSearchReq) (*api.ArcPassedSearchReply, error) {
	var oneWord bool
	kwLen := utf8.RuneCountInString(req.Keyword)
	for _, val := range s.ac.Search.OneWordLens { // 几个字以内用 title.item 做单字分词
		if kwLen == val {
			oneWord = true
			break
		}
	}
	var kwFields []string
	for _, val := range req.KwFields {
		if oneWord && val == api.KwField_title {
			kwFields = append(kwFields, "title.item")
			continue
		}
		kwFields = append(kwFields, val.String())
	}
	if len(kwFields) == 0 {
		for _, val := range api.KwField_name {
			if oneWord && val == "title" {
				kwFields = append(kwFields, "title.item")
				continue
			}
			kwFields = append(kwFields, val)
		}
	}
	var (
		searchReply *model.ArcSearchReply
		tagReply    map[int64]int64
	)
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		searchReply, err = s.arcPassedSearch(ctx, req.Mid, req.Tid, req.Keyword, kwFields, req.Highlight, int(req.Pn), int(req.Ps), req.Order, req.Without, convertSort(req.Sort))
		return err
	})
	if req.HasTags {
		g.Go(func(ctx context.Context) error {
			var err error
			if tagReply, err = s.dao.ArcSearchTag(ctx, req.Mid, req.Keyword, kwFields, req.Without); err != nil {
				log.Error("ArcPassedSearch req:%+v error:%+v", req, err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	reply := &api.ArcPassedSearchReply{}
	if searchReply != nil && searchReply.Page != nil {
		reply.Total = searchReply.Page.Total
	}
	var aids []int64
	for _, val := range searchReply.Result {
		aids = append(aids, val.Aid)
	}
	if len(aids) == 0 {
		return reply, nil
	}
	archives, err := s.multiArcs(ctx, aids)
	if err != nil {
		return nil, err
	}
	for _, val := range searchReply.Result {
		arc, ok := archives[val.Aid]
		if !ok || arc == nil || !arc.IsNormal() {
			log.Error("日志告警 ArcPassedSearch 稿件数据不对 mid:%d aid:%d", req.Mid, val.Aid)
			continue
		}
		a := model.CopyFromArc(arc)
		// 高亮
		h := val.Highlight
		if h != nil {
			if h.Title != "" {
				a.Title = h.Title
			}
			if h.Content != "" {
				a.Desc = h.Content
			}
		}
		reply.Archives = append(reply.Archives, a)
	}
	for tid, count := range tagReply {
		tag, ok := s.types[int32(tid)]
		if !ok {
			log.Error("日志告警 ArcPassedSearch 标签数据不对 mid:%d tid:%d", req.Mid, tid)
			continue
		}
		reply.Tags = append(reply.Tags, &api.Tag{Tid: tid, Name: tag.GetName(), Count: count})
	}
	return reply, nil
}

func (s *Service) arcPassedSearch(ctx context.Context, mid, tid int64, keyword string, kwFields []string, highlight bool, pn, ps int, order api.SearchOrder, without []api.Without, sort api.Sort) (*model.ArcSearchReply, error) {
	var (
		reply, backupReply *model.ArcSearchReply
		backupErr          error
	)
	degrade := s.ac.Search.Degrade
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		searchReply, err := s.dao.ArcSearch(ctx, mid, tid, keyword, kwFields, highlight, pn, ps, order, without, sort)
		if err != nil {
			return err
		}
		if searchReply == nil {
			return nil
		}
		reply = &model.ArcSearchReply{Page: searchReply.Page}
		for _, val := range searchReply.Result {
			r := &model.ArcsResult{
				Aid:       val.Aid,
				Highlight: val.Highlight,
			}
			reply.Result = append(reply.Result, r)
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		if !degrade {
			return nil
		}
		without := api.Without_no_space
		total, err := s.arcPassedTotal(ctx, mid, without)
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("arcPassedSearch mid:%d,without:%v error:%+v", mid, without, err)
			}
			backupErr = err
			return nil
		}
		start := (pn - 1) * ps
		end := start + ps - 1
		isAsc := sort == api.Sort_asc
		aids, err := s.dao.CacheArcPassed(ctx, mid, int64(start), int64(end), isAsc, without)
		if err != nil {
			log.Error("arcPassedSearch mid:%d,start:%d,end:%d,isAsc:%v,without:%v error:%+v", mid, start, end, isAsc, without, err)
			backupErr = err
			return nil
		}
		backupReply = &model.ArcSearchReply{Page: &model.Page{Total: total}}
		for _, val := range aids {
			r := &model.ArcsResult{
				Aid: val,
			}
			backupReply.Result = append(backupReply.Result, r)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("arcPassedSearch mid:%d,tid:%d,keyword:%s,kwFields:%+v,highlight:%v,pn:%d,ps:%d,order:%v,without:%v,sort:%v error:%+v", mid, tid, keyword, kwFields, highlight, pn, ps, order, without, sort, err)
		if degrade {
			log.Warn("arcPassedSearch 降级成功,mid:%d,tid:%d,keyword:%s,kwFields:%+v,highlight:%v,pn:%d,ps:%d,order:%v,without:%v,sort:%v error:%+v", mid, tid, keyword, kwFields, highlight, pn, ps, order, without, sort, backupErr)
			return backupReply, backupErr
		}
		return nil, err
	}
	return reply, nil
}

func (s *Service) ArcsPassedSort(ctx context.Context, req *api.ArcsPassedSortReq) (*api.ArcsPassedSortReply, error) {
	if s.searchGray(req.Mids[0]) {
		return s.arcsPassedSort(ctx, req)
	}
	searchReply, err := s.dao.ArcsSearchSort(ctx, req.Mids, req.Tid, int(req.Ps), req.Order, req.Sort)
	if err != nil {
		log.Error("ArcsPassedSort req:%+v error:%+v", req, err)
		return nil, err
	}
	reply := &api.ArcsPassedSortReply{
		Archives: map[int64]*api.ArcPassedSortReply{},
	}
	for mid, aids := range searchReply {
		var arcs []*api.SortArc
		for _, aid := range aids {
			arcs = append(arcs, &api.SortArc{Aid: aid})
		}
		reply.Archives[mid] = &api.ArcPassedSortReply{Archive: arcs}
	}
	return reply, nil
}

func (s *Service) ArcPassedExist(ctx context.Context, req *api.ArcPassedExistReq) (*api.ArcPassedExistReply, error) {
	var ok bool
	req.Without, ok = s.convertWithout(req.Without, false)
	if ok && req.Tid == 0 && req.Order == api.SearchOrder_pubtime {
		exist, err := s.dao.CacheArcPassedExist(ctx, req.Mid, req.Aid, req.Without[0])
		if err != nil {
			log.Error("ArcPassedExist req:%+v error:%+v", req, err)
			return nil, err
		}
		return &api.ArcPassedExistReply{
			Exist: exist,
		}, nil
	}
	scoreReply, err := s.dao.ArcSearchScore(ctx, req.Mid, req.Aid, req.Tid, req.Order, req.Without)
	if err != nil {
		log.Error("ArcPassedExist req:%+v error:%+v", req, err)
		return nil, err
	}
	return &api.ArcPassedExistReply{
		Exist: scoreReply != nil,
	}, nil
}

func (s *Service) searchGray(mid int64) bool {
	return mid%100 < s.ac.SearchGray.Bucket
}

func (s *Service) arcsPassedSort(ctx context.Context, req *api.ArcsPassedSortReq) (*api.ArcsPassedSortReply, error) {
	reply := &api.ArcsPassedSortReply{
		Archives: map[int64]*api.ArcPassedSortReply{},
	}
	midm := map[int64]struct{}{}
	for _, mid := range req.Mids {
		midm[mid] = struct{}{}
	}
	var mutex sync.Mutex
	eg := errgroup.WithContext(ctx)
	for mid := range midm {
		tmpMid := mid
		eg.Go(func(ctx context.Context) error {
			searchReply, err := s.dao.ArcSearch(ctx, tmpMid, req.Tid, "", nil, false, 1, int(req.Ps), req.Order, []api.Without{api.Without_staff}, req.Sort)
			if err != nil {
				log.Error("ArcsPassedSort req:%+v error:%+v", req, err)
				return nil
			}
			if searchReply == nil || len(searchReply.Result) == 0 {
				return nil
			}
			var arcs []*api.SortArc
			for _, val := range searchReply.Result {
				arcs = append(arcs, &api.SortArc{Aid: val.Aid})
			}
			mutex.Lock()
			reply.Archives[tmpMid] = &api.ArcPassedSortReply{Archive: arcs}
			mutex.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) multiArcs(ctx context.Context, aids []int64) (map[int64]*arcapi.Arc, error) {
	const _count = 100
	var shard int
	if len(aids) < _count {
		shard = 1
	} else {
		shard = len(aids) / _count
		if len(aids)%(shard*_count) != 0 {
			shard++
		}
	}
	aidss := make([][]int64, shard)
	for i, aid := range aids {
		aidss[i%shard] = append(aidss[i%shard], aid)
	}
	arcms := make([]map[int64]*arcapi.Arc, len(aidss))
	g := errgroup.WithCancel(ctx)
	for idx, aids := range aidss {
		if len(aids) == 0 {
			continue
		}
		tmpIdx, tmpAids := idx, aids
		g.Go(func(ctx context.Context) error {
			arcs, err := s.archiveGRPC.Arcs(ctx, &arcapi.ArcsRequest{Aids: tmpAids})
			if err != nil {
				return err
			}
			arcms[tmpIdx] = arcs.GetArcs()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	res := map[int64]*arcapi.Arc{}
	for _, arcm := range arcms {
		for aid, arc := range arcm {
			res[aid] = arc
		}
	}
	return res, nil
}

func (s *Service) Types(ctx context.Context) (map[int32]*arcapi.Tp, error) {
	reply, err := s.archiveGRPC.Types(ctx, &arcapi.NoArgRequest{})
	if err != nil {
		return nil, err
	}
	return reply.GetTypes(), nil
}

func (s *Service) convertWithout(withoutReq []api.Without, withoutStaffReq bool) (without []api.Without, ok bool) {
	if withoutStaffReq {
		withoutReq = append(withoutReq, api.Without_staff)
	}
	fm := make(map[api.Without]struct{})
	for _, val := range withoutReq {
		if val == api.Without_none {
			continue
		}
		if _, ok := fm[val]; ok {
			continue
		}
		without = append(without, val)
		fm[val] = struct{}{}
	}
	if len(without) == 0 {
		without = append(without, api.Without_none)
	}
	if len(without) == 1 {
		switch without[0] {
		case api.Without_none, api.Without_staff, api.Without_no_space:
			return without, true
		default:
		}
	}
	return without, false
}

func convertSort(sort string) api.Sort {
	new := api.Sort_desc
	if sort == "asc" {
		new = api.Sort_asc
	}
	return new
}
