package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/system"
	"go-gateway/app/web-svr/activity/interface/tool"
	"strconv"
	"strings"
	"time"
)

// 用户授权企业微信
func (s *Service) WXAuth(ctx context.Context, req *model.WXAuthArgs) (sessionToken string, err error) {
	// 来源限制
	if _, ok := s.c.System.CORPSecret[req.From]; !ok {
		err = ecode.SystemFromParamsErr
		return
	}
	// 获取access_token
	var accessToken string
	if accessToken, err = s.dao.GetWXAccessToken(ctx, req.From); err != nil {
		return
	}
	// 通过code获取userID
	var userID string
	if userID, err = s.dao.GetWXUserUserIDByAccessTokenAndCode(ctx, accessToken, req.Code); err != nil {
		return
	}
	// 因为bug导致增加置换成工卡号步骤
	var uid string
	uid, err = s.dao.GetUIDWithWXUserID(ctx, userID)
	if err != nil {
		return
	}
	var DBUserInfo *model.DBUserInfo
	if DBUserInfo, err = s.dao.GetDBUserInfoByUID(ctx, uid); err != nil {
		err = ecode.SystemGetDBUserInfoFailed
		return
	}
	// 可以查询到此用户 直接下发token
	sessionToken = tool.MD5(uid + tool.RandStringRunes(10))
	if DBUserInfo.ID > 0 {
		//if err = s.dao.UpdateUserInfoByUID(ctx, uid, sessionToken); err != nil {
		//	sessionToken = ""
		//	err = ecode.SystemUpdateWXUserInfoFailed
		//	return
		//}
		sessionToken = DBUserInfo.Token
		return
	} else {
		// 查询不到此用户 新增
		if err = s.dao.CreateUserInfo(ctx, uid, sessionToken); err != nil {
			sessionToken = ""
			err = ecode.SystemCreateWXUserInfoFailed
			return
		}
	}

	return
}

// 服务启动加载信息
func (s *Service) InitEmployeesInfo(ctx context.Context) (err error) {
	data, err := s.dao.GetOAAllUsersInfo(ctx)
	if err != nil {
		return
	}
	// 因为OA中不包含外包同学信息 所以这部分信息去DB表中补充
	// 企业微信中 根据code拿到user_id这个字段可以相当于工号 不过前缀是V开头
	extraData, err := s.dao.GetExtraUsersInfo(ctx)
	if err != nil {
		return
	}
	// data 和 extraData 进行合并
	newData := make([]*model.User, 0)
	for _, v := range data {
		newData = append(newData, v)
	}
	for _, v := range extraData {
		newData = append(newData, v)
	}
	// 工号处理用newData 外包同学存在工号
	if err = s.dao.BuildMapKeyWorkCode(ctx, newData); err != nil {
		return
	}
	// id处理用oa获取的正式员工数据
	if err = s.dao.BuildMapKeyOAID(ctx, data); err != nil {
		return
	}
	return
}

func (s *Service) GetConfig(ctx context.Context, req *model.GetConfigArgs) (res *model.GetConfigRes, err error) {
	// 来源限制
	if _, ok := s.c.System.CORPSecret[req.From]; !ok {
		err = ecode.SystemFromParamsErr
		return
	}
	var JSAPITicket string
	nonceStr := tool.RandStringRunes(8)
	ts := time.Now().Unix()
	url := req.Url

	if JSAPITicket, err = s.dao.GetWXJSAPITicket(ctx, req.From); err != nil {
		return
	}

	joinStr := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s", JSAPITicket, nonceStr, ts, url)
	sha1Str := tool.SHA1(joinStr)

	res = &model.GetConfigRes{
		CORPID:    s.c.System.CORPID,
		Timestamp: ts,
		NonceStr:  nonceStr,
		Signature: sha1Str,
	}

	return
}

// 获取用户信息通过uid
func (s *Service) GetUserInfoByUID(ctx context.Context, uid string) (res *model.SystemUser, err error) {
	user, err := s.dao.GetUserInfoWithUID(ctx, uid)
	if err != nil {
		return
	}
	res = &model.SystemUser{UID: uid, NickName: user.NickName, Avatar: user.Avatar, DepartmentName: user.DepartmentName, LastName: user.LastName, UseKind: user.UseKind}
	return
}

