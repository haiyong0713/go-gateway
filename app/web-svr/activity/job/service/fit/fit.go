package fit

import (
	"context"
	"encoding/json"
	"fmt"
	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	tunnelV2Mdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	"sync"
	"time"

	"go-common/library/database/sql"
	"go-gateway/app/app-svr/archive/service/api"
	apiAI "go-gateway/app/web-svr/activity/interface/api"
	fmdl "go-gateway/app/web-svr/activity/job/model/fit"
	favgrpc "go-main/app/community/favorite/service/api"
	"strconv"
	"strings"
)

const (
	LIMIT                    = 50
	SetMemberExpireTime      = 600
	sql4FetchReservedMIDList = `
SELECT id, mid
FROM act_reserve_%v
WHERE sid = ?
	AND id > ?
    AND state = 1
ORDER BY id ASC
LIMIT 1000
`
	_defaultTimeStr  = "2006-01-02 15:04:05"
	_defaultCardDays = 10000
	_databusMidBatch = 100
	_setMemberBatch  = 500
	_5WReward        = 7
	_PIFUReward      = 15
	_7WReward        = 20
	_8WReward        = 30
	videoNumLimit    = 20
)

// FlushPlanCardData 刷数据半天一刷？
func (s *Service) FlushPlanCardData(ctx context.Context) error {
	// 查询全部plan计划
	allPlans := make([]*fmdl.PlanRecordRes, 0)
	allPlans, err := s.dao.GetPlanList(ctx, 0, LIMIT)
	if err != nil {
		log.Errorc(ctx, "Job FlushPlanCardData err, err is (%+v)", err)
		return err
	}

	// 分别拿到播单id，批量查询播单视频
	for _, plan := range allPlans {
		var views int32 = 0
		var danmaku int32 = 0
		var (
			mlids []int64
			aids  []int64
		)
		bodanStr := plan.BodanId
		if bodanStr == "" {
			continue
		}
		bodanIds := strings.Split(bodanStr, "-")
		for _, bodanId := range bodanIds {
			tmp, err := strconv.ParseInt(bodanId, 10, 64)
			if err != nil {
				log.Errorc(ctx, "Job service.FlushPlanCardData err, err is (%+v)", err)
				return err
			}
			mlids = append(mlids, tmp)

		}
		// grpc获取播单详情 type=2代表视频类收藏夹
		folderReply, err := s.dao.Folders(ctx, mlids, 2)
		if err != nil {
			log.Errorc(ctx, "JOB FlushPlanCardData & get fav.Folders err!error is (%v)", err)
			return err
		}
		// 根据每个收藏夹详情拿当前计划卡片全部aids
		for _, folder := range folderReply.Res {
			fvideos := &favgrpc.FavoritesReply{}
			fvideos, err = s.dao.FavoritesAll(ctx, 2, 1, folder.Mid, folder.ID, 1, videoNumLimit)
			if err != nil {
				log.Errorc(ctx, "job FlushPlanCardData & get fav.FavoritesAll err!"+
					"error is (%v),mid is (%v),uid is (%v),folderid is (%v)", err, 1, folder.Mid, folder.ID)
				continue
			}
			if fvideos.Res.List == nil {
				log.Errorc(ctx, "job FlushPlanCardData & get fav.FavoritesAll result is nil,"+
					"folder id is (%v),mid is (%v).", folder.ID, folder.Mid)
				continue
			}
			for _, v := range fvideos.Res.List {
				aids = append(aids, v.Oid)
			}

		}
		// 获取视频信息
		var archive map[int64]*api.Arc
		if len(aids) > 0 {
			archive, err = s.arcs(ctx, aids, 2)
			if err != nil {
				log.Errorc(ctx, "job FlushPlanCardData get arcs err(%v)", err)
				continue
			}
			// 累加一个卡片视频播放数、弹幕数
			for _, v := range archive {
				views += v.Stat.View
				danmaku += v.Stat.Danmaku
			}
		}
		// 更新db
		_, err = s.dao.UpdateOnePlanById(ctx, plan.ID, views, danmaku)
		if err != nil {
			return err
		}

	}
	// 更新缓存
	return nil
}

