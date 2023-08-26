package fit

import (
	"context"
	"encoding/json"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/model/fit"
	favgrpc "go-main/app/community/favorite/service/api"
	"strconv"
	"strings"
	"time"
)

const _PlanListLimit = 20

// TaskHistoryProgress 用户打卡任务完成历史数据
func (s *Service) TaskHistoryCountProgress(ctx context.Context, mid int64, activityId int64) (*fit.UserSignDaysRes, error) {
	var start []byte
	res := &fit.UserSignDaysRes{IsJoin: 1}
	for {
		var (
			countReply *actPlat.GetCounterResResp
		)
		countReply, err := client.ActPlatClient.GetCounterRes(ctx, &actPlat.GetCounterResReq{
			Activity: strconv.FormatInt(activityId, 10),
			Counter:  s.c.FitCounter.HasLimit,
			Mid:      mid,
			Time:     0,
			Start:    start,
		})
		if err != nil {
			log.Errorc(ctx, "fit activity get grpc client.ActPlatClient.GetCounterRes() mid(%d) error(%+v)", mid, err)
			return nil, err
		}
		if countReply == nil || countReply.CounterList == nil {
			log.Warnc(ctx, "fit activity get client.ActPlatClient.GetCounterRes() mid(%d) historyReply is nil", mid)
			res.SignDays = 0
			return res, nil
		}
		// 计算已经打卡天数
		for _, v := range countReply.CounterList {
			res.SignDays += v.Val
		}

		start = countReply.Next
		if len(start) == 0 || countReply.Next == nil {
			break
		}
	}
	return res, nil
}

// GetPlanCardList 获取系列计划列表
func (s *Service) GetPlanCardList(ctx context.Context, pn, ps int) (*fit.PlanRecordListRes, error) {
	res := &fit.PlanRecordListRes{}
	// 读缓存
	var err error
	res.PlanList, err = s.fitDao.CacheGetPlanList(ctx)
	if err == nil && res.PlanList != nil {
		res.Page = pn
		res.Size = len(res.PlanList)
		res.Total = len(res.PlanList)
		return res, nil
	} else {
		log.Errorc(ctx, "service.GetPlanCardList get CacheGetPlanList err,error is (%v).or res is nil.", err)
	}

	// 回源，去db中取数据
	offset := (pn - 1) * ps
	limit := ps
	list, err := s.fitDao.GetPlanList(ctx, offset, limit)
	if err != nil {
		log.Errorc(ctx, "service.GetPlanCardList err,error is (%v).", err)
		return nil, ecode.GetPlanListErr
	}
	if list == nil {
		return nil, nil
	}
	res.PlanList = list
	// 塞缓存
	err = s.fitDao.CacheSetPlanList(ctx, res.PlanList)
	if err != nil {
		log.Errorc(ctx, "service.GetPlanCardList.fitDao.CacheSetPlanList err,error is (%v).", err)
	}
	res.Page = pn
	res.Size = len(list)
	// 第一期全部去除，所以可以直接读取，后续分页取出另外再加获取全部数据量
	res.Total = len(list)

	return res, nil
}

// GetPlanCardDetail
func (s *Service) GetPlanCardDetail(ctx context.Context, planId int64, mid int64, activityId int64) (*fit.PlanWeekBodanList, error) {
	res := &fit.PlanWeekBodanList{}
	// 从缓存中获取
	res, err := s.fitDao.CacheGetPlanDeatailById(ctx, planId)
	if err == nil && res.List != nil {
		// 更新一遍缓存中的isviewed
		// 获取用户今日观看视频aid集合
		aidsMap := make(map[int64]int64)
		aidsMap, err = s.todayUserViewedAids(ctx, mid, activityId, s.c.FitCounter.NoLimit)
		if err != nil {
			// 打日志不阻碍后面列表产出
			log.Errorc(ctx, "GetPlanCardDetail get todayUserViewedAids err , error is (%v).", err)
		}
		for _, v := range res.List {
			for _, v1 := range v.List {
				flag := s.isAidInMap(v1.Aid, aidsMap)
				v1.IsViewed = flag
			}
		}
		return res, nil
	} else {
		log.Errorc(ctx, "service.GetPlanCardList get CacheGetPlanList err,error is (%v).or res is nil.", err)
	}
	// 回源
	res, err = s.GetPlanCardDetailFromDB(ctx, planId, mid, activityId)
	if err == nil && res != nil && len(res.List) > 0 {
		// 塞缓存
		err = s.fitDao.CacheSetPlanDeatailById(ctx, planId, res)
		if err != nil {
			log.Errorc(ctx, "service.GetPlanCardDetail.fitDao.CacheSetPlanDeatailById err,error is (%v).", err)
		}
	}

	return res, nil

}

