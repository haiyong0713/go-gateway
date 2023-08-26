package article

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thoas/go-funk"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/metric"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/article"
	"go-gateway/app/app-svr/hkt-note/service/model/note"
	"reflect"
	"strconv"
	"strings"
	"sync"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/hkt-note/common"
	noteSrv "go-gateway/app/app-svr/hkt-note/service/service/note"
)

const (
	AutoPullCvidMaxSize          = 200
	AutoPullCvidArrayShouldLenth = 2
)

var (
	Metric_ArcTagCount    = metric.NewBusinessMetricCount("note_arc_tag", "type")
	Metric_ArcForbidCount = metric.NewBusinessMetricCount("note_arc_forbid", "type")
)

func (s *Service) PublishListInUser(c context.Context, req *api.NoteListReq) (*api.NoteListReply, error) {
	// 公开笔记总数
	total, err := s.artDao.ArtCountInUser(c, req.Mid)
	if err != nil {
		log.Error("artError PublishListInUser req(%+v) err(%+v)", req, err)
		return nil, err
	}
	page := &api.Page{
		Total: int64(total),
		Size_: req.Ps,
		Num:   req.Pn,
	}
	if page.Total == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	min, max := note.ToPage(req.Pn, req.Ps)
	// ps, pn超过total上限
	if min >= page.Total {
		return &api.NoteListReply{Page: page}, nil
	}
	// 当前页笔记ids
	ids, err := s.artDao.ArtListInUser(c, min, max, req.Mid)
	if err != nil {
		log.Error("artError PublishListInUser req(%+v) err(%+v)", req, err)
		return nil, err
	}
	if len(ids) == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	cvids, _ := article.ToArtKeys(ids)
	// 批量获取笔记详情
	cacheDetails, err := s.artDao.ArtDetails(c, cvids, article.TpArtDetailCvid)
	if err != nil {
		log.Error("artError PublishListInUser req(%+v) err(%+v)", req, err)
		return nil, err
	}
	// 获取笔记所属稿件详情
	var (
		eg2        = errgroup.WithContext(c)
		aids, sids = article.ToVideoIds(cacheDetails)
		arcs       map[int64]*arcapi.Arc
		ssns       map[int32]*cssngrpc.SeasonCard
	)
	if len(aids) > 0 {
		eg2.Go(func(c context.Context) (err error) {
			arcs, err = s.noteDao.Arcs(c, aids)
			return err
		})
	}
	if len(sids) > 0 {
		eg2.Go(func(c context.Context) (err error) {
			ssns, err = s.noteDao.CheeseSeasons(c, sids)
			return err
		})
	}
	if err = eg2.Wait(); err != nil {
		log.Error("artError PublishListInUser req(%+v) err(%+v)", req, err)
		return nil, err
	}
	res := note.DealArtListItem(page, cvids, cacheDetails, arcs, ssns, s.c.NoteCfg.WebPubUrlFromSpace)
	return res, nil
}