func (s *Service) arcs(c context.Context, aids []int64, retryCnt int) (arcs map[int64]*api.Arc, err error) {
	var arcsRly *api.ArcsReply
	for i := 0; i < retryCnt; i++ {
		if arcsRly, err = s.arcClient.Arcs(c, &api.ArcsRequest{Aids: aids}); err == nil {
			arcs = arcsRly.Arcs
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

// SetMemberIntToRWHttp 给任务同步视频池视频 20分钟过期 10分钟跑一次？
func (s *Service) SetMemberIntToRWHttp(ctx context.Context) error {
	// 全部aid集合
	aids := make([]int64, 0)
	// 1.从系列计划中读取播单，再读取视频aid
	allPlans := make([]*fmdl.PlanRecordRes, 0)
	allPlans, err := s.dao.GetPlanList(ctx, 0, LIMIT)
	if err != nil {
		log.Errorc(ctx, "Job SetMemberIntToRWHttp err, err is (%+v)", err)
		return err
	}
	// 分别拿到播单id，批量查询播单视频
	for _, plan := range allPlans {
		mlids := make([]int64, 0)
		bodanStr := plan.BodanId
		if bodanStr == "" {
			continue
		}
		bodanIds := strings.Split(bodanStr, "-")
		for _, bodanId := range bodanIds {
			tmp, err := strconv.ParseInt(bodanId, 10, 64)
			if err != nil {
				log.Errorc(ctx, "Job SetMemberIntToRWHttp err, err is (%+v)", err)
				return err
			}
			mlids = append(mlids, tmp)

		}
		// grpc获取播单详情 type=2代表视频类收藏夹
		folderReply, err := s.dao.Folders(ctx, mlids, 2)
		if err != nil {
			log.Errorc(ctx, "JOB SetMemberIntToRWHttp & get fav.Folders err!error is (%v)", err)
			return err
		}
		// 根据每个收藏夹详情拿当前计划卡片全部aids
		for _, folder := range folderReply.Res {
			fvideos := &favgrpc.FavoritesReply{}
			fvideos, err = s.dao.FavoritesAll(ctx, 2, 1, folder.Mid, folder.ID, 1, videoNumLimit)
			if err != nil {
				log.Errorc(ctx, "job SetMemberIntToRWHttp & get fav.FavoritesAll err!"+
					"error is (%v),mid is (%v),uid is (%v),folderid is (%v)", err, 1, folder.Mid, folder.ID)
				continue
			}
			if fvideos.Res.List == nil {
				log.Errorc(ctx, "job SetMemberIntToRWHttp & get fav.FavoritesAll result is nil,"+
					"folder id is (%v),mid is (%v).", folder.ID, folder.Mid)
				continue
			}
			for _, v := range fvideos.Res.List {
				aids = append(aids, v.Oid)
			}

		}
	}
	// 2.从conf中读取热门视频aid
	for _, aid := range s.c.FitHotVideo {
		aids = append(aids, aid...)
	}
	// 3.分批将aids同步给任务
	for i := 0; i < len(aids); i += _setMemberBatch {
		subAids := make([]int64, 0)
		if i+_setMemberBatch <= len(aids) {
			// 满一批
			subAids = aids[i : i+_setMemberBatch]
		} else {
			// 不足一批
			subAids = aids[i:]
		}
		values := make([]*actPlat.SetMemberInt, 0)
		for _, aid := range subAids {
			vl := &actPlat.SetMemberInt{
				Value:      aid,
				ExpireTime: SetMemberExpireTime,
			}
			values = append(values, vl)
		}
		req := &actPlat.SetMemberIntReq{
			Activity: strconv.FormatInt(s.c.FitJobConfig.DaKaActivityId, 10),
			Name:     "set",
			Values:   values,
		}
		_, err = s.actPlatClient.AddSetMemberInt(ctx, req)
		if err != nil {
			log.Errorc(ctx, "SetMemberIntToRWHttp s.actPlatClient.AddSetMemberInt err,error is (%v).", err)
			return err
		}
	}
	log.Infoc(ctx, "SetMemberIntToRWHttp success!")
	return err
}

// FlushPlanData 说系列计划播放数和弹幕数
func (service *Service) FlushPlanData() {
	ctx := context.Background()
	log.Infoc(ctx, "FlushPlanData start!")
	if err := service.FlushPlanCardData(ctx); err != nil {
		log.Errorc(ctx, "FlushPlanData err :%v", err)
	}
	return
}

func (s *Service) SendGiftToUserRailGun(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	log.Info("SendGiftToUserRailGun start!")
	historyMsg := &fmdl.ActPlatHistoryTopicMsg{}
	if err := json.Unmarshal(msg.Payload(), historyMsg); err != nil {
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: historyMsg.MID,
		Item:  historyMsg,
	}, nil
}

// SendAwardRailGun 给满足条件用户刷礼物
func (s *Service) SendAwardRailGun(c context.Context, item interface{}) railgun.MsgPolicy {
	msg, ok := item.(*fmdl.ActPlatHistoryTopicMsg)
	if !ok {
		log.Error("SendAwardRailGun item  error(%+v)", ok)
		return railgun.MsgPolicyIgnore
	}
	actId, err := strconv.ParseInt(msg.Activity, 10, 64)
	if err != nil || actId != s.c.FitJobConfig.DaKaActivityId || msg.Counter != s.c.FitJobConfig.GetCounterResHasLimitCounter {
		log.Error("SendAwardRailGun item  error(%+v)", ok)
		return railgun.MsgPolicyIgnore
	}
	log.Infoc(c, "SendAwardRailGun databus ActPlatHistory data(%+v)", msg)
	// 查看参与活动的用户连续打卡天数
	signDaysObj := &fmdl.UserSignDaysRes{}
	signDaysObj, err = s.TaskHistoryCountProgress(c, msg.MID, actId, s.c.FitJobConfig.GetCounterResHasLimitCounter)
	if err != nil {
		log.Error("SendAwardRailGun TaskHistoryCountProgress error(%+v)", err)
		return railgun.MsgPolicyIgnore
	}
	// 根据规则发放奖励
	// 判断今天是否应该发皮肤
	if signDaysObj != nil && signDaysObj.SignDays == s.c.FitJobConfig.PiFuDay {
		// 未发放进行奖品发放 发奖品本身会做去重
		_ = SendAward(c, msg.MID, s.c.FitJobConfig.AwardId, msg.Activity)
	}
	return railgun.MsgPolicyNormal
}

// SendGiftToUser 给满足条件用户刷礼物
func (service *Service) SendGiftToUser() {
	defer service.waiter.Done()
	ctx := context.Background()
	log.Infoc(ctx, "SendGiftToUser start!")
	// 订阅用户参与活动消息推送
	if service.fitActivityHistorySub == nil {
		log.Infoc(ctx, "nil quit")
		return
	}
	for {
		if service.closed {
			log.Infoc(ctx, "SendGiftToUser closed!")
			return
		}
		msg, ok := <-service.fitActivityHistorySub.Messages()
		if !ok {
			log.Infoc(ctx, "SendGiftToUser databus exit!")
			return
		}
		msg.Commit()
		m := &fmdl.ActPlatHistoryTopicMsg{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("SendGiftToUser json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		actId, err := strconv.ParseInt(m.Activity, 10, 64)
		if err != nil {
			log.Error("SendGiftToUser strconv.ParseInt error(%+v)", err)
			continue
		}
		if actId != service.c.FitJobConfig.DaKaActivityId || m.Counter != service.c.FitJobConfig.GetCounterResNoLimitCounter {
			continue
		}
		log.Infoc(ctx, "SendGiftToUser databus:actPlatHistorySub ActPlatHistory data(%+v)", m)
		// 查看参与活动的用户连续打卡天数
		signDaysObj := &fmdl.UserSignDaysRes{}
		signDaysObj, err = service.TaskHistoryCountProgress(ctx, m.MID, actId, service.c.FitJobConfig.GetCounterResHasLimitCounter)
		if err != nil {
			log.Error("SendGiftToUser TaskHistoryCountProgress error(%+v)", err)
			continue
		}
		// 根据规则发放奖励
		// 判断今天是否应该发皮肤
		if signDaysObj != nil && signDaysObj.SignDays == service.c.FitJobConfig.PiFuDay {
			// 未发放进行奖品发放 发奖品本身会做去重
			log.Infoc(ctx, "SendGiftToUser SendAward mid:(%v)", m.MID)
			_ = SendAward(ctx, m.MID, service.c.FitJobConfig.AwardId, m.Activity)
		}

	}
}

// SendAward grpc
func SendAward(ctx context.Context, mid int64, awardId int64, activityId string) error {
	req := &apiAI.RewardsSendAwardReq{}
	req.Mid = mid
	req.AwardId = awardId
	req.Sync = false
	req.UpdateCache = true
	req.Business = "task"
	req.UniqueId = fmt.Sprintf("%v-%d", activityId, mid)
	_, err := client.ActivityClient.RewardsSendAward(ctx, req)
	if err != nil {
		log.Errorc(ctx, "SendGiftToUser SendAward() mid is (%v),error(%+v)", mid, err)
		return err
	}
	return nil
}

// TaskHistoryProgress 用户打卡任务完成历史数据
func (service *Service) TaskHistoryCountProgress(ctx context.Context, mid int64, activityId int64, counter string) (*fmdl.UserSignDaysRes, error) {
	var start []byte
	res := &fmdl.UserSignDaysRes{}
	for {
		var (
			countReply *actPlat.GetCounterResResp
		)
		countReply, err := client.ActplatClient.GetCounterRes(ctx, &actPlat.GetCounterResReq{
			Activity: strconv.FormatInt(activityId, 10),
			Counter:  counter,
			Mid:      mid,
			Time:     0,
			Start:    start,
		})
		if err != nil {
			log.Errorc(ctx, "SendGiftToUserget grpc client.ActPlatClient.GetCounterRes() mid(%d) error(%+v)", mid, err)
			return nil, err
		}
		if countReply == nil || countReply.CounterList == nil {
			log.Warnc(ctx, "SendGiftToUser get client.ActPlatClient.GetCounterRes() mid(%d) historyReply is nil", mid)
			return nil, ecode.FitActivityUserNotJoin
		}
		// 计算已经打卡天数
		for _, v := range countReply.CounterList {
			res.SignDays += v.Val
			res.Time = v.Time
		}
		start = countReply.Next
		if len(start) == 0 || countReply.Next == nil {
			break
		}
	}
	return res, nil
}

// SendTianMaCard 推送天马卡订阅cron调用
func (service *Service) SendTianMaCard() {
	ctx := context.Background()
	log.Infoc(ctx, "SendTianMaCard start!")
	if err := service.SendTianMaCardHttp(ctx); err != nil {
		log.Errorc(ctx, "SendTianMaCard err :%v", err)
	}
	return
}

// SendTianMaCard 推送天马卡订阅http调用
func (service *Service) SendTianMaCardHttp(ctx context.Context) (err error) {
	// 查找已订阅的用户
	reservedMIDs := service.fetchReservedUsers(ctx, service.c.FitJobConfig.DingYueActivityId)
	if len(reservedMIDs) <= 0 {
		log.Infoc(ctx, "SendTianMaCard reserved users is zero !")
		return nil
	}
	// 生成不同卡片的用户集合
	var (
		pushUser = make(map[int64][]int64, 0)
		l1       = sync.Mutex{}
	)
	eg := errgroup.WithContext(ctx)
	for _, user := range reservedMIDs {
		tmpUser := user
		// 限流
		eg.Go(func(ctx context.Context) (err error) {
			service.waiterFollowRWLimit.Wait()
			flag, err := service.IsFollowing(ctx, service.c.FitJobConfig.DaKaActivityId, tmpUser.MID)
			if err != nil {
				log.Errorc(ctx, "SendTianMaCard service.IsFollowing err,err is (%v),user mid is (%v)!", err, tmpUser.MID)
				return err
			}
			// 对每个用户进行分类
			if flag == true {
				// 查询用户打卡天数
				signDays, _ := service.TaskHistoryCountProgress(ctx, tmpUser.MID, service.c.FitJobConfig.DaKaActivityId,
					service.c.FitJobConfig.GetCounterResHasLimitCounter)
				if signDays != nil {
					// 对参与打卡的用户进行天数分类
					l1.Lock()
					if v, ok := pushUser[signDays.SignDays]; ok {
						pushUser[signDays.SignDays] = append(v, tmpUser.MID)
					} else {
						pushUser[signDays.SignDays] = append(make([]int64, 0), tmpUser.MID)
					}
					l1.Unlock()
				} else {
					flag = false
				}

			}
			if flag == false {
				l1.Lock()
				if v, ok := pushUser[_defaultCardDays]; ok {
					pushUser[_defaultCardDays] = append(v, tmpUser.MID)
				} else {
					pushUser[_defaultCardDays] = append(make([]int64, 0), tmpUser.MID)
				}
				l1.Unlock()
			}
			return nil
		})

	}
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "SendTianMaCardHttp eg.Wait error(%v)", err)
	}
	// 创建推荐流事件
	tunnelV2Req := &tunnelV2Mdl.AddEventReq{
		BizId:    1001,
		UniqueId: service.c.FitJobConfig.DingYueActivityId,
		Title:    fmt.Sprintf("健身订阅直播提醒%d", service.c.FitJobConfig.DingYueActivityId),
	}

	_, err = client.TunnelClient.AddEvent(ctx, tunnelV2Req)
	if xecode.Cause(err).Code() == fmdl.TunnelV2EventAlready { // 事件已注册不用返回错误
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "SendTianMaCard client.TunnelClient.AddEvent() error(%+v)", err)
		return err
	}

	// 新建卡片
	if err := service.newCard(ctx, pushUser); err != nil {
		log.Error("SendTianMaCardHttp SendTianMaCard service.newCard error,err is (%v)!", err)
		return err
	}

	// 发databus推送
	t := time.Now()
	eg2 := errgroup.WithContext(ctx)
	for day, users := range pushUser {
		tmpDay := day
		tmpUsers := users
		// 限流
		eg2.Go(func(ctx context.Context) error {
			service.waiterDatabusLimit.Wait()
			cardUniId := generateCardUniqueId(tmpDay)
			// 查询卡片状态
			if !service.isHasCard(ctx, service.c.FitJobConfig.DingYueActivityId, cardUniId) {
				log.Errorc(ctx, "SendTianMaCard service.isHasCard result is card not ready!uniqueid is (%v),carduniqid is (%v)!",
					service.c.FitJobConfig.DingYueActivityId, cardUniId)
				return ecode.TianMaCardNotReady
			}
			// -参与打卡：小于10天的每天推送一次；大于10天的，T+7推送一次 -未参与打卡用户集：出兜底文案
			if tmpDay <= 10 || tmpDay == _defaultCardDays {
				service.SendTunnelDatabus(ctx, tmpUsers, service.c.FitJobConfig.DingYueActivityId, cardUniId)
			} else {
				if int(t.Weekday()) == service.c.FitTianMaCardContentConf.NeedPushWeekDay {
					err = service.SendTunnelDatabus(ctx, tmpUsers, service.c.FitJobConfig.DingYueActivityId, cardUniId)
					if err != nil {
						log.Errorc(ctx, "SendTianMaCard service.SendTunnelDatabus return error,err is (%v)!", err)
					}
					return err
				}
			}
			return nil
		})

	}
	if err = eg2.Wait(); err != nil {
		log.Errorc(ctx, "SendTianMaCardHttp eg2.Wait error(%v)", err)
		return err
	}
	return nil

}

func (service *Service) newCard(ctx context.Context, pushUser map[int64][]int64) (err error) {
	// 生成content
	for day := range pushUser {
		var tempId int64
		params := make(map[string]string, 0)
		if day == _defaultCardDays {
			tempId = service.c.FitTianMaCardContentConf.DefaultTemplateId
			params["bonus"] = "20W"
		} else {
			tempId = service.c.FitTianMaCardContentConf.ChangeTemplateId
			params["day"] = strconv.FormatInt(day, 10)
		}
		cardUniId := generateCardUniqueId(day)
		CardContent := &tunnelCommon.FeedTemplateCardContent{
			TemplateId: tempId,
			Params:     params,
			Icon:       service.c.FitTianMaCardContentConf.Icon,
			Link:       service.c.FitTianMaCardContentConf.Link,
			Button: &tunnelCommon.FeedButton{
				Type: "text",
				Text: service.c.FitTianMaCardContentConf.ButtonText,
				Link: service.c.FitTianMaCardContentConf.ButtonLink,
			},
			Trace:       &tunnelCommon.FeedTrace{SubGoTo: "fit"},
			ShowTimeTag: tunnelCommon.HideTimeTag,
		}
		err = service.createTianMaCard(ctx, service.c.FitJobConfig.DingYueActivityId, cardUniId, CardContent, day)

	}
	return err

}

// generateCardUniqueId ...
func generateCardUniqueId(signedDays int64) int64 {
	// 获取当天八点时间
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.Parse("2006-01-02", timeStr)
	timeNumber := t.Unix() // 8点

	// 转换成string 与 signedDays进行拼接
	timeNumberStr := strconv.FormatInt(timeNumber, 10)
	signedDaysStr := strconv.FormatInt(signedDays, 10)
	str := timeNumberStr + signedDaysStr

	cardUniqueId, _ := strconv.ParseInt(str, 10, 64)
	return cardUniqueId

}

func (service *Service) createTianMaCard(ctx context.Context, uniqueId int64, cardUniqId int64, cardContent *tunnelCommon.FeedTemplateCardContent, signedDays int64) (err error) {
	// 创建天马卡
	timS := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), service.c.FitTianMaCardContentConf.TianMaSTime, 0, 0, 0, time.Local)
	timE := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), service.c.FitTianMaCardContentConf.TianMaEndSTime, 0, 0, 0, time.Local)
	feedTemplateReq := &tunnelV2Mdl.UpsertCardFeedTemplateReq{
		BizId:        1001,
		UniqueId:     service.c.FitJobConfig.DingYueActivityId,
		CardUniqueId: cardUniqId,
		TriggerType:  "time",
		StartTime:    timS.Format(_defaultTimeStr),
		EndTime:      timE.Format(_defaultTimeStr),

		CardContent: cardContent,
		Description: fmt.Sprintf("健身打卡活动,连续打卡(%d)天用户卡", signedDays),
	}

	for i := 0; i < 3; i++ {
		_, err = client.TunnelClient.UpsertCardFeedTemplate(ctx, feedTemplateReq)
		log.Infoc(ctx, "SendTianMaCard client.TunnelClient.UpsertCardMsgTemplate req(%+v) err(%+v)", feedTemplateReq, err)
		if err == nil {
			break
		}
	}
	return
}

