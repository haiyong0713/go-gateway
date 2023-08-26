package selected

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"

	"go-common/library/sync/errgroup.v2"

	api "git.bilibili.co/bapis/bapis-go/archive/service"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favGRGP "git.bilibili.co/bapis/bapis-go/community/service/favorite"

	"github.com/pkg/errors"
)

const (
	_serieNoOP        = 0
	_serieOperated    = 1
	_seriePassed      = 2
	_serieModified    = 3
	_serieBackoff     = 4 // 兜底数据状态
	_ActionAdd        = "add"
	_ActionUpdate     = "update"
	_ActionDel        = "delete"
	_ActionPub        = "publish"
	_ActionTaskUpdate = "taskUpdate"
	_dateFmt          = "2006-01-02"
)

// ArcTitle returns arc structure
func (s *Service) ArcTitle(c context.Context, aid int64) (arc *api.Arc, err error) {
	if arc, err = s.arcDao.Arc(c, aid); err != nil {
		err = fmt.Errorf("无效ID")
		return
	}
	var noHotAids map[int64]struct{}
	if noHotAids, _, err = s.arcDao.FlowJudge(c, []int64{aid}, s.c.WeeklySelected.FlowCtrl); err != nil {
		return
	}
	if _, ok := noHotAids[aid]; ok {
		err = fmt.Errorf("该视频热门禁止")
		return
	}
	return
}

// OpSerie 拒绝、新增、删除、新增卡片等操作会变更当期的状态
func (s *Service) OpSerie(c context.Context, serie *selected.Serie) (err error) {
	switch serie.Status {
	case _serieNoOP:
		if err = s.dao.UpdateSerieStatus(c, serie.ID, _serieOperated); err != nil {
			log.Error("OpSerie ID %d, Status %d", serie.ID, _serieOperated)
		}
	case _seriePassed:
		if err = s.dao.UpdateSerieStatus(c, serie.ID, _serieModified); err != nil {
			log.Error("OpSerie ID %d, Status %d", serie.ID, _serieOperated)
		}
	case _serieBackoff:
		if err = s.dao.UpdateSerieStatus(c, serie.ID, _serieModified); err != nil {
			log.Error("OpSerie ID %d, Status %d", serie.ID, _serieOperated)
		}
	}
	return
}

