package service

import (
	"context"
	"sort"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/interface/model"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	thumbupmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"go-common/library/sync/errgroup.v2"
)

const (
	_dyTypeCoin          = -1
	_dyTypeLike          = -2
	_dyTypeMerge         = -3
	_businessArchiveLike = "archive"
	//动态服务(文字)
	_businessDynamicLike = "dynamic"
	//动态相簿服务
	_businessDyAlbumLike = "album"
	//动态小视频
	_businessDyclipLike = "clip"
	//付费系列(动态用)
	_businessDyCheeseLike = "cheese"
	_businessArticleLike  = "article"
	_likeVideoCnt         = 100
	_dyDefaultQn          = 16
	_dyInsertIndex        = 3
)

// DynamicList get dynamic list.
func (s Service) DynamicList(c context.Context, arg *model.DyListArg) (dyTotal *model.DyTotal, err error) {
	var (
		list          []*model.DyItem
		actDyItem     *model.DyItem
		dyList        *model.DyList
		topDy         *model.DyCard
		topErr, dyErr error
		hasDy, top    bool
	)
	fp := arg.Pn == 1
	repeatDyIDs := make(map[int64]int64, 1)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if topDy, topErr = s.topDynamic(ctx, arg.Vmid, arg.Qn); topErr == nil && topDy != nil {
			top = fp
			repeatDyIDs[topDy.Desc.DynamicID] = topDy.Desc.DynamicID
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if dyList, dyErr = s.dao.DynamicList(ctx, arg.Mid, arg.Vmid, arg.DyID, arg.Qn, arg.Pn); dyErr != nil {
			log.Error("s.dao.DynamicList(mid:%d,vmid:%d,dyID:%d,qn:%d,pn:%d) error(%+v)", arg.Mid, arg.Vmid, arg.DyID, arg.Qn, arg.Pn, dyErr)
		}
		return nil
	})
	// only show special type in first page
	if fp {
		group.Go(func(ctx context.Context) error {
			actDyItem = s.actDyItem(ctx, arg.Mid, arg.Vmid)
			return nil
		})
	}
	if e := group.Wait(); e != nil {
		log.Error("DynamicList group.Wait mid(%d) error(%v)", arg.Vmid, e)
	}
	// rm repeat data
	if dyErr == nil && dyList != nil && len(dyList.Cards) > 0 {
		for _, v := range dyList.Cards {
			if _, ok := repeatDyIDs[v.Desc.DynamicID]; ok {
				continue
			}
			item := new(model.DyResult)
			item.FromCard(v)
			list = append(list, &model.DyItem{Type: v.Desc.Type, Card: item, Ctime: v.Desc.Timestamp})
		}
		hasDy = dyList.HasMore == 1
	}
	dyTotal = new(model.DyTotal)
	if top {
		topItem := new(model.DyResult)
		topItem.FromCard(topDy)
		dyTotal.List = append(dyTotal.List, &model.DyItem{Type: topDy.Desc.Type, Top: true, Card: topItem, Ctime: topDy.Desc.Timestamp})
	}
	dyTotal.HasMore = hasDy
	dyTotal.List = append(dyTotal.List, list...)
	// first page insert special card to third
	if fp && actDyItem != nil {
		var tmp []*model.DyItem
		if len(dyTotal.List) > _dyInsertIndex {
			tmp = append(tmp, dyTotal.List[:_dyInsertIndex-1]...)
			tmp = append(tmp, actDyItem)
			tmp = append(tmp, dyTotal.List[_dyInsertIndex:]...)
		} else {
			tmp = append(tmp, actDyItem)
		}
		dyTotal.List = tmp
	}
	return
}

