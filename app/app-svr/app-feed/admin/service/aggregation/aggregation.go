package aggregation

import (
	"context"
	"sort"
	"strings"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/aggregation"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"

	tag "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"go-common/library/sync/errgroup.v2"
)

const (

	// 	_maxAids   = 100
	_maxCnt    = 100
	_manualTag = "人工"
)

func judgeParam(param aggregation.AggPub, tagID []int64) (err error) {
	if len(tagID) == 0 {
		return ecode.Error(ecode.RequestErr, "关联TAG不能为空")
	}
	if param.HotTitle == "" || len([]rune(param.HotTitle)) > 10 {
		return ecode.Error(ecode.RequestErr, "热词名称不能为空/热词名称不能超过10个字")
	}
	if param.Title == "" || len([]rune(param.Title)) > 25 {
		return ecode.Error(ecode.RequestErr, "入口文案不能为空/入口文案不能超过25个字")
	}
	if param.SubTitle == "" || len([]rune(param.SubTitle)) > 25 {
		return ecode.Error(ecode.RequestErr, "页面副标题不能为空/页面副标题不能超过25个字")
	}
	return
}

func (s *Service) isOnlyHotWord(ctx context.Context, hotTitle string) (err error) {
	var cnt int
	if cnt, err = s.dao.HotWordCount(ctx, hotTitle); err != nil {
		log.Error("[judgeParam] s.dao.HotWordCount() error(%v)", err)
		return
	}
	if cnt != 0 {
		err = ecode.Error(ecode.RequestErr, "该热词已存在")
	}
	return
}