// SelResAdd adds a card
func (s *Service) SelResAdd(c context.Context, req *selected.ReqSelAdd, creator *selected.Operator) (err error) {
	var (
		serie  *selected.Serie
		dupCnt int
	)
	if serie, err = s.actSerieByNb(c, &selected.FindSerie{
		Number: req.Number,
		Type:   req.Type,
	}); err != nil {
		return
	}
	var aid int64
	if aid, err = common.GetAvID(req.RID); err != nil {
		return
	}
	if dupCnt, err = s.dao.DuplicateCheck(c, serie.ID, aid, req.Rtype, 0); err != nil {
		return
	}
	if dupCnt > 0 {
		return xecode.AppSelectedIDExist
	}
	resource := &selected.Resource{}
	resource.FromReq(req, serie.ID, creator.Uname)
	if err = s.dao.AddRes(c, resource); err != nil { // add the new resource
		log.Error("SelResAdd AddRes Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	if err = util.AddLogs(common.LogSelectedResource, creator.Uname, creator.UID, serie.Number, _ActionAdd, resource); err != nil { // add operation log
		log.Error("SelResAdd AddRes AddLog error(%v)", err)
		return
	}
	return
}

// actSerieByNb picks the serie with the number and the type
func (s *Service) actSerieByNb(c context.Context, req *selected.FindSerie) (serie *selected.Serie, err error) {
	if serie, err = s.dao.PickSerie(c, req); err != nil {
		log.Error("SelResAdd PickSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	if err = s.OpSerie(c, serie); err != nil { // modify serie's status
		log.Error("SelResAdd OpSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	return
}

// actSerieByID picks the resource and its serie to change the serie's status
func (s *Service) actSerieByID(c context.Context, resID int64) (resource *selected.Resource, serie *selected.Serie, err error) {
	if resource, err = s.dao.PickRes(c, resID); err != nil {
		log.Error("actSerieByID PickRes ID %d, Err %v", resID, err)
		return
	}
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{ID: resource.SerieID}); err != nil {
		log.Error("actSerieByID PickSerie ID %d, Err %v", resID, err)
		return
	}
	if err = s.OpSerie(c, serie); err != nil { // modify serie's status
		log.Error("actSerieByID OpSerie Serie %v, Err %v", serie, err)
		return
	}
	return
}

// SelResEdit def.
func (s *Service) SelResEdit(c context.Context, req *selected.ReqSelEdit, editor *selected.Operator) (err error) {
	var (
		resource *selected.Resource
		serie    *selected.Serie
		dupCnt   int
	)
	if resource, serie, err = s.actSerieByID(c, req.ID); err != nil {
		return
	}
	var aid int64
	if aid, err = common.GetAvID(req.RID); err != nil {
		return
	}
	req.RIDInt = aid
	if dupCnt, err = s.dao.DuplicateCheck(c, resource.SerieID, aid, req.Rtype, resource.ID); err != nil {
		return
	}
	if dupCnt > 0 {
		return xecode.AppSelectedIDExist
	}
	if err = s.dao.UpdateRes(c, resource, req); err != nil {
		log.Error("SelResEdit OpSerie Err %v", err)
		return
	}
	if err = util.AddLogs(common.LogSelectedResource, editor.Uname, editor.UID, serie.Number, _ActionUpdate, req); err != nil { // add operation log
		log.Error("SelResEdit AddLog error(%v)", err)
		return
	}
	return
}

// SelResReject rejects the card
func (s *Service) SelResReject(c context.Context, id int64, editor *selected.Operator) (err error) {
	var (
		resource *selected.Resource
		serie    *selected.Serie
	)
	if resource, serie, err = s.actSerieByID(c, id); err != nil {
		return
	}
	if err = s.dao.RejectRes(c, resource); err != nil {
		log.Error("SelResEdit OpSerie Err %v", err)
		return
	}
	if err = util.AddLogs(common.LogSelectedResource, editor.Uname, editor.UID, serie.Number, common.OptionReject, resource); err != nil { // add operation log
		log.Error("SelResEdit AddLog error(%v)", err)
		return
	}
	return
}

// SelResDel deletes the card
func (s *Service) SelResDel(c context.Context, id int64, editor *selected.Operator) (err error) {
	var (
		resource *selected.Resource
		serie    *selected.Serie
	)
	if resource, serie, err = s.actSerieByID(c, id); err != nil {
		return
	}
	if err = s.dao.DelRes(c, id); err != nil {
		log.Error("SelResEdit OpSerie Err %v", err)
		return
	}
	if err = util.AddLogs(common.LogSelectedResource, editor.Uname, editor.UID, serie.Number, _ActionDel, resource); err != nil { // add operation log
		log.Error("SelResEdit AddLog error(%v)", err)
		return
	}
	return
}

// SelResSort def.
func (s *Service) SelResSort(c context.Context, req *selected.SelSortReq, editor *selected.Operator) (err error) {
	if len(req.CardIDs) == 0 {
		return ecode.RequestErr
	}
	var (
		serie  *selected.Serie
		resCnt int
	)
	if serie, err = s.actSerieByNb(c, &selected.FindSerie{
		Number: req.Number,
		Type:   req.Type,
	}); err != nil {
		return
	}
	if resCnt, err = s.dao.CntRes(c, serie.ID); err != nil {
		log.Error("SelResSort SID %d, Err %v", serie.ID, err)
		return
	}
	if len(req.CardIDs) != resCnt {
		log.Error("SelResSort Resources Len %d, Req Len %d", resCnt, len(req.CardIDs))
		return ecode.RequestErr
	}
	if err = s.dao.SortRes(c, serie.ID, req.CardIDs); err != nil {
		log.Error("SelResSort Serie_ID %d, CardIDs %v, Err %v", serie.ID, req.CardIDs, err)
		return
	}
	if err = util.AddLogs(common.LogSelectedResource, editor.Uname, editor.UID, serie.Number, _ActionUpdate, req.CardIDs); err != nil { // add operation log
		log.Error("SelResEdit AddLog error(%v)", err)
		return
	}
	return
}

// SelAudit def.
func (s *Service) SelAudit(c context.Context, req *selected.PreviewReq, editor *selected.Operator) (err error) {
	var (
		serie *selected.Serie
		valid bool
	)
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		Type:   req.Type,
		Number: req.Number,
	}); err != nil {
		log.Error("SelAudit PickSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	// 判断前三位稿件是否有寄语
	if valid, err = s.dao.SerieValid(c, serie.ID); err != nil {
		log.Error("SelAudit Valid Sid %d, Err %v", serie.ID, err)
		return
	}
	if !valid {
		err = xecode.AppSelectedNotValid
		return
	}
	// 获取配置的稿件资源
	resources, err := s.dao.PickResBySerieID(c, serie.ID)
	if err != nil {
		log.Error("SelAudit PickResBySerieID  serieId(%d), Err(%+v)", serie.ID, err)
		return ecode.Errorf(ecode.ServerErr, "每周必看审核：获取周期精选资源服务出错(%+v)", err)
	}
	var rids []int64
	for _, res := range resources {
		// 过滤非稿件非启用数据
		if !res.IsArc() || res.Status != selected.ResourceStatusOn {
			continue
		}
		rids = append(rids, res.RID)
	}
	// 没有稿件，退出发布流程
	if len(rids) == 0 {
		err = ecode.Errorf(ecode.ServerErr, "每周必看审核：符合要求的稿件为0 Number(%d)", serie.Number)
		return err
	}
	// 生成播单
	err = s.genMediaList(c, serie, rids)
	if err != nil {
		log.Error("SelAudit genMediaList  serieId(%d), Err(%+v)", serie.ID, err)
		return ecode.Errorf(ecode.ServerErr, "每周必看审核：生成播单出错(%+v)", err)
	}
	// 同步到 TV
	s.sendOOT(c, rids, serie)
	// 通过审核
	if err = s.dao.SeriePass(c, serie.ID); err != nil {
		log.Error("SelAudit SeriePass Sid %d, Err %v", serie.ID, err)
		return ecode.Errorf(ecode.ServerErr, "每周必看审核：通过审核出错(%+v)", err)
	}
	// 当前时间在发布时间之后
	if time.Now().After(serie.Pubtime.Time()) {
		// 同步到荣誉稿件, 需要在更新缓存之前同步，原因：修改的稿件有数据库与缓存数据进行对比
		var modifyRes []*selected.Resource
		modifyRes, err = s.getModifyRes(c, serie, resources)
		if err != nil {
			log.Error("SelAudit getModifyRes serieId(%d), Err(%+v)", serie.ID, err)
			return ecode.Errorf(ecode.ServerErr, "每周必看审核：获取播单缓存数据出错(%+v)", err)
		}
		err = s.sendArchiveHonor(c, serie, modifyRes)
		if err != nil {
			log.Error("SelAudit sendArchiveHonor  serieId(%d), Err(%+v)", serie.ID, err)
			return ecode.Errorf(ecode.ServerErr, "每周必看审核：同步荣誉稿件出错(%+v)", err)
		}
		// 同步每周必看数据到缓存
		err = s.refCache(c, serie.Type, serie.Number)
		if err != nil {
			log.Error("SelAudit refCache  serieId(%d), Err(%+v)", serie.ID, err)
			return ecode.Errorf(ecode.ServerErr, "每周必看审核：同步每周必看数据到缓存出错(%+v)", err)
		}
	}
	if err = util.AddLogs(common.LogSelectedSerie, editor.Uname, editor.UID, serie.Number, _ActionPub, serie.ID); err != nil { // add operation log
		log.Error("SelAudit AddLog error(%v)", err)
		return ecode.Errorf(ecode.ServerErr, "每周必看审核：添加操作日志出错(%+v)", err)
	}
	// 审核通过时间 大于 发布时间, 并小于发布那周的周日0点，触发每周必看天马推送
	if req.Type == selected.SERIE_TYPE_WEEKLY_SELECTED &&
		time.Now().After(serie.Pubtime.Time()) && time.Now().Before(serie.Pubtime.Time().Add(30*time.Hour)) {
		log.Warn("Service SelAudit WeeklySelectedTunnel")
		if err = s.dao.WeeklySelectedTunnel(c, req.Number); err != nil {
			log.Error("SelAudit WeeklySelectedTunnel err(%+v)", err)
			return ecode.Errorf(ecode.ServerErr, "每周必看审核：天马小卡推送错误(%+v)", err)
		}
	}
	return
}

// 发送到 TV OOT，
func (s *Service) sendOOT(c context.Context, rids []int64, serie *selected.Serie) {
	// 并发上报 TV
	eg := errgroup.WithContext(context.Background())
	for i := 0; i < len(rids); i++ {
		msg := &selected.OTTSeriesMsg{Aid: rids[i], Number: serie.Number}
		eg.Go(func(ctx context.Context) error {
			_ = s.dao.PubOTT(c, msg)
			return nil
		})
	}
	_ = eg.Wait()
}

// 发布每周必看
func (s *Service) PublishWeekly(c context.Context) (err error) {
	serie, err := s.dao.PickPublishSerie(c, selected.SERIE_TYPE_WEEKLY_SELECTED)
	if err != nil {
		return errors.WithMessage(err, "PublishWeekly PickPublishSerie")
	}
	log.Warn("PublishWeekly serie(%+v)", serie)
	if serie == nil {
		return fmt.Errorf("publish serie not found")
	}
	var resources []*selected.Resource
	switch serie.Status {
	// 通过审核, 第一次发布获取通过审核的稿件
	case _seriePassed:
		resources, err = s.dao.PickValidResBySerieID(c, serie.ID)
		if err != nil {
			return errors.WithMessagef(err, "PublishWeekly dao RntRes sid(%d)", serie.ID)
		}
	// 无操作，但是本期未操作且AI卡片足够，则出兜底数据
	case _serieNoOP:
		resources, err = s.dao.PickResBySidSource(c, serie.ID, selected.ResourceSourceAI)
		if err != nil {
			return errors.WithMessagef(err, "PublishWeekly dao RntRes sid(%d)", serie.ID)
		}
		if len(resources) < s.c.WeeklySelected.RecoveryNb {
			return fmt.Errorf("PublishWeekly sid(%d) Number(%d) recouce count(%d) < %d", serie.ID, serie.Number, len(resources), s.c.WeeklySelected.RecoveryNb)
		}
		if len(resources) > s.c.WeeklySelected.MaxNumber {
			resources = resources[:s.c.WeeklySelected.MaxNumber]
		}
		err = s.dao.UpdateSerieStatus(c, serie.ID, _serieBackoff)
		if err != nil {
			return errors.WithMessagef(err, "PublishWeekly dao UpdateSerieStatus sid(%d)", serie.ID)
		}
		log.Error("【日志报警】每周必看使用兜底数据")
	// 已经使用兜底数据发布了
	case _serieBackoff:
		log.Warn("【日志报警】每周必看使用兜底数据")
		return
	default:
		log.Error("【日志报警】每周必看未生成")
		return
	}
	log.Warn("PublishWeekly resources length: %d", len(resources))
	// 自动发布流程
	err = s.publishSerieProcess(c, serie, resources)
	if err != nil {
		return err
	}
	// 提高 每周必看在热门分类入口设置中的位置
	err = s.updateEntranceRank(c)
	if err != nil {
		return errors.WithMessage(err, "PublishWeekly updateEntranceRank")
	}
	log.Warn("PublishWeekly 每周必看发布成功")
	return
}

// 定时自动发布流程
func (s *Service) publishSerieProcess(c context.Context, serie *selected.Serie, resources []*selected.Resource) (err error) {
	var rids []int64
	var cids []int64
	for _, res := range resources {
		// 过滤非稿件非启用数据
		if !res.IsArc() || res.Status != selected.ResourceStatusOn {
			continue
		}
		rids = append(rids, res.RID)
	}
	// 没有稿件，退出发布流程
	if len(rids) == 0 {
		err = fmt.Errorf("符合要求的稿件为0 Number(%d)", serie.Number)
		return err
	}
	//获取稿件cid，推到弹幕后台
	cids, err = s.arcDao.AidsToCids(c, rids)
	if err != nil {
		//推送稿件cid失败只做提示，不影响别的流程
		log.Error("publishSerieProcess AidsToCids err(%v)", err)
		err = nil
	} else {
		s.damuDao.AddWeekViewDanmuV2(c, cids)
	}
	// 生成播单
	err = s.genMediaList(c, serie, rids)
	if err != nil {
		return errors.WithMessage(err, "PublishWeekly genMediaList")
	}
	// 同步到荣誉稿件(必须在同步每周必看缓存之前，避免同步缓存成功后未同步荣誉稿件，而导致手动发布时通过缓存判断同步荣誉稿件的数据错误)
	err = s.sendArchiveHonor(c, serie, resources)
	if err != nil {
		return errors.WithMessage(err, "PublishWeekly sendArchiveHonor")
	}
	// 同步每周必看数据到缓存
	err = s.refCache(c, serie.Type, serie.Number)
	if err != nil {
		return errors.WithMessage(err, "PublishWeekly refCache")
	}
	// 同步到 TV
	s.sendOOT(c, rids, serie)
	return err
}

// 获取修改的资源
func (s *Service) getModifyRes(c context.Context, serie *selected.Serie, resources []*selected.Resource) (modifyRes []*selected.Resource, err error) {
	var (
		cacheResMap = map[int64]interface{}{} // 缓存中的稿件数据 map
		serieCache  *selected.SerieFull       // 缓存中的数据
	)
	serieCache, err = s.dao.PickSerieCache(c, serie.Type, serie.Number)
	if err != nil {
		return nil, errors.WithMessagef(err, "Service getModifyRes PickSerieCache")
	}
	if serieCache == nil {
		log.Warn("Service getModifyRes PickSerieCache is empty")
		return
	}
	for _, res := range serieCache.List {
		cacheResMap[res.RID] = nil
	}
	for _, resource := range resources {
		_, ok := cacheResMap[resource.RID]
		// 稿件从启用转为禁用
		if ok && resource.Status == selected.ResourceStatusReject {
			modifyRes = append(modifyRes, resource)
		} else if resource.Status == selected.ResourceStatusOn {
			// 通过状态稿件
			modifyRes = append(modifyRes, resource)
		}
	}
	return
}
func (s *Service) sendArchiveHonor(c context.Context, serie *selected.Serie, resources []*selected.Resource) (err error) {
	var archiveHonors []*selected.ArchiveHonor
	log.Warn("sendArchiveHonor resources length1: %d", len(resources))
	for _, res := range resources {
		if !res.IsArc() {
			continue
		}
		archiveHonor := &selected.ArchiveHonor{
			Aid:   res.RID,
			Type:  selected.ArchiveHonorTypeWeekly,
			Url:   s.c.WeeklySelected.HonorLink + fmt.Sprintf("?num=%d&navhide=1", serie.Number),
			NaUrl: s.c.WeeklySelected.HonorLinkV2 + fmt.Sprintf("?current_tab=week-%d", serie.Number),
			Desc:  fmt.Sprintf("第%d期每周必看", serie.Number),
		}
		if res.Status == 1 {
			archiveHonor.Action = selected.ArchiveHonorActionUpdate // 插入、更新
		} else {
			archiveHonor.Action = selected.ArchiveHonorActionDelete // 删除
		}
		archiveHonors = append(archiveHonors, archiveHonor)
	}
	log.Warn("sendArchiveHonor resources length2: %d", len(archiveHonors))
	// 并发上报荣誉稿件
	eg := errgroup.WithContext(c)
	for i := 0; i < len(archiveHonors); i++ {
		honor := archiveHonors[i]
		eg.Go(func(ctx context.Context) error {
			// 重试3次，避免因为网络抖动等导致上报失败
			err2 := retry.WithAttempts(context.Background(), "dao get weekly selected", 3, netutil.DefaultBackoffConfig, func(c context.Context) (err error) {
				err = s.dao.PubArchiveHonor(context.Background(), honor)
				return err
			})
			return err2
		})
	}
	if err = eg.Wait(); err != nil {
		return errors.WithMessage(err, "errgroup wait ")
	}
	return
}

func (s *Service) genMediaList(c context.Context, serie *selected.Serie, aids []int64) (err error) {
	// 验证稿件 id
	arcs, err := s.arcDao.Arcs(c, aids)
	if err != nil {
		return ecode.Errorf(ecode.ServerErr, "获取稿件信息服务出错(%+v)", err)
	}
	// 判断 aid 是否存在
	var aidNotExist []int64
	for _, aid := range aids {
		if _, ok := arcs[aid]; !ok {
			aidNotExist = append(aidNotExist, aid)
		}
	}
	if len(aidNotExist) > 0 {
		return ecode.Errorf(ecode.RequestErr, "稿件 id 不存在（%v）", aidNotExist)
	}
	// 播单不存在, 创建收藏夹作为播单
	if serie.MediaID == 0 {
		log.Warn("serie info: subject:%s, description:%s", serie.Subject, serie.ShareSubtitle)
		// 创建收藏夹作为播单
		fid, err2 := s.favDao.AddFolder(c, &favGRGP.AddFolderReq{
			Name:        serie.MediaListTitle(),
			Description: serie.ShareSubtitle,
			Typ:         favmdl.TypeVideo,
			Mid:         s.c.WeeklySelected.PlaylistMid,
			Public:      favmdl.AttrDefaultPublic,
			Cover:       arcs[aids[0]].Pic, // 第一个稿件的图片
		})
		if err2 != nil {
			return ecode.Errorf(ecode.ServerErr, "创建收藏夹服务错误(%+v)", err2)
		}
		serie.MediaID = fid
		// 跟新Serie播单id
		err2 = s.dao.UpdateSerieMediaId(c, serie.ID, serie.MediaID)
		if err2 != nil {
			return ecode.Errorf(ecode.ServerErr, "更新 mediaId 服务错误(%+v)", err2)
		}
	}
	req := &favGRGP.MultiReplaceReq{
		Typ:  favmdl.TypeVideo,
		Mid:  s.c.WeeklySelected.PlaylistMid,
		Oids: aids,
		Fid:  serie.MediaID,
	}
	// 替换播单数据
	err = s.favDao.MultiReplace(c, req)
	if err != nil {
		return ecode.Errorf(ecode.ServerErr, "往播单里添加数据服务错误(%+v) req(%+v)", err, req)
	}
	return err
}

// SerieEdit def.
func (s *Service) SerieEdit(c context.Context, req *selected.SerieEditReq, editor *selected.Operator) (err error) {
	var serie *selected.Serie
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		Type:   req.Type,
		Number: req.Number,
	}); err != nil { // modify serie's status
		log.Error("SelAudit PickSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	if err = s.OpSerie(c, serie); err != nil { // modify serie's status
		log.Error("SelResAdd OpSerie Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	req.SerieDB.ID = serie.ID // pick the PK
	if err = s.dao.SerieUpdate(c, &req.SerieDB); err != nil {
		log.Error("SelAudit SerieUpdate Type %s, Number %d, Err %v", req.Type, req.Number, err)
		return
	}
	if err = util.AddLogs(common.LogSelectedSerie, editor.Uname, editor.UID, serie.Number, _ActionUpdate, serie.ID); err != nil { // add operation log
		log.Error("SelAudit AddLog error(%v)", err)
		return
	}
	return
}

// SerieEdit def.
func (s *Service) UpdateTaskStatus(c context.Context, id int64, taskStatus int, editor *selected.Operator) (err error) {
	var serie *selected.Serie
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		ID: id,
	}); err != nil { // modify serie's status
		log.Error("UpdateTaskStatus PickSerie ID %d, Err %v", id, err)
		return
	}
	//nolint:gomnd
	if serie.TaskStatus == 11 {
		return nil
	}
	serie.TaskStatus = taskStatus
	if err = s.dao.SerieUpdateTaskStatus(c, serie); err != nil {
		log.Error("UpdateTaskStatus fail ID %d, Err %v", id, err)
		return
	}

	logDetail := make(map[string]int)
	logDetail["id"] = int(id)
	logDetail["taskStatus"] = taskStatus

	if err = util.AddLogs(common.LogSelectedSerie, editor.Uname, editor.UID, serie.Number, _ActionTaskUpdate, logDetail); err != nil { // add operation log
		log.Error("UpdateTaskStatus AddLog error(%v)", err)
		return
	}
	return
}

func (s *Service) UpdatePushTaskStatus(c context.Context, id int64, pushTaskStatus int, editor *selected.Operator) (err error) {
	var serie *selected.Serie
	if serie, err = s.dao.PickSerie(c, &selected.FindSerie{
		ID: id,
	}); err != nil { // modify serie's status
		log.Error("UpdateTaskStatus PickSerie ID %d, Err %v", id, err)
		return
	}
	//nolint:gomnd
	serie.TaskStatus = serie.TaskStatus/10*10 + pushTaskStatus

	if err = s.dao.SerieUpdateTaskStatus(c, serie); err != nil {
		log.Error("UpdateTaskStatus fail ID %d, Err %v", id, err)
		return
	}

	logDetail := make(map[string]int)
	logDetail["id"] = int(id)
	logDetail["taskStatus"] = serie.TaskStatus

	if err = util.AddLogs(common.LogSelectedSerie, editor.Uname, editor.UID, serie.Number, _ActionTaskUpdate, logDetail); err != nil { // add operation log
		log.Error("UpdateTaskStatus AddLog error(%v)", err)
		return
	}
	return
}

// SelValidBeforeTouchUsers def.
// 在推送用户之前进行验证
func (s *Service) SelValidBeforeTouchUsers(c context.Context, req *selected.FindSerie) (serie *selected.Serie, err error) {
	if serie, err = s.dao.PickSerie(c, req); err != nil {
		log.Error("[SerieTouchUser] PickSerieID Sid %d, Err %v", req.ID, err)
		return nil, ecode.Error(-400, "获取当期每周必看失败，请确认传入参数！")
	} else {
		//nolint:gomnd
		if serie.Status != 2 {
			return nil, ecode.Error(-400, "当期每周必看尚未审核通过，请先通过审核！")
		}
		if serie.MediaID == 0 {
			return nil, ecode.Error(-400, "当期每周必看尚未生成播单，请联系网关同学！")
		}
		//nolint:gomnd
		if serie.TaskStatus == 11 {
			return nil, ecode.Error(-400, "当期每周必看已通知用户，请勿重复触发！")
		}
		//nolint:gomnd
		if serie.TaskStatus == 13 {
			return nil, ecode.Error(-400, "当期每周必看正在执行Push任务，请勿重复触发！")
		}
	}
	return serie, nil
}

// SelEditPushInfo def.
// 修改push信息
func (s *Service) SelEditPushInfo(c context.Context, req *selected.SeriePush) (err error) {
	if _, err := s.dao.PickSerie(c, &selected.FindSerie{ID: req.ID}); err != nil {
		log.Error("[SelEditPushInfo] PickSerieID Sid %d, Err %v", req.ID, err)
		return ecode.Error(-400, "获取当期每周必看失败，请确认传入参数！")
	}
	if err = s.dao.SerieUpdatePush(c, req); err != nil {
		log.Error("[SelEditPushInfo] SerieUpdatePush Sid %d, Err %v", req.ID, err)
		return ecode.Error(-500, "修改push信息失败！")
	}
	return nil
}

const (
	_pushBatch  = 1000
	_tagUsersPs = 1000
)

// SelPushUsers def.
// pushSerie picks the elements of serie to push to the subscribers
func (s *Service) PushSerie(ctx context.Context, serie *selected.Serie) (err error) {
	var (
		mids       []int64
		pageCnt    float64
		successCnt float64
		failCount  float64
	)
	if mids, err = s.subscribers(ctx); err != nil {
		log.Error("[PushSerie] Sid %d, Get Subscribers Err %v", serie.ID, err)
		return
	}
	lenMids := len(mids)
	if lenMids > _pushBatch { // record to push mids
		log.Info("[PushSerie] PushMids %v", mids[0:_pushBatch])
	} else {
		log.Info("[PushSerie] PushMids %v", mids)
	}
	if serie.PushSubtitle == "" {
		if serie.PushSubtitle, err = s.getPushBody(ctx, serie); err != nil {
			log.Error("[PushSerie] Sid %d, GetPushBody Err %v", serie.ID, err)
			return
		}
	}

	for { // push by piece of 100K
		var toPushMids []int64
		if len(mids) == 0 { // when the length of mid % 100K == 0
			break
		} else if len(mids) > _pushBatch { // cut the slice
			toPushMids = mids[0:_pushBatch]
			mids = mids[_pushBatch:]
		} else { // the last slice
			toPushMids = mids
			mids = []int64{}
		}
		pageCnt += 1
		if err = s.pushDao.NoticeUser(toPushMids, serie.UUID(toPushMids[0]), serie); err != nil {
			log.Error("[PushSerie] Sid %d, PushPage %f Err %v", serie.ID, pageCnt, err)
			failCount += 1
			continue
		}
		successCnt += 1
	}
	if pageCnt == 0 {
		return ecode.Error(-500, "push量为0")
	} else {
		//nolint:gomnd
		percent := successCnt / pageCnt * 100
		//nolint:gomnd
		if percent < 95 {
			return ecode.Error(-500, fmt.Sprintf("push成功率低于95%%, 当前为%.2f", percent))
		}
	}
	return
}

func (s *Service) subscribers(c context.Context) (mids []int64, err error) {
	var (
		offset      int
		midCnt      int
		subscribers []int64
	)

	subscribedTag := s.c.WeeklySelected.SubscribedTag

	calDate := time.Now()
	// 12点后拿当天，12点前拿昨日
	//if calDate.Hour() < 12 {
	duration, _ := time.ParseDuration("-24h")
	calDate = calDate.Add(duration)
	//}

	if midCnt, err = s.dao.GetSubscriberCnt(c, subscribedTag, calDate.Format("2006-01-02")); err != nil {
		log.Error("s.dao.GetSubscriberCnt error(%v)", err)
		return mids, err
	}

	if midCnt == 0 {
		return mids, nil
	}

	for {
		retryErr := retry.WithAttempts(c, "tagSub", s.c.WeeklySelected.AttemptCount, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			var detailErr error
			if subscribers, detailErr = s.dao.GetSubscriberDetail(c, subscribedTag, offset, _tagUsersPs, calDate.Format("2006-01-02")); detailErr != nil {
				log.Error("s.dao.GetSubscriberDetail error(%v)", detailErr)
				return detailErr
			}
			return nil
		})
		if retryErr != nil {
			return nil, retryErr
		}
		offset += _tagUsersPs
		mids = append(mids, subscribers...)
		if offset > midCnt { // the cursor reaches the end
			break
		}
	}
	return mids, nil
}

// getPushBody picks the first archive's title and the label of the serie to combine to body of push
func (s *Service) getPushBody(ctx context.Context, serie *selected.Serie) (title string, err error) {
	var (
		resources []*selected.Resource
		aids      []int64
		arcs      map[int64]*api.Arc
	)
	if resources, err = s.dao.PickValidResBySerieID(ctx, serie.ID); err != nil {
		log.Error("[PushSerie] Get Resources Sid %d Err %v", serie.ID, err)
		return
	}
	for _, v := range resources {
		if v.IsArc() {
			aids = append(aids, v.RID)
		}
	}
	if len(aids) == 0 {
		err = ecode.NothingFound
		log.Error("[PushSerie] Sid %d, Arc Empty!!", serie.ID)
		return
	}
	if arcs, err = s.arcDao.Arcs(ctx, aids); err != nil {
		log.Error("[PushSerie] SerieID %d, Arc %d Error %v!", serie.ID, resources[0].RID, err)
		return
	}
	for _, v := range aids { // pick the first arc's title to build the body
		if arc, ok := arcs[v]; ok {
			title = serie.PushBody() + arc.Title
			return
		}
	}
	err = ecode.NothingFound
	return
}

func (s *Service) newSerie(typ string) (err error) {
	ctx := context.Background()
	lastSerie, err := s.dao.GetLastValidSerieByType(ctx, typ)
	if err != nil {
		log.Error("newSerie type(%s), error(%+v)", typ, err)
		return
	}
	stime, _ := time.ParseInLocation(_dateFmt, time.Now().Format(_dateFmt), time.Local) // 周五0点
	// 判断新的 Serie Stime 不大于当前最大 number 的 Stime，说明已经创建新的Serie
	if !stime.After(lastSerie.Stime.Time()) {
		log.Info("newSerie 新一期的 Serie 已经创建")
		return
	}
	etime := stime.AddDate(0, 0, 7).Add(-1 * time.Second)                               // 下周四的23:59:59
	pubtime := stime.AddDate(0, 0, 7).Add(time.Duration(s.c.WeeklySelected.UpdateTime)) // 下周五的晚18点
	serie := &selected.Serie{
		Type:    typ,
		Number:  lastSerie.Number + 1,
		Stime:   xtime.Time(stime.Unix()),
		Etime:   xtime.Time(etime.Unix()),
		Pubtime: xtime.Time(pubtime.Unix()),
	}
	if err = s.dao.CreateSerie(ctx, serie); err != nil {
		log.Error("CreateSerie error(%+v)", err)
		return
	}
	return
}