// 获取用户信息通过Cookie
func (s *Service) GetUserInfoByCookie(ctx context.Context, sessionToken string) (res *model.SystemUser, err error) {
	if res, err = s.dao.GetUserInfoWithCookie(ctx, sessionToken); err != nil {
		return
	}
	// 判白逻辑 年会后要干掉
	if res.UseKind != "全职" && res.UseKind != "实习生转正" && tool.InStrSlice(res.UID, s.c.Party2021.UserExtra) == false {
		res = nil
		err = ecode.SystemNotIn2021PartyMembersErr
		return
	}
	return
}

// 签到
func (s *Service) Sign(ctx context.Context, aid int64, sessionToken string, location string) (err error) {
	actSubject, err := s.dao.GetActivityInfo(ctx, aid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	if actSubject.ID <= 0 {
		err = ecode.SystemNoActivityErr
		return
	}

	if actSubject.Type != model.SystemActivityTypeSign {
		err = ecode.SystemNoActivityErr
		return
	}

	ts := time.Now().Unix()
	// 未开始
	if ts < actSubject.Stime.Time().Unix() {
		err = ecode.SystemActivityNotStartErr
		return
	}
	// 已结束
	if ts > actSubject.Etime.Time().Unix() {
		err = ecode.SystemActivityIsEndErr
		return
	}

	// 获取用户信息
	var uid string
	if uid, err = s.dao.GetUserIDWithCookie(ctx, sessionToken); err != nil {
		return
	}

	// 查询是否签到过
	var record *model.ActivitySign
	record, err = s.dao.ActivitySigned(ctx, aid, uid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}
	if record.ID > 0 {
		err = ecode.SystemActivitySignedErr
		return
	}

	// 签到
	if err = s.dao.DoActivitySign(ctx, aid, uid, location); err != nil {
		err = ecode.SystemActivitySignErr
		return
	}

	return
}

// 获取活动基本信息
func (s *Service) ActivityInfo(ctx context.Context, aid int64, sessionToken string) (res *model.ActivityInfoRes, err error) {

	actSubject, err := s.dao.GetActivityInfo(ctx, aid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	if actSubject.ID <= 0 {
		err = ecode.SystemNoActivityErr
		return
	}

	// 获取用户信息
	var uid string
	if uid, err = s.dao.GetUserIDWithCookie(ctx, sessionToken); err != nil {
		return
	}

	res = new(model.ActivityInfoRes)
	res.ID = actSubject.ID
	res.Name = actSubject.Name
	res.Type = actSubject.Type
	res.Stime = actSubject.Stime
	res.Etime = actSubject.Etime
	res.Config = actSubject.Config

	// 签到活动
	if actSubject.Type == model.SystemActivityTypeSign {
		config := new(model.SystemActivitySignConfig)
		if err = json.Unmarshal([]byte(actSubject.Config), config); err != nil {
			log.Errorc(ctx, "ActivityInfo json.Unmarshal Err config:%v err:%v", config, err)
			err = ecode.SystemActivityConfigErr
			return
		}

		outPutConfig := make(map[string]interface{})
		// 签到定位开关
		outPutConfig["location"] = config.Location
		// 跳转页
		outPutConfig["jump_url"] = config.JumpURL
		// 是否展现座位表
		outPutConfig["show_seat"] = 0
		// 座位文案
		outPutConfig["seat_text"] = config.SeatText
		// 需要展现座位
		if config.ShowSeat == 1 {
			var seatInfo *model.SystemActivitySeat
			if seatInfo, err = s.dao.GetSeatInfo(ctx, aid, uid); err != nil {
				err = ecode.SystemNetWorkBuzyErr
				return
			}
			// 存在座位表信息
			if seatInfo != nil && seatInfo.ID > 0 {
				outPutConfig["show_seat"] = 1
				outPutConfig["seat_content"] = seatInfo.Content
			}
		}

		res.Config = outPutConfig
	}

	// 投票活动
	if actSubject.Type == model.SystemActivityTypeVote {
		config := new(model.SystemActivityVoteConfig)
		if err = json.Unmarshal([]byte(actSubject.Config), config); err != nil {
			log.Errorc(ctx, "ActivityInfo json.Unmarshal Err config:%v err:%v", config, err)
			err = ecode.SystemActivityConfigErr
			return
		}
		res.Config = config

		// 查询是否投票过
		isVote := 0
		var record *model.ActivityVote
		record, err = s.dao.ActivityVoted(ctx, aid, uid)
		if err != nil {
			err = ecode.SystemNetWorkBuzyErr
			return
		}
		if record.ID > 0 {
			isVote = 1
		}
		res.Extra = map[string]interface{}{
			"isVote": isVote,
		}
	}

	return
}

// 批量获取用户信息通过uid
func (s *Service) GetUsersInfoByUIDs(ctx context.Context, uids []string) (res []*model.SystemUser, err error) {
	res = make([]*model.SystemUser, 0)
	for _, uid := range uids {
		var user *model.User
		user, err = s.dao.GetUserInfoWithUID(ctx, uid)
		if err == ecode.SystemNoUserInOAErr {
			continue
		}
		if err != nil {
			return
		}
		item := &model.SystemUser{UID: uid, NickName: user.NickName, Avatar: user.Avatar, DepartmentName: user.DepartmentName, LastName: user.LastName, UseKind: user.UseKind, LoginID: user.LoginID}
		res = append(res, item)
	}
	return
}

// 获取用户信息通过Cookie
func (s *Service) AddV(ctx context.Context, userID string, departmentName string, name string) (err error) {
	detail, err := s.dao.GetWXUserInfo(ctx, userID)
	if err != nil {
		log.Errorc(ctx, "AddV s.dao.GetWXUserInfo(ctx, %v) Err err(%+v)", userID, err)
		return
	}

	data := &model.User{
		Avatar:         detail.Avatar,
		DepartmentName: departmentName,
		LastName:       name,
		LoginID:        "",
		NickName:       detail.Alias,
		WorkCode:       userID,
		UseKind:        "外包",
	}

	return s.dao.SetExtraUsersInfo(ctx, data)
}

// 投票
func (s *Service) Vote(ctx context.Context, aid int64, sessionToken string, content string) (err error) {
	actSubject, err := s.dao.GetActivityInfo(ctx, aid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	if actSubject.ID <= 0 {
		err = ecode.SystemNoActivityErr
		return
	}

	if actSubject.Type != model.SystemActivityTypeVote {
		err = ecode.SystemNoActivityErr
		return
	}

	ts := time.Now().Unix()
	// 未开始
	if ts < actSubject.Stime.Time().Unix() {
		err = ecode.SystemActivityNotStartErr
		return
	}
	// 已结束
	if ts > actSubject.Etime.Time().Unix() {
		err = ecode.SystemActivityIsEndErr
		return
	}

	// 获取用户信息
	var uid string
	if uid, err = s.dao.GetUserIDWithCookie(ctx, sessionToken); err != nil {
		return
	}

	// 查询是否投票过
	var record *model.ActivityVote
	record, err = s.dao.ActivityVoted(ctx, aid, uid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}
	if record.ID > 0 {
		err = ecode.SystemActivityVotedErr
		return
	}

	// 投票
	if err = s.DoSystemVote(ctx, aid, uid, actSubject.Config, content); err != nil {
		return
	}

	return
}

func (s *Service) DoSystemVote(ctx context.Context, aid int64, uid string, config string, content string) (err error) {
	// 解析配置文件
	configData := new(model.SystemActivityVoteConfig)
	if err = json.Unmarshal([]byte(config), configData); err != nil {
		log.Errorc(ctx, "DoSystemVote json.Unmarshal([]byte(%+v), %+v) Err err(%+v)", config, configData, err)
		err = ecode.SystemActivityConfigErr
		return
	}
	// 解析用户投票的数据
	voteData := make([]*model.VoteEachItem, 0)
	if err = json.Unmarshal([]byte(content), &voteData); err != nil {
		log.Errorc(ctx, "DoSystemVote json.Unmarshal([]byte(%+v), %+v) Err err(%+v)", content, voteData, err)
		err = ecode.SystemActivityParamsErr
		return
	}
	// 对题目长度校验
	if len(configData.Items) != len(voteData) {
		log.Errorc(ctx, "DoSystemVote len(configData.Items) != len(voteData.Items) Err configData(%+v) voteData(%+v)", configData, voteData)
		err = ecode.SystemActivityVoteErr
		return
	}

	rows := make([]*model.ActivityVote, 0)

	// 对每道题目允许最大选项做校验
	for k, voteItem := range voteData {
		configItem := configData.Items[k]
		// 如果是投票类型的item
		if configItem.Type == model.SystemActivityVoteSelect {
			options := strings.Split(voteItem.Options, ",")
			// 限制数不等
			if int64(len(options)) > configItem.Options.LimitNum {
				log.Errorc(ctx, "DoSystemVote len(%+v) != configItem.Options.LimitNum Err", len(options), configItem.Options.LimitNum)
				err = ecode.SystemActivityVoteErr
				return
			}
			if voteItem.Score > configItem.Options.Score {
				log.Errorc(ctx, "DoSystemVote voteItem.Score > configItem.Options.Score Err voteItem.Score(%+v) configItem.Options.Score(%+v)", voteItem.Score, voteItem.Score)
				err = ecode.SystemActivityVoteErr
				return
			}

			// 解析每道题数据存放db
			for _, v := range options {
				row := new(model.ActivityVote)
				row.AID = aid
				row.UID = uid
				row.ItemID = int64(k)
				row.OptionID, _ = strconv.ParseInt(v, 10, 64)
				row.Score = voteItem.Score

				rows = append(rows, row)
			}
		}
	}

	return s.dao.InsertVote(ctx, rows)
}

// 获奖用户通知
func (s *Service) Notify(ctx context.Context, uids string, message string, from string) (err error) {
	token, err := s.dao.GetWXAccessToken(ctx, from)
	if err != nil {
		return
	}
	uids = strings.Join(strings.Split(uids, ","), "|")
	return s.dao.SendMessage(ctx, token, uids, message)
}

// 提问
func (s *Service) Question(ctx context.Context, aid int64, sessionToken string, content string) (err error) {
	defer func() {
		log.Errorc(ctx, "Question err(%+v)", err)
	}()

	actSubject, err := s.dao.GetActivityInfo(ctx, aid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	if actSubject.ID <= 0 {
		err = ecode.SystemNoActivityErr
		return
	}

	if actSubject.Type != model.SystemActivityTypeQuestion {
		err = ecode.SystemNoActivityErr
		return
	}

	ts := time.Now().Unix()
	// 未开始
	if ts < actSubject.Stime.Time().Unix() {
		err = ecode.SystemActivityNotStartErr
		return
	}
	// 已结束
	if ts > actSubject.Etime.Time().Unix() {
		err = ecode.SystemActivityIsEndErr
		return
	}

	// 获取用户信息
	var uid string
	if uid, err = s.dao.GetUserIDWithCookie(ctx, sessionToken); err != nil {
		return
	}

	// 开始处理问卷
	question := make(map[int64]string, 0)
	if err = json.Unmarshal([]byte(content), &question); err != nil {
		err = errors.Wrap(err, "json.Unmarshal err")
		return
	}

	var questions []model.QuestionEachItem
	if len(question) > 0 {
		for k, v := range question {
			questions = append(questions, model.QuestionEachItem{QID: k, Question: v})
		}
	}

	if len(questions) > 0 {
		return s.dao.InsertQuestion(ctx, aid, uid, questions)
	}

	return
}

// 提问
func (s *Service) QuestionList(ctx context.Context, aid int64, sessionToken string) (res map[int64][]*model.ActivitySystemQuestionList, err error) {
	res = make(map[int64][]*model.ActivitySystemQuestionList, 0)

	defer func() {
		log.Errorc(ctx, "QuestionList err(%+v)", err)
	}()

	actSubject, err := s.dao.GetActivityInfo(ctx, aid)
	if err != nil {
		err = ecode.SystemNetWorkBuzyErr
		return
	}

	if actSubject.ID <= 0 {
		err = ecode.SystemNoActivityErr
		return
	}

	if actSubject.Type != model.SystemActivityTypeQuestion {
		err = ecode.SystemNoActivityErr
		return
	}

	config := &model.SystemQuestionConfig{}
	if err = json.Unmarshal([]byte(actSubject.Config), config); err != nil {
		err = errors.Wrap(err, " json.Unmarshal([]byte(actSubject.Config) err")
		return
	}

	var filterSwitch int64
	filterSwitch = config.FilterSwitch

	// 获取用户信息
	var uid string
	if uid, err = s.dao.GetUserIDWithCookie(ctx, sessionToken); err != nil {
		return
	}

	// 查询问卷信息
	data := make([]*model.ActivitySystemQuestion, 0)
	if data, err = s.dao.GetQuestionList(ctx, aid); err != nil {
		err = errors.Wrap(err, " s.dao.GetQuestionList")
		return
	}

	for _, v := range data {
		var isSelf int64
		if uid == v.UID {
			isSelf = 1
		}
		if filterSwitch == 1 {
			if isSelf != 1 {
				continue
			}
		}
		item := &model.ActivitySystemQuestionList{*v, isSelf}
		if _, ok := res[v.QID]; ok {
			res[v.QID] = append(res[v.QID], item)
		} else {
			res[v.QID] = []*model.ActivitySystemQuestionList{item}
		}
	}

	return
}
