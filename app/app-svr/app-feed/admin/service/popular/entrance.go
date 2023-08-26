package popular

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	taGrpcModel "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"go-common/library/sync/errgroup.v2"

	"github.com/jinzhu/gorm"
)

const (
	_addEntrance           = "add"
	_editEntrance          = "up"
	_forbiddenEntrance     = "forbidden"
	_activationEntrance    = "activation"
	_activationEntranceNum = 0
	_fbEntranceNum         = 1
	_deleteEntrance        = "del"
	_delEntranceNum        = 2
	_topPhotoBuildLimit    = "粉版ios&安卓"
	_topPhotoLocationName  = "热门中间页"
	_maxCout               = 500
	_manualTag             = "人工"
	_defaultState          = 1
	_maxAids               = 100
	_maxRedDotContentLen   = 3
)

func paramJudge(param *show.EntranceSave) (err error) {
	var wList, bList []int64
	//nolint:gomnd
	if len([]rune(param.Title)) > 8 {
		err = ecode.Error(ecode.RequestErr, "入口标题超过八个字符")
		return
	}
	if param.Grey > 100 || param.Grey < 0 {
		err = ecode.Error(ecode.RequestErr, "灰度区间为[0,100]")
		return
	}
	if wList, err = xstr.SplitInts(param.WhiteList); err != nil || len(wList) > 200 {
		err = ecode.Error(ecode.RequestErr, "白名单填写错误")
		return
	}
	if bList, err = xstr.SplitInts(param.WhiteList); err != nil || len(bList) > 200 {
		err = ecode.Error(ecode.RequestErr, "黑名单填写错误")
		return
	}
	if param.RedDot != 0 && (param.RedDotText == "" || len([]rune(param.RedDotText)) > 3) {
		err = ecode.Error(ecode.RequestErr, "红点文案不能为空且不超过三个字符")
	}
	return
}

func versionJudge(version string) (err error) {
	var verCtl []*show.VersionControl
	if err = json.Unmarshal([]byte(version), &verCtl); err != nil {
		log.Error("[PopEntranceSave] json.Unmarshal() error(%v)", err)
		return
	}
	for _, v := range verCtl {
		if v.BuildStart != 0 && v.BuildEnd != 0 && v.BuildStart > v.BuildEnd {
			err = ecode.Error(ecode.RequestErr, "版本控制信息存在问题")
			return
		}
	}
	return
}

// PopEntranceSave .
func (s *Service) PopEntranceSave(ctx context.Context, param *show.EntranceSave, uid int64, uname string) (err error) {
	var id int64
	if err = paramJudge(param); err != nil {
		log.Error("[PopEntranceSave] paramJudge() error(%v)", err)
		return
	}
	if err = versionJudge(param.BuildLimit); err != nil {
		log.Error("[PopEntranceSave] versionJudge() error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"hot_title": param.Title,
	}
	if param.ID == 0 {
		if id, err = s.showDao.PopEntranceAdd(ctx, param); err != nil {
			log.Error("[PopEntranceSave] s.PopEntranceAdd() moduleID(%s) error(%v)", param.ModuleID, err)
			return
		}
		if err = util.AddLogs(common.LogEntrance, uname, uid, id, _addEntrance, obj); err != nil {
			log.Error("[PopEntranceAdd] AddLogs error(%v)", err)
		}
	} else {
		if err = s.showDao.PopEntranceEdit(ctx, param); err != nil {
			log.Error("[PopEntranceSave] s.PopEntranceEdit() error(%v)", err)
			return
		}
		if err = util.AddLogs(common.LogEntrance, uname, uid, param.ID, _editEntrance, obj); err != nil {
			log.Error("[PopEntranceEdit] AddLogs error(%v)", err)
		}
	}
	return
}

// PopularEntrance .
func (s *Service) PopularEntrance(ctx context.Context, state, pn, ps int) (res *show.EntranceListRes, err error) {
	if res, err = s.showDao.PopEntrance(ctx, state, pn, ps); err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			res.Items = make([]*show.EntranceList, 0)
			log.Error("[PopularEntrance] NO Record")
			return
		}
		log.Error("[PopularEntrance]s.PopularEnter() error(%v)", err)
		return
	}
	for _, v := range res.Items {
		if res, err := s.showDao.CacheAIChannelRes(ctx, v.ID); err == nil {
			v.VideoCount = len(res)
		}
		var verCtl []*show.VersionControl
		if jsonErr := json.Unmarshal([]byte(v.BuildLimit), &verCtl); jsonErr != nil {
			log.Error("[PopularEntrance] json.Unmarshal() jsonErr(%v)", jsonErr)
			continue
		}
		v.BuildLimitSc = verCtl
	}
	return
}

