package rank

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	actclient "git.bilibili.co/bapis/bapis-go/activity/service"
	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"

	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
)

// 添加新榜单
func (s Service) AddNewRank(req *rankModel.RankConfigReq, uname string, uid int64) (err error) {
	if err = s.dao.InsertRankConfig(req, uname); err != nil {
		log.Error("AddNewRank s.dao.InsertRankConfig req(%v) err(%v)", req, err)
	}

	return

}

// 编辑榜单配置
func (s Service) EditRankConfig(req *rankModel.EditRankConfigReq, uname string, uid int64) (err error) {
	if err = s.dao.UpdateRankConfig(req, uname); err != nil {
		log.Error("EditRankConfig s.dao.UpdateRankConfig req(%v) err(%v)", req, err)
	}

	return
}

// 发榜
func (s Service) RankPublish(c *bm.Context, rankid int, username string) (err error) {
	var (
		count      int
		configList []*rankModel.RankConfig
	)
	if configList, count, err = s.dao.QueryRankConfigList(rankid, "", -1, 0, 1, 20); err != nil {
		log.Error("GetRankConfig s.dao.QueryRankConfigList error(%v)", err)
	}

	if count == 0 {
		err = ecode.Error(ecode.RequestErr, "榜单不存在")
		return
	}

	config := configList[0]

	//nolint:gomnd
	if config.State == 3 {
		err = ecode.Error(ecode.RequestErr, "已结榜的榜单不能发榜")
		return
	}

	// 得到真实排名
	var rankVideos []int64
	rankVideos, _, err = s.GetShowAvidList(c, rankid)
	if err != nil {
		log.Error("s.GetShowAvidList rankId(%v) error(%v)", rankid, err)
		//nolint:ineffassign,staticcheck
		rankVideos = []int64{}
		return
	}
	var avrank []string

	length := len(rankVideos)
	//nolint:gomnd
	if length >= 100 {
		length = 100
	}
	for i := 0; i < length; i++ {
		avrank = append(avrank, strconv.FormatInt(rankVideos[i], 10))
	}

	if err = s.dao.RankPublish(rankid, avrank, username); err != nil {
		log.Error("dao RankPublish error(%v) rankid(%v) ", err, rankid)
		return
	}

	return

}

// 结榜
func (s Service) RankTerminate(c *bm.Context, tmcontent *rankModel.TernimateContent, username string) (err error) {

	// 得到真实排名
	var rankVideos []int64
	_, rankVideos, err = s.GetShowAvidList(c, tmcontent.Id)
	if err != nil {
		log.Error("s.GetShowAvidList rankId(%v) error(%v)", tmcontent.Id, err)
		//nolint:ineffassign,staticcheck
		rankVideos = []int64{}
		return
	}

	var (
		count      int
		configList []*rankModel.RankConfig
	)
	if configList, count, err = s.dao.QueryRankConfigList(tmcontent.Id, "", -1, 0, 1, 20); err != nil {
		log.Error("GetRankConfig s.dao.QueryRankConfigList error(%v)", err)
	}

	if count == 0 {
		err = ecode.Error(ecode.RequestErr, "榜单不存在")
		return
	}

	config := configList[0]

	if config.HistoryId == 0 {
		err = ecode.Error(ecode.RequestErr, "当前榜单从未发过榜单，请先发榜，再结榜")
		return
	}

	avrank := []string{}

	length := len(rankVideos)
	//nolint:gomnd
	if length >= 100 {
		length = 100
	}
	for i := 0; i < length; i++ {
		avrank = append(avrank, strconv.FormatInt(rankVideos[i], 10))
	}

	if err = s.dao.RankTerminate(config.HistoryId, tmcontent, avrank, username); err != nil {
		log.Error("dao RankPublish error(%v) rankid(%v) ", err, tmcontent.Id)
		return
	}

	// 将状态修改为"已结榜"
	if err = s.dao.UpdateRankState(config.ID, 3); err != nil {
		log.Error("JobUpdateRankState s.dao.UpdateRankState error(%v)", err)
		return
	}

	return

}

