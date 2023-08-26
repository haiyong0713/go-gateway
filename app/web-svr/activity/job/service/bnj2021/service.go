package bnj2021

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"go-common/library/queue/databus"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/tool"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/model/bnj"
	innerTool "go-gateway/app/web-svr/activity/job/tool"

	"github.com/fsnotify/fsnotify"
)

const (
	filenameOfBnjReward         = "%v/bnj2021_reward.json"
	filenameOfDefaultReward     = "%v/bnj2021_default_reward.json"
	filenameOfReserveRewardRule = "%v/bnj2021_reserve_reward_rule.json"
	filenameOfLiveDurationRule  = "%v/bnj2021_live_duration_rule.json"
	filenameOfReserveLiveAward  = "%v/bnj2021_award_live.json"
	filenameOfExamStatsRule     = "%v/exam_stats_rule.json"
	filenameOfBizLimitRule      = "%v/biz_limit_rule.json"

	drawType4LiveAR       = 2
	drawType4Reserve      = 3
	drawType4LiveDuration = 2

	businessOfPayReward = "bnj2021Live"

	metricKey4BnjDrawLottery = "bnj2021_draw_lottery"

	srvName = "activity_job_bnj2021"
)

var (
	ARRewardConsumerCfg  *databus.Config
	BnjRewardCfg         *bnj.RewardConfig
	BnjDefaultReward     []*api.RewardsSendAwardReply
	BnjAward4ReserveLive *api.RewardsSendAwardReply

	BizLimitRule map[string]int64
)

func init() {
	BizLimitRule = make(map[string]int64, 0)
	BnjRewardCfg = new(bnj.RewardConfig)
	BnjDefaultReward = make([]*api.RewardsSendAwardReply, 0)

	BnjAward4ReserveLive = new(api.RewardsSendAwardReply)
	{
		BnjAward4ReserveLive.AwardId = 128
		BnjAward4ReserveLive.Name = "“不问天”装扮 普通套装（3天）"
		BnjAward4ReserveLive.ActivityId = 3
		BnjAward4ReserveLive.ActivityName = "2021拜年纪活动"
		BnjAward4ReserveLive.Type = "GarbSuit"
		BnjAward4ReserveLive.Icon = "https://i0.hdslb.com/bfs/activity-plat/static/8a3e1fa14e30dc3be9c5324f604e5991/1gO8BmIZN_w500_h500.png"
		BnjAward4ReserveLive.ExtraInfo = make(map[string]string, 1)
	}
}

func RegisterFileWatcher() {
	list := []string{
		filenameOfReserveLiveAward,
		filenameOfBnjReward,
		filenameOfDefaultReward,
		filenameOfReserveRewardRule,
		filenameOfLiveDurationRule,
		filenameOfExamStatsRule,
		filenameOfBizLimitRule,
	}
	for _, v := range list {
		if d := genWatchedFilename(v); d != "" {
			_ = tool.RegisterWatchHandlerV1(d, UpdateConfigurationByFilename)
		}
	}
}

func genWatchedFilename(old string) (new string) {
	if dir := os.Getenv("CONF_PATH"); dir != "" {
		new = fmt.Sprintf(old, dir)
	}

	return
}

func UpdateConfigurationByFilename(ctx context.Context, event fsnotify.Event) {
	if event.Op.String() != fsnotify.Write.String() {
		return
	}

	switch event.Name {
	case genWatchedFilename(filenameOfReserveLiveAward):
		_ = UpdateBnjReserveLiveAwardCfg()
	case genWatchedFilename(filenameOfBnjReward):
		_ = UpdateBnjRewardCfg()
	case genWatchedFilename(filenameOfDefaultReward):
		_ = UpdateBnjDefaultReward()
	case genWatchedFilename(filenameOfReserveRewardRule):
		_ = UpdateReserveRewardRule()
	case genWatchedFilename(filenameOfLiveDurationRule):
		_ = UpdateLiveDurationRule()
	case genWatchedFilename(filenameOfExamStatsRule):
		_ = UpdateExamStatRule()
	case genWatchedFilename(filenameOfBizLimitRule):
		_ = UpdateBizLimitRule()
	}
}

func UpdateReserveRewardRule() (err error) {
	filename := genWatchedFilename(filenameOfReserveRewardRule)
	if filename == "" {
		return
	}

	var (
		cfg map[int64]*bnj.ReserveRewardRuleFor2021
		bs  []byte
	)
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, &cfg)
	if err == nil {
		reserveRewardRuleM = cfg
	}

	return
}

func UpdateBizLimitRule() (err error) {
	filename := genWatchedFilename(filenameOfBizLimitRule)
	if filename == "" {
		return
	}

	var bs []byte
	cfg := make(map[string]int64, 0)
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, &cfg)
	if err == nil {
		resetLimiters(cfg)
		BizLimitRule = cfg
	}

	return
}

func UpdateExamStatRule() (err error) {
	filename := genWatchedFilename(filenameOfExamStatsRule)
	if filename == "" {
		return
	}

	var (
		cfg *bnj.ExamStatsRule
		bs  []byte
	)
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, &cfg)
	if err == nil {
		examStatsRule = cfg
	}

	return
}

