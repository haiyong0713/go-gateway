package wishes_2021_spring

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sync"
	"time"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/activity/interface/client"

	"go-common/library/ecode"

	innerEcode "go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/component"
	dao "go-gateway/app/web-svr/activity/interface/dao/wishes_2021_spring"
	model "go-gateway/app/web-svr/activity/interface/model/wishes_2021_spring"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const (
	cacheKey4UserCommit = "activity:common:user_commit:mq:%v"

	maxByte4UserContent = 1024 * 1024

	sceneID4UserCommitWish       = int64(1)
	sceneID4UserCommitManuScript = int64(2)
)

var (
	innerActivityMap map[string]*model.CommonActivityConfig
)

func init() {
	innerActivityMap = make(map[string]*model.CommonActivityConfig, 0)
}

func InitManuScriptActivityMap(m map[string]*model.CommonActivityConfig) {
	innerActivityMap = m
}

func FetchUserCommitContentInLive(ctx context.Context, req *model.UserCommitListRequestInLive) (
	resp *model.UserCommitListRespInLive, err error) {
	list := make([]map[string]interface{}, 0)
	resp = new(model.UserCommitListRespInLive)
	{
		resp.Ps = 20
		resp.Data = list
	}
	if !req.Validate() {
		err = ecode.RequestErr

		return
	}

	activityCfg, ok := innerActivityMap[req.ActivityUniqID]
	if !ok {
		err = ecode.RequestErr

		return
	}

	req.ActivityID = activityCfg.ActivityID
	list, err = dao.FetchUserCommitContentListInLive(ctx, req)
	if err == nil {
		midMap := genMidMapByUserCommitList(list)
		if len(midMap) > 0 {
			userInfoMap := new(sync.Map)
			userInfoMap, err = fetchUserProfileMapByMidList(ctx, midMap)
			if err == nil {
				list = rebuildUserCommitListByUserProfileMap(list, userInfoMap)
				resp.Data = list
				resp.Total, err = dao.FetchUserCommitContentCountInLive(ctx, req)
			}
		}
	}

	return
}

func rebuildUserCommitListByUserProfileMap(list []map[string]interface{}, m *sync.Map) []map[string]interface{} {
	for _, v := range list {
		if d, ok := v["mid"].(int64); ok {
			if userInfo, ok := m.Load(d); ok {
				v["user_info"] = userInfo
			}
		}
	}

	return list
}

func genMidMapByUserCommitList(list []map[string]interface{}) (m map[int64]int64) {
	m = make(map[int64]int64, 0)
	for _, v := range list {
		if d, ok := v["mid"].(int64); ok {
			m[d] = 1
		}
	}

	return
}

func fetchUserProfileMapByMidList(ctx context.Context, midMap map[int64]int64) (m *sync.Map, err error) {
	m = new(sync.Map)
	eg := errgroup.WithContext(ctx)
	for k := range midMap {
		tmpMid := k
		eg.Go(func(ctx context.Context) error {
			userProfile, err := client.AccountClient.Profile3(ctx, &accapi.MidReq{Mid: tmpMid})
			if err != nil {
				log.Errorc(ctx, "UserCommitContent4Aggregation client.AccountClient.Profile3(%v) error(%v)", tmpMid, err)
				return err
			}
			tmpInfo := new(model.UserInfo)
			{
				tmpInfo.Mid = tmpMid
				tmpInfo.Nickname = userProfile.Profile.Name
				tmpInfo.Face = userProfile.Profile.Face
				tmpInfo.Identification = userProfile.Profile.Identification == 1
				tmpInfo.Silence = userProfile.Profile.Silence == 1
				tmpInfo.TelStatus = userProfile.Profile.TelStatus == 1
			}

			m.Store(tmpMid, tmpInfo)

			return nil
		})
	}

	err = eg.Wait()

	return
}

