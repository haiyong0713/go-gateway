package like

import (
	"context"
	"encoding/json"
	"net"
	"sort"
	"sync"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	tagmdl "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	thumbupmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/database/bfs"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	dao "go-gateway/app/web-svr/activity/interface/dao/like"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/pkg/idsafe/bvid"

	figureapi "git.bilibili.co/bapis/bapis-go/account/service/figure"
	spyapi "git.bilibili.co/bapis/bapis-go/account/service/spy"
	fliapi "git.bilibili.co/bapis/bapis-go/filter/service"
)

const (
	_aidBulkSize     = 50
	_tagBlkSize      = 50
	_tagArcType      = 3
	_tagLikePoint    = 100
	_orderTypeCtime  = "ctime"
	_orderTypeRandom = "random"
	_specialLikeRate = 1000
	_businessLike    = "archive"
	_businessLikeArt = "article"
	_tpWebDataVideo  = 1
	_tpWebDataInfo   = 2
)

var (
	_emptyLikeList = make([]*likemdl.Like, 0)
	_emptyArcs     = make([]*api.Arc, 0)
	_emptyViewData = make([]*likemdl.WebDataRes, 0)
)

// UpdateActSourceList update act arc list.
func (s *Service) updateActSourceList(c context.Context, sid int64, typ string) (err error) {
	var (
		likes []*likemdl.Item
	)
	if likes, err = s.dao.LikeList(c, sid); err != nil {
		log.Error("UpdateActSourceList s.dao.LikeList(%d) error(%v)", sid, err)
		return
	}
	s.cache.Do(c, func(c context.Context) {
		if typ == _typeAll {
			s.updateActCacheList(c, sid, likes)
		}
		if typ == _typeRegion {
			s.updateActRegionList(c, sid, likes)
		}
	})
	return
}

func (s *Service) updateActCacheList(c context.Context, sid int64, likes []*likemdl.Item) (err error) {
	var (
		aids []int64
		tags map[int64][]*tagmdl.Tag
		arcs map[int64]*api.Arc
	)
	likeMap := make(map[int64]*likemdl.Item, len(likes))
	for _, v := range likes {
		if v.Wid > 0 {
			aids = append(aids, v.Wid)
			likeMap[v.Wid] = v
		}
	}
	if len(aids) == 0 {
		return
	}
	tags = s.arcTags(c, aids)

	log.Infoc(c, "updateActCacheList arcTags :%v", len(tags))
	if arcs, err = s.archives(c, aids); err != nil {
		return
	}
	arcTagMap := make(map[int64][]*likemdl.Item, len(s.dialectTags))
	tagLikePtTmp := make(map[int64]int32, len(s.dialectTags))
	for aid, arcTag := range tags {
		for _, tag := range arcTag {
			if _, ok := s.dialectTags[tag.Id]; ok {
				arcTagMap[tag.Id] = append(arcTagMap[tag.Id], likeMap[aid])
				if arc, ok := arcs[aid]; ok && arc.IsNormal() {
					tagLikePtTmp[tag.Id] += arc.Stat.Like
				}
			}
		}
	}

	arcTagMapStr, _ := json.Marshal(arcTagMap)
	log.Infoc(c, "updateActCacheList arcTagMapStr :%s", arcTagMapStr)
	tagLikePtTmpStr, _ := json.Marshal(tagLikePtTmp)
	log.Infoc(c, "updateActCacheList tagLikePtTmpStr :%s", tagLikePtTmpStr)

	tagPtMap := make(map[int64]int32, len(s.dialectTags))
	for tagID, v := range arcTagMap {
		s.dao.SetLikeTagCache(c, sid, tagID, v)
		if like, ok := tagLikePtTmp[tagID]; ok {
			tagPt := int32(len(v)*_tagLikePoint) + like
			tagPtMap[tagID] = tagPt
		}
	}
	s.dao.SetTagLikeCountsCache(c, sid, tagPtMap)
	regionMap := make(map[int32][]*likemdl.Item, len(s.dialectRegions))
	for _, arc := range arcs {
		if region, ok := s.arcType[arc.TypeID]; ok {
			if _, ok := s.dialectRegions[region.Pid]; ok {
				regionMap[region.Pid] = append(regionMap[region.Pid], likeMap[arc.Aid])
			}
		}
	}
	for rid, v := range regionMap {
		s.dao.SetLikeRegionCache(c, sid, rid, v)
	}
	return
}

func (s *Service) updateActRegionList(c context.Context, sid int64, likes []*likemdl.Item) (err error) {
	var (
		aids []int64
		arcs map[int64]*api.Arc
	)
	likeMap := make(map[int64]*likemdl.Item, len(likes))
	for _, v := range likes {
		if v.Wid > 0 {
			aids = append(aids, v.Wid)
			likeMap[v.Wid] = v
		}
	}
	if len(aids) == 0 {
		return
	}
	if arcs, err = s.archives(c, aids); err != nil {
		return
	}
	regionMap := make(map[int32][]*likemdl.Item)
	for _, arc := range arcs {
		if region, ok := s.arcType[arc.TypeID]; ok {
			regionMap[region.Pid] = append(regionMap[region.Pid], likeMap[arc.Aid])
		}
	}
	for rid, v := range regionMap {
		s.dao.SetLikeRegionCache(c, sid, rid, v)
	}
	return
}