// PopEntranceOperate .
func (s *Service) PopEntranceOperate(ctx context.Context, id, uid int64, state int, uname, title string) (err error) {
	var action string
	if err = s.showDao.PopEntranceOperate(ctx, id, state); err != nil {
		log.Error("[PopEntranceOperate]s.PopEntranceOperate() id(%d) error(%v)", id, err)
		return
	}
	switch state {
	case _activationEntranceNum:
		action = _activationEntrance
	case _fbEntranceNum:
		action = _forbiddenEntrance
	case _delEntranceNum:
		action = _deleteEntrance
	}
	obj := map[string]interface{}{
		"hot_title": title,
	}
	if err = util.AddLogs(common.LogEntrance, uname, uid, id, action, obj); err != nil {
		log.Error("PopEntranceOperate error(%v)", err)
	}
	return
}

// RedDotUpdate .
func (s *Service) RedDotUpdate(ctx context.Context, operator, moduleID string, id int64) (err error) {
	var rowsAffected int64
	if rowsAffected, err = s.showDao.RedDotUpdate(ctx, moduleID, id); err != nil {
		log.Error("[RedDotUpdate] s.showDao.RedDotUpdate() moduleID(%s) operator(%s) rowsAffected(%d) error(%v)", moduleID, operator, rowsAffected, err)
		return
	}
	if rowsAffected == 0 {
		return ecode.NotModified
	}
	log.Info("[RedDotUpdate] moduleID(%s) operator(%s) rows(%d) 更新红点成功!", moduleID, operator, rowsAffected)
	return
}

// RedDotUpdateDisposable .
func (s *Service) RedDotUpdateDisposable(ctx context.Context, operator, moduleID string, id int64, content string) (err error) {
	var rowsAffected int64
	if len(content) > _maxRedDotContentLen {
		err = ecode.Error(ecode.RequestErr, "红点文案最多三个字")
	}
	if rowsAffected, err = s.showDao.RedDotUpdateDisposable(ctx, moduleID, id, content); err != nil {
		log.Error("[RedDotUpdateDisposable] s.showDao.RedDotUpdateDisposable() id(%d) operator(%s) rowsAffected(%d) error(%v)", id, operator, rowsAffected, err)
		return
	}
	if rowsAffected == 0 {
		return ecode.NotModified
	}
	log.Info("[RedDotUpdateDisposable] id(%d) operator(%s) rows(%d) 更新红点成功!", id, operator, rowsAffected)
	return
}