func (service *Service) fetchReservedUsers(ctx context.Context, activityId int64) []*fmdl.ReservedUser {
	var lastID int64
	lastID = 0
	suffix := activityId % 100
	list := make([]*fmdl.ReservedUser, 0)
	for {
		// 分批1000/每次的取
		tmpList, dbErr := fetchReservedMIDFromDB(ctx, activityId, lastID, fmt.Sprintf("%02d", suffix))
		if dbErr != nil {
			log.Error("FetchReservedMID err:,err is (%v)!", dbErr)
			break
		}

		if len(tmpList) > 0 {
			lastID = tmpList[len(tmpList)-1].ID
			list = append(list, tmpList...)
		}
		if len(tmpList) < 1000 {
			break
		}
	}
	return list
}

func fetchReservedMIDFromDB(ctx context.Context, activityID, lastID int64, suffix string) (list []*fmdl.ReservedUser, err error) {
	list = make([]*fmdl.ReservedUser, 0)
	query := fmt.Sprintf(sql4FetchReservedMIDList, suffix)

	var rows *sql.Rows
	rows, err = component.GlobalDBOfRead.Query(ctx, query, activityID, lastID)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		tmp := new(fmdl.ReservedUser)
		if tmpErr := rows.Scan(&tmp.ID, &tmp.MID); tmpErr == nil {
			list = append(list, tmp)
		}
	}
	err = rows.Err()

	return
}