func (s *Service) PublishListInArc(c context.Context, req *api.NoteListReq) (*api.NoteListReply, error) {
	// 公开笔记总数
	total, err := s.artDao.ArtCountInArc(c, req.Oid, req.OidType)
	if err != nil {
		log.Error("artError PublishListInArc req(%+v) err(%+v)", req, err)
		return nil, err
	}
	page := &api.Page{
		Total: total,
		Size_: req.Ps,
		Num:   req.Pn,
	}
	if page.Total == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	min, max := note.ToPage(req.Pn, req.Ps)
	// ps, pn超过total上限
	if min >= page.Total {
		return &api.NoteListReply{Page: page}, nil
	}
	// 当前页笔记ids
	ids, err := s.artDao.ArtListInArc(c, min, max, req.Oid, req.OidType)
	if err != nil {
		log.Error("artError PublishListInArc req(%+v) err(%+v)", req, err)
		return nil, err
	}
	if len(ids) == 0 {
		return &api.NoteListReply{Page: page}, nil
	}
	// 获取该稿件uper的笔记，若无则设置无up公开笔记，否则在第一页置顶
	// 在后续页如果有uper的笔记noteid，过滤
	uperCvidNoteIdStr := s.findUperNote(c, req.UperMid, req.Oid, req.OidType)
	targetIds := make([]string, 0)
	if len(uperCvidNoteIdStr) > 0 {
		//存在uper主的笔记
		// 先将ids过滤一遍，去掉uperNoteId
		for _, id := range ids {
			if id != uperCvidNoteIdStr {
				targetIds = append(targetIds, id)
			}
		}
		if req.Pn == 1 {
			//第一页数据，将up主的笔记插在第一条
			targetIds = append([]string{uperCvidNoteIdStr}, targetIds...)
		}
	} else {
		targetIds = ids
	}
	var (
		eg           = errgroup.WithContext(c)
		cvids, _     = article.ToArtKeys(targetIds)
		cacheDetails map[int64]*article.ArtDtlCache
		artMetas     map[int64]*artmdl.Meta
		hasLikeMap   map[int64]*thumbupgrpc.UserLikeState
	)
	// 批量获取笔记详情
	eg.Go(func(c context.Context) error {
		var e error
		cacheDetails, e = s.artDao.ArtDetails(c, cvids, article.TpArtDetailCvid)
		return e
	})
	// 批量获取专栏点赞信息
	eg.Go(func(c context.Context) error {
		var e error
		if artMetas, e = s.artDao.ArticleMetasSimple(c, cvids); e != nil {
			log.Warn("artWarn err(%+v)", e)
			artMetas = make(map[int64]*artmdl.Meta)
		}
		return nil
	})
	// 用户点赞信息
	if req.Mid > 0 {
		eg.Go(func(c context.Context) error {
			var e error
			if hasLikeMap, e = s.artDao.HasLike(c, cvids, req.Mid); e != nil {
				log.Warn("artWarn err(%+v)", e)
				hasLikeMap = make(map[int64]*thumbupgrpc.UserLikeState)
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("artError PublishListInArc req(%+v) err(%+v)", req, err)
		return nil, err
	}
	list := make([]*api.NoteSimple, 0, len(cacheDetails))
	for _, cvid := range cvids {
		dtl, ok := cacheDetails[cvid]
		if !ok {
			log.Warn("artWarn PublishListInArc cvid(%d) detail not found,skip", cvid)
			continue
		}
		if dtl.PubStatus != article.PubStatusPassed {
			log.Warn("artInfo PublishListInArc cvid(%d) detail(%+v) isn't pass,skip", cvid, dtl)
			continue
		}
		var (
			likes   int64
			hasLike bool
		)
		if meta, metaOk := artMetas[cvid]; metaOk {
			if meta != nil && meta.Stats != nil {
				likes = meta.Stats.Like
			}
		}
		if hasLikeMap != nil {
			if thumbupState, tOk := hasLikeMap[cvid]; tOk {
				hasLike = thumbupState.State == thumbupgrpc.State_STATE_LIKE
			}
		}
		webUrl := fmt.Sprintf(s.c.NoteCfg.WebPubUrlFromArc, dtl.Cvid)
		list = append(list, dtl.ToCard(webUrl, likes, hasLike))
	}
	return &api.NoteListReply{Page: page, List: list}, nil
}

func (s *Service) findUperNote(c context.Context, uperMid, oid, oidType int64) (cvidNoteIdStr string) {
	if uperMid == 0 {
		return ""
	}
	userArcNoteIds, err := s.noteDao.NoteAid(c, &api.NoteListInArcReq{
		Mid:     uperMid,
		Oid:     oid,
		OidType: oidType,
	})
	log.Warnc(c, "findUperNote userArcNoteIds  %v uperMid %v oid %v oidType %v", userArcNoteIds, uperMid, oid, oidType)
	if err != nil {
		log.Errorc(c, "findUperNote NoteAid err(%+v) uperMid %v oid %v oidType %v", err, uperMid, oid, oidType)
		return ""
	}
	if len(userArcNoteIds) == 0 {
		return ""
	}
	// 判断是否公开
	artDtl, err := s.artDao.RawArtDetail(c, userArcNoteIds[0], article.TpArtDetailNoteId, article.PubStatusPassed, 0)
	if err != nil {
		log.Errorc(c, "findUperNote NoteAid err(%+v) uperMid %v oid %v oidType %v", err, uperMid, oid, oidType)
		return ""
	}
	if artDtl != nil && artDtl.Cvid > 0 {
		return article.ToArtListVal(artDtl.Cvid, userArcNoteIds[0])
	}
	return ""
}

func (s *Service) PublishNoteInfo(c context.Context, req *api.PublishNoteInfoReq) (*api.PublishNoteInfoReply, error) {
	dtl, err := s.artDao.ArtDetail(c, req.Cvid, article.TpArtDetailCvid)
	if err != nil {
		log.Error("artError err(%+v)", err)
		return nil, err
	}
	if dtl.Cvid == -1 || dtl.Deleted == 1 || dtl.PubStatus == article.PubStatusLock {
		log.Error("artError PublishNoteInfo req(%+v) dtl(%+v) nil", req, dtl)
		return nil, xecode.ArtDetailNotFound
	}
	var (
		eg                  = errgroup.WithContext(c)
		cont                *article.ArtContCache
		artCvidCnt          int64
		hasPubSuccessBefore bool
	)
	eg.Go(func(c context.Context) error {
		var e error
		cont, e = s.artDao.ArtContent(c, req.Cvid, dtl.PubVersion)
		if e != nil {
			return e
		}
		if cont.Cvid == -1 || cont.Deleted == 1 {
			log.Error("artError PublishNoteInfo req(%+v) cont(%+v) nil", req, cont)
			return xecode.ArtContentNotFound
		}
		return nil
	})
	eg.Go(func(c context.Context) error {
		var e error
		artCvidCnt, e = s.artDao.ArtCountInArc(c, dtl.Oid, int64(dtl.OidType))
		return e
	})
	eg.Go(func(c context.Context) error {
		hasPubSuccessBefore, _ = s.artDao.GetPubSuccessCvidsBeforeAssignedVersion(c, dtl.Cvid, dtl.PubVersion)
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("artError err(%+v)", err)
		return nil, err
	}
	var (
		tags   []*api.NoteTag
		cidCnt int64
	)
	tags, cidCnt, err = s.noteDao.ToTags(c, dtl.Oid, dtl.NoteId, cont.Tag, dtl.OidType)
	if err != nil {
		log.Error("artError PublishNoteInfo req(%+v) err(%+v)", req, err)
		return nil, err
	}
	res := &api.PublishNoteInfoReply{
		Cvid:                dtl.Cvid,
		NoteId:              dtl.NoteId,
		Title:               dtl.Title,
		Summary:             dtl.Summary,
		Content:             cont.Content,
		Tags:                tags,
		CidCount:            cidCnt,
		PubStatus:           int64(dtl.PubStatus),
		PubReason:           dtl.PubReason,
		Oid:                 dtl.Oid,
		OidType:             int64(dtl.OidType),
		Mid:                 dtl.Mid,
		ArcCvidCnt:          artCvidCnt,
		Mtime:               int64(dtl.Mtime),
		PubTime:             int64(dtl.Pubtime),
		HasPubSuccessBefore: hasPubSuccessBefore,
	}
	return res, nil
}

func (s *Service) SimpleArticles(c context.Context, req *api.SimpleArticlesReq) (*api.SimpleArticlesReply, error) {
	dtl, err := s.artDao.ArtDetails(c, req.Cvids, article.TpArtDetailCvid)
	if err != nil {
		log.Error("ArtError SimpleArticles err(%+v)", err)
		return nil, err
	}
	res := make(map[int64]*api.SimpleArticleCard)
	for _, d := range dtl {
		res[d.Cvid] = d.ToSimpleCard()
	}
	return &api.SimpleArticlesReply{Items: res}, nil
}

func (s *Service) UpArc(c context.Context, req *api.UpArcReq) (*api.UpArcReply, error) {
	// 公开笔记总数 从art_detail中获取该稿件下公开的笔记（不包含pubstatus lock）
	total, err := s.artDao.ArtCountInArc(c, req.Oid, req.OidType)
	if err != nil {
		log.Error("artError UpArc req(%+v) err(%+v)", req, err)
		return &api.UpArcReply{}, nil
	}
	if total == 0 {
		return &api.UpArcReply{}, nil
	}
	var (
		eg               = errgroup.WithContext(c)
		allArcPubCvids   []int64 // 该稿件下公开笔记cvids
		allArcPubNoteIds []int64 // 该稿件下公开笔记ids
		userArcNoteIds   []int64 // 某用户在该稿件下笔记ids
	)
	eg.Go(func(c context.Context) error {
		ids, e := s.artDao.ArtListInArc(c, 0, total, req.Oid, req.OidType)
		if e != nil {
			log.Warn("artWarn UpArc req(%+v) err(%+v)", req, err)
			return nil
		}
		allArcPubCvids, allArcPubNoteIds = article.ToArtKeys(ids)
		return nil
	})
	eg.Go(func(c context.Context) error {
		var e error
		//note-detail表，uper在oid下写的笔记
		userArcNoteIds, e = s.noteDao.NoteAid(c, &api.NoteListInArcReq{
			Mid:     req.UpperId,
			Oid:     req.Oid,
			OidType: req.OidType,
		})
		if e != nil {
			log.Warn("artWarn UpArc req(%+v) err(%+v)", req, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("artError UpArc req(%+v) err(%+v)", req, err)
		return nil, err
	}
	if len(allArcPubNoteIds) == 0 || len(userArcNoteIds) == 0 {
		return &api.UpArcReply{}, nil
	}
	exist := make(map[int64]struct{})
	for _, id := range userArcNoteIds {
		if id <= 0 {
			continue
		}
		exist[id] = struct{}{}
	}
	for idx, id := range allArcPubNoteIds {
		if _, ok := exist[id]; ok {
			var link string
			if len(allArcPubCvids)-1 >= idx {
				link = fmt.Sprintf(s.c.NoteCfg.UpPubUrl, allArcPubCvids[idx])
			}
			return &api.UpArcReply{NoteId: id, JumpLink: link}, nil
		}
	}
	return &api.UpArcReply{}, nil
}

func (s *Service) ArcTag(ctx context.Context, req *api.ArcTagReq, noteSrv *noteSrv.Service) (resp *api.ArcTagReply, err error) {
	if !s.arcTagAllower.Allow() {
		err = ecode.Error(ecode.LimitExceed, "ArcTag 发送的太快了,请稍后再试!")
		return &api.ArcTagReply{}, err
	}
	Metric_ArcTagCount.Inc("req")
	// 是否命中直接拉起笔记
	//autoPullCvid := s.autoPullCvid(ctx, req.Oid)
	var autoPullCvid int64
	resp = &api.ArcTagReply{
		AutoPullCvid: autoPullCvid,
	}
	isLoginUserUper := false //当前登录用户是否是稿件up主
	if req.LoginMid > 0 && req.LoginMid == req.UpperId {
		isLoginUserUper = true
	}
	// 公开笔记总数 从art_detail中获取该稿件下公开的笔记 如果笔记对应最新一篇专栏是审核锁定，则不算
	total, err := s.artDao.ArtCountInArc(ctx, req.Oid, req.OidType)
	if err != nil {
		log.Errorc(ctx, "ArcTag ArtCountInArc err %v req(%+v) ", err, req)
		return resp, nil
	}
	if total == 0 {
		//该稿件下无公开笔记
		resp = s.processArcTagEdit(ctx, isLoginUserUper, req, noteSrv, total, autoPullCvid)
		return resp, nil
	}
	var (
		eg               = errgroup.WithContext(ctx)
		allArcPubCvids   []int64 // 该稿件下公开笔记cvids
		allArcPubNoteIds []int64 // 该稿件下公开笔记ids
		userArcNoteIds   []int64 // 某用户在该稿件下笔记ids
	)
	eg.Go(func(c context.Context) error {
		ids, e := s.artDao.ArtListInArc(c, 0, total, req.Oid, req.OidType)
		if e != nil {
			log.Warnc(ctx, "artWarn UpArc req(%+v) err(%+v)", req, err)
			return nil
		}
		allArcPubCvids, allArcPubNoteIds = article.ToArtKeys(ids)
		return nil
	})
	eg.Go(func(c context.Context) error {
		var e error
		//note-detail表，uper在oid下写的笔记
		userArcNoteIds, e = s.noteDao.NoteAid(c, &api.NoteListInArcReq{
			Mid:     req.UpperId,
			Oid:     req.Oid,
			OidType: req.OidType,
		})
		if e != nil {
			log.Warnc(ctx, "artWarn UpArc req(%+v) err(%+v)", req, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "artError UpArc req(%+v) err(%+v)", req, err)
		return nil, err
	}
	if len(allArcPubNoteIds) == 0 || len(userArcNoteIds) == 0 {
		resp = s.processArcTagEdit(ctx, isLoginUserUper, req, noteSrv, total, autoPullCvid)
		return resp, nil
	}
	exist := make(map[int64]struct{})
	for _, id := range userArcNoteIds {
		if id <= 0 {
			continue
		}
		exist[id] = struct{}{}
	}
	for idx, id := range allArcPubNoteIds {
		if _, ok := exist[id]; ok {
			var link string
			if len(allArcPubCvids)-1 >= idx {
				link = fmt.Sprintf(s.c.NoteCfg.UpPubUrl, allArcPubCvids[idx])
			}
			Metric_ArcTagCount.Inc("hitDetail")
			return &api.ArcTagReply{
				NoteId:       id,
				JumpLink:     link,
				TagShowText:  "up主笔记",
				NotesCount:   total,
				AutoPullCvid: autoPullCvid,
			}, nil
		}
	}
	resp = s.processArcTagEdit(ctx, isLoginUserUper, req, noteSrv, total, autoPullCvid)
	return resp, nil
}

func (s *Service) processArcTagEdit(ctx context.Context, isLoginUserUper bool, req *api.ArcTagReq, noteSrv *noteSrv.Service, total int64, autoPullCvid int64) (resp *api.ArcTagReply) {
	if isLoginUserUper && s.isArcSatisfyEditTag(ctx, req, noteSrv) {
		tagLink, tagShowText := s.buildArcTagNoteEdit(req.Oid)
		Metric_ArcTagCount.Inc("hitEdit")
		return &api.ArcTagReply{
			NoteId:       0,
			JumpLink:     tagLink,
			TagShowText:  tagShowText,
			NotesCount:   total,
			AutoPullCvid: autoPullCvid,
		}
	}
	return &api.ArcTagReply{
		NotesCount:   total,
		AutoPullCvid: autoPullCvid,
	}
}

// 稿件是否满足笔记编辑器的展示要求
func (s *Service) isArcSatisfyEditTag(ctx context.Context, req *api.ArcTagReq, noteSrv *noteSrv.Service) (res bool) {
	//二级分区是否在白名单
	if !funk.ContainsInt64(s.c.NoteCfg.ArcTagCfg.AllowTypeIds, int64(req.SubTypeId)) {
		log.Warnc(ctx, "%v isArcSatisfyEditTag not in allowTypeId", req.Oid)
		return false
	}
	// 请求forbid判断稿件是否被管控
	Metric_ArcForbidCount.Inc("sceneArcTag")
	arcsForbidReq := &api.ArcsForbidReq{
		Aids: []int64{req.Oid},
	}
	arcsForbidReply, err := noteSrv.ArcsForbid(ctx, arcsForbidReq)
	if err != nil || arcsForbidReply == nil {
		log.Warnc(ctx, "%v isArcSatisfyEditTag forbid err %v", req.Oid, err)
		//请求失败视为被管控
		return false
	}
	if arcForbidReply, ok := arcsForbidReply.Items[req.Oid]; ok {
		//明确获取forbid为true视为被管控
		if arcForbidReply {
			return false
		}
	}
	return true
}

func (s *Service) buildArcTagNoteEdit(aid int64) (link string, tagShowText string) {
	link = fmt.Sprintf(s.c.NoteCfg.ArcTagCfg.EditNoteTagLink, aid)
	tagShowText = s.c.NoteCfg.ArcTagCfg.TagShowText
	return link, tagShowText
}

// 返回值cvid >0 意味着有直接要拉起展示的笔记
/*func (s *Service) autoPullCvid(ctx context.Context, aid int64) (cvid int64) {
	if aid <= 0 {
		return 0
	}
	// 是否命中拉起白名单
	cvid, err := s.artDao.GetAutoPullCvid(ctx, aid)
	if err != nil {
		return 0
	}
	return cvid
}*/

// 用于pre环境导白名单
func (s *Service) AutoPullCvid(ctx context.Context, req *api.AutoPullAidCivdReq) (resp *api.AutoPullAidCivdReply, err error) {
	resp = &api.AutoPullAidCivdReply{}
	log.Warnc(ctx, "AutoPullCvid req  size %v", len(req.AidToCvids))
	if len(req.AidToCvids) > AutoPullCvidMaxSize {
		log.Warnc(ctx, "AutoPullCvid req too long  bvid size %v", len(req.AidToCvids))
		return resp, nil
	}
	//将bvids分十组
	partSize := 5
	aidCvidPart := make([][]string, partSize)
	for index, aidToCvidStr := range req.AidToCvids {
		targetIndex := index % partSize
		aidCvidPart[targetIndex] = append(aidCvidPart[targetIndex], aidToCvidStr)
	}
	errG := errgroup.WithContext(ctx)
	errG.GOMAXPROCS(partSize)
	failedAid := make([]string, 0)
	lock := sync.Mutex{}
	for i := 0; i < partSize; i++ {
		if len(aidCvidPart[i]) == 0 {
			continue
		}
		curAidCvidPart := aidCvidPart[i]
		errG.Go(func(ctx context.Context) error {
			//本次处理的bvids
			for _, aidToCvid := range curAidCvidPart {
				temp := strings.Split(aidToCvid, ":")
				if len(temp) != AutoPullCvidArrayShouldLenth {
					lock.Lock()
					failedAid = append(failedAid, aidToCvid)
					lock.Unlock()
					continue
				}
				aid, err := strconv.ParseInt(temp[0], 10, 64)
				if err != nil {
					log.Errorc(ctx, "AutoPullCvid ParseInt aid  err %v aidToCvid %v", err, aidToCvid)
					return err
				}
				cvid, err := strconv.ParseInt(temp[1], 10, 64)
				if err != nil {
					log.Errorc(ctx, "AutoPullCvid ParseInt cvid  err %v aidToCvid %v", err, aidToCvid)
					return err
				}
				err = s.artDao.SetAutoPullCvid(ctx, aid, cvid)
				return err
			}
			return nil
		})
	}
	if err = errG.Wait(); err != nil {
		log.Errorc(ctx, "AutoPullCvid errG err %v", err)
	}
	log.Warnc(ctx, "AutoPullCvid failed info %v", failedAid)
	return resp, err
}

func (s *Service) ArcNotesCount(ctx context.Context, req *api.ArcNotesCountReq) (resp *api.ArcNotesCountReply, err error) {
	resp = &api.ArcNotesCountReply{}
	// 公开笔记总数 从art_detail中获取该稿件下公开的笔记 如果笔记对应最新一篇专栏是审核锁定，则不算
	total, err := s.artDao.ArtCountInArc(ctx, req.Oid, req.OidType)
	if err != nil {
		log.Errorc(ctx, "ArcNotesCount ArtCountInArc err %v req(%+v) ", err, req)
		return resp, xecode.ArtNoteCountFail
	}
	resp.NotesCount = total
	return resp, nil
}

func (s *Service) BatchGetReplyRenderInfo(ctx context.Context, req *api.BatchGetReplyRenderInfoReq) (resp *api.BatchGetReplyRenderInfoRes, err error) {
	resp = &api.BatchGetReplyRenderInfoRes{
		Items: make(map[int64]*api.ReplyRenderInfoItem),
	}
	if len(req.Cvids) == 0 {
		return resp, xecode.BatchGetReplyRenderInfoReqInvalid
	}
	if len(req.Cvids) > common.Batch_Reply_Max_Limit {
		return resp, xecode.BatchGetReplyRenderInfoReqCvidOverSize
	}
	var (
		// key是cvid
		artDetailMap map[int64]*article.ArtDtlCache
		// key是oid
		arcInfoMap map[int64]*arcapi.Arc
	)
	// 批量获取cvid对应的articleDetail 和 oid对应的Arcs信息
	artDetailMap, err = s.artDao.ArtDetails(ctx, req.Cvids, article.TpArtDetailCvid)
	if err != nil {
		log.Errorc(ctx, "BatchGetReplyRenderInfo getArtDetails error (%+v) req %v", err, req)
		return resp, xecode.BatchGetReplyRenderInfoFail
	}
	oids := parseOidsByArticleDetail(artDetailMap)
	arcInfoMap, err = s.noteDao.Arcs(ctx, oids)
	if err != nil {
		log.Errorc(ctx, "BatchGetReplyRenderInfo get Arcs Info err %v oids %v req %v", err, oids, req)
		return resp, xecode.BatchGetReplyRenderInfoFail
	}

	// cvids分成几部分，并发处理
	cvidsPart := make([][]int64, common.Batch_Reply_Concurrent_Size)
	for index, curCvid := range req.Cvids {
		targetIndex := index % common.Batch_Reply_Concurrent_Size
		cvidsPart[targetIndex] = append(cvidsPart[targetIndex], curCvid)
	}
	log.Warnc(ctx, "BatchGetReplyRenderInfo cvidsPart %v ", cvidsPart)
	errGroup := errgroup.WithContext(ctx)
	errGroup.GOMAXPROCS(common.Batch_Reply_Concurrent_Size)
	lock := sync.Mutex{}
	for _, curPartList := range cvidsPart {
		func(cvidList []int64) {
			errGroup.Go(func(ctx context.Context) error {
				for _, cvid := range cvidList {
					if cvid <= 0 {
						continue
					}
					item := s.buildReplyRenderInfoItem(ctx, cvid, artDetailMap, arcInfoMap)
					if item != nil {
						lock.Lock()
						resp.Items[cvid] = item
						lock.Unlock()

					}
				}
				return nil
			})
		}(curPartList)
	}
	if err = errGroup.Wait(); err != nil {
		log.Errorc(ctx, "BatchGetReplyRenderInfo errG err %v", err)
		return resp, xecode.BatchGetReplyRenderInfoFail
	}
	return resp, nil
}

func parseOidsByArticleDetail(artDetailMap map[int64]*article.ArtDtlCache) (oids []int64) {
	oids = make([]int64, 0)
	for _, value := range artDetailMap {
		oids = append(oids, value.Oid)
	}
	return oids
}

func (s *Service) buildReplyRenderInfoItem(ctx context.Context, cvid int64, artDetailMap map[int64]*article.ArtDtlCache, arcInfoMap map[int64]*arcapi.Arc) (item *api.ReplyRenderInfoItem) {
	if _, found := artDetailMap[cvid]; !found {
		log.Warnc(ctx, "buildReplyRenderInfoItem not found artDetail by cvid %v", cvid)
		return nil
	}
	// 获取cvid对应的article_detail
	curArtDetail := artDetailMap[cvid]
	if curArtDetail.Cvid == -1 || curArtDetail.Deleted == 1 || curArtDetail.PubStatus == article.PubStatusLock {
		log.Errorc(ctx, "buildReplyRenderInfoItem not found valid article detail cvid %v artDetail %v", cvid, curArtDetail)
		return nil
	}
	// 构建item基础信息
	var lastMTimeText string
	hasPubSuccessBefore, _ := s.artDao.GetPubSuccessCvidsBeforeAssignedVersion(ctx, cvid, curArtDetail.PubVersion)
	log.Warnc(ctx, "buildReplyRenderInfoItem hasPubSuccessBefore cvid %v pubVersion %v hasPubSuccessBefore %v", cvid, curArtDetail.PubVersion, hasPubSuccessBefore)

	if hasPubSuccessBefore {
		lastMTimeText = fmt.Sprintf("%s%s", common.Publish_Info_Last_Time_Text_Prefix, curArtDetail.Pubtime.Time().Format("2006-01-02"))
	} else {
		lastMTimeText = curArtDetail.Pubtime.Time().Format("2006-01-02")
	}
	item = &api.ReplyRenderInfoItem{
		Summary:       curArtDetail.Summary,
		ClickUrl:      fmt.Sprintf(s.c.NoteCfg.ReplyCfg.WebUrl, cvid),
		LastMtimeText: lastMTimeText,
	}
	if len(item.Summary) == 0 {
		item.Summary = s.c.NoteCfg.ReplyCfg.ReplySummaryTextDefault
	}
	// 判断是否展示图片
	// 判断分区是否在白名单中
	if _, found := arcInfoMap[curArtDetail.Oid]; !found {
		log.Warnc(ctx, "buildReplyRenderInfoItem not found arcInfo by cvid %v oid %v", cvid, curArtDetail.Oid)
		return item
	}
	if !funk.ContainsInt64(s.c.NoteCfg.ReplyCfg.ImageAllowTypeIds, int64(arcInfoMap[curArtDetail.Oid].TypeID)) {
		log.Warnc(ctx, "buildReplyRenderInfoItem not in imageAllowTypeId req %v oid %v", cvid, curArtDetail.Oid)
		return item
	}
	// 构建图片信息
	item.Images = s.buildReplyRenderInfoItemImage(ctx, cvid, curArtDetail.PubVersion)
	return item
}

// 获取笔记内容中的截屏图片
func (s *Service) buildReplyRenderInfoItemImage(ctx context.Context, cvid int64, pubVersion int64) (videoImages []string) {
	videoImages = make([]string, 0)
	artContent, err := s.artDao.ArtContent(ctx, cvid, pubVersion)
	if err != nil {
		log.Errorc(ctx, "buildReplyRenderInfoItemImage get article content err %v cvid %v", err, cvid)
		return videoImages
	}
	if artContent.Cvid == -1 || artContent.Deleted == 1 {
		log.Errorc(ctx, "buildReplyRenderInfoItemImage not found valid article content cvid %v artContent %v", cvid, artContent)
		return videoImages
	}

	contentItems := make([]*common.ContentBody, 0)
	if err := json.Unmarshal([]byte(artContent.Content), &contentItems); err != nil {
		log.Errorc(ctx, "buildReplyRenderInfoItemImage unmarshal content err %v contentStr %v", err, artContent.Content)
		return []string{}
	}

	imageShowInReplyMaxLimit := s.c.NoteCfg.ReplyCfg.ImageShowInReplyMaxLimit
	if imageShowInReplyMaxLimit == 0 {
		imageShowInReplyMaxLimit = common.ImageShowInReplyMaxLimitDefault
	}
	for _, curItem := range contentItems {
		if int32(len(videoImages)) >= imageShowInReplyMaxLimit {
			return videoImages
		}
		if curItem.Insert == nil {
			continue
		}
		if reflect.TypeOf(curItem.Insert).Name() == "string" {
			continue
		}
		//判断是否是图片
		contentInsert := &common.ContentInsert{}
		contentInsertBytes, err := json.Marshal(curItem.Insert)
		if err != nil {
			log.Errorc(ctx, "buildReplyRenderInfoItemImage marshal curItem.Insert err %v value %v", err, curItem.Insert)
			continue
		}
		if err := json.Unmarshal(contentInsertBytes, &contentInsert); err != nil {
			log.Errorc(ctx, "buildReplyRenderInfoItemImage unmarshal contentInsert err %v contentInsertBytes %v", err, contentInsertBytes)
			continue
		}
		if contentInsert.ImageUpload == nil {
			continue
		}
		if contentInsert.ImageUpload.Source == common.ImageUploadSourceVideo {
			videoImages = append(videoImages, "https:"+contentInsert.ImageUpload.Url)
		}
	}
	return videoImages
}

func (s *Service) GetAttachedRpid(ctx context.Context, req *api.GetAttachedRpidReq) (resp *api.GetAttachedRpidReply, err error) {
	if !s.getAttachedRpidAllower.Allow() {
		err = ecode.Error(ecode.LimitExceed, "GetAttachedRpid 发送的太快了,请稍后再试!")
		return nil, err
	}
	resp = &api.GetAttachedRpidReply{
		Rpid: 0,
	}
	rpidInfo := &common.TaishanCvidMappingRpidInfo{}
	key := fmt.Sprintf(common.Cvid_Mapping_Rpid_Taishan_Key, req.Cvid)
	record, err := s.artDao.GetTaishan(ctx, key, common.TaishanConfig.NoteReply)
	if err != nil {
		log.Errorc(ctx, "GetAttachedRpid query user info from taiShan failed:(%v), key:%v", err, key)
		return
	}
	if record == nil || len(record.Columns) == 0 || len(record.Columns[0].Value) == 0 {
		log.Infoc(ctx, "GetAttachedRpid query cvidInfo from taiShan record isEmpty:(%v), key:%v", record, key)
		return
	}
	err = json.Unmarshal(record.Columns[0].Value, rpidInfo)
	if err != nil {
		log.Errorc(ctx, "GetAttachedRpid get rpidInfo from taiShan Unmarshal err %v and key %v", err, key)
		return
	}
	if rpidInfo.Status != common.Cvid_Rpid_Attached {
		log.Warnc(ctx, "GetAttachedRpid rpidInfo Status from taishan is deleted rpidInfo %v", rpidInfo)
		return
	}
	resp.Rpid = rpidInfo.Rpid
	return
}