// 更改榜单状态
func (s Service) RankOption(c *bm.Context, req *rankModel.RankCommonQuery) (err error) {
	// 参数检查
	var rank *rankModel.RankConfigRes
	rank, err = s.GetRankConfig(c, req)
	if err != nil {
		log.Error("RankOption s.GetRankConfig err(%v)", err)
		return
	}

	if req.State == 1 {
		if time.Now().Unix() > rank.ETime.Time().Unix() {
			// log.Error("RankOption s.dao.RankOption req(%v) err(%v)", req, err)
			err = ecode.Error(-1, "当前时间已经大于结束时间!")
			return
		}
	}

	if err = s.dao.RankOption(req.Id, req.State); err != nil {
		log.Error("RankOption s.dao.RankOption req(%v) err(%v)", req, err)
		return
	}
	return

}

// 获取所有榜单
func (s Service) GetRankList(ctx context.Context, req *rankModel.RankCommonQuery, uname string, uid int64) (pager *rankModel.RankListPager, err error) {
	var (
		count      int
		configList []*rankModel.RankConfig
		// tagIds     []int64
		// actIds     []int64
		// tagMap   map[int64]*tag.Tag
	)

	if configList, count, err = s.dao.QueryRankConfigList(req.Id, req.Keyword, req.State, req.Time, req.Page, req.Size); err != nil {
		log.Error("s.dao.QueryRankConfigList error(%v)", err)
		return
	}

	pager = &rankModel.RankListPager{
		Page: &common.Page{
			Total: count,
			Size:  req.Size,
			Num:   req.Page,
		},
	}

	// 获取所有的分区信息
	var TypesReply *arcgrpc.TypesReply
	if TypesReply, err = s.arcClient.Types(ctx, &arcgrpc.NoArgRequest{}); err != nil {
		return
	}
	for _, config := range configList {
		var (
			tids   []*rankModel.IdAndName
			actIds []*rankModel.IdAndName
			tags   []*rankModel.IdAndName
		)

		// 根据分区id填写分区name
		for _, tid := range string2Int64Array(config.Tids) {
			types, ok := TypesReply.Types[int32(tid)]
			if ok {

				tids = append(tids, &rankModel.IdAndName{
					ID:   int(tid),
					Name: types.Name,
				})
			} else {
				tids = append(tids, &rankModel.IdAndName{
					ID:   int(tid),
					Name: "错误!未能获取TID信息!",
				})
			}
		}
		// 根据活动ID填写活动name,以及所有活动ID下对应的tag name
		for _, actid := range string2Int64Array(config.ActIds) {
			req := actclient.ActSubProtocolReq{
				Sid: actid,
			}
			var actSubProtocolReply *actclient.ActSubProtocolReply
			if actSubProtocolReply, err = s.actClient.ActSubProtocol(ctx, &req); err != nil {
				actIds = append(actIds, &rankModel.IdAndName{
					ID:   int(actid),
					Name: "错误!未能获取actid信息!",
				})
				// 这里需要手动处理错误
				err = nil
				continue
			}
			actIds = append(actIds, &rankModel.IdAndName{
				ID:   int(actid),
				Name: actSubProtocolReply.Subject.Name,
			})

			actType := actSubProtocolReply.Subject.Type
			// 只在这四种情况下填写tag name,其中是22是打卡,4/13/16是视频源
			if actType == 4 || actType == 13 || actType == 16 || actType == 22 {
				if actSubProtocolReply.Protocol != nil {
					tags = append(tags, &rankModel.IdAndName{
						Name: actSubProtocolReply.Protocol.Tags,
					})
				}
				if actSubProtocolReply.Rules != nil {
					for _, v := range actSubProtocolReply.Rules {
						tags = append(tags, &rankModel.IdAndName{
							Name: v.Tags,
						})
					}
					tags = append(tags, &rankModel.IdAndName{
						Name: actSubProtocolReply.Protocol.Tags,
					})
				}
			}

		}

		pager.Item = append(pager.Item, &rankModel.RankListItem{
			Id:     config.ID,
			Title:  config.Title,
			Tids:   tids,
			ActIds: actIds,
			Tags:   tags,
			State:  config.State,
			STime:  config.STime,
			ETime:  config.ETime,
			CUser:  config.CUser,
		})
	}

	return
}

// 获取当前时刻的avid列表
func (s Service) GetShowAvidList(ctx context.Context, rankId int) (avidList, avidListDedup []int64, err error) {
	var (
		allList      []rankModel.RankDetailAVItem
		allListDedup []rankModel.RankDetailAVItem
	)
	if allList, allListDedup, err = s.GetRankAVShowList(ctx, rankId); err != nil {
		log.Error("s.GetRankAVShowList id(%v) error(%v)", rankId, err)
		return
	}

	for _, item := range allList {
		if item.ShowRank > 0 {
			avidList = append(avidList, item.Avid)
		}
	}

	for _, item := range allListDedup {
		if item.ShowRank > 0 {
			avidListDedup = append(avidListDedup, item.Avid)
		}
	}

	return
}