// PopularView .
func (s *Service) PopularView(ctx context.Context, id int64) (res *show.EntranceView, err error) {
	var (
		cacheData   []*show.PopularCard
		resource    []*show.PopChannelResource
		channelTags []*show.PopChannelTag
	)
	res = new(show.EntranceView)
	res.ID = id
	eg := errgroup.WithContext(ctx) // 获取选项
	eg.Go(func(c context.Context) (err error) {
		res.HeadImage, err = s.showDao.EntranceGetTopPhoto(ctx, id)
		if err != nil {
			log.Error("PopularView EntranceGetTopPhoto %+v", err)
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		cacheData, err = s.showDao.CacheAIChannelRes(ctx, id)
		if err != nil {
			log.Error("PopularView CacheAIChannelRes %+v", err)
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		resource, err = s.showDao.PopCRFindByTEID(ctx, id)
		if err != nil {
			log.Error("PopularView PopCRFindByTEID %+v", err)
		}
		return
	})
	eg.Go(func(c context.Context) (err error) {
		channelTags, err = s.showDao.PopCTFindByTEID(ctx, id)
		if err != nil {
			log.Error("PopularView PopCTFindByTEID %+v", err)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res.Tags, res.Videos, err = s.handlerVideosTags(ctx, cacheData, resource, channelTags)
	return
}

//nolint:gocognit
func (s *Service) handlerVideosTags(ctx context.Context, cacheData []*show.PopularCard, resource []*show.PopChannelResource,
	channelTags []*show.PopChannelTag) (returnTags []*show.Tag, videos []*show.Video, err error) {
	var (
		resM   []*show.Video // 手动新增部分
		resD   []*show.Video // 手动删除部分
		resAI  []*show.Video // AI返回部分
		aids   []int64
		tagIDs = make(map[int64]interface{})
		arcs   = make(map[int64]*api.Arc)
		tags   map[int64]*taGrpcModel.Tag
		exist  = make(map[int64]struct{}) // 用于判断AI部分是不是存在于人工部分
	)
	for _, item := range resource {
		aids = append(aids, item.RID)
		if item.TagID > 0 {
			tagIDs[item.TagID] = nil
		}
		if item.State == 1 {
			resM = append(resM, &show.Video{
				RID:   item.RID,
				State: item.State,
				TagID: item.TagID,
			})
		} else {
			resD = append(resD, &show.Video{
				RID:   item.RID,
				State: item.State,
				TagID: item.TagID,
			})
		}
		exist[item.RID] = struct{}{}
	}
	for _, item := range cacheData {
		item.TagId, _ = xstr.SplitInts(item.TagIdStr)
		if _, ok := exist[item.Value]; ok {
			for i := 0; i < len(resM); i++ {
				if resM[i].RID == item.Value {
					resM[i].TagIDs = item.TagId
				}
			}
			for i := 0; i < len(resD); i++ {
				if resD[i].RID == item.Value {
					resD[i].TagIDs = item.TagId
				}
			}
			continue
		}
		aids = append(aids, item.Value)
		var tagIdTemp int64
		if len(item.TagId) > 0 {
			for _, tagId := range item.TagId {
				tagIDs[tagId] = nil
			}
			tagIdTemp = item.TagId[0]
		}
		resAI = append(resAI, &show.Video{
			RID:    item.Value,
			State:  _defaultState,
			TagID:  tagIdTemp,
			TagIDs: item.TagId,
		})
	}
	for _, item := range channelTags {
		if item == nil {
			continue
		}
		if item.TagID > 0 {
			tagIDs[item.TagID] = nil
		}
	}
	if len(aids) > 0 { // Arcs只支持100个，需要分页查询
		pag := len(aids)/_maxAids + 1
		for i := 0; i < pag; i++ {
			maxIndix := (i + 1) * _maxAids
			if maxIndix > len(aids) {
				maxIndix = len(aids)
			}
			aidTemp := aids[i*_maxAids : maxIndix]
			// aids的长度为100/200或其他类似值时，会有空数组发送，因此处理该bug
			if len(aidTemp) == 0 {
				continue
			}
			var arcsTemp map[int64]*api.Arc
			if arcsTemp, err = s.arrDao.Arcs(ctx, aidTemp); err != nil {
				return
			}
			if len(arcsTemp) > 0 {
				for k, v := range arcsTemp {
					if _, ok := arcs[k]; !ok {
						arcs[k] = v
					}
				}
			}
		}

	}
	if len(tagIDs) > 0 {
		var tids []int64
		for tagId := range tagIDs {
			tids = append(tids, tagId)
		}
		if tags, err = s.TagGrpc(ctx, tids); err != nil {
			return
		}
	}
	videos = append(videos, resM...)
	videos = append(videos, resAI...)
	videos = append(videos, resD...)
	if len(videos) > _maxCout {
		videos = videos[:_maxCout]
	}
	for i := 0; i < len(videos); i++ {
		if videos[i].TagID == 0 {
			videos[i].TagName = _manualTag
		}
		for _, itemTemp := range videos[i].TagIDs {
			if item, ok := tags[itemTemp]; ok {
				videos[i].TagName += item.Name + "、"
			}
		}
		videos[i].TagName = strings.TrimRight(videos[i].TagName, "、")
		if item, ok := arcs[videos[i].RID]; ok {
			videos[i].Title = item.Title
			videos[i].Author = item.Author.Name
		}
		if videos[i].RID > 0 {
			videos[i].BvID, _ = common.GetBvID(videos[i].RID)
		}
	}
	for _, item := range channelTags {
		if tagItem, ok := tags[item.TagID]; ok {
			returnTags = append(returnTags, &show.Tag{
				ID:   tagItem.Id,
				Name: tagItem.Name,
			})
		}
	}
	return
}

// PopularViewSave .
func (s *Service) PopularViewSave(ctx context.Context, id int64, topPhoto string) (err error) {
	err = s.showDao.EntranceTopPhotoUpdate(ctx, id, topPhoto)
	return
}

// PopularViewAdd .
func (s *Service) PopularViewAdd(ctx context.Context, id int64, rid []int64) (err error) {
	err = s.showDao.PopChannelResourceAddM(ctx, id, rid)
	return
}

// PopularViewOperate .
func (s *Service) PopularViewOperate(ctx context.Context, id, rid, tagID int64, state int) (err error) {
	err = s.showDao.PopChannelResourceState(ctx, id, rid, tagID, state)
	return
}

// PopularTagAdd .
func (s *Service) PopularTagAdd(ctx context.Context, id int64, tagID []int64) (err error) {
	err = s.showDao.PopChannelTagAddM(ctx, id, tagID)
	return
}

// PopularTagDel .
func (s *Service) PopularTagDel(ctx context.Context, id, tagID int64) (err error) {
	err = s.showDao.PopChannelTagDelete(ctx, id, tagID)
	return
}

// PopularMiddleSave .
func (s *Service) PopularMiddleSave(ctx context.Context, id, locationId int64, topPhoto string) (err error) {
	if id == 0 {
		err = s.showDao.PopTopPhotoAdd(ctx, &show.PopTopPhotoAD{
			TopPhoto:   topPhoto,
			LocationId: locationId,
			Deleted:    common.NotDeleted,
		})
		return
	}
	err = s.showDao.PopTopPhotoUpdate(ctx, id, topPhoto)
	return
}

// PopularMiddleList .
func (s *Service) PopularMiddleList(ctx context.Context, pn, ps int) (res []*show.MiddleTopPhoto, err error) {
	var list []*show.PopTopPhoto
	if list, err = s.showDao.PopTPFind(ctx, pn, ps); err != nil {
		return
	}
	for _, item := range list {
		res = append(res, &show.MiddleTopPhoto{
			ID:           item.ID,
			LocationId:   item.LocationId,
			LocationName: _topPhotoLocationName,
			TopPhoto:     item.TopPhoto,
			BuildLimit:   _topPhotoBuildLimit,
		})
	}
	return
}

// TagGrpc .
func (s *Service) TagGrpc(ctx context.Context, ids []int64) (tags map[int64]*taGrpcModel.Tag, err error) {
	tags = map[int64]*taGrpcModel.Tag{}
	var maxCount = 50
	for i := 0; i < len(ids); i += maxCount {
		var (
			reply *taGrpcModel.TagsReply
		)
		arg := &taGrpcModel.TagsReq{
			Mid: 0,
		}
		// 判断边界值小于等于长度
		if i+maxCount <= len(ids) {
			// 获取 maxCount 个 tid
			arg.Tids = ids[i : i+maxCount]
		} else {
			// 获取剩余tids
			arg.Tids = ids[i:]
		}
		if reply, err = s.tagClient.Tags(ctx, arg); err != nil {
			return
		}
		if reply == nil || len(reply.Tags) == 0 {
			err = fmt.Errorf("参数错误，ID为%q的tag找不到", ids)
			return
		}
		for tid, tag := range reply.Tags {
			tags[tid] = tag
		}
	}
	return
}
