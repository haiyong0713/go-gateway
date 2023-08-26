package feedback

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	bm "go-common/library/net/http/blademaster"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	mdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// FeedbackList get app feedback list.
func (s *Service) FeedbackList(c context.Context, req *mdl.FeedbackReq) (res *model.FeedbackList, err error) {
	var count int
	if count, err = s.fkDao.FeedbackCount(c, req.AppKey, req.VersionCode, req.Buvid, req.Brand, req.Model, req.Osver, req.Province, req.Isp, req.Description, req.Remark, req.Business, req.Principal, req.CrashReason, req.Operator, req.Mid, req.Status, req.RobotKey, req.IsBug, req.CrashStartTime, req.CrashEndTime, req.CreateStartTime, req.CreateEndTime); err != nil {
		log.Errorc(c, "%+v", err)
		return
	}
	if count == 0 {
		return
	}
	var rows []*mdl.FeedbackDB
	if rows, err = s.fkDao.FeedbackList(c, req.AppKey, req.VersionCode, req.Buvid, req.Brand, req.Model, req.Osver, req.Province, req.Isp, req.Description, req.Remark, req.Business, req.Principal, req.CrashReason, req.Operator, req.Mid, req.Status, req.RobotKey, req.IsBug, req.CrashStartTime, req.CrashEndTime, req.CreateStartTime, req.CreateEndTime, req.Pn, req.Ps); err != nil {
		log.Errorc(c, "%+v", err)
		return
	}
	var resp []*mdl.FeedbackRes
	uArr := uniqUsers(rows)
	uMap, _ := s.userInfoMap(c, uArr)
	for _, r := range rows {
		resp = append(resp, r.Convert2Resp(uMap))
	}
	res = &model.FeedbackList{
		PageInfo: &model.PageInfo{
			Total: count,
			Pn:    req.Pn,
			Ps:    req.Ps,
		},
		Items: resp,
	}
	return
}

// FeedbackAdd add a feedback
func (s *Service) FeedbackAdd(c *bm.Context, req *mdl.FeedbackReq) (interface{}, error) {
	if req.Status == fawkes.NilInt {
		req.Status = 0
	}
	uname := utils.GetUsername(c)
	// bot id -> webhookzx
	var (
		webhooks []string
		err      error
	)
	if webhooks, err = s.RobotIdToWebHook(c, req.WxRobotIds); err != nil {
		log.Errorc(c, "%v", err)
		return nil, err
	}
	req.WxRobots = webhooks
	row := req.Convert2DB(uname)
	insert, err := s.fkDao.FeedbackInsert(c, row)

	return struct {
		Id int64 `json:"id"`
	}{insert}, err
}

func (s *Service) FeedbackInfo(c *bm.Context, req *mdl.FeedbackReq) (res *mdl.FeedbackRes, err error) {
	var r mdl.FeedbackDB
	if r, err = s.fkDao.FeedbackQueryByPk(c, req.ID); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	uArr := uniqUsers([]*mdl.FeedbackDB{&r})
	uMap, _ := s.userInfoMap(c, uArr)
	res = r.Convert2Resp(uMap)
	return
}

func (s *Service) FeedbackDel(c *bm.Context, req *mdl.FeedbackReq) (resp interface{}, err error) {
	effected, err := s.fkDao.FeedbackDeleteByPk(c, req.ID)
	return struct {
		IsDelete   bool  `json:"isDelete"`
		EffectRows int64 `json:"effectRows"`
	}{effected > 0, effected}, err
}

func (s *Service) FBTapdBugCreate(c *bm.Context, req *mdl.FeedbackTapdBug) (bugID string, err error) {
	bugID, err = s.fkDao.CreateTapdBug(c, req, s.c.TAPD.Token)
	return
}

// FeedbackUpdate update a feedback
func (s *Service) FeedbackUpdate(c *bm.Context, req *mdl.FeedbackReq) (resp interface{}, err error) {
	var origin mdl.FeedbackDB
	if origin, err = s.fkDao.FeedbackQueryByPk(c, req.ID); err != nil {
		log.Errorc(c, "%+v", err)
		return
	}
	var webhooks []string
	if webhooks, err = s.RobotIdToWebHook(c, req.WxRobotIds); err != nil {
		log.Errorc(c, "%v", err)
		return nil, err
	}
	req.WxRobots = webhooks
	row := req.Convert2DB(nil)
	var eff int64
	if eff, err = s.fkDao.FeedbackUpdateByPk(c, row); err != nil {
		log.Errorc(c, "%+v", err)
		return
	}
	s.event.Publish(UpdateEvent, utils.CopyTrx(c), &origin)
	return struct {
		IsUpdate   bool  `json:"isUpdate"`
		EffectRows int64 `json:"effectRows"`
	}{eff > 0, eff}, err
}