// BehaviorList get user coin and like total list.
func (s Service) BehaviorList(c context.Context, mid, vmid, lastTime int64, ps int) (list []*model.DyItem) {
	var (
		coinList, likeList, actList []*model.DyActItem
		aids                        []int64
		coinErr, likeErr            error
		coinPcy, likePcy, pageCheck bool
	)
	group := errgroup.WithContext(c)
	privacy := s.privacy(c, vmid)
	if value, ok := privacy[model.PcyCoinVideo]; ok && value != _defaultPrivacy {
		coinPcy = true
	}
	if value, ok := privacy[model.PcyLikeVideo]; ok && value != _defaultPrivacy {
		likePcy = true
	}
	// coin video
	if mid == vmid || !coinPcy {
		group.Go(func(ctx context.Context) error {
			coinList, coinErr = s.coinVideos(ctx, vmid, coinPcy)
			return nil
		})
	}
	// like video
	if mid == vmid || !likePcy {
		group.Go(func(ctx context.Context) error {
			likeList, likeErr = s.likeVideos(ctx, vmid, _likeVideoCnt, likePcy)
			return nil
		})
	}
	if e := group.Wait(); e != nil {
		log.Error("BehaviorList mid(%d) vmid(%d) group wait error(%v)", mid, vmid, e)
	}
	if coinErr == nil {
		if l := len(coinList); l > 0 {
			actList = append(actList, coinList...)
		}
	}
	if likeErr == nil {
		if l := len(likeList); l > 0 {
			actList = append(actList, likeList...)
		}
	}
	if len(actList) == 0 {
		return
	}
	sort.Slice(actList, func(i, j int) bool { return actList[i].ActionTime > actList[j].ActionTime })
	if s.c.Rule.Merge {
		actList = mergeDyActItem(actList)
	}
	if lastTime > 0 {
		for i, v := range actList {
			if v.ActionTime < lastTime {
				actList = actList[i:]
				pageCheck = true
				break
			}
		}
		if !pageCheck {
			return
		}
	}
	if len(actList) > ps {
		actList = actList[:ps-1]
	}
	for _, v := range actList {
		if v.Aid > 0 {
			aids = append(aids, v.Aid)
		}
	}
	if len(aids) > 0 {
		if arcsReply, e := s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); e != nil {
			log.Error("BehaviorList s.arcClient.Arcs(%v) error(%v)", aids, e)
		} else {
			for _, v := range actList {
				if arc, ok := arcsReply.Arcs[v.Aid]; ok && arc != nil && arc.IsNormal() {
					video := new(model.VideoItem)
					video.FromArchive(arc)
					video.ActionTime = v.ActionTime
					list = append(list, &model.DyItem{Type: v.Type, Archive: video, Ctime: v.ActionTime, Privacy: v.Privacy})
				}
			}
		}
	}
	return
}