func UserCommitContent4Aggregation(ctx context.Context, mid int64, uniqID string) (resp *model.UserCommit4AggregationWithUserInfo, err error) {
	resp = new(model.UserCommit4AggregationWithUserInfo)
	{
		resp.UserCommit4Aggregation = model.NewUserCommit4Aggregation()
		resp.UserInfo = new(model.UserInfo)
	}
	activityCfg, ok := innerActivityMap[uniqID]
	if !ok {
		err = ecode.RequestErr

		return
	}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		userProfile, err := client.AccountClient.Profile3(ctx, &accapi.MidReq{Mid: mid})
		if err != nil {
			log.Errorc(ctx, "UserCommitContent4Aggregation client.AccountClient.Profile3(%v) error(%v)", mid, err)
			return err
		}
		resp.UserInfo.Mid = mid
		resp.UserInfo.Nickname = userProfile.Profile.Name
		resp.UserInfo.Face = userProfile.Profile.Face
		resp.UserInfo.Identification = userProfile.Profile.Identification == 1
		resp.UserInfo.Silence = userProfile.Profile.Silence == 1
		resp.UserInfo.TelStatus = userProfile.Profile.TelStatus == 1
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		resp.UserCommit4Aggregation, err = dao.FetchUserCommitContent(ctx, mid, activityCfg.ActivityID)
		return
	})

	eg.Go(func(ctx context.Context) (err error) {
		resp.LastId, err = dao.CountUserManuScriptList(ctx, activityCfg.ActivityID)
		return
	})
	err = eg.Wait()
	return
}

func CommitUserManuScript(ctx context.Context, req *api.CommonActivityUserCommitReq) (
	m map[string]interface{}, err error) {
	m = make(map[string]interface{}, 0)
	{
		m["id"] = 0
	}
	var cfg *model.CommonActivityConfig
	cfg, err = CheckActivity(ctx, req, sceneID4UserCommitManuScript)
	if err != nil {
		return
	}

	var times int64
	times, err = dao.FetchUserCommitManuScriptTimesFromDB(ctx, req.MID, req.ActivityID)
	if err == nil {
		if times >= cfg.MaxUploadTimes {
			err = innerEcode.BwsOnlineTimeUsed
		} else {
			m["id"], err = dao.InsertUserCommitManuScript(ctx, req, true)
		}
	}

	return
}

func CheckActivity(ctx context.Context, req *api.CommonActivityUserCommitReq, sceneID int64) (
	cfg *model.CommonActivityConfig, err error) {
	var ok bool
	cfg, ok = innerActivityMap[req.UniqID]
	if !ok {
		err = ecode.RequestErr

		return
	}

	now := time.Now().Unix()
	if now < cfg.StartTime || now >= cfg.EndTime {
		err = ecode.RequestErr

		return
	}

	req.ActivityID = cfg.ActivityID
	if ok := tool.IsStringOverSizeLimit(req.Content, maxByte4UserContent); ok {
		err = ecode.RequestErr

		return
	}

	err = validateUserCommitContent(req.Content)
	if err != nil {
		return
	}

	req.SceneID = sceneID
	var maxCommitTimes, commitTimes4Now int64
	switch sceneID {
	case sceneID4UserCommitWish:
		// TODO: can insert or update
	case sceneID4UserCommitManuScript:
		commitTimes4Now, err = dao.FetchUserCommitManuScriptTimes(ctx, req.MID, req.ActivityID)
		if err != nil {
			break
		}

		maxCommitTimes = cfg.MaxUploadTimes
	}

	if err != nil {
		return
	}

	switch sceneID {
	case sceneID4UserCommitWish:
		// TODO:
	case sceneID4UserCommitManuScript:
		if commitTimes4Now >= maxCommitTimes {
			err = innerEcode.BwsOnlineTimeUsed
		}
	}

	return
}

func validateUserCommitContent(content string) (err error) {
	m := make(map[string]interface{}, 0)
	err = json.Unmarshal([]byte(content), &m)

	return
}