// IsFollowing 是否预约
func (service *Service) IsFollowing(c context.Context, sid int64, mid int64) (flag bool, err error) {
	res, err := client.ActivityClient.ReserveFollowing(c, &apiAI.ReserveFollowingReq{
		Sid: sid,
		Mid: mid,
	})
	if err != nil {
		log.Errorc(c, "client.ActivityClient.ReserveFollowing err(%v)", err)
		return
	}
	return res.IsFollow, nil
}

func (service *Service) isHasCard(ctx context.Context, uniqueID, cardUniqueID int64) bool {
	// 查询卡片状态
	flag := false
	cardReq := &tunnelV2Mdl.CardReq{
		BizId:        1001,
		UniqueId:     uniqueID,
		CardUniqueId: cardUniqueID,
	}
	for i := 0; i < 10; i++ {
		cardRes, err := client.TunnelClient.Card(ctx, cardReq)
		log.Infoc(ctx, "SendTianMaCard client.TunnelClient.Card req(%+v) err(%+v)", cardReq, err)

		if err == nil && cardRes != nil && cardRes.State == tunnelCommon.CardStateDelivering {
			flag = true
			break
		}
		if err != nil {
			log.Errorc(ctx, "isHasCard error, uniqueID(%d) CardUniqueId(%v),err is (%v)!", uniqueID, cardUniqueID, err)
		}
		time.Sleep(time.Second * 5)
	}
	return flag
}

