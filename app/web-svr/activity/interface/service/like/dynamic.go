package like

import (
	"context"
	"strconv"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	arccli "go-gateway/app/app-svr/archive/service/api"

	appecode "go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	dynmdl "go-gateway/app/web-svr/activity/interface/model/dynamic"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/pkg/idsafe/bvid"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
)

// ActLiked .
func (s *Service) ActLiked(c context.Context, arg *lmdl.ParamAddLikeAct, mid int64) (res *dynmdl.LikedReply, err error) {
	var (
		rely *lmdl.ActReply
	)
	if rely, err = s.LikeAct(c, arg, mid); err != nil {
		switch {
		case xecode.EqualError(ecode.ActivityLikeHasEnd, err):
			err = appecode.AppActHasEnd
		case xecode.EqualError(ecode.ActivityLikeNotStart, err):
			err = appecode.AppActNotStart
		case xecode.EqualError(ecode.ActivityOverLikeLimit, err):
			err = appecode.AppActOverLikeLimit
		case xecode.EqualError(ecode.ActivityLikeIPFrequence, err), xecode.EqualError(ecode.ActivityLikeScoreLower, err),
			xecode.EqualError(ecode.ActivityLikeRegisterLimit, err), xecode.EqualError(ecode.ActivityLikeBeforeRegister, err),
			xecode.EqualError(ecode.ActivityTelValid, err), xecode.EqualError(ecode.ActivityLikeLevelLimit, err):
			err = appecode.AppNoLikeCondition
		}
		log.Error("s.ActLiked(%v) error(%v)", arg, err)
		return
	}
	if rely == nil {
		return
	}
	res = &dynmdl.LikedReply{Score: rely.Score, Toast: "投票成功，已为稿件增加" + strconv.FormatInt(rely.Score, 10) + "票"}
	return
}

func (s *Service) toAvIds(ids []string) (avIds map[string]int64) {
	avIds = make(map[string]int64, len(ids))
	for _, id := range ids {
		if strings.HasPrefix(id, "BV1") {
			if avid, err := bvid.BvToAv(id); err == nil && avid > 0 {
				avIds[id] = avid
			}
		} else {
			if k, err := strconv.ParseInt(id, 10, 64); err == nil && k > 0 {
				avIds[id] = k
			}
		}
	}
	return
}

// VideoAct .
func (s *Service) VideoAct(c context.Context, arg *dynmdl.ParamVideoAct, mid int64) (rly *dynmdl.VideoActReply, err error) {
	var (
		aids    map[string]int64
		rids    []*dynmdl.RidInfo
		dyReply *dynmdl.DyResult
	)
	aids = s.toAvIds(arg.IDs)
	lg := len(aids)
	if lg == 0 {
		return
	}
	rids = make([]*dynmdl.RidInfo, 0, lg)
	for _, v := range aids {
		rids = append(rids, &dynmdl.RidInfo{Rid: v, Type: arg.Type})
	}
	if dyReply, err = s.dynamicDao.Dynamic(c, &dynmdl.Resources{Array: rids}, mid); err != nil || dyReply == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rids, err)
		return
	}
	if len(dyReply.Cards) == 0 {
		return
	}
	rly = &dynmdl.VideoActReply{List: make(map[string]*dynmdl.Item, lg)}
	for k, v := range aids {
		if _, ok := dyReply.Cards[v]; !ok {
			continue
		}
		rly.List[k] = &dynmdl.Item{DyCard: &dynmdl.DyCard{Card: dyReply.Cards[v].Card, Desc: dyReply.Cards[v].Desc}}
	}
	return
}

// ActList .
func (s *Service) ActList(c context.Context, arg *dynmdl.ParamActList, mid int64) (rly *dynmdl.ActReply, err error) {
	var (
		likeList *lmdl.ActLikes
		subPs    = arg.Ps + 6
	)
	argLike := &lmdl.ArgActLikes{Sid: arg.Sid, Mid: mid, SortType: arg.SortType, Ps: subPs, Offset: arg.Offset, Zone: arg.Zone}
	if likeList, err = s.ActLikes(c, argLike); err != nil || likeList == nil {
		log.Error("s.ActLikes(%v) error(%v)", argLike, err)
		return
	}
	lg := len(likeList.List)
	if lg == 0 || likeList.Sub == nil {
		// 没有获取导数据
		rly = &dynmdl.ActReply{}
		return
	}
	switch arg.Goto {
	case "resource": //资源小卡
		rly, err = s.ResourceCard(c, likeList, arg.Offset, arg.Attribute, arg.Ps)
	default:
		rly, err = s.DynamicCard(c, likeList, mid, arg.Offset, arg.Ps)
	}
	if err != nil || rly == nil {
		// 没有获取导数据
		rly = &dynmdl.ActReply{HasMore: likeList.HasMore, Offset: likeList.Offset}
		err = nil
		return
	}
	tempMou := &api.NativeModule{Attribute: arg.Attribute}
	if tempMou.IsAttrLast() != api.AttrModuleYes && tempMou.IsAttrHideMore() != api.AttrModuleYes {
		//查看更多标签
		if len(rly.Items) > 0 && rly.HasMore >= 1 {
			if arg.PageID <= 0 || arg.Ukey == "" {
				return
			}
			//查看更多按钮h5统一跳转app话题活动页
			tmpMore := &dynmdl.Item{}
			tmpMore.FromVideoMore()
			rly.Items = append(rly.Items, tmpMore)
		}
	}
	return
}