func UpdateLiveDurationRule() (err error) {
	filename := genWatchedFilename(filenameOfLiveDurationRule)
	if filename == "" {
		return
	}

	var (
		cfg map[int64]*bnj.LotteryRuleFor2021
		bs  []byte
	)
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, &cfg)
	if err == nil {
		liveDurationRuleM = cfg
	}

	return
}

func UpdateBnjDefaultReward() (err error) {
	filename := genWatchedFilename(filenameOfDefaultReward)
	if filename == "" {
		return
	}

	var (
		cfg []*api.RewardsSendAwardReply
		bs  []byte
	)
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, &cfg)
	if err == nil {
		BnjDefaultReward = cfg
	}

	return
}

func UpdateBnjReserveLiveAwardCfg() (err error) {
	filename := genWatchedFilename(filenameOfReserveLiveAward)
	if filename == "" {
		return
	}

	var (
		newCfg *api.RewardsSendAwardReply
		bs     []byte
	)
	bs, err = readByteSliceByFilename(filename)
	if err == nil {
		err = json.Unmarshal(bs, &newCfg)
		if err == nil {
			BnjAward4ReserveLive = newCfg
		}
	}

	return
}

func UpdateBnjRewardCfg() (err error) {
	filename := genWatchedFilename(filenameOfBnjReward)
	if filename == "" {
		return
	}

	var newCfg *bnj.RewardConfig
	newCfg, err = readBnjRewardByFilename(filename)
	if err == nil {
		BnjRewardCfg = newCfg
	}

	return
}

func readBnjRewardByFilename(filename string) (cfg *bnj.RewardConfig, err error) {
	cfg = new(bnj.RewardConfig)
	var bs []byte
	bs, err = readByteSliceByFilename(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(bs, cfg)

	return
}

func readByteSliceByFilename(filename string) (bs []byte, err error) {
	var f *os.File
	bs = make([]byte, 0)
	f, err = os.Open(filename)
	if err != nil {
		return
	}

	defer func() {
		_ = f.Close()
	}()

	bs, err = ioutil.ReadAll(f)

	return
}

/*
*

	activity_id: 80 // 宣发页面抽奖
	activity_id: 81 // 直播间预约抽奖
	activity_id: 82 // 直播间观看时长达标抽奖
	activity_id: 83 // 直播间游戏奖券抽奖
*/
func batchARDraw(mid, count, opType, activityID int64, defaultAward bool) (list []*api.RewardsSendAwardReply) {
	list = make([]*api.RewardsSendAwardReply, 0)
	req := new(api.Bnj2021LotteryReq)
	var (
		resp *api.Bnj2021LotteryReply
		err  error
	)
	if defaultAward {
		goto DefaultAward
	}

	{
		req.Mid = mid
		req.Count = count
		req.Type = opType
		req.Debug = true
		req.ActivityId = activityID
	}

	waitBizLimit(limitKeyOfDraw, limitKeyOfDraw)
	resp, err = client.ActivityClient.Bnj2021Lottery(context.Background(), req)
	if err == nil {
		list = resp.List
	} else {
		tool.IncrCommonBizStatus(metricKey4BnjDrawLottery, tool.StatusOfFailed)
	}

DefaultAward:
	if d := count - int64(len(list)); d > 0 && len(BnjDefaultReward) > 0 {
		for i := int64(0); i < d; i++ {
			if activityID == 81 {
				tmp := DefaultRewardDeepCopy(BnjAward4ReserveLive, mid)
				list = append(list, tmp)

				continue
			}

			rand.Seed(time.Now().Unix())
			tmp := DefaultRewardDeepCopy(BnjDefaultReward[rand.Intn(len(BnjDefaultReward))], mid)
			list = append(list, tmp)
		}
	}

	return
}

func DefaultRewardDeepCopy(reward *api.RewardsSendAwardReply, mid int64) (newOne *api.RewardsSendAwardReply) {
	newOne = new(api.RewardsSendAwardReply)
	{
		newOne.ActivityId = reward.ActivityId
		newOne.ActivityName = reward.ActivityName
		newOne.Type = reward.Type
		newOne.Name = reward.Name
		newOne.Icon = reward.Icon
		newOne.AwardId = reward.AwardId
		newOne.Mid = mid

		if reward.ExtraInfo == nil {
			reward.ExtraInfo = make(map[string]string, 0)
		}

		extra := make(map[string]string, 0)
		{
			for k, v := range reward.ExtraInfo {
				extra[k] = v
			}
		}
		newOne.ExtraInfo = extra
	}

	return
}

func PayDrawReward(mid, rewardID int64, business, uniqueID string) (err error) {
	if rewardID == 0 {
		return
	}

	req := new(api.RewardsSendAwardReq)
	{
		req.Mid = mid
		req.AwardId = rewardID
		req.Business = business
		req.UniqueId = uniqueID
		req.UpdateDb = true //保证消息队列丢数据下的一致性
		req.Sync = false
	}
	if _, rpcErr := client.ActivityClient.RewardsSendAward(context.Background(), req); rpcErr != nil {
		tool.IncrCommonBizStatus(bizNameOfUserLottery4GRPC, innerTool.BizStatusOfFailed)
	}

	return
}