// 获取榜单下的视频列表,已经经过干预排序
func (s Service) GetRankAVList(ctx context.Context, req *rankModel.RankCommonQuery) (rankAVList *rankModel.RankDetailPager, err error) {
	var (
		allAvList     []rankModel.RankDetailAVItem
		count         int
		configList    []*rankModel.RankConfig
		rankhistory   *rankModel.RankHistoryDB
		publish_state int
		bvstr         string
		logDate       string
		jobFinishTime int64
	)
	ps := req.Size
	pn := req.Page

	if configList, count, err = s.dao.QueryRankConfigList(req.Id, "", -1, 0, 1, 20); err != nil {
		log.Error("GetRankConfig s.dao.QueryRankConfigList error(%v)", err)
	}
	if count == 0 {
		return
	}
	config := configList[0]
	title := config.Title
	if rankhistory, err = s.dao.FindRankHistoryConfig(config.HistoryId); err != nil {
		log.Error("GetRankConfig s.dao.FindRankHistoryConfig error(%v)", err)
		return
	}
	jobPublishTime := rankhistory.LogData
	// 有数据没发榜->1
	// 有数据已发榜->0
	// 无数据->2
	if originalMaxLogDate, jobMTime, dateError := s.dao.GetOriginalRankScoreListTime(config.ID); dateError != nil {
		log.Error("s.dao.GetOriginalRankScoreListTime rankId(%v) error(%v)", config.ID, dateError)
		logDate = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	} else {
		logDate = originalMaxLogDate
		jobFinishTime = jobMTime.Unix()
	}

	if _, count, err = s.dao.GetOriginalRankScoreList(config.ID, logDate); err != nil {
		log.Error("s.dao.GetOriginalRankScoreList rankId(%v) logDate(%v) error(%v)", config.ID, logDate, err)
		return
	}
	if count == 0 {
		// 说明没有数据
		publish_state = 2
	} else {
		// 获取当天0点的unix时间戳,与最新的发榜时间比较
		year, month, day := time.Now().Date()
		today := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
		if today > jobPublishTime.Time().Unix() {
			// 说明没有发榜
			publish_state = 1
		} else {
			// 说明已经发榜
			publish_state = 0
		}
	}

	// 获取当前页的视频列表,已经经过干预排序
	if allAvList, _, err = s.GetRankAVShowList(ctx, req.Id); err != nil {
		log.Error("s.GetRankAVShowList id(%v) error(%v)", req.Id, err)
		return
	}

	// 按条件过滤稿件
	// avid mid is_hidden
	if req.Avid != 0 {
		tempAVlist := []rankModel.RankDetailAVItem{}
		for _, v := range allAvList {
			if v.Avid == req.Avid {
				tempAVlist = append(tempAVlist, v)
			}
		}
		allAvList = tempAVlist
	}

	if req.Mid != 0 {
		tempAVlist := []rankModel.RankDetailAVItem{}
		for _, v := range allAvList {
			if v.User.Uid == req.Mid {
				tempAVlist = append(tempAVlist, v)
			}
		}
		allAvList = tempAVlist
	}

	if req.IsHidden != -1 {
		tempAVlist := []rankModel.RankDetailAVItem{}
		for _, v := range allAvList {
			if v.IsHidden == req.IsHidden {
				tempAVlist = append(tempAVlist, v)
			}
		}
		allAvList = tempAVlist
	}

	end := ps * pn
	if end >= len(allAvList) {
		end = len(allAvList)
	}
	currentAvList := allAvList[ps*(pn-1) : end]

	// 根据视频列表填充相应的用户信息
	//nolint:staticcheck
	archives := &arcgrpc.ArcsReply{}
	var aids []int64
	for i := 0; i < len(currentAvList); i++ {
		aids = append(aids, currentAvList[i].Avid)
	}
	if aids == nil {
		currentAvList = nil
	} else {
		if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: aids}); err != nil {
			log.Error("s.arcClient.Arcs error %v", err)
			return
		}
		for i := 0; i < len(currentAvList); i++ {
			archive, ok := archives.Arcs[currentAvList[i].Avid]
			if ok {
				currentAvList[i].Title = archive.Title
				currentAvList[i].User.Uname = archive.Author.Name
				currentAvList[i].User.Uid = archive.Author.GetMid()
				if bvstr, err = bvav.ToBvStr(strconv.FormatInt(currentAvList[i].Avid, 10)); err != nil {
					return
				}
				currentAvList[i].Bvid = bvstr
			} else {
				log.Error("can't get archive info from rank (%v),avid: (%v) error %v", config.ID, currentAvList[i].Avid, err)
				currentAvList[i].Title = "错误,未能获取标题!"
				currentAvList[i].User.Uname = "错误,未能获取用户名!"
				currentAvList[i].User.Uid = -1
				currentAvList[i].Bvid = "-1"
			}
		}
	}

	rankAVList = &rankModel.RankDetailPager{
		Title:         title,
		JobFinishTime: jobFinishTime,
		PublishState:  publish_state,
		List:          currentAvList,
		Page: common.Page{
			Num:  pn,
			Size: ps,
			// 上线后要特别注意此处
			Total: len(allAvList),
		},
	}

	return
}