// GetPlanCardDetailFromDB db获取系列计划详情
func (s *Service) GetPlanCardDetailFromDB(ctx context.Context, planId int64, mid int64, activityId int64) (*fit.PlanWeekBodanList, error) {
	res := &fit.PlanWeekBodanList{}
	// 获取计划详情
	planInfo, err := s.fitDao.GetPlanById(ctx, planId)
	if err != nil {
		log.Errorc(ctx, "service.GetPlanCardDetail err!error is (%v)", err)
		return nil, ecode.GetPlanDetailErr
	}
	if planInfo == nil {
		return nil, nil
	}

	var (
		bodanStr    string
		bodanIds    []string
		mlids       []int64
		sortMlidMap = make(map[int64]*fit.BodanDetail, 0)
		count       int
		reply       *favgrpc.FoldersReply
	)
	fvideos := &favgrpc.FavoritesReply{}

	// 获取用户今日观看视频aid集合
	aidsMap := make(map[int64]int64)
	aidsMap, err = s.todayUserViewedAids(ctx, mid, activityId, s.c.FitCounter.NoLimit)
	if err != nil {
		// 打日志不阻碍后面列表产出
		log.Errorc(ctx, "GetPlanCardDetail get todayUserViewedAids err , error is (%v).", err)
	}
	// 拿到播单
	bodanStr = planInfo.BodanId
	if bodanStr != "" {
		bodanIds = strings.Split(bodanStr, "-")
		for _, bodanId := range bodanIds {
			tmp, err := strconv.ParseInt(bodanId, 10, 64)
			if err != nil {
				return nil, ecode.StringToInt64Err
			}
			mlids = append(mlids, tmp)
		}
		// grpc获取播单详情 type=2代表视频类收藏夹
		reply, err = s.favDao.Folders(ctx, mlids, 2)
		if err != nil {
			log.Errorc(ctx, "service.GetPlanCardDetail & get fav.Folders err!error is (%v)", err)
			return nil, ecode.GetPlanDetailErr
		}
		for _, folder := range reply.Res {

			if folder.RecentOids == nil {
				continue
			}
			var aids []int64
			// 排序用
			aidsSortMap := map[int64]*fit.VideoDetail{}

			tmpBodanDetail := &fit.BodanDetail{}
			tmpBodanDetail.BodanId = folder.Mlid
			tmpBodanDetail.BodanTitle = folder.Name
			tmpBodanDetail.BodanDesc = folder.Description

			// grpc获取播单里的视频列表
			fvideos, err = s.favDao.FavoritesAll(ctx, 2, mid, folder.Mid, folder.ID, 1, _PlanListLimit)
			if err != nil {
				log.Errorc(ctx, "service.GetPlanCardDetail & get fav.FavoritesAll err!"+
					"error is (%v),mid is (%v),uid is (%v),folderid is (%v)", err, mid, folder.Mid, folder.ID)
				return nil, err
			}
			if fvideos.Res.List == nil {
				log.Errorc(ctx, "service.GetPlanCardDetail & get fav.FavoritesAll result is nil")
				return nil, nil
			}
			for _, v := range fvideos.Res.List {
				aids = append(aids, v.Oid)
			}
			// 获取视频信息
			var archive map[int64]*api.Arc
			if len(aids) > 0 {
				archive, err = s.archive.AllArchiveInfo(ctx, aids)
				if err != nil {
					log.Errorc(ctx, "s.GetPlanCardDetail.getAllArchiveInfo err(%v)", err)
					continue
				}
			}
			// 填充视频字段
			for aid, v := range archive {
				flag := s.isAidInMap(aid, aidsMap)
				aidsSortMap[aid] = &fit.VideoDetail{
					Aid:       aid,
					Title:     v.Title,
					Duration:  v.Duration,
					Pic:       v.Pic,
					View:      v.Stat.View,
					Reply:     v.Stat.Reply,
					Danmaku:   v.Stat.Danmaku,
					ShortLink: v.ShortLinkV2,
					IsViewed:  flag,
				}
			}
			for _, v := range aids {
				tmpBodanDetail.List = append(tmpBodanDetail.List, aidsSortMap[v])
			}
			sortMlidMap[tmpBodanDetail.BodanId] = tmpBodanDetail
			count += int(folder.Count)
		}
		// mlids重新排序播单
		for _, mlid := range mlids {
			res.List = append(res.List, sortMlidMap[mlid])
		}
	}
	res.Count = count
	return res, nil

}