func uniqUsers(rows []*mdl.FeedbackDB) (re []string) {
	set := make(map[string]bool)
	for _, v := range rows {
		for _, e := range strings.Split(v.Editor, ",") {
			set[e] = true
		}
		set[v.Principal] = true
		set[v.Operator] = true
	}
	delete(set, "")
	for k := range set {
		re = append(re, k)
	}
	return
}

func (s *Service) userInfoMap(c context.Context, uArr []string) (uMap map[string]*mdl.UserInfo, err error) {
	infos, err := s.fkDao.UserInfo(c, uArr)
	if err != nil {
		log.Errorc(c, "%+v", err)
		return
	}
	uMap = make(map[string]*mdl.UserInfo)
	for _, v := range infos {
		uMap[v.UserName] = v
	}
	return
}

func (s *Service) AlertFeedback() {
	var (
		feedbacks []*mdl.FeedbackDB
		robots    []*mdl.Robot
		c         = context.Background()
		err       error
	)
	robots, err = s.fkDao.AppRobotList(c, "", "", "社区问题反馈-弹幕", 1)
	if len(robots) == 0 || err != nil {
		log.Errorc(c, "AlertFeedback get AppRobotList error: %v", err)
		return
	}
	danmuRobot := robots[0].WebHook
	feedbacks, err = s.fkDao.FeedbackAlert(c, danmuRobot)
	if err != nil {
		log.Error("FeedbackAlert error: %v", err)
		return
	}
	var (
		danmuNotShowCount  = 0 //弹幕不展示问题
		danmuCatonCount    = 0 //弹幕卡顿问题
		danmuNumCount      = 0 //弹幕数量问题
		danmuSettingCount  = 0 //弹幕设置问题
		danmuSubtitleCount = 0 //字幕问题
		otherIssueCount    = 0 //弹幕其它问题
	)
	for _, fd := range feedbacks {
		if mdl.DanmuNotShow.MatchString(fd.Description) {
			danmuNotShowCount++
		} else if mdl.DanmuCaton.MatchString(fd.Description) {
			danmuCatonCount++
		} else if mdl.DanmuNum.MatchString(fd.Description) {
			danmuNumCount++
		} else if mdl.DanmuSetting.MatchString(fd.Description) {
			danmuSettingCount++
		} else if mdl.DanmuSubtitle.MatchString(fd.Description) {
			danmuSubtitleCount++
		} else {
			otherIssueCount++
		}
	}
	msgMarkdown := fmt.Sprintf(`弹幕问题统计 %v 日0点截至当前 %v
	>不展示问题：<font color="#E6A23C">%v</font>
	>卡顿问题：<font color="#E6A23C">%v</font>
	>数量问题：<font color="#E6A23C">%v</font>
	>设置问题：<font color="#E6A23C">%v</font>
	>字幕问题：<font color="#E6A23C">%v</font>
	>其它问题：<font color="#E6A23C">%v</font>
	`, time.Now().Format("2006-01-02"), time.Now().Format("15:04"), danmuNotShowCount, danmuCatonCount, danmuNumCount, danmuSettingCount, danmuSubtitleCount, otherIssueCount)
	err = s.fkDao.RobotNotify(mdl.DanmuNotifRobot, &mdl.Markdown{
		Content: msgMarkdown,
	})
	if err != nil {
		log.Errorc(c, "wx robot notify Markdown error: %v", err)
	}
	// 由于markdown 无法通过手机号 @用户，所以多发一条text格式 @木名（zhongyihong）晴风(xuqiang01) 花泽真菜(zhangjinhao)
	err = s.fkDao.RobotNotify(mdl.DanmuNotifRobot, &mdl.Text{
		Content:             "",
		MentionedMobileList: []string{"18621719461", "18831993662", "18817671669"},
	})
	if err != nil {
		log.Errorc(c, "wx robot notify Text error: %v", err)
	}
}

func (s *Service) RobotIdToWebHook(c context.Context, robotIds []int64) (webhooks []string, err error) {
	mu := sync.Mutex{}
	group := errgroup.WithContext(c)
	for _, botId := range robotIds {
		var (
			bot *appmdl.Robot
			bid = botId
		)
		group.Go(func(ctx context.Context) error {
			if bot, err = s.fkDao.AppRobotInfoById(c, bid); err != nil {
				log.Errorc(c, "%v", err)
				return err
			}
			mu.Lock()
			if bot != nil {
				webhooks = append(webhooks, bot.WebHook)
			}
			mu.Unlock()
			return nil
		})
	}
	if err = group.Wait(); err != nil {
		log.Errorc(c, "group.Wait %v", err)
	}
	return
}