// SendTunnelDatabus 发消息推送
func (service *Service) SendTunnelDatabus(ctx context.Context, mids []int64, uniqueID, cardUniqueID int64) (err error) {
	if len(mids) < 0 {
		return
	}
	for i := 0; i < len(mids); i += _databusMidBatch {
		subMids := make([]int64, 0)
		if i+_databusMidBatch <= len(mids) {
			// 满一批
			subMids = mids[i : i+_databusMidBatch]
		} else {
			// 不足一批
			subMids = mids[i:]
		}
		reqParam := struct {
			BizID        int64   `json:"biz_id"`
			UniqueID     int64   `json:"unique_id"`
			Mids         []int64 `json:"mids"`
			State        int8    `json:"state"`
			CardUniqueId int64   `json:"card_unique_id"`
			timestamp    int64   `json:"timestamp"`
		}{1001, uniqueID, subMids, 1, cardUniqueID, time.Now().Unix()}
		if err = service.fitTunnelPub.Send(ctx, strconv.FormatInt(time.Now().UnixNano()/1e6, 10), reqParam); err != nil {
			log.Errorc(ctx, "SendTunnelDatabus fitTunnelPub.Send error, mids(%v) uniqueID(%d) CardUniqueId(%v) error(%+v)", subMids, uniqueID, cardUniqueID, err)
		}
		log.Infoc(ctx, "SendTunnelDatabus fitTunnelPub.Send success, mids(%v) uniqueID(%d) CardUniqueId(%v)!", mids, uniqueID, cardUniqueID)
	}

	return
}

