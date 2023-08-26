package http

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/rate/limit/quota"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/admin/client"
	"go-gateway/app/web-svr/activity/interface/api"
	"strconv"
	"time"

	"gopkg.in/go-playground/validator.v9"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/activity/admin/service/rewards"
)

var v *validator.Validate
var _fanOut *fanout.Fanout

var awardRetryQuota quota.Waiter

const waiterNameFmt = "%s.%s.main.web-svr.activity-admin|Rewards|Retry|total"

func getRewardsRetryWaiterName() string {
	return fmt.Sprintf(waiterNameFmt, env.DeployEnv, env.Zone)
}

func init() {
	quota.Init()
	v = validator.New()
	_fanOut = fanout.New("general")
	awardRetryQuota = quota.NewWaiter(&quota.WaiterConfig{
		ID: getRewardsRetryWaiterName(),
	})

}

func addInternalRewardsRouter(group *bm.RouterGroup) {
	rewardsGroup := group.Group("/rewards", authSrv.Permit2("ACTIVITY_REWARD_ADMIN"))
	{
		rewardsGroup.GET("/awards/list", RewardsListAwards)
		rewardsGroup.POST("/awards/add", RewardsAddAward)
		rewardsGroup.POST("/awards/addj", RewardsAddAwardJson)
		rewardsGroup.POST("/awards/del", RewardsDelAward)
		rewardsGroup.POST("/awards/update", RewardsUpdateAward)
		rewardsGroup.POST("/awards/updatej", RewardsUpdateAwardJson)
		rewardsGroup.POST("/awards/cdkey", RewardsUploadCdKey)
		rewardsGroup.GET("/awards/cdkey/count", RewardsGetCdkeyCount)

		rewardsGroup.POST("/awards/retry", RewardsRetrySendAward)
		rewardsGroup.POST("/awards/send", RewardsSendAward)
		rewardsGroup.POST("/activity/add", RewardsAddActivity)
		rewardsGroup.GET("/activity/detail", RewardsGetActivityDetail)
		rewardsGroup.POST("/activity/del", RewardsDelActivity)
		rewardsGroup.GET("/activity/list", RewardsListActivity)
		rewardsGroup.POST("/activity/update", RewardsUpdateActivity)
	}
	rewardsCheckGroup := group.Group("/rewards_check")
	{
		rewardsCheckGroup.POST("/mall", RewardsCheckSentStatus)
	}
}

func RewardsListAwards(ctx *bm.Context) {
	v := &api.RewardsListAwardReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsListAward(ctx, v))
}

func RewardsAddAward(ctx *bm.Context) {
	v := &api.RewardsAddAwardReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	extraStr := ctx.Request.Form.Get("extra_info")
	extra := make(map[string]string)
	if extraStr != "" {
		if err := json.Unmarshal([]byte(extraStr), &extra); err != nil {
			log.Errorc(ctx, "json.Unmarshal(Extra: %v) failed. error(%v)", extra, err)
			err = ecode.Error(ecode.RequestErr, "extra参数有误")
			return
		}
		v.ExtraInfo = extra
	}

	ctx.JSON(client.ActivityClient.RewardsAddAward(ctx, v))
}

func RewardsAddAwardJson(ctx *bm.Context) {
	v := &api.RewardsAddAwardReq{}
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsAddAward(ctx, v))
}

func RewardsDelAward(ctx *bm.Context) {
	v := &api.RewardsDelAwardReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsDelAward(ctx, v))
}

func RewardsUpdateAward(ctx *bm.Context) {
	v := &api.RewardsAwardInfo{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	extraStr := ctx.Request.Form.Get("extra_info")
	extra := make(map[string]string)
	if extraStr != "" {
		if err := json.Unmarshal([]byte(extraStr), &extra); err != nil {
			log.Errorc(ctx, "json.Unmarshal(Extra: %v) failed. error(%v)", extra, err)
			err = ecode.Error(ecode.RequestErr, "extra参数有误")
			return
		}
		v.ExtraInfo = extra
	}

	ctx.JSON(client.ActivityClient.RewardsUpdateAward(ctx, v))
}

func RewardsUpdateAwardJson(ctx *bm.Context) {
	v := &api.RewardsAwardInfo{}
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsUpdateAward(ctx, v))
}

func RewardsAddActivity(ctx *bm.Context) {
	v := &api.RewardsAddActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsAddActivity(ctx, v))
}

func RewardsDelActivity(ctx *bm.Context) {
	v := &api.RewardsDelActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsDelActivity(ctx, v))
}

func RewardsListActivity(ctx *bm.Context) {
	v := &api.RewardsListActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsListActivity(ctx, v))
}

func RewardsUpdateActivity(ctx *bm.Context) {
	v := &api.RewardsUpdateActivityReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsUpdateActivity(ctx, v))
}

func RewardsGetActivityDetail(ctx *bm.Context) {
	v := &api.RewardsGetActivityDetailReq{}
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.ActivityClient.RewardsGetActivityDetail(ctx, v))
}