func (s *Service) archives(c context.Context, aids []int64) (archives map[int64]*api.Arc, err error) {
	var (
		mutex         = sync.Mutex{}
		aidsLen       = len(aids)
		group, errCtx = errgroup.WithContext(c)
	)
	archives = make(map[int64]*api.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func() (err error) {
			var arcs *api.ArcsReply
			arg := &api.ArcsRequest{Aids: partAids}
			if arcs, err = client.ArchiveClient.Arcs(errCtx, arg); err != nil || arcs == nil {
				log.Error("s.arcRPC.Archives(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
	}
	err = group.Wait()
	return
}

func (s *Service) arcTags(c context.Context, aids []int64) (tags map[int64][]*tagmdl.Tag) {
	var (
		tagErr error
		mutex  = sync.Mutex{}
	)
	group, errCtx := errgroup.WithContext(c)
	aidsLen := len(aids)
	tags = make(map[int64][]*tagmdl.Tag, aidsLen)
	for i := 0; i < aidsLen; i += _tagBlkSize {
		var partAids []int64
		if i+_tagBlkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_tagBlkSize]
		}
		group.Go(func() (err error) {
			var reply *tagmdl.ResTagsReply
			arg := &tagmdl.ResTagsReq{Oids: partAids, Type: _tagArcType}
			if reply, tagErr = s.tagGRPC.ResTags(errCtx, arg); tagErr != nil {
				log.Errorc(c, "ResTags接口错误 s.tag.ResTag(%+v) error(%v)", arg, tagErr)
				return tagErr
			}
			mutex.Lock()
			if reply != nil {
				for aid, tmpTags := range reply.ResTags {
					tags[aid] = append(tags[aid], tmpTags.Tags...)
				}
			}
			mutex.Unlock()
			return nil
		})
	}
	group.Wait()
	return
}

// TagArcList tag arc list.
func (s *Service) TagArcList(c context.Context, sid, tagID int64, pn, ps int, typ, ip string) (list []*likemdl.Like, cnt int, err error) {
	var (
		likes      []*likemdl.Item
		start, end int
		aids       []int64
		archives   map[int64]*api.Arc
	)
	if sid != s.c.Rule.DialectSid {
		err = xecode.RequestErr
		return
	}
	if _, ok := s.dialectTags[tagID]; !ok {
		err = xecode.RequestErr
		return
	}
	if cnt, err = s.dao.LikeTagCnt(c, sid, tagID); err != nil {
		log.Error("TagArcList s.dao.LikeTagCnt sid(%d) tagID(%d) error(%v)", sid, tagID, err)
		return
	}
	if start, end, err = s.fmtStartEnd(pn, ps, cnt, typ); err != nil {
		err = nil
		list = _emptyLikeList
		return
	}
	if likes, err = s.dao.LikeTagCache(c, sid, tagID, start, end); err != nil {
		log.Error("TagArcList s.dao.LikeTagCache sid(%d) tagID(%d) start(%d) end(%d) error(%+v)", sid, tagID, start, end, err)
		return
	}
	for _, v := range likes {
		if v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	if len(aids) == 0 {
		list = _emptyLikeList
		return
	}
	if archives, err = s.archives(c, aids); err != nil {
		log.Error("TagArcList s.archives aids(%v) error(%+v)", aids, err)
		return
	}
	for _, v := range likes {
		if arc, ok := archives[v.Wid]; ok && arc.IsNormal() {
			list = append(list, &likemdl.Like{Item: v, Archive: arc})
		}
	}
	l := len(list)
	if l == 0 {
		list = _emptyLikeList
		return
	}
	if typ == _orderTypeRandom {
		s.shuffle(l, func(i, j int) {
			list[i], list[j] = list[j], list[i]
		})
	}
	return
}

// RegionArcList region arc list.
func (s *Service) RegionArcList(c context.Context, sid int64, rid int32, pn, ps int, typ, ip string) (list []*likemdl.Like, cnt int, err error) {
	var (
		likes      []*likemdl.Item
		start, end int
		aids       []int64
		archives   map[int64]*api.Arc
	)
	if sid != s.c.Rule.DialectSid {
		err = xecode.RequestErr
		return
	}
	if _, ok := s.dialectRegions[rid]; !ok {
		err = xecode.RequestErr
		return
	}
	if cnt, err = s.dao.LikeRegionCnt(c, sid, rid); err != nil {
		log.Error("RegionArcList s.dao.LikeRegionCnt sid(%d) rid(%d) error(%v)", sid, rid, err)
		return
	}
	if start, end, err = s.fmtStartEnd(pn, ps, cnt, typ); err != nil {
		err = nil
		list = _emptyLikeList
		return
	}
	if likes, err = s.dao.LikeRegionCache(c, sid, rid, start, end); err != nil {
		log.Error("RegionArcList s.dao.LikeRegionCache sid(%d) rid(%d) start(%d) end(%d) error(%+v)", sid, rid, start, end, err)
		return
	}
	for _, v := range likes {
		if v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	if len(aids) == 0 {
		list = _emptyLikeList
		return
	}
	if archives, err = s.archives(c, aids); err != nil {
		log.Error("RegionArcList s.archives aids(%v) error(%+v)", aids, err)
		return
	}
	for _, v := range likes {
		if arc, ok := archives[v.Wid]; ok && arc.IsNormal() {
			list = append(list, &likemdl.Like{Item: v, Archive: arc})
		}
	}
	l := len(list)
	if l == 0 {
		list = _emptyLikeList
		return
	}
	if typ == _orderTypeRandom {
		s.shuffle(l, func(i, j int) {
			list[i], list[j] = list[j], list[i]
		})
	}
	return
}

// TagLikeCounts .
func (s *Service) TagLikeCounts(c context.Context, sid int64) (data map[int64]int32, err error) {
	if sid != s.c.Rule.DialectSid {
		err = xecode.RequestErr
		return
	}
	return s.dao.TagLikeCountsCache(c, sid, s.c.Rule.DialectTags)
}

func (s *Service) fmtStartEnd(pn, ps, cnt int, typ string) (start, end int, err error) {
	if typ == _orderTypeCtime {
		start = (pn - 1) * ps
		end = start + ps - 1
		if start > cnt {
			err = xecode.NothingFound
			return
		}
		if end > cnt {
			end = cnt
		}
	} else {
		if ps >= cnt-1 {
			start = 0
		} else {
			start = s.r.Intn(cnt - ps - 1)
		}
		end = start + ps - 1
	}
	return
}

func (s *Service) shuffle(l int, swap func(i, j int)) {
	for i := l - 1; i > 0; i-- {
		j := s.r.Intn(i + 1)
		swap(i, j)
	}
}

// LikeInitialize initialize like cache data .
func (s *Service) LikeInitialize(c context.Context, lid int64) (err error) {
	if lid < 0 {
		lid = 0
	}
	var likesItem []*likemdl.Item
	for {
		if likesItem, err = s.dao.LikeListMoreLid(c, lid); err != nil {
			log.Error("dao.LikeInitialize(%d) error(%+v)", lid, err)
			break
		}
		if len(likesItem) == 0 {
			log.Info("LikeInitialize end success")
			break
		}
		for _, val := range likesItem {
			item := val
			if lid < item.ID {
				lid = item.ID
			}
			id := item.ID
			//the likes offline is stored with empty data
			if item.State != 1 {
				item = &likemdl.Item{}
			}
			s.cache.Do(c, func(c context.Context) {
				s.dao.AddCacheLike(c, id, item)
			})
		}
	}
	s.cache.Do(c, func(c context.Context) {
		s.LikeMaxIDInitialize(c)
	})
	return
}

// LikeMaxIDInitialize likes max id initialize
func (s *Service) LikeMaxIDInitialize(c context.Context) (err error) {
	var likeItem *likemdl.Item
	if likeItem, err = s.dao.LikeMaxID(c); err != nil {
		log.Error("s.dao.LikeMaxID() error(%+v)", err)
		return
	}
	if likeItem.ID >= 0 {
		if err = s.dao.AddCacheLikeMaxID(c, likeItem.ID); err != nil {
			log.Error("s.dao.AddCacheLikeMaxID(%d),error(%v)", likeItem.ID, err)
		}
	}
	return
}

// LikeUp update likes cache and like maxID cache
func (s *Service) LikeUp(c context.Context, lid int64) (err error) {
	var (
		likeItem  *likemdl.Item
		likeMaxID int64
	)
	group, ctx := errgroup.WithContext(c)
	group.Go(func() (e error) {
		if likeItem, e = s.dao.RawLike(ctx, lid); e != nil {
			log.Error("LikeUp:s.dao.RawLike(%d) error(%+v)", lid, e)
		}
		return
	})
	group.Go(func() (e error) {
		if likeMaxID, e = s.dao.CacheLikeMaxID(ctx); e != nil {
			log.Error("LikeUp:s.dao.CacheLikeMaxID() error(%v)", e)
		}
		return
	})
	if err = group.Wait(); err != nil {
		log.Error("LikeUp: group.Wait() error(%v)", err)
		return
	}
	if likeMaxID < lid {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheLikeMaxID(c, lid)
		})
	}
	if likeItem.ID == 0 {
		likeItem = &likemdl.Item{}
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.AddCacheLike(c, lid, likeItem)
	})
	return
}