// todayUserViewedAids 用户今日观看的aid
func (s *Service) todayUserViewedAids(ctx context.Context, mid int64, activityId int64, counter string) (aidsMap map[int64]int64, err error) {
	var (
		start []byte
	)
	aidsMap = make(map[int64]int64)
	timeTPL := "2006-01-02"
	today := time.Now().Format(timeTPL)
	for {
		var (
			historyReply  *actPlat.GetHistoryResp
			historySource *fit.HistorySource
		)
		if historyReply, err = client.ActPlatClient.GetHistory(ctx, &actPlat.GetHistoryReq{
			Activity: strconv.FormatInt(activityId, 10),
			Counter:  counter,
			Mid:      mid,
			Start:    start,
		}); err != nil {
			log.Errorc(ctx, "todayIsViewed:client.ActPlatClient.GetHistory() mid(%d) error(%+v)", mid, err)
			return
		}
		if historyReply == nil {
			log.Warnc(ctx, "todayIsViewed:client.ActPlatClient.GetHistory() mid(%d) historyReply is nil", mid)
			return
		}

		// 抽取aid
		for _, v := range historyReply.History {
			if today == time.Unix(v.Timestamp, 0).Format(timeTPL) {
				if err = json.Unmarshal([]byte(v.Source), &historySource); err != nil {
					continue
				}
				aidsMap[historySource.Aid] = historySource.Aid
			}
		}
		start = historyReply.Next
		if len(start) == 0 {
			break
		}
	}
	return
}

// isAidInMap 某个aid是否在map[aid]aid里
func (s *Service) isAidInMap(aid int64, aidsMap map[int64]int64) bool {
	if aidsMap == nil {
		return false
	}
	if _, ok := aidsMap[aid]; ok {
		return true
	}
	return false
}

// GetHotTags 获取热门视频标签
func (s *Service) GetHotTags(ctx context.Context) *fit.HotTagsListRes {
	res := &fit.HotTagsListRes{}
	for k, v := range s.c.FitHotVideoConf {
		res.List = append(res.List, &fit.TagInfo{
			Title:   k,
			BodanId: v,
		})
	}
	// 排序
	bubbleSort(res.List)
	return res
}

func bubbleSort(data []*fit.TagInfo) {
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data)-i-1; j++ {
			if data[j].BodanId > data[j+1].BodanId {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}

// GetHotVideosByTag 根据tag标签获取热门视频列表
func (s *Service) GetHotVideosByTag(ctx context.Context, mlid string, mid int64, activityId int64, pn, ps int) (*fit.HotVideosRes, error) {
	res := &fit.HotVideosRes{}
	aids := []int64{}
	// 排序用
	aidsSortMap := map[int64]*fit.VideoDetail{}

	// 获取用户今日观看视频aid集合
	aidsMap := make(map[int64]int64)
	aidsMap, err := s.todayUserViewedAids(ctx, mid, activityId, s.c.FitCounter.NoLimit)
	if err != nil {
		// 打日志不阻碍后面列表产出
		log.Errorc(ctx, "GetHotVideosByTag get todayUserViewedAids err , error is (%v).", err)
	}
	// conf中获取当前tag页aid列表
	aids = s.CutCurrentPageAids(s.c.FitHotVideoList[mlid], pn, ps)
	// 查询video详情
	var archive map[int64]*api.Arc
	if len(aids) > 0 {
		archive, err = s.archive.AllArchiveInfo(ctx, aids)
		if err != nil {
			log.Errorc(ctx, "s.GetHotVideosByTag.getAllArchiveInfo err(%v)", err)
			return nil, err
		}
	}

	// 填充返回视频字段
	for aid, v := range archive {
		flag := s.isAidInMap(aid, aidsMap)
		aidsSortMap[aid] = &fit.VideoDetail{
			Aid:       aid,
			Title:     v.Title,
			Duration:  v.Duration,
			Pic:       v.Pic,
			View:      v.Stat.View,
			Reply:     v.Stat.Reply,
			Danmaku:   v.Stat.Danmaku,
			ShortLink: v.ShortLinkV2,
			IsViewed:  flag,
		}
	}
	// 乱序需要重新排序
	for _, v := range aids {
		res.VideoList = append(res.VideoList, aidsSortMap[v])
	}

	res.Total = len(s.c.FitHotVideoList[mlid])
	res.Size = len(res.VideoList)
	res.Page = pn
	return res, nil
}

// CutCurrentPageAids 获取分页里的aids
func (s *Service) CutCurrentPageAids(aids []int64, pn, ps int) []int64 {
	res := []int64{}
	if (pn-1)*ps >= len(aids) {
		return res
	}
	if len(aids)-(pn-1)*ps <= ps {
		res = aids[(pn-1)*ps:]
	} else {
		res = aids[(pn-1)*ps : pn*ps]
	}
	return res
}