// ExportUser 导出用户
func (service *Service) ExportUser(ctx context.Context) (map[int64][]int64, error) {
	// 查找所有参与打卡用户
	reservedMIDs := service.fetchReservedUsers(ctx, service.c.FitJobConfig.DaKaActivityId)
	if len(reservedMIDs) <= 0 {
		log.Infoc(ctx, "ExportUser signup reserved users is zero !")
		return nil, nil
	}
	// 查询每个用户的连续打卡情况
	userResult := make(map[int64][]int64, 0)
	for _, user := range reservedMIDs {
		// 查询用户打卡天数
		userSignInfo, _ := service.TaskHistoryCountProgress(ctx, user.MID, service.c.FitJobConfig.DaKaActivityId,
			service.c.FitJobConfig.GetCounterResHasLimitCounter)
		if userSignInfo != nil {
			// 漏斗
			if userSignInfo.SignDays >= _5WReward {
				userResult[_5WReward] = append(userResult[_5WReward], user.MID)
			}
			if userSignInfo.SignDays >= _PIFUReward {
				userResult[_PIFUReward] = append(userResult[_PIFUReward], user.MID)
			}
			if userSignInfo.SignDays >= _7WReward {
				userResult[_7WReward] = append(userResult[_7WReward], user.MID)
			}
			if userSignInfo.SignDays >= _8WReward {
				userResult[_8WReward] = append(userResult[_8WReward], user.MID)
			}
		}
	}
	return userResult, nil
}

// SetMemberIntToRW 给任务刷视频池子
func (service *Service) SetMemberIntToRW() {
	ctx := context.Background()
	log.Infoc(ctx, "SetMemberIntToRW start!")
	if err := service.SetMemberIntToRWHttp(ctx); err != nil {
		log.Errorc(ctx, "SetMemberIntToRW err :%v", err)
	}
	return
}