// AddLikeCtimeCache add cache .
func (s *Service) AddLikeCtimeCache(c context.Context, lid int64) (err error) {
	var (
		likeItem *likemdl.Item
		cItems   = make([]*likemdl.Item, 0, 1)
	)
	if likeItem, err = s.dao.RawLike(c, lid); err != nil {
		log.Error("LikeUp:s.dao.RawLike(%d) error(%+v)", lid, err)
		return
	}
	if likeItem.ID > 0 && likeItem.StickTop == 0 {
		eg, errCtx := errgroup.WithContext(c)
		cItems = append(cItems, likeItem)
		eg.Go(func() (e error) {
			e = s.dao.LikeListCtime(errCtx, likeItem.Sid, cItems)
			return
		})
		eg.Go(func() (e error) {
			// 初始化排行榜数据
			e = s.dao.SetRedisCache(c, likeItem.Sid, lid, 0, likeItem.Type)
			return
		})
		if err = eg.Wait(); err != nil {
			log.Error("AddLikeCtimeCache eg.Wait() error(%+v)", err)
		}
	}
	return
}

// DelLikeCtimeCache delete ctime cache.
func (s *Service) DelLikeCtimeCache(c context.Context, lid, sid int64, likeType int64) (err error) {
	var (
		cItems = make([]*likemdl.Item, 0, 1)
	)
	likeItem := &likemdl.Item{
		ID:   lid,
		Sid:  sid,
		Type: likeType,
	}
	cItems = append(cItems, likeItem)
	eg, errCtx := errgroup.WithContext(c)
	eg.Go(func() (e error) {
		if e = s.dao.DelLikeListCtime(errCtx, likeItem.Sid, cItems); e != nil {
			log.Error("s.dao.DelLikeListCtime(%v) error (%v)", likeItem, e)
		}
		return
	})
	eg.Go(func() (e error) {
		if e = s.dao.DelLikeListLikes(errCtx, likeItem.Sid, cItems); e != nil {
			log.Error("s.dao.DelLikeListLikes(%v) error (%v)", likeItem, e)
		}
		return
	})
	err = eg.Wait()
	return
}

// ActSetReload .
func (s *Service) ActSetReload(c context.Context, lid int64) (err error) {
	var (
		likeItem *likemdl.Item
	)
	if likeItem, err = s.dao.RawLike(c, lid); err != nil {
		log.Error("ActSetReload:s.dao.RawLike(%d) error(%+v)", lid, err)
		return
	}
	if likeItem.ID == 0 {
		return
	}
	//获取lid的点赞数，回源到热度排序的集合中
	return s.dao.SetLikesReload(c, lid, likeItem.Sid, likeItem.Type)
}

// SubjectStat get subject stat .
func (s *Service) SubjectStat(c context.Context, sid int64) (score *likemdl.SubjectScore, err error) {
	if sid == s.c.Rule.S8Sid {
		var arcScore, artScore int64
		group, errCtx := errgroup.WithContext(c)
		group.Go(func() error {
			var (
				stat   *likemdl.SubjectStat
				arcErr error
			)
			if stat, arcErr = s.dao.CacheSubjectStat(errCtx, s.c.Rule.S8ArcSid); arcErr != nil {
				log.Error("s.dao.CacheSubjectStat sid(%d) error(%v)", sid, arcErr)
			}
			if stat == nil {
				stat = new(likemdl.SubjectStat)
			}
			arcScore = stat.Count*_specialLikeRate + stat.Like
			return nil
		})
		group.Go(func() error {
			var (
				stat   *likemdl.SubjectStat
				artErr error
			)
			if stat, artErr = s.dao.CacheSubjectStat(errCtx, s.c.Rule.S8ArtSid); artErr != nil {
				log.Error("s.dao.CacheSubjectStat sid(%d) error(%v)", sid, artErr)
			}
			if stat == nil {
				stat = new(likemdl.SubjectStat)
			}
			artScore = stat.Count*_specialLikeRate + stat.Like
			return nil
		})
		group.Wait()
		score = &likemdl.SubjectScore{Score: arcScore + artScore}
	} else {
		var stat *likemdl.SubjectStat
		if stat, err = s.dao.CacheSubjectStat(c, sid); err != nil {
			log.Error("s.dao.CacheSubjectStat sid(%d) error(%v)", sid, err)
			err = nil
		}
		if stat == nil {
			stat = new(likemdl.SubjectStat)
		}
		if sid == s.c.Rule.KingStorySid {
			score = &likemdl.SubjectScore{Score: stat.View + stat.Fav + stat.Coin + stat.Like}
		} else {
			score = &likemdl.SubjectScore{Score: stat.Count*_specialLikeRate + stat.Like}
		}
	}
	return
}

