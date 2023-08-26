package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"strconv"
	"time"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"

	tunnelCommon "git.bilibili.co/bapis/bapis-go/platform/common/tunnel"
	tunnelV2Mdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
)

const (
	_imagePre           = "https://i0.hdslb.com"
	_imagePreUat        = "https://uat-i0.hdslb.com"
	_contestTwoTeam     = 2
	_tunnelBGroupSource = "esports"
	_defaultTimeStr     = "2006-01-02 15:04:05"
)

// InitTunnelEvent 创建事件 .
func (d *dao) InitTunnelEvent(ctx context.Context, contest *model.ContestModel) (err error) {
	// 注册事件 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-1.1%E6%B3%A8%E5%86%8C%E4%BA%8B%E4%BB%B6
	season, err := d.GetSeasonByID(ctx, contest.Sid)
	if err != nil {
		log.Errorc(ctx, "InitTunnelEvent d.getSeasonByID() contestID(%d) error(%+v)", contest.ID, err)
		return
	}
	tunnelV2Req := &tunnelV2Mdl.AddEventReq{
		BizId:    d.conf.TunnelBGroup.TunnelBizID,
		UniqueId: contest.ID,
		Title:    fmt.Sprintf("赛事订阅直播提醒%d(%s-%s)", contest.ID, season.Title, contest.GameStage),
	}
	_, err = d.tunnelV2Client.AddEvent(ctx, tunnelV2Req)
	if xecode.Cause(err).Code() == model.TunnelV2EventAlready { // 事件已注册不用返回错误
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "InitTunnelEvent d.tunnelV2Client.AddEvent() contestID(%d) error(%+v)", contest.ID, err)
		return fmt.Errorf("小卡注册事件出错(%+v)", err)
	}
	return
}

// UpsertTunnelCard 创建卡片 .
func (d *dao) UpsertTunnelCard(ctx context.Context, contest *model.ContestModel) (err error) {
	// 新增/更新Feed订阅卡-模板模式 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0Feed%E8%AE%A2%E9%98%85%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0Feed%E8%AE%A2%E9%98%85%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F
	var (
		teamsMap         map[int64]*model.TeamModel
		season           *model.SeasonModel
		params           = make(map[string]string)
		imgPre           string
		cartInitialState = "active"
	)
	if env.DeployEnv == env.DeployEnvUat {
		imgPre = _imagePreUat
	} else {
		imgPre = _imagePre
	}
	if contest.Status == model.FreezeTrue {
		cartInitialState = "frozen"
	}
	if season, teamsMap, err = d.fetchSeasonTeamsByContest(ctx, contest); err != nil {
		log.Errorc(ctx, "UpsertTunnelCard d.fetchSeasonTeamsByContest() contest(%+v) error(%+v)", contest, err)
		return
	}
	// 主客队不存在不发送
	if len(teamsMap) < _contestTwoTeam {
		return
	}
	homeTeam, ok := teamsMap[contest.HomeID]
	if !ok {
		return
	}
	awayTeam, ok := teamsMap[contest.AwayID]
	if !ok {
		return
	}
	// params.
	params["teamA"] = homeTeam.Title
	params["teamB"] = awayTeam.Title
	params["stage"] = contest.GameStage
	params["season"] = season.Title
	// cardContent.
	twoLink := fmt.Sprintf(d.conf.TunnelBGroup.Link, contest.LiveRoom)
	cardContent := &tunnelCommon.FeedTemplateCardContent{
		TemplateId: d.conf.TunnelBGroup.NewTemplateID,
		Params:     params,
		Link:       twoLink,
		Icon:       imgPre + season.Logo,
		Button: &tunnelCommon.FeedButton{
			Type: "text",
			Text: d.conf.TunnelBGroup.NewCardText,
			Link: twoLink,
		},
		Trace: &tunnelCommon.FeedTrace{
			SubGoTo:  "esports",
			Param:    season.ID,
			SubParam: contest.ID,
		},
		ShowTimeTag: tunnelCommon.HideTimeTag, // 不展示时间.
	}
	userInfoStruct := struct {
		Name     string `json:"name"`
		Business string `json:"business"`
	}{
		strconv.FormatInt(contest.ID, 10),
		d.conf.TunnelBGroup.NewBusiness}
	userInfo, _ := json.Marshal(userInfoStruct)
	feedTemplateReq := &tunnelV2Mdl.UpsertCardFeedTemplateReq{
		BizId:        d.conf.TunnelBGroup.TunnelBizID,
		UniqueId:     contest.ID,
		CardUniqueId: contest.ID,
		TriggerType:  "time",
		StartTime:    time.Unix(contest.Stime, 0).Format(_defaultTimeStr),
		EndTime:      time.Unix(contest.Etime, 0).Format(_defaultTimeStr),
		TargetUserGroup: &tunnelCommon.TargetUserGroup{
			UserType: tunnelCommon.BGroup,
			UserInfo: string(userInfo),
		},
		CardContent:  cardContent,
		Description:  fmt.Sprintf("赛季(%d-%s),赛程(%d),阶段(%s)", season.ID, season.Title, contest.ID, contest.GameStage),
		InitialState: cartInitialState,
	}
	_, err = d.tunnelV2Client.UpsertCardFeedTemplate(ctx, feedTemplateReq)
	errCode := xecode.Cause(err).Code()
	if errCode == model.TunnelV2NotExists || errCode == model.TunnelV2CardStatusErr {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "[Grpc][UpsertCardFeedTemplate][Retry][Error], err:%+v", err)
		return
	}
	return
}