// NewActList .
func (s *Service) NewActList(c context.Context, arg *dynmdl.ParamNewActList, mid int64) (rly *dynmdl.ActReply, err error) {
	var (
		likeList *lmdl.ActLikes
		subPs    = arg.Ps + 6
	)
	argLike := &lmdl.ArgActLikes{Sid: arg.Sid, Mid: mid, SortType: arg.SortType, Ps: subPs, Offset: arg.Offset, Zone: arg.Zone}
	if likeList, err = s.ActLikes(c, argLike); err != nil || likeList == nil {
		log.Error("s.ActLikes(%v) error(%v)", argLike, err)
		return
	}
	lg := len(likeList.List)
	if lg == 0 || likeList.Sub == nil {
		// 没有获取导数据
		rly = &dynmdl.ActReply{}
		return
	}
	rly, err = s.newActVideo(c, likeList, arg.Offset, arg.Ps)
	if err != nil {
		// 没有获取导数据
		rly = &dynmdl.ActReply{HasMore: likeList.HasMore, Offset: likeList.Offset}
		err = nil
	}
	return
}

// newActVideo 新视频卡act模式 .
func (s *Service) newActVideo(c context.Context, likeList *lmdl.ActLikes, offset int64, ps int) (*dynmdl.ActReply, error) {
	var ids []int64
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			ids = append(ids, v.Item.Wid)
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
	)
	if len(ids) > 0 {
		arcRes, e := client.ArchiveClient.Arcs(c, &arccli.ArcsRequest{Aids: ids})
		if e != nil {
			log.Error("s.arcClient.Arcs aids(%v) error(%v)", ids, e)
			return nil, e
		}
		if arcRes != nil {
			arcRly = arcRes.Arcs
		}
	}
	rly := &dynmdl.ActReply{}
	for _, v := range likeList.List {
		// 补位逻辑自己计算offset
		offset++
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		if va, ok := arcRly[v.Item.Wid]; !ok || va == nil {
			continue
		}
		bvidStr, _ := bvid.AvToBv(v.Item.Wid)
		tmp.FromUgcVideo(arcRly[v.Item.Wid], bvidStr)
		rly.Items = append(rly.Items, tmp)
		// 补位结束
		if len(rly.Items) >= ps {
			break
		}
	}
	// 补位后的offset
	rly.Offset = int64(offset)
	rly.HasMore = likeList.HasMore
	if likeList.HasMore == 0 && rly.Offset < likeList.Offset {
		rly.HasMore = 1
	}
	return rly, nil
}

func (s *Service) ResourceCard(c context.Context, likeList *lmdl.ActLikes, offset, attribute int64, ps int) (*dynmdl.ActReply, error) {
	switch {
	case likeList.Sub.IsARTICLE():
		return s.ResourceArt(c, likeList, offset, attribute, ps)
	default: //12,4,16
		return s.ResourceVideo(c, likeList, offset, attribute, ps)
	}
}