// SetSubjectStat set subject stat .
func (s *Service) SetSubjectStat(c context.Context, stat *likemdl.SubjectStat) (err error) {
	return s.dao.AddCacheSubjectStat(c, stat.Sid, stat)
}

// ViewRank get view rank arcs.
func (s *Service) ViewRank(c context.Context, sid int64, pn, ps int, typ string) (list []*api.Arc, count int, err error) {
	var (
		aidsCache       string
		aids, pieceAids []int64
		arcs            map[int64]*api.Arc
	)
	if aidsCache, err = s.dao.CacheViewRank(c, sid, typ); err != nil {
		log.Error("ViewRank s.dao.CacheViewRank(%d,%s) error(%v)", sid, typ, err)
		return
	}
	if aids, err = xstr.SplitInts(aidsCache); err != nil {
		log.Error("ViewRank xstr.SplitInts(%d,%s) error(%v)", sid, aidsCache, err)
		return
	}
	count = len(aids)
	start := (pn - 1) * ps
	end := start + ps - 1
	if count < start {
		list = _emptyArcs
		return
	}
	if count > end {
		pieceAids = aids[start : end+1]
	} else {
		pieceAids = aids[start:]
	}
	if arcs, err = s.archives(c, pieceAids); err != nil {
		log.Error("ViewRank s.archives(%v) error(%v)", aids, err)
		return
	}
	for _, aid := range pieceAids {
		if arc, ok := arcs[aid]; ok && arc.IsNormal() {
			HideArcAttribute(arc)
			list = append(list, arc)
		}
	}
	if len(list) == 0 {
		list = _emptyArcs
	}
	return
}

// SetViewRank set view rank arcs.
func (s *Service) SetViewRank(c context.Context, sid int64, aids []int64, typ string) (err error) {
	aidsStr := xstr.JoinInts(aids)
	if err = s.dao.AddCacheViewRank(c, sid, aidsStr, typ); err != nil {
		log.Error("SetViewRank s.dao.AddCacheViewRank(%d,%s,%s) error(%v)", sid, aidsStr, typ, err)
	}
	return
}

// ObjectGroup group like data.
func (s *Service) ObjectGroup(c context.Context, sid int64, ck string) (data map[int64][]*likemdl.GroupItem, err error) {
	var sids []int64
	if sids, err = s.dao.SourceItemData(c, sid); err != nil {
		log.Error("ObjectGroup SourceItemData(%d) error(%+v)", sid, err)
		return
	}
	if len(sids) == 0 {
		log.Warn("ObjectGroup sid(%d) len(sids) == 0", sid)
		err = xecode.NothingFound
		return
	}
	data = make(map[int64][]*likemdl.GroupItem, len(sids))
	group, errCtx := errgroup.WithContext(c)
	mutex := sync.Mutex{}
	for _, v := range sids {
		groupSid := v
		group.Go(func() error {
			item, e := s.dao.GroupItemData(errCtx, groupSid, ck)
			if e != nil {
				log.Error("ObjectGroup s.dao.GroupItemData(%d) error(%+v)", groupSid, e)
			} else {
				mutex.Lock()
				data[groupSid] = item
				mutex.Unlock()
			}
			return nil
		})
	}
	group.Wait()
	return
}

// UpListGroup group up list data.
func (s *Service) UpListGroup(c context.Context, sid, mid int64) (data map[int64][]*likemdl.List, err error) {
	var (
		sids               []int64
		isEnt, isEntSecond bool
	)
	for _, v := range s.c.Ent.UpSids {
		if sid == v {
			isEnt = true
			break
		}
	}
	for _, v := range s.c.Ent.SecondSids {
		if sid == v {
			isEntSecond = true
			break
		}
	}
	if isEnt {
		sids = s.c.Ent.UpSids
	} else if isEntSecond {
		sids = s.c.Ent.SecondSids
	} else {
		if sids, err = s.dao.SourceItemData(c, sid); err != nil {
			log.Error("UpListGroup SourceItemData(%d) error(%+v)", sid, err)
			return
		}
	}
	if len(sids) == 0 {
		log.Warn("UpListGroup sid(%d) len(sids) == 0", sid)
		err = xecode.NothingFound
		return
	}
	data = make(map[int64][]*likemdl.List, len(sids))
	group, errCtx := errgroup.WithContext(c)
	mutex := sync.Mutex{}
	for _, v := range sids {
		groupSid := v
		group.Go(func() error {
			arg := &likemdl.ParamList{Sid: groupSid, Type: dao.ActOrderCtime, Pn: 1}
			// 娱乐大赏活动只需3个数据
			if isEnt || isEntSecond {
				arg.Type = dao.ActOrderLike
				arg.Ps = 3
			}
			item, e := s.StoryKingList(errCtx, arg, mid)
			if e != nil {
				log.Error("UpListGroup s.StoryKingList(%d) error(%+v)", groupSid, e)
			} else if item != nil {
				mutex.Lock()
				data[groupSid] = item.List
				mutex.Unlock()
			}
			return nil
		})
	}
	group.Wait()
	return
}

// SetLikeContent .
func (s *Service) SetLikeContent(c context.Context, lid int64) (err error) {
	var (
		conts map[int64]*likemdl.LikeContent
	)
	if conts, err = s.dao.RawLikeContent(c, []int64{lid}); err != nil {
		log.Error("s.dao.RawLikeContent(%d) error(%+v)", lid, err)
		return
	}
	if _, ok := conts[lid]; !ok {
		conts = make(map[int64]*likemdl.LikeContent, 1)
		conts[lid] = &likemdl.LikeContent{}
	}
	if err = s.dao.AddCacheLikeContent(c, conts); err != nil {
		log.Error("s.dao.AddCacheLikeContent(%d) error(%+v)", lid, err)
	}
	return
}