// UpsertTunnelMsgCard 创建私信卡片 .
func (d *dao) UpsertTunnelMsgCard(ctx context.Context, contest *model.ContestModel) (err error) {
	// 新增/更新私信通知卡-模板模式 https://info.bilibili.co/pages/viewpage.action?pageId=184989531#id-%E4%B8%9C%E9%A3%8E%E7%BB%9F%E4%B8%80%E5%8D%8F%E8%AE%AE%E4%B8%9A%E5%8A%A1%E5%AF%B9%E6%8E%A5-%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0%E7%A7%81%E4%BF%A1%E9%80%9A%E7%9F%A5%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F%E6%96%B0%E5%A2%9E/%E6%9B%B4%E6%96%B0%E7%A7%81%E4%BF%A1%E9%80%9A%E7%9F%A5%E5%8D%A1-%E6%A8%A1%E6%9D%BF%E6%A8%A1%E5%BC%8F
	var (
		teamsMap         map[int64]*model.TeamModel
		season           *model.SeasonModel
		params           []string
		imgPre           string
		cartInitialState = "active"
		notifyCode       string
	)
	if env.DeployEnv == env.DeployEnvUat {
		imgPre = _imagePreUat
	} else {
		imgPre = _imagePre
	}
	if contest.Status == model.FreezeTrue {
		cartInitialState = "frozen"
	}
	if season, teamsMap, err = d.fetchSeasonTeamsByContest(ctx, contest); err != nil {
		log.Errorc(ctx, "UpsertTunnelMsgCard d.fetchSeasonTeamsByContest() contest(%+v) error(%+v)", contest, err)
		return
	}
	if params, notifyCode, err = d.getCardMsgParams(contest, season, teamsMap); err != nil {
		log.Errorc(ctx, "UpsertTunnelMsgCard  contest special type contestID(%d) error(%+v)", contest.ID, err)
		err = nil
		return
	}
	// cardMsgContent.
	cardLink := fmt.Sprintf(d.conf.TunnelCardMsg.Link, contest.LiveRoom)
	cardMsgContent := &tunnelCommon.MsgTemplateCardContent{
		NotifyCode: notifyCode,
		Params:     params,
		SenderUid:  season.MessageSenduid,
		JumpUriConfig: &tunnelCommon.MsgUriPlatform{
			AllUri: cardLink,
			Text:   d.conf.TunnelCardMsg.JumpText,
		},
		Notifier: &tunnelCommon.MsgNotifier{
			Nickname:  season.Title,
			AvatarUrl: imgPre + season.Logo,
			JumpUrl:   cardLink,
		},
	}
	userInfoStruct := struct {
		Name     string `json:"name"`
		Business string `json:"business"`
	}{
		strconv.FormatInt(contest.ID, 10),
		d.conf.TunnelCardMsg.NewBusiness}
	userInfo, _ := json.Marshal(userInfoStruct)
	feedTemplateReq := &tunnelV2Mdl.UpsertCardMsgTemplateReq{
		BizId:        d.conf.TunnelCardMsg.TunnelBizID,
		UniqueId:     contest.ID,
		CardUniqueId: d.conf.TunnelCardMsg.CardUniqueId,
		TriggerType:  "time",
		StartTime:    time.Unix(contest.Stime, 0).Format(_defaultTimeStr),
		EndTime:      time.Unix(contest.Etime, 0).Format(_defaultTimeStr),
		TargetUserGroup: &tunnelCommon.TargetUserGroup{
			UserType: tunnelCommon.BGroup,
			UserInfo: string(userInfo),
		},
		CardContent:  cardMsgContent,
		Description:  fmt.Sprintf("赛季(%d-%s),赛程(%d),阶段(%s)", season.ID, season.Title, contest.ID, contest.GameStage),
		InitialState: cartInitialState,
	}
	_, err = d.tunnelV2Client.UpsertCardMsgTemplate(ctx, feedTemplateReq)
	errCode := xecode.Cause(err).Code()
	if errCode == model.TunnelV2NotExists || errCode == model.TunnelV2CardStatusErr {
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "[Grpc][UpsertTunnelMsgCard][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) getCardMsgParams(contest *model.ContestModel, season *model.SeasonModel, teamsMap map[int64]*model.TeamModel) (params []string, notifyCode string, err error) {
	if season == nil {
		err = fmt.Errorf("getCardMsgParams contestID(%d) season is nil", contest.ID)
		return
	}
	if season.MessageSenduid == 0 {
		err = fmt.Errorf("getCardMsgParams contestID(%d) season MessageSenduid is zero", contest.ID)
		return
	}
	switch contest.Special {
	case model.DefaultContest:
		// 主客队不存在不发送
		if len(teamsMap) < _contestTwoTeam {
			err = fmt.Errorf("getCardMsgParams contestID(%d) team not two", contest.ID)
			return
		}
		homeTeam, ok := teamsMap[contest.HomeID]
		if !ok {
			err = fmt.Errorf("getCardMsgParams homeTeam(%d) not found", contest.HomeID)
			return
		}
		awayTeam, ok := teamsMap[contest.AwayID]
		if !ok {
			err = fmt.Errorf("getCardMsgParams awayTeam(%d) not found", contest.AwayID)
			return
		}
		// params.
		params = append(params, season.Title)
		params = append(params, homeTeam.Title)
		params = append(params, awayTeam.Title)
		// notifyCode
		notifyCode = d.conf.TunnelCardMsg.DefaultContestNotifyCode
	case model.ContestSpecial:
		params = append(params, season.Title)
		params = append(params, contest.GameStage)
		// notifyCode
		notifyCode = d.conf.TunnelCardMsg.SpecialContestNotifyCode
	default:
		err = fmt.Errorf("getCardMsgParams contest special type (%d) error", contest.ID)
		return
	}
	if notifyCode == "" {
		err = fmt.Errorf("getCardMsgParams notifyCode (%d) error", contest.ID)
		return
	}
	return
}

func (d *dao) BGroupDataBusPub(ctx context.Context, mid, contestID, state int64) (err error) {
	reqParam := struct {
		Mid       int64  `json:"mid"`
		Source    string `json:"source"`
		Name      string `json:"name"`
		State     int64  `json:"state"`
		Timestamp int64  `json:"timestamp"`
	}{
		mid,
		_tunnelBGroupSource,
		strconv.FormatInt(contestID, 10),
		state,
		time.Now().Unix()}
	key := strconv.FormatInt(mid, 10)
	buf, _ := json.Marshal(reqParam)
	if err = retry.WithAttempts(ctx, "interface_contest_fav_send_event", 3, netutil.DefaultBackoffConfig,
		func(c context.Context) error {
			return d.BGroupMessagePub.Send(ctx, key, buf)
		}); err != nil {
		log.Errorc(ctx, "[Dao][BGroupDataBusPub][Error] d.BGroupMessagePub.Send mid(%d) contestID(%d) reqParam(%+v) error(%+v)", mid, contestID, reqParam, err)
	}
	log.Infoc(ctx, "[Dao][BGroupDataBusPub][Info] d.BGroupMessagePub.Send mid(%d) contestID(%d) reqParam(%+v) success", mid, contestID, reqParam)
	return
}