func (s *Service) actDyItem(c context.Context, mid, vmid int64) (item *model.DyItem) {
	var (
		coinPcy, likePcy            bool
		likeList, coinList, preList []*model.DyActItem
		coinErr, likeErr            error
	)
	group := errgroup.WithContext(c)
	privacy := s.privacy(c, vmid)
	if value, ok := privacy[model.PcyCoinVideo]; ok && value != _defaultPrivacy {
		coinPcy = true
	}
	if value, ok := privacy[model.PcyLikeVideo]; ok && value != _defaultPrivacy {
		likePcy = true
	}
	// coin video
	if mid == vmid || !coinPcy {
		group.Go(func(ctx context.Context) error {
			coinList, coinErr = s.coinVideos(ctx, vmid, coinPcy)
			return nil
		})
	}
	// like video
	if mid == vmid || !likePcy {
		group.Go(func(ctx context.Context) error {
			likeList, likeErr = s.likeVideos(ctx, vmid, _samplePs, likePcy)
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if coinErr == nil && len(coinList) > 0 {
		preList = append(preList, coinList...)
	}
	if likeErr == nil && len(likeList) > 0 {
		preList = append(preList, likeList...)
	}
	if len(preList) == 0 {
		return
	}
	sort.Slice(preList, func(i, j int) bool { return preList[i].ActionTime > preList[j].ActionTime })
	if s.c.Rule.Merge {
		preList = mergeDyActItem(preList)
	}
	actItem := preList[0]
	arc, err := s.arcClient.Arc(c, &arcmdl.ArcRequest{Aid: actItem.Aid})
	if err != nil {
		return
	}
	if arc != nil && arc.Arc != nil && arc.Arc.IsNormal() {
		video := new(model.VideoItem)
		video.FromArchive(arc.Arc)
		video.ActionTime = actItem.ActionTime
		item = &model.DyItem{Type: actItem.Type, Archive: video, Ctime: actItem.ActionTime, Privacy: actItem.Privacy}
	}
	return
}

func mergeDyActItem(preList []*model.DyActItem) (mergeList []*model.DyActItem) {
	type privacy struct {
		Num  int
		Coin bool
		Like bool
	}
	aidNumMap := make(map[int64]*privacy, len(preList))
	aidExist := make(map[int64]struct{}, len(preList))
	for _, v := range preList {
		if _, exist := aidNumMap[v.Aid]; !exist {
			aidNumMap[v.Aid] = new(privacy)
		}
		aidNumMap[v.Aid].Num++
		switch v.Type {
		case _dyTypeCoin:
			aidNumMap[v.Aid].Coin = v.Privacy
		case _dyTypeLike:
			aidNumMap[v.Aid].Like = v.Privacy
		}
	}
	for _, v := range preList {
		num := aidNumMap[v.Aid].Num
		if num > 1 {
			if _, ok := aidExist[v.Aid]; !ok {
				v.Type = _dyTypeMerge
				v.Privacy = aidNumMap[v.Aid].Coin && aidNumMap[v.Aid].Like
				mergeList = append(mergeList, v)
			}
			aidExist[v.Aid] = struct{}{}
		} else {
			mergeList = append(mergeList, v)
		}
	}
	return
}

func (s *Service) coinVideos(c context.Context, vmid int64, pcy bool) (list []*model.DyActItem, err error) {
	var (
		coinReply *coinmdl.ListReply
		aids      []int64
	)
	if coinReply, err = s.coinClient.List(c, &coinmdl.ListReq{Mid: vmid, Business: _businessCoin, Ts: time.Now().Unix()}); err != nil {
		log.Error("s.coinClient.List(%d) error(%v)", vmid, err)
		return
	}
	existArcs := make(map[int64]*coinmdl.ModelList, len(coinReply.List))
	for _, v := range coinReply.List {
		if len(aids) > _coinVideoLimit {
			break
		}
		if _, ok := existArcs[v.Aid]; ok {
			continue
		}
		if v.Aid > 0 {
			list = append(list, &model.DyActItem{Aid: v.Aid, Type: _dyTypeCoin, ActionTime: v.Ts, Privacy: pcy})
			existArcs[v.Aid] = v
		}
	}
	return
}

func (s *Service) likeVideos(c context.Context, mid int64, cnt int, pcy bool) (list []*model.DyActItem, err error) {
	var (
		rep *thumbupmdl.UserLikesReply
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	arg := &thumbupmdl.UserLikesReq{Mid: mid, Business: _businessArchiveLike, Pn: _samplePn, Ps: int64(cnt), IP: ip}
	if rep, err = s.thumbupClient.UserLikes(c, arg); err != nil {
		log.Error("s.thumbupClient.UserLikes(%d) error(%v)", mid, err)
		return
	}
	if rep == nil || len(rep.Items) == 0 {
		return
	}
	for _, v := range rep.Items {
		list = append(list, &model.DyActItem{Aid: v.MessageID, Type: _dyTypeLike, ActionTime: int64(v.Time), Privacy: pcy})
	}
	return
}

// topDynamic get top dynamic.
func (s *Service) topDynamic(c context.Context, mid int64, qn int) (res *model.DyCard, err error) {
	var (
		dyID int64
	)
	if dyID, err = s.dao.TopDynamic(c, mid); err != nil {
		return
	}
	if dyID == 0 {
		err = ecode.NothingFound
		return
	}
	if res, err = s.dao.Dynamic(c, mid, dyID, qn); err != nil || res == nil {
		log.Error("Dynamic s.dao.Dynamic mid(%d) dyID(%d) error(%v)", mid, dyID, err)
		err = ecode.NothingFound
	}
	return
}

// SetTopDynamic set top dynamic.
func (s *Service) SetTopDynamic(c context.Context, mid, dynamicID int64) (err error) {
	var (
		dynamic *model.DyCard
		preDyID int64
	)
	if dynamic, err = s.dao.Dynamic(c, mid, dynamicID, _dyDefaultQn); err != nil || dynamic == nil {
		log.Error("SetTopDynamic s.dao.Dynamic(%d) error(%v)", dynamicID, err)
		return
	}
	if dynamic.Desc.UID != mid {
		err = ecode.RequestErr
		return
	}
	if preDyID, err = s.dao.TopDynamic(c, mid); err != nil {
		return
	}
	if preDyID == dynamicID {
		err = ecode.NotModified
		return
	}
	if err = s.dao.AddTopDynamic(c, mid, dynamicID); err == nil {
		_ = s.dao.AddCacheTopDynamic(c, mid, dynamicID)
	}
	return
}

// CancelTopDynamic cancel top dynamic.
func (s *Service) CancelTopDynamic(c context.Context, mid int64, now time.Time) (err error) {
	var dyID int64
	if dyID, err = s.dao.TopDynamic(c, mid); err != nil {
		return
	}
	if dyID == 0 {
		err = ecode.RequestErr
		return
	}
	if err = s.dao.DelTopDynamic(c, mid, now); err == nil {
		_ = s.dao.AddCacheTopDynamic(c, mid, -1)
	}
	return
}

func (s *Service) DynamicSearch(ctx context.Context, mid int64, dynSearchArg *model.DynamicSearchArg) (*model.DynamicSearchRes, error) {
	dynamicIDs, total, err := s.dao.DynamicSearch(ctx, mid, dynSearchArg.Mid, dynSearchArg.Keyword, dynSearchArg.Pn, dynSearchArg.Ps)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	cards, err := s.dao.DynamicDetail(ctx, mid, dynamicIDs)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &model.DynamicSearchRes{
		Cards: cards,
		Total: total,
	}, nil
}
