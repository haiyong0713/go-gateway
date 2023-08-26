package tips

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/tips"
	Log "go-gateway/app/app-svr/app-feed/admin/util"
)

func (s *Service) OpenSearchTips(c context.Context, ids []int64, startTs, endTs xtime.Time, searchWord string, status int, ps, pn int) (tipsList []model.SearchTipRes, total int, err error) {
	var (
		tipsListDB []model.SearchTipDB
		tipsIdList []int64
		queryMap   map[int64][]model.SearchTipQueryDB
	)
	// 查出tips
	if tipsListDB, total, err = s.dao.FindTipList(c, ids, startTs, endTs, searchWord, status, ps, pn); err != nil {
		return
	}

	for _, tipItem := range tipsListDB {
		tipsIdList = append(tipsIdList, tipItem.ID)
	}

	// 查出query的map
	if queryMap, err = s.dao.FindQueryMap(c, tipsIdList); err != nil {
		return
	}

	var normalPlat = []model.PlatRes{{
		PlatType: 0,
		PlatName: "android",
	}, {
		PlatType: 1,
		PlatName: "ios",
	}, {
		PlatType: 30,
		PlatName: "web",
	}}

	// 拼装结果
	for _, v := range tipsListDB {
		if queryItem, ok := queryMap[v.ID]; ok {
			plat := normalPlat
			if hasSuicideWord(queryItem) {
				plat = normalPlat[2:3]
			}
			tipsList = append(tipsList, model.SearchTipRes{
				ID:          v.ID,
				Title:       v.Title,
				SubTitle:    v.SubTitle,
				IsImmediate: v.IsImmediate,
				SearchWord:  queryItem,
				STime:       v.STime,
				CUser:       v.CUser,
				Status:      v.Status,
				HasBgImg:    v.HasBgImg,
				JumpUrl:     v.JumpUrl,
				Plat:        plat,
			})
		}
	}
	return
}

func hasSuicideWord(queries []model.SearchTipQueryDB) bool {
	for _, q := range queries {
		if q.SearchWord == "自杀" {
			return true
		}
	}
	return false
}

func (s *Service) SearchTipAdd(c context.Context, tip *model.SearchTipRes, username string, uid int64) (err error) {
	var (
		insertTip = model.SearchTipDB{
			Title:    tip.Title,
			SubTitle: tip.SubTitle,
			STime:    tip.STime,
			CUser:    username,
			Status:   tip.Status,
			HasBgImg: tip.HasBgImg,
			JumpUrl:  tip.JumpUrl,
		}
		insertQuery = tip.SearchWord
		plat        = 0
	)
	for _, v := range tip.Plat {
		//nolint:gomnd
		if plat >= 3 {
			plat = 3
			break
		}
		plat += v.PlatType
	}
	insertTip.Plat = plat

	if tip.IsImmediate != 0 {
		insertTip.IsImmediate = 1
		insertTip.STime = xtime.Time(time.Now().Unix())
		insertTip.Status = 1
	}

	var pass bool
	if pass, err = s.dao.CheckConflict(c, 0, insertTip.STime, insertTip.Plat, insertQuery); err != nil || !pass {
		return
	}

	if err = s.dao.InsertTip(c, insertTip, insertQuery); err != nil {
		return
	}

	obj := map[string]interface{}{
		"value": tip,
		"id":    0,
	}
	if err = Log.AddLogs(common.LogSearchTips, username, uid, 0, "SearchTipAdd", obj); err != nil {
		log.Error("search tips SearchTipAdd AddLog error(%v)", err)
		return
	}

	log.Info("searchTips service SearchTipAdd(%v) success", tip)

	return
}

func (s *Service) SearchTipUpdate(c context.Context, tip *model.SearchTipRes, username string, uid int64) (err error) {
	if tip.ID == 0 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}
	var (
		updateTip = model.SearchTipDB{
			ID:       tip.ID,
			Title:    tip.Title,
			SubTitle: tip.SubTitle,
			STime:    tip.STime,
			Status:   tip.Status,
			HasBgImg: tip.HasBgImg,
			JumpUrl:  tip.JumpUrl,
		}
		updateQuery = tip.SearchWord
		plat        = 0
	)

	//for range tip.Plat {
	//	//nolint:gomnd
	//	if plat >= 3 {
	//		plat = 3
	//		break
	//	}
	//	plat += 1
	//}
	updateTip.Plat = plat

	var pass bool
	if pass, err = s.dao.CheckConflict(c, tip.ID, updateTip.STime, updateTip.Plat, updateQuery); err != nil || !pass {
		return
	}

	// 立即生效，status 就是 1，否则认为是定时生效，全都置为待生效，等待 job 去变更
	if tip.IsImmediate != 0 {
		updateTip.IsImmediate = 1
		updateTip.STime = xtime.Time(time.Now().Unix())
		updateTip.Status = 1
	} else {
		updateTip.Status = 0
		updateTip.ETime = 0
	}

	if err = s.dao.UpdateTip(c, updateTip, updateQuery); err != nil {
		return
	}

	obj := map[string]interface{}{
		"value": tip,
		"id":    tip.ID,
	}
	if err = Log.AddLogs(common.LogSearchTips, username, uid, tip.ID, "SearchTipUpdate", obj); err != nil {
		log.Error("search tips SearchTipUpdate AddLog error(%v)", err)
		return
	}

	log.Info("searchTips service SearchTipUpdate(%v) success", tip)

	return
}

func (s *Service) SearchTipOperate(c context.Context, id int64, operation int, username string, uid int64) (err error) {
	if id == 0 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}
	if err = s.dao.UpdateTipOperation(c, []int64{id}, operation); err != nil {
		return
	}

	tipsList, _, _ := s.OpenSearchTips(c, []int64{id}, 0, 0, "", -1, 1, 1)

	if len(tipsList) == 0 {
		tipsList = []model.SearchTipRes{{
			ID: id,
		}}
	}

	obj := map[string]interface{}{
		"value": tipsList[0],
		"id":    id,
	}
	action := "SearchTipOperate"
	//nolint:gomnd
	switch operation {
	case 1:
		{
			action = "SearchTipOnline"
		}
	case 2:
		{
			action = "SearchTipOffline"
		}
	}
	if err = Log.AddLogs(common.LogSearchTips, username, uid, id, action, obj); err != nil {
		log.Error("search tips SearchTipOperate AddLog error(%v)", err)
		return
	}

	log.Info("searchTips service SearchTipOperate() id(%v) operation(%v) success", id, operation)

	return
}

func (s *Service) PublishMonitor() {
	for {
		err := s.PublishJob()
		if err != nil {
			log.Error("tips PublishMonitor job error(%v)", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func (s *Service) PublishJob() (err error) {
	// 查出所有 status=0 的配置，检查是否到上线时间，到了就更改 status=1
	var (
		tipsListDB    []model.SearchTipDB
		c             = context.Background()
		needUpdateIds []int64
	)
	if tipsListDB, _, err = s.dao.FindTipList(c, nil, 0, 0, "", 0, 1000, 1); err != nil {
		return
	}
	currentTime := xtime.Time(time.Now().Unix())
	for _, v := range tipsListDB {
		if v.STime <= currentTime {
			needUpdateIds = append(needUpdateIds, v.ID)
		}
	}
	if len(needUpdateIds) > 0 {
		if err = s.dao.UpdateTipOperation(c, needUpdateIds, 1); err != nil {
			return
		}
		log.Info("searchTips service PublishJob() ids(%v) success", needUpdateIds)
	}
	return
}