func CommitUserContent(ctx context.Context, req *api.CommonActivityUserCommitReq) (err error) {
	if _, err = CheckActivity(ctx, req, sceneID4UserCommitWish); err != nil {
		return
	}

	err = pubCommitInfoIntoMQ(ctx, req)

	return
}

func pubCommitInfoIntoMQ(ctx context.Context, req *api.CommonActivityUserCommitReq) (err error) {
	bs, _ := json.Marshal(req)
	_, err = component.BackUpMQ.Do(ctx, "LPUSH", genCacheKey4UserCommit(req.MID), string(bs))

	return
}

func InsertUserCommit2DB(ctx context.Context, req *api.CommonActivityUserCommitReq) (err error) {
	activityCfg, ok := innerActivityMap[req.UniqID]
	if !ok {
		err = ecode.RequestErr

		return
	}

	switch req.SceneID {
	case sceneID4UserCommitWish:
		var commitID int64
		commitID, err = dao.FetchUserCommitContentIDFromDB(ctx, req.MID, req.ActivityID)
		if err != nil {
			break
		}

		if commitID > 0 {
			err = dao.UpdateUserCommitContent(ctx, commitID, req, true)
		} else {
			err = dao.InsertUserCommitLogAndContent(ctx, req, true)
		}
	case sceneID4UserCommitManuScript:
		var times int64
		times, err = dao.FetchUserCommitManuScriptTimesFromDB(ctx, req.MID, req.ActivityID)
		if err == nil {
			if times >= activityCfg.MaxUploadTimes {
				err = innerEcode.BwsOnlineTimeUsed
			} else {
				_, err = dao.InsertUserCommitManuScript(ctx, req, true)
			}
		}
	}

	return
}

func genCacheKey4UserCommit(mid int64) (key string) {
	key = fmt.Sprintf(cacheKey4UserCommit, mid%10)

	return
}

func ActAuditMertialPub(ctx context.Context, req *api.CommonActivityAuditPubReq) (err error) {
	log.Infoc(ctx, "ActAuditMertialPub TableName:%v , ActionType:%v , RawMessage:%s", req.TableName, req.ActionType, req.RawMessage)
	userCommitInfo := new(model.UserCommitManuscriptDB)
	if err = json.Unmarshal(req.RawMessage, userCommitInfo); err != nil || userCommitInfo == nil {
		return errors.Wrap(err, "RawMessage not type of  userCommitInfo")
	}
	var activityCfg *model.CommonActivityConfig
	for _, v := range innerActivityMap {
		if v.ActivityID > 0 && v.ActivityID == userCommitInfo.ActivityId {
			activityCfg = v
			break
		}
	}
	if activityCfg == nil || len(activityCfg.SendPropertys) == 0 {
		log.Infoc(ctx, "ActAuditMertialPub empty sync config:%+v ", activityCfg)
		return ecode.RequestErr
	}

	if len(userCommitInfo.Content) > 0 {
		record := model.AuditInfo{}
		var contentBytes []byte
		if contentBytes, err = base64.StdEncoding.DecodeString(userCommitInfo.Content); err != nil {
			log.Warnc(ctx, "decode user commit content base64, value:%s ,  err:%+v", userCommitInfo.Content, err)
			contentBytes = []byte(userCommitInfo.Content)
		}
		if err = json.Unmarshal(contentBytes, &record); err != nil {
			return errors.Wrap(err, "userCommitInfo Content err")
		}
		record.CTimeStr = userCommitInfo.CTimeStr
		record.MtimeStr = userCommitInfo.MtimeStr
		newRecord := new(model.AuditInfo)
		if err = tool.SimpleCopyProperties(newRecord, record, activityCfg.SendPropertys, "json"); err != nil {
			log.Warnc(ctx, "SimpleCopyProperties err:%+v", err)
		}
		err = dao.ManuScriptAuditPub(ctx, newRecord)
	}
	return
}