func (s *Service) ResourceVideo(c context.Context, likeList *lmdl.ActLikes, offset, attribute int64, ps int) (*dynmdl.ActReply, error) {
	var ids []int64
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			ids = append(ids, v.Item.Wid)
		}
	}
	var (
		arcRly map[int64]*arccli.Arc
	)
	if len(ids) > 0 {
		arcRes, e := client.ArchiveClient.Arcs(c, &arccli.ArcsRequest{Aids: ids})
		if e != nil {
			log.Error("s.arcClient.Arcs aids(%v) error(%v)", ids, e)
			return nil, e
		}
		if arcRes != nil {
			arcRly = arcRes.Arcs
		}
	}
	tempMou := &api.NativeModule{Attribute: attribute}
	arcDisplay := tempMou.IsAttrDisplayVideoIcon() == api.AttrModuleYes
	rly := &dynmdl.ActReply{}
	for _, v := range likeList.List {
		// 补位逻辑自己计算offset
		offset++
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		if va, ok := arcRly[v.Item.Wid]; !ok || va == nil {
			continue
		}
		bvidStr, _ := bvid.AvToBv(v.Item.Wid)
		tmp.FromResourceArc(arcRly[v.Item.Wid], arcDisplay, bvidStr, nil)
		rly.Items = append(rly.Items, tmp)
		// 补位结束
		if len(rly.Items) >= ps {
			break
		}
	}
	// 补位后的offset
	rly.Offset = int64(offset)
	rly.HasMore = likeList.HasMore
	if likeList.HasMore == 0 && rly.Offset < likeList.Offset {
		rly.HasMore = 1
	}
	return rly, nil
}

func (s *Service) ResourceArt(c context.Context, likeList *lmdl.ActLikes, offset, attribute int64, ps int) (*dynmdl.ActReply, error) {
	var cvids []int64
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			cvids = append(cvids, v.Item.Wid)
		}
	}
	var (
		artRly map[int64]*artmdl.Meta
	)
	if len(cvids) > 0 {
		artRes, e := s.artClient.ArticleMetas(c, &artapi.ArticleMetasReq{Ids: cvids, From: 2})
		if e != nil {
			log.Error("s.dao.ArticleMeta cvids(%v) error(%v)", cvids, e)
			return nil, e
		}
		if artRes != nil {
			artRly = artRes.Res
		}
	}
	tempMou := &api.NativeModule{Attribute: attribute}
	artDisplay := tempMou.IsAttrDisplayArticleIcon() == api.AttrModuleYes
	rly := &dynmdl.ActReply{}
	for _, v := range likeList.List {
		// 补位逻辑自己计算offset
		offset++
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		tmp := &dynmdl.Item{}
		if va, ok := artRly[v.Item.Wid]; !ok || va == nil {
			continue
		}
		tmp.FromResourceArt(artRly[v.Item.Wid], artDisplay)
		rly.Items = append(rly.Items, tmp)
		// 补位结束
		if len(rly.Items) >= ps {
			break
		}
	}
	// 补位后的offset
	rly.Offset = int64(offset)
	rly.HasMore = likeList.HasMore
	if likeList.HasMore == 0 && rly.Offset < likeList.Offset {
		rly.HasMore = 1
	}
	return rly, nil
}

// DynamicCard .
func (s *Service) DynamicCard(c context.Context, likeList *lmdl.ActLikes, mid, offset int64, ps int) (*dynmdl.ActReply, error) {
	rids := make([]*dynmdl.RidInfo, 0)
	itemObj := make(map[int64]*lmdl.ItemObj)
	widType := dynmdl.VideoType
	if likeList.Sub.IsARTICLE() {
		widType = dynmdl.ARTICLETYPE
	}
	for _, v := range likeList.List {
		if v.Item != nil && v.Item.Wid > 0 {
			rids = append(rids, &dynmdl.RidInfo{Rid: v.Item.Wid, Type: widType})
			itemObj[v.Item.Wid] = v
		}
	}
	rous := &dynmdl.Resources{Array: rids}
	dyReply, err := s.dynamicDao.Dynamic(c, rous, mid)
	if err != nil || dyReply == nil {
		log.Error("s.dynamicDao.Dynamic(%v) error(%v)", rous, err)
		return nil, err
	}
	list := &dynmdl.ActReply{}
	for _, v := range likeList.List {
		// 补位逻辑自己计算offset
		offset++
		if v.Item == nil || v.Item.Wid == 0 {
			continue
		}
		if _, ok := dyReply.Cards[v.Item.Wid]; !ok {
			continue
		}
		if _, k := itemObj[v.Item.Wid]; !k {
			continue
		}
		temp := &dynmdl.Item{}
		if likeList.Sub.IsVideoLike() {
			temp.FromVideoLike(dyReply.Cards[v.Item.Wid], itemObj[v.Item.Wid])
		} else {
			temp.FromVideo(dyReply.Cards[v.Item.Wid])
		}
		list.Items = append(list.Items, temp)
		// 补位结束
		if len(list.Items) >= ps {
			break
		}
	}
	// 补位后的offset
	list.Offset = int64(offset)
	list.HasMore = likeList.HasMore
	if likeList.HasMore == 0 && list.Offset < likeList.Offset {
		list.HasMore = 1
	}
	return list, nil
}