func RewardsSendAward(ctx *bm.Context) {
	v := new(struct {
		Mid     int64 `form:"mid" validate:"required"`
		AwardId int64 `form:"award_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	tsNow := time.Now().UnixNano()/1e6 - 1600000000000
	req := &api.RewardsSendAwardReq{}
	req.Mid = v.Mid
	req.AwardId = v.AwardId
	req.Sync = true
	req.UpdateCache = true
	req.Business = "adminManual"
	req.UniqueId = fmt.Sprintf("admin-%v", tsNow)
	ctx.JSON(client.ActivityClient.RewardsSendAward(ctx, req))
}

func RewardsUploadCdKey(c *bm.Context) {
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	awardIdStr := c.Request.Form.Get("award_id")
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "csv文件解析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件解析失败"))
		return
	}
	awardId, err := strconv.ParseInt(awardIdStr, 10, 64)
	if err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "award_id 不合法"))
		return
	}

	//reader := csv.NewReader(bom.NewReader(file))
	reader := csv.NewReader(file)
	keys := make([]string, 0)
	records, err := reader.ReadAll()
	if err != nil {
		log.Errorc(c, "csv文件读取析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件读取失败"))
		return
	}
	for _, line := range records {
		if len(line) <= 0 {
			continue
		}
		keys = append(keys, line[0])
	}
	go rewards.Client.UploadCdKey(context.Background(), userName, awardId, keys)
	c.JSON(nil, err)
}

func RewardsRetrySendAward(c *bm.Context) {
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "csv文件解析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件解析失败"))
		return
	}
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Errorc(c, "csv文件读取析失败， error(%v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "csv文件读取失败"))
		return
	}
	err = _fanOut.Do(c, func(ctx context.Context) {
		_ = rewardsRetrySendAwardByList(ctx, records)
	})
	c.JSON("", err)

}

func rewardsRetrySendAwardByList(ctx context.Context, records [][]string) (err error) {
	log.Errorc(ctx, "rewardsRetrySendAwardByList start")
	defer func() {
		log.Errorc(ctx, "rewardsRetrySendAwardByList finish")
		if err != nil {
			log.Errorc(ctx, "rewardsRetrySendAwardByList error: %v", err)
		}
	}()
	for idx, line := range records {
		if len(line) <= 0 {
			continue
		}
		mid := int64(0)
		aid := int64(0)
		mid, err = strconv.ParseInt(line[0], 10, 0)
		if err != nil {
			return err
		}
		aid, err = strconv.ParseInt(line[1], 10, 0)
		if err != nil {
			return err
		}
		if mid == 0 || aid == 0 || line[2] == "" || line[3] == "" {
			err = fmt.Errorf(fmt.Sprintf("line format error for line %v content %v", idx, line))
			return
		}
		awardRetryQuota.Wait()
		_, err = client.ActivityClient.RetryRewardsSendAward(ctx, &api.RetryRewardsSendAwardReq{
			Mid:      mid,
			UniqueId: line[2],
			Business: line[3],
			AwardId:  aid,
		})
		if err != nil {
			err = fmt.Errorf("send %v error: %v", line, err)
			return err
		}
	}
	return
}

type VipMallRewardsCheckSentStatusParam struct {
	AssetRequest *VipMallRewardsCheckSentStatusAssetRequest `json:"assetRequest" validate:"required"`
}

func (p *VipMallRewardsCheckSentStatusParam) GetMidAndUniqueId() (mid int64, uniqueId string, err error) {
	if p.AssetRequest == nil {
		err = ecode.RequestErr
		return
	}
	mid, err = p.AssetRequest.GetMid()
	if err != nil {
		return
	}
	uniqueId, err = p.AssetRequest.GetUniqueId()
	return
}

type VipMallRewardsCheckSentStatusAssetRequest struct {
	ReferenceId string `json:"referenceId"`
	SourceBizId string `json:"sourceBizId"`
	Mid         int64  `json:"mid"`
	Uid         string `json:"uid"`
}

func (r *VipMallRewardsCheckSentStatusAssetRequest) GetUniqueId() (uniqueId string, err error) {
	uniqueId = r.ReferenceId
	if uniqueId == "" {
		uniqueId = r.SourceBizId
	}
	if uniqueId == "" {
		err = ecode.RequestErr
		return
	}
	return
}

func (r *VipMallRewardsCheckSentStatusAssetRequest) GetMid() (mid int64, err error) {
	mid = r.Mid
	if mid == 0 {
		mid, err = strconv.ParseInt(r.Uid, 10, 64)
	}
	if mid == 0 {
		err = ecode.RequestErr
		return
	}
	return
}

func RewardsCheckSentStatus(ctx *bm.Context) {
	v := &VipMallRewardsCheckSentStatusParam{}
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	mid, uniqueId, err := v.GetMidAndUniqueId()
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	res, err := client.ActivityClient.RewardsCheckSentStatus(ctx, &api.RewardsCheckSentStatusReq{
		Mid:      mid,
		UniqueId: uniqueId,
	})
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	data := make(map[string]interface{})
	data["result"] = res.Result
	ctx.JSONMap(data, err)
	return
}

func RewardsGetCdkeyCount(c *bm.Context) {
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	awardIdStr := c.Request.Form.Get("award_id")
	awardId, err := strconv.ParseInt(awardIdStr, 10, 64)
	if err != nil {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "award_id 不合法"))
		return
	}
	c.JSON(rewards.Client.GetCdkeyCount(context.Background(), awardId))
}