// AddAggregation .
func (s *Service) AddAggregation(ctx context.Context, param aggregation.AggPub, tagID []int64, name string, uid int64) (err error) {
	var hotID int64
	if err = judgeParam(param, tagID); err != nil {
		return
	}
	if err = s.isOnlyHotWord(ctx, param.HotTitle); err != nil {
		return
	}
	if hotID, err = s.dao.AddAggregation(ctx, param, tagID); err != nil {
		log.Error("[AddAggregation] s.dao.AddAggregation() error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"hot_title": param.HotTitle,
	}
	if err = util.AddLogs(common.LogAggregation, name, uid, hotID, aggregation.AggAdd, obj); err != nil {
		log.Error("[AddAggregation] AddLog error(%v)", err)
	}
	return
}

// UpdateAggregation .
func (s *Service) UpdateAggregation(ctx context.Context, param aggregation.AggPub, tagID []int64, name string, uid int64) (err error) {
	var res *aggregation.AggPub
	if err = judgeParam(param, tagID); err != nil {
		return
	}
	if res, err = s.dao.FindNameByID(ctx, param.ID); err != nil {
		log.Error("[UpdateAggregation] s.dao.FindNameByID() error(%v)", err)
		return
	}
	if res.HotTitle != param.HotTitle {
		if err = s.isOnlyHotWord(ctx, param.HotTitle); err != nil {
			return
		}
	}
	if err = s.dao.UpdateAggregation(ctx, param, tagID); err != nil {
		log.Error("[UpdateAggregation] s.dao.UpdateAggregation() id(%d) error(%v)", param.ID, err)
		return
	}
	obj := map[string]interface{}{
		"hot_title": param.HotTitle,
	}
	if err = util.AddLogs(common.LogAggregation, name, uid, param.ID, aggregation.AggEdit, obj); err != nil {
		log.Error("[UpdateAggregation] AddLog error(%v)", err)
	}
	return
}

// AggOperate .
func (s *Service) AggOperate(ctx context.Context, id, uid int64, state int, name, hotTitle string) (err error) {
	var action string
	if err = s.dao.AggOperate(ctx, id, state); err != nil {
		log.Error("[AggOperate] s.dao.AggOperate() id(%d) state(%d) error(%v)", id, state, err)
		return
	}
	switch state {
	case aggregation.AggreAuditNum: // 重新审核
		action = aggregation.AggreAudit
	case aggregation.AggPassNum: // 通过
		action = aggregation.AggPass
	case aggregation.AggRefuseNum: // 拒绝
		action = aggregation.AggRefuse
	case aggregation.AggOfflineNum: // 下线
		action = aggregation.AggOffline
	case aggregation.AggDelNum: // 删除
		action = aggregation.AggDel
	}
	obj := map[string]interface{}{
		"hot_title": hotTitle,
	}
	if err = util.AddLogs(common.LogAggregation, name, uid, id, action, obj); err != nil {
		log.Error("AggOperate AddLog error(%v)", err)
	}

	return
}

// AggregationList .
func (s *Service) AggregationList(ctx context.Context, param *aggregation.AggListReq) (res *aggregation.AggListReply, err error) {
	var (
		hotIDs      []int64
		tagIDs      []int64
		aggHotIDRes []*aggregation.AggTag
		tagReply    *tag.TagReply
	)
	if param.TagName != "" {
		if tagReply, err = s.dao.TagIDByName(ctx, param.TagName); err != nil {
			log.Error("[AggregationList] s.dao.NameByTagID() error(%v)", err)
			return
		}
		if tagReply != nil && tagReply.Tag != nil {
			tagIDs = append(tagIDs, tagReply.Tag.Id)
		}
	}
	tagIDs = append(tagIDs, param.TagID)
	if len(tagIDs) != 0 {
		if aggHotIDRes, err = s.dao.FindByTagIDs(ctx, tagIDs); err != nil {
			log.Error("[AggregationList] s.dao.FindByTagIDs() tag_id(%d) error(%v)", param.TagID, err)
			return
		}
	}
	for _, v := range aggHotIDRes {
		hotIDs = append(hotIDs, v.HotwordID)
	}
	if res, err = s.dao.AggList(ctx, param, hotIDs); err != nil {
		log.Error("[AggregationList]s.dao.AggList() error(%v)", err)
		return
	}
	for _, v := range res.Items {
		var (
			aggRes  []*aggregation.AggTag
			tagsRes *tag.TagsReply
			tags    []int64
			tagSc   []*aggregation.Tag
		)
		// TODO 改成批量查询
		if aggRes, err = s.dao.TagIDByID(ctx, v.ID); err != nil && len(aggRes) == 0 {
			log.Error("[AggregationList] s.dao.TagIDByID hotID(%d) error(%v)", v.ID, err)
			err = nil
			continue
		}
		for _, tid := range aggRes {
			tags = append(tags, tid.TagID)
		}
		if len(tags) != 0 {
			if tagsRes, err = s.dao.NameByTagID(ctx, tags); err != nil {
				log.Error("[AggregationList] s.dao.NameByTagID() tag_id(%s) error(%v)", xstr.JoinInts(tags), err)
				err = nil
				continue
			}
		}
		if tagsRes != nil && len(tagsRes.Tags) != 0 {
			var tids []int
			// 这里为了不让tag顺序固定，不来回跳
			for tid := range tagsRes.Tags {
				tids = append(tids, int(tid))
			}
			sort.Ints(tids)
			for _, t := range tids {
				if tagR, ok := tagsRes.Tags[int64(t)]; ok {
					a := &aggregation.Tag{}
					a.ID = int64(t)
					a.Name = tagR.Name
					tagSc = append(tagSc, a)
				}
			}
		}
		v.Tag = tagSc
	}
	return
}

// Tag .
func (s *Service) Tag(ctx context.Context, tagName string) (tagReply *tag.TagReply, err error) {
	if tagReply, err = s.dao.TagIDByName(ctx, tagName); err != nil {
		log.Error("[Tag] s.dao.TagIDByName() tagName(%s) error(%v)", tagName, err)
	}
	return
}

func (s *Service) AggTagAdd(ctx context.Context, id int64, tagID []int64) error {
	return s.dao.AggTagAddM(ctx, id, tagID)
}

func (s *Service) AggTagDel(ctx context.Context, id, tagID int64) error {
	return s.dao.AggTagDelete(ctx, id, tagID)
}

// AggViewAdd .
func (s *Service) AggViewAdd(ctx context.Context, id int64, rids []int64) error {
	return s.dao.HotwordAggResourceAddM(ctx, id, rids)
}

func (s *Service) AggViewOp(ctx context.Context, id, rid, tagID int64, state int) error {
	return s.dao.HotwordAggResourceState(ctx, id, rid, tagID, state)
}

// AggView .
func (s *Service) AggView(ctx context.Context, id int64) (views *aggregation.ViewsReply, err error) {
	var (
		items      []*aggregation.Views
		dbRes      []*aggregation.HotwordAggResource
		cardAIRes  []*aggregation.CardList
		tagRes     []*aggregation.AggTag
		returnTags []*aggregation.Tag
	)
	eg := errgroup.WithContext(ctx) // 获取选项
	eg.Go(func(c context.Context) (err error) {
		cardAIRes, err = s.dao.AggView(ctx, id)
		if err != nil {
			log.Error("AggView id(%d) cardAIRes err(%+v)", id, err)
		}
		return nil
	})
	eg.Go(func(c context.Context) (err error) {
		dbRes, err = s.dao.HwResourceByHwID(ctx, id)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		tagRes, err = s.dao.TagIDByID(ctx, id)
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	returnTags, items, err = s.handlerAggView(ctx, cardAIRes, dbRes, tagRes)
	views = &aggregation.ViewsReply{
		Items: items,
		Tags:  returnTags,
		Total: len(items),
	}
	return
}

//nolint:gocognit
func (s *Service) handlerAggView(ctx context.Context, aiResource []*aggregation.CardList, dbResource []*aggregation.HotwordAggResource,
	tagRes []*aggregation.AggTag) (returnTags []*aggregation.Tag, videos []*aggregation.Views, err error) {
	var (
		resM   []*aggregation.Views // 手动新增部分
		resD   []*aggregation.Views // 手动删除部分
		resAI  []*aggregation.Views // AI返回部分
		aids   []int64
		tagIDs []int64
		//nolint:ineffassign
		arcs  = make(map[int64]*api.Arc)
		tags  map[int64]*tag.Tag
		exist = make(map[int64]struct{}) // 用于判断AI部分是不是存在于人工部分
	)
	for _, item := range dbResource {
		aids = append(aids, item.Oid)
		if item.TagID > 0 {
			tagIDs = append(tagIDs, item.TagID)
		}
		if item.State == aggregation.DefaultState {
			resM = append(resM, &aggregation.Views{
				AvID:  item.Oid,
				TagID: item.TagID,
				State: item.State,
			})
		}
		if item.State != aggregation.DefaultState {
			resD = append(resD, &aggregation.Views{
				AvID:  item.Oid,
				State: item.State,
				TagID: item.TagID,
			})
		}
		exist[item.Oid] = struct{}{}
	}
	for _, item := range aiResource {
		item.TagIDs, _ = xstr.SplitInts(item.Tag)
		if _, ok := exist[item.ID]; ok {
			for i := 0; i < len(resM); i++ {
				if resM[i].AvID == item.ID {
					resM[i].TagIDs = item.TagIDs
				}
			}
			for i := 0; i < len(resD); i++ {
				if resD[i].AvID == item.ID {
					resD[i].TagIDs = item.TagIDs
				}
			}
			continue
		}
		aids = append(aids, item.ID)
		var tagIdTemp int64
		if len(item.TagIDs) > 0 {
			tagIDs = append(tagIDs, item.TagIDs...)
			tagIdTemp = item.TagIDs[0]
		}
		resAI = append(resAI, &aggregation.Views{
			AvID:   item.ID,
			State:  aggregation.DefaultState,
			TagID:  tagIdTemp,
			TagIDs: item.TagIDs,
		})
	}
	for _, item := range tagRes {
		if item == nil {
			continue
		}
		if item.TagID > 0 {
			tagIDs = append(tagIDs, item.TagID)
		}
	}
	if arcs, err = s.arcDao.ArcsWithPage(ctx, aids); err != nil {
		log.Error("[AggView] s.dao.ArcsWithPage aids %v, err %v", aids, err)
		return
	}
	if len(tagIDs) > 0 {
		filterTagIDs := aggregation.FilterDupIDs(tagIDs)
		if tags, err = s.dao.NamesByTagIDs(ctx, filterTagIDs); err != nil {
			log.Error("[AggView] s.dao.NamesByTagIDs tids %v, err %v", tagIDs, err)
			return
		}
	}
	videos = append(videos, resM...)
	videos = append(videos, resAI...)
	videos = append(videos, resD...)
	if len(videos) > _maxCnt {
		videos = videos[:_maxCnt]
	}
	for i := 0; i < len(videos); i++ {
		if videos[i].TagID == 0 || videos[i].State == aggregation.BlockState {
			videos[i].TagName = _manualTag + "、"
		}
		for _, itemTemp := range videos[i].TagIDs {
			if item, ok := tags[itemTemp]; ok {
				videos[i].TagName += item.Name + "、"
			}
		}
		videos[i].TagName = strings.TrimRight(videos[i].TagName, "、")
		if item, ok := arcs[videos[i].AvID]; ok {
			videos[i].Title = item.Title
			videos[i].UpName = item.Author.Name
		}
		if videos[i].AvID > 0 {
			videos[i].BvID, _ = common.GetBvID(videos[i].AvID)
		}
	}
	for _, item := range tagRes {
		if tagItem, ok := tags[item.TagID]; ok {
			returnTags = append(returnTags, &aggregation.Tag{
				ID:   tagItem.Id,
				Name: tagItem.Name,
			})
		}
	}
	return
}