// AddLikeActCache .
func (s *Service) AddLikeActCache(c context.Context, sid, lid, score int64) (err error) {
	var (
		likeItem *likemdl.Item
	)
	if likeItem, err = s.dao.Like(c, lid); err != nil {
		log.Error("AddLikeActCache:s.dao.Like(%d) error(%+v)", lid, err)
		return
	}
	if likeItem.ID == 0 {
		return
	}
	if err = s.dao.SetRedisCache(c, sid, lid, score, likeItem.Type); err != nil {
		log.Error("AddLikeActCache:s.dao.SetRedisCache(%d,%d,%d) error(%+v)", sid, lid, score, err)
	}
	return
}

// LikeActCache .
func (s *Service) LikeActCache(c context.Context, sid, lid int64) (res int64, err error) {
	return s.dao.LikeActZscore(c, sid, lid)
}

// arcTag get archive and tags.
func (s *Service) arcTag(c context.Context, list []*likemdl.List, order string, mid int64) (err error) {
	var (
		arcsReply    map[int64]*api.Arc
		lt           = len(list)
		wids         = make([]int64, 0, lt)
		tagRes       map[int64][]string
		hasLikeReply *thumbupmdl.HasLikeReply
	)
	for _, v := range list {
		if v.Wid > 0 {
			wids = append(wids, v.Wid)
		}
	}
	eg, errCtx := errgroup.WithContext(c)
	if len(wids) > 0 {
		eg.Go(func() (e error) {
			arcsReply, e = s.archives(errCtx, wids)
			return
		})
	}
	eg.Go(func() (e error) {
		tagRes, e = s.dao.MultiTags(errCtx, wids)
		return
	})
	if mid != 0 && (order == dao.EsOrderLikes || order == dao.ActOrderCtime) {
		eg.Go(func() (e error) {
			hasLikeReply, e = s.thumbupClient.HasLike(errCtx, &thumbupmdl.HasLikeReq{Business: _businessLike, MessageIds: wids, Mid: mid})
			return
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("arcTag:eg.Wait() error(%+v)", err)
		return
	}
	for _, v := range list {
		if v.Wid == 0 {
			continue
		}
		obj := new(likemdl.ArgTag)
		if ac, ok := arcsReply[v.Wid]; ok && ac.IsNormal() {
			if v.Sid == s.c.Rule.LimitArcSid && order == dao.ActOrderStochastic {
				if ac.AttrVal(api.AttrBitNoRecommend) == api.AttrYes {
					// filter
					v.State = 0
					continue
				}
			}
			HideArcAttribute(ac)
			obj.Archive = ac
			if bvidStr, e := bvid.AvToBv(ac.Aid); e == nil && bvidStr != "" {
				obj.Bvid = bvidStr
			}
		}
		if _, ok := tagRes[v.Wid]; ok {
			obj.Tags = tagRes[v.Wid]
		}
		v.Object = obj
		if hasLikeReply != nil {
			if stat, ok := hasLikeReply.States[v.Wid]; ok && stat != nil {
				v.HasLikes = int32(stat.State)
			}
		}
	}
	return
}

// LikeOidsInfo .
func (s *Service) LikeOidsInfo(c context.Context, sType int, oids []int64) (res map[int64]*likemdl.Item, err error) {
	ps := len(oids)
	pn := 1
	if res, err = s.dao.OidInfoFromES(c, oids, sType, ps, pn); err != nil {
		log.Error("s.dao.OidInfoFromES(%v,%d) error(%v)", oids, sType, err)
	}
	return
}

// LikeAddOther .
func (s *Service) LikeAddOther(c context.Context, arg *likemdl.ParamOther, mid int64) (res int64, err error) {
	var (
		subject   *likemdl.SubjectItem
		memberRly *accapi.ProfileReply
		nowTime   = time.Now().Unix()
		imageURL  string
		item      *likemdl.Item
		cont      *likemdl.LikeContent
	)
	if subject, err = s.dao.ActSubject(c, arg.Sid); err != nil {
		log.Error("LikeAddText:s.dao.ActSubject(%d) error(%+v)", arg.Sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if int64(subject.Stime) > nowTime {
		err = ecode.ActivityNotStart
		return
	}
	if int64(subject.Etime) < nowTime {
		err = ecode.ActivityOverEnd
		return
	}
	if subject.TextType() {
		log.Error("LikeAddText:type is not support")
		return
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.upJudgeUser(c, subject, memberRly.Profile); err != nil {
		return
	}
	// 判断用户等级和手机绑定情况
	if memberRly.Profile.Level < 1 {
		err = ecode.ActivityUpLevelLimit
		return
	}
	if memberRly.Profile.TelStatus != 1 {
		err = ecode.ActivityTelValid
		return
	}
	//上传图片
	if imageURL, err = s.bfs.Upload(c, &bfs.Request{Bucket: s.c.Rule.Bucket, Filename: "", ContentType: arg.FileType, File: arg.Image}); err != nil {
		log.Error("s.dao.Upload error(%v)", err)
		return
	}
	item = &likemdl.Item{
		Type:    arg.Type,
		Wid:     arg.Wid,
		Sid:     arg.Sid,
		Mid:     mid,
		State:   0,
		Referer: arg.RefererURI,
	}
	cont = &likemdl.LikeContent{
		Plat:    arg.Plat,
		Device:  arg.Device,
		Message: arg.Message,
		IPv6:    []byte{},
		Image:   imageURL,
	}
	// 特殊活动免审核
	if arg.Sid == s.c.Staff.PicSid {
		item.State = 1
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		cont.IPv6 = IPv6
	}
	if res, err = s.dao.ItemAndContent(c, item, cont); err != nil {
		log.Error(" s.dao.ItemAndContent() error(%d)", err)
	}
	return
}

// LikeAddText .
func (s *Service) LikeAddText(c context.Context, arg *likemdl.ParamText, mid int64) (res int64, err error) {
	var (
		subject   *likemdl.SubjectItem
		memberRly *accapi.ProfileReply
		nowTime   = time.Now().Unix()
		item      *likemdl.Item
		cont      *likemdl.LikeContent
	)
	if subject, err = s.dao.ActSubject(c, arg.Sid); err != nil {
		log.Error("LikeAddText:s.dao.ActSubject(%d) error(%+v)", arg.Sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if int64(subject.Stime) > nowTime {
		err = ecode.ActivityNotStart
		return
	}
	if int64(subject.Etime) < nowTime {
		err = ecode.ActivityOverEnd
		return
	}
	if !subject.TextType() {
		log.Error("LikeAddText:type is not support")
		return
	}
	if memberRly, err = s.accClient.Profile3(c, &accapi.MidReq{Mid: mid}); err != nil {
		log.Error(" s.acc.Profile3(c,&accmdl.ArgMid{Mid:%d}) error(%v)", mid, err)
		return
	}
	if err = s.upJudgeUser(c, subject, memberRly.Profile); err != nil {
		return
	}
	// 新预约数据源-新表处理数据和缓存
	if subject.CacheReserve() {
		res, err = s.Reserve(c, subject.ID, mid, 1)
		return
	}
	IsQuestionnaire := subject.IsQuestionnaire()
	if len(arg.Message) < 1 || (!IsQuestionnaire && len(arg.Message) > 2000) {
		err = xecode.RequestErr
		return
	}
	if s.checkRepeatSubmit(subject) {
		var repet int
		if repet, err = s.dao.TextOnly(c, subject.ID, mid); err != nil {
			log.Error("s.dao.CacheTextOnly(%d,%d) error(%+v)", subject.ID, mid, err)
			return
		}
		if repet > 0 {
			err = ecode.ActivityRepeatSubmit
			return
		}
	}
	item = &likemdl.Item{
		Type:    arg.Type,
		Wid:     arg.Wid,
		Sid:     arg.Sid,
		Mid:     mid,
		State:   0,
		Referer: arg.RefererURI,
	}
	cont = &likemdl.LikeContent{
		Plat:    arg.Plat,
		Message: arg.Message,
		IPv6:    []byte{},
	}
	if IPv6 := net.ParseIP(metadata.String(c, metadata.RemoteIP)); IPv6 != nil {
		cont.IPv6 = IPv6
	}
	if !subject.AttrFlag(likemdl.FLAGFIRST) {
		var filRly *fliapi.FilterReply
		if filRly, err = client.FilterClient.Filter(c, &fliapi.FilterReq{Area: "activity", Message: cont.Message}); err == nil {
			item.State = 1
			cont.Message = filRly.Result
		}
	}
	if IsQuestionnaire {
		if res, err = s.dao.ItemAndContentNew(c, item, cont); err != nil {
			log.Error(" s.dao.ItemAndContentNew() error(%d)", err)
			return
		}
	} else {
		if res, err = s.dao.ItemAndContent(c, item, cont); err != nil {
			log.Error(" s.dao.ItemAndContent() error(%d)", err)
			return
		}
	}
	if s.checkRepeatSubmit(subject) {
		s.dao.AddCacheTextOnly(c, subject.ID, 1, mid)
		s.cache.Do(c, func(c context.Context) {
			s.dao.IncrCacheLikeTotal(c, arg.Sid)
			s.dao.AddCacheLikeCheck(c, mid, item, arg.Sid)
		})
	}
	return
}

func (s *Service) checkRepeatSubmit(subject *likemdl.SubjectItem) bool {
	return !subject.IsDuplicateSubmit() && (subject.CacheOnly() || subject.ID == s.c.Bdf.ImageSid)
}

// LikeTotal .
func (s *Service) LikeTotal(c context.Context, sid int64) (total int64, err error) {
	var subject *likemdl.SubjectItem
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		log.Error("LikeTotal:s.dao.ActSubject(%d) error(%+v)", sid, err)
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if subject.Type != likemdl.QUESTION && subject.Type != likemdl.RESERVATION {
		err = xecode.RequestErr
		return
	}
	if total, err = s.dao.LikeTotal(c, sid); err != nil {
		log.Error("s.dao.LikeTotal(%d) error(%v)", sid, err)
		err = nil
	}
	return
}

// upJudgeUser judge user could like or not .
func (s *Service) upJudgeUser(c context.Context, subject *likemdl.SubjectItem, member *accapi.Profile) (err error) {
	if member.Silence == _silenceForbid {
		err = ecode.ActivityMemberBlocked
		return
	}
	if subject.Flag == 0 {
		return
	}
	if !subject.AttrFlag(likemdl.FLAGUPSPY) {
		var reply *spyapi.InfoReply
		if reply, err = s.spyClient.Info(c, &spyapi.InfoReq{Mid: member.Mid}); err != nil {
			log.Error("s.spyClient.Info(%d) error(%v)", member.Mid, err)
			return
		}
		if reply.Ui == nil || int64(reply.Ui.Score) <= subject.UpScore {
			err = ecode.ActivityUpScoreLower
			return
		}
	}
	if subject.Type == likemdl.CLOCKIN && subject.UpFigureScore > 0 {
		figureReply, _ := s.figureClient.UserFigure(c, &figureapi.UserFigureReq{MID: member.Mid})
		if int64(figureReply.GetPercentage()) > subject.UpFigureScore {
			err = ecode.ActivityUpScoreLower
			return
		}
		return
	}
	if !subject.AttrFlag(likemdl.FLAGUPUSTIME) {
		if subject.UpUstime <= xtime.Time(member.JoinTime) {
			err = ecode.ActivityUpRegisterLimit
			return
		}
	}
	if !subject.AttrFlag(likemdl.FLAGUPUETIME) {
		if subject.UpUetime >= xtime.Time(member.JoinTime) {
			err = ecode.ActivityUpBeforeRegister
			return
		}
	}
	if !subject.AttrFlag(likemdl.FLAGUPPHONEBIND) {
		if member.TelStatus != 1 {
			err = ecode.ActivityTelValid
			return
		}
	}
	if !subject.AttrFlag(likemdl.FLAGUPLEVEL) {
		if subject.UpLevel > int64(member.Level) {
			err = ecode.ActivityUpLevelLimit
			return
		}
	}
	if !subject.AttrFlag(likemdl.FLAGFANLIMIT) {
		var stat *relationapi.StatReply
		if subject.FanLimitMin > 0 || subject.FanLimitMax > 0 {
			if stat, err = s.relClient.Stat(c, &relationapi.MidReq{Mid: member.Mid}); err != nil {
				log.Error("s.relClient.Stat(%d) error(%v)", member.Mid, err)
				return
			}
			if subject.FanLimitMin > 0 && stat.Follower < subject.FanLimitMin {
				err = ecode.ActivityUpFanLimit
				return
			}
			if subject.FanLimitMax > 0 && stat.Follower > subject.FanLimitMax {
				err = ecode.ActivityUpFanLimit
				return
			}
		}
	}
	if !subject.AttrFlag(likemdl.FLAGVIPLIMIT) {
		if !member.Vip.IsValid() {
			err = ecode.ActivityUpVipLimit
			return
		}
	}
	if !subject.AttrFlag(likemdl.FLAGYEARVIPLIMIT) {
		if !member.Vip.IsAnnual() {
			err = ecode.ActivityUpYearVipLimit
		}
	}
	return
}

// LikeCheckJoin .
func (s *Service) LikeCheckJoin(c context.Context, mid, sid int64) (join int, err error) {
	var (
		subject *likemdl.SubjectItem
		item    *likemdl.Item
	)
	if subject, err = s.dao.ActSubject(c, sid); err != nil {
		return
	}
	if subject.ID == 0 {
		err = ecode.ActivityHasOffLine
		return
	}
	if item, err = s.dao.LikeCheck(c, mid, sid); err != nil {
		log.Error("s.dao.LikeCheck mid(%d) sid(%d) error(%v)", mid, sid, err)
		return
	}
	if item != nil && item.Sid != 0 {
		join = 1
	}
	return
}

func (s *Service) LikeArcTypeCount(c context.Context, sid int64) (map[int64]int64, error) {
	subject, err := s.dao.ActSubject(c, sid)
	if err != nil {
		log.Error("LikeArcTypeCount:s.dao.ActSubject(%d) error(%v)", sid, err)
		return nil, err
	}
	if subject.ID == 0 {
		return nil, ecode.ActivityHasOffLine
	}
	var checkType bool
	for _, v := range likemdl.VIDEOALL {
		if subject.Type == v {
			checkType = true
		}
	}
	if !checkType {
		return nil, xecode.RequestErr
	}
	data, err := s.dao.CacheLikeTypeCount(c, sid)
	if err != nil {
		log.Error("LikeArcTypeCount s.dao.LikeTypeCount(%d) error(%v)", sid, err)
		return nil, err
	}
	return data, nil
}

func (s *Service) ListActivityArcs(ctx context.Context, req *pb.ListActivityArcsReq) (*pb.ListActivityArcsReply, error) {
	arcs, err := s.dao.ActivityArchives(ctx, req.Sid, req.Mid)
	if err != nil {
		return nil, err
	}
	if len(arcs) == 0 {
		return nil, xecode.NothingFound
	}
	sort.SliceStable(arcs, func(i, j int) bool {
		return arcs[i].Ctime > arcs[j].Ctime
	})
	out := make([]int64, 0, len(arcs))
	for _, arc := range arcs {
		out = append(out, arc.Wid)
	}
	return &pb.ListActivityArcsReply{
		Aid: out,
	}, nil
}

func (s *Service) YellowGreenVote(ctx context.Context, sid int64) (res *likemdl.YgVote, err error) {
	res = new(likemdl.YgVote)
	period := s.yellowGreenYingYuanPeriod(sid)
	if period == nil {
		err = xecode.Errorf(xecode.RequestErr, "应援活动id不存在")
		return
	}
	for i := 0; i < 3; i++ {
		if res, err = s.dao.CacheYellowGreenVote(ctx, period); err == nil {
			break
		}
		if err != nil {
			log.Errorc(ctx, "YellowGreenVote s.dao.CacheYellowGreenVote error(%+v)", err)
		}
	}
	if res == nil {
		res = &likemdl.YgVote{}
	}
	return
}

func (s *Service) yellowGreenYingYuanPeriod(yingYuanSid int64) *likemdl.YellowGreenPeriod {
	for _, period := range s.c.YellowAndGreen.Period {
		if period.YellowYingYuanSid == yingYuanSid || period.GreenYingYuanSid == yingYuanSid {
			return period
		}
	}
	return nil
}

func (s *Service) ActFilter(ctx context.Context, param *likemdl.ParamFilter) (res *likemdl.ActSensitive, err error) {
	var keys []string
	res = &likemdl.ActSensitive{}
	for _, key := range param.Keys {
		keys = append(keys, "act:"+key)
	}
	if len(keys) == 0 {
		log.Errorc(ctx, "ActFilter keys empty param(%+v)", param)
		err = xecode.RequestErr
		return
	}
	if res.IsSensitive, err = s.filterDao.ActFilter(ctx, param.Area, param.Message, keys, param.Level); err != nil {
		log.Errorc(ctx, "s.filterDao.ActFilter param(%+v) error(%+v)", param, err)
		return
	}
	return
}

func (s *Service) ViewData(ctx context.Context, param *likemdl.ParamViewData) (res []*likemdl.WebDataRes, count int, err error) {
	var (
		list []*likemdl.WebData
		ok   bool
	)
	// 限制需要判断时间的活动只能是_tpWebDataInfo
	if param.Type != _tpWebDataInfo && s.isLimitSid(param.Sid) {
		err = xecode.RequestErr
		log.Errorc(ctx, "ViewData  s.dao.ViewDataCache param(%+v) sid only 2", param)
		return
	}
	if list, ok = s.webViewData[param.Sid]; !ok {
		res = _emptyViewData
		return
	}
	if param.Type == _tpWebDataVideo || param.Type == _tpWebDataInfo {
		if res, count, err = s.arcViewData(ctx, param, list); err != nil {
			log.Errorc(ctx, "ViewData  s.arcViewData param(%+v) error(%+v)", param, err)
			return
		}
	} else {
		res, count = s.commonViewData(param, list)
	}
	if res == nil {
		res = _emptyViewData
	}
	return
}

func (s *Service) isLimitSid(sid int64) bool {
	for _, v := range s.c.OperationSource.InfoSids {
		if sid == v {
			return true
		}
	}
	return false
}

func (s *Service) arcViewData(ctx context.Context, param *likemdl.ParamViewData, list []*likemdl.WebData) (res []*likemdl.WebDataRes, count int, err error) {
	var (
		resList []*likemdl.WebData
		tmpList []*likemdl.WebData
		stime   time.Time
		start   = (param.Pn - 1) * param.Ps
		end     = start + param.Ps - 1
	)
	// 判断开始时间
	if param.Type == _tpWebDataInfo {
		for _, viewData := range list {
			if stime, err = time.ParseInLocation(`2006-01-02 15:04:05`, viewData.Stime, time.Local); err != nil {
				log.Errorc(ctx, "ViewData time.ParseInLocation sid(%d) viewData(%+v) error(%v)", param.Sid, viewData, err)
				return
			}
			if time.Now().Unix() >= stime.Unix() {
				resList = append(resList, viewData)
			}
		}
	} else {
		resList = list
	}
	count = len(resList)
	if count == 0 || count < start {
		res = _emptyViewData
		return
	}
	if count > end+1 {
		tmpList = resList[start : end+1]
	} else {
		tmpList = resList[start:]
	}
	// 返回结果
	for _, dataArc := range tmpList {
		if dataArc == nil {
			log.Errorc(ctx, "ViewData arcViewData sid(%d) dataArc is null", param.Sid)
			continue
		}
		res = append(res,
			&likemdl.WebDataRes{
				ID:      dataArc.ID,
				Vid:     dataArc.Vid,
				Data:    dataArc.OutData,
				Archive: dataArc.Arc,
				Name:    dataArc.Name,
				Stime:   dataArc.Stime,
				Etime:   dataArc.Etime,
				Ctime:   dataArc.Ctime,
				Mtime:   dataArc.Mtime,
			})
	}
	return
}

func (s *Service) commonViewData(param *likemdl.ParamViewData, list []*likemdl.WebData) (res []*likemdl.WebDataRes, count int) {
	var (
		tmpList []*likemdl.WebData
		start   = (param.Pn - 1) * param.Ps
		end     = start + param.Ps - 1
	)
	count = len(list)
	if count == 0 || count < start {
		res = _emptyViewData
		return
	}
	if count > end+1 {
		tmpList = list[start : end+1]
	} else {
		tmpList = list[start:]
	}
	for _, data := range tmpList {
		res = append(res, &likemdl.WebDataRes{
			ID:    data.ID,
			Vid:   data.Vid,
			Data:  data.OutData,
			Name:  data.Name,
			Stime: data.Stime,
			Etime: data.Etime,
			Ctime: data.Ctime,
			Mtime: data.Mtime,
		})
	}
	return
}

//func (s *Service) loadOperationData() {
//	ctx := context.Background()
//	ticker := time.NewTicker(s.c.OperationSource.UpdateTicker)
//	defer func() {
//		ticker.Stop()
//	}()
//	for {
//		select {
//		case <-ticker.C:
//			s.OperationDataSet(ctx)
//		}
//	}
//}

func (s *Service) OperationDataSet(ctx context.Context) {
	tmpWebData := make(map[int64][]*likemdl.WebData)
	for _, sid := range s.c.OperationSource.OperationSids {
		list, err := s.dao.ViewDataCache(ctx, sid)
		if err != nil {
			log.Errorc(ctx, "OperationDataSet  s.dao.ViewDataCache sid(%d) error(%+v)", sid, err)
			continue
		}
		if len(list) == 0 {
			log.Infoc(ctx, "OperationDataSet s.dao.ViewDataCache(%d) count 0", sid)
			continue
		}
		tmpWebData[sid] = viewDataList(ctx, list)
	}
	s.webViewData = tmpWebData
}

func viewDataList(ctx context.Context, list []*likemdl.WebData) (res []*likemdl.WebData) {
	var (
		err       error
		aids      []int64
		arcsReply *api.ArcsReply
	)
	for _, data := range list {
		var (
			bv  string
			aid int64
		)
		mapData := make(map[string]interface{}, 0)
		if data.Data != "" {
			if err = json.Unmarshal([]byte(data.Data), &mapData); err != nil {
				// 忽略错误，打日志
				log.Errorc(ctx, "OperationDataSet viewDataList json.Unmarshal error(%v)", err)
			} else {
				if tmpBvid, ok := mapData["bvid"]; ok {
					bv = tmpBvid.(string)
					if tmpAid, e := bvid.BvToAv(bv); e != nil {
						// 忽略错误，打日志
						log.Errorc(ctx, "OperationDataSet bvid.BvToAv bv(%s) error(%v)", bv, e)
					} else {
						aids = append(aids, tmpAid)
						aid = tmpAid
					}
				}
			}
		}
		res = append(res,
			&likemdl.WebData{
				ID:      data.ID,
				Vid:     data.Vid,
				Data:    data.Data,
				OutData: mapData,
				Bvid:    bv,
				Aid:     aid,
				Name:    data.Name,
				Stime:   data.Stime,
				Etime:   data.Etime,
				Ctime:   data.Ctime,
				Mtime:   data.Mtime,
			})
	}
	if len(aids) == 0 {
		return
	}
	if arcsReply, err = client.ArchiveClient.Arcs(ctx, &api.ArcsRequest{Aids: aids}); err != nil {
		log.Errorc(ctx, "OperationDataSet arcViewData client.ArchiveClient.Arcs aids(%v) error(%v)", aids, err)
		return
	}
	if arcsReply == nil {
		log.Errorc(ctx, "OperationDataSet arcViewData client.ArchiveClient.Arcs aids(%v) reply.Arcs is nil", aids)
		return
	}
	for _, resData := range res {
		var arc *api.Arc
		if resData == nil || resData.Aid == 0 {
			continue
		}
		if archive, arcOK := arcsReply.Arcs[resData.Aid]; arcOK {
			arc = archive
		}
		resData.Arc = arc
	}
	return
}