// 获取榜单的具体配置
func (s Service) GetRankConfig(ctx context.Context, req *rankModel.RankCommonQuery) (configDetail *rankModel.RankConfigRes, err error) {
	var (
		count      int
		configList []*rankModel.RankConfig
	)
	if configList, count, err = s.dao.QueryRankConfigList(req.Id, "", -1, 0, 1, 20); err != nil {
		log.Error("GetRankConfig s.dao.QueryRankConfigList error(%v)", err)
	}

	if count == 0 {
		return
	}

	config := configList[0]

	var (
		tids      []*rankModel.IdAndName
		actIds    = []*rankModel.IdAndName{}
		blacklist []*rankModel.UserItem
	)
	// 获取所有的分区信息,填写分区ID和name
	var TypesReply *arcgrpc.TypesReply
	if TypesReply, err = s.arcClient.Types(ctx, &arcgrpc.NoArgRequest{}); err != nil {
		return
	}
	for _, v := range string2Int64Array(config.Tids) {
		name := TypesReply.Types[int32(v)].Name
		tids = append(tids, &rankModel.IdAndName{
			ID:   int(v),
			Name: name,
		})
	}

	// 填写活动ID和name,以及对应的TAG name
	for _, v := range string2Int64Array(config.ActIds) {
		req := actclient.ActSubProtocolReq{
			Sid: v,
		}
		var actSubProtocolReply *actclient.ActSubProtocolReply
		if actSubProtocolReply, err = s.actClient.ActSubProtocol(ctx, &req); err != nil {
			actIds = append(actIds, &rankModel.IdAndName{
				ID:   int(v),
				Name: strconv.FormatInt(v, 10),
			})
			continue
		}
		actIds = append(actIds, &rankModel.IdAndName{
			ID:   int(v),
			Name: actSubProtocolReply.Subject.Name,
		})
	}
	// 根据黑名单返回用户ID
	for _, v := range string2Int64Array(config.Blacklist) {
		req := account.MidReq{
			Mid: v,
		}
		var InfoReply *account.InfoReply
		if InfoReply, err = s.accClient.Info3(ctx, &req); err != nil {
			return
		}
		blacklist = append(blacklist, &rankModel.UserItem{
			Uid:   v,
			Uname: InfoReply.Info.Name,
		})
	}

	var scoreConfig []*rankModel.ScoreConfig
	//nolint:staticcheck
	err = json.Unmarshal([]byte(config.ScoreConfig), &scoreConfig)

	var description []*rankModel.Description
	err = json.Unmarshal([]byte(config.Description), &description)
	configDetail = &rankModel.RankConfigRes{
		ID:                config.ID,
		Title:             config.Title,
		STime:             config.STime,
		ETime:             config.ETime,
		Cycle:             config.Cycle,
		PerUpdate:         config.PerUpdate,
		Tids:              tids,
		ActIds:            actIds,
		ArchiveStime:      config.ArchiveStime,
		ArchiveEtime:      config.ArchiveEtime,
		ArchiveSelectMode: config.ArchiveSelectMode,
		ScoreConfig:       scoreConfig,
		Blacklist:         blacklist,
		Cover:             config.Cover,
		Description:       description,
		AvManuallyList:    string2Int64Array(config.AvManuallyList),
		State:             config.State,
	}

	return
}
