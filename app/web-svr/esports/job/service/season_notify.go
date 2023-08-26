package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/tool"
)

const (
	seasonStatusOfNotStart       = "season_not_start"
	seasonStatusOfEnd            = "season_end"
	seasonStatusOfMatchInPlay    = "match_in_play"
	seasonStatusOfMatchOutOfPlay = "match_out_of_play"

	seasonStatusDisplayOfNotStart       = "赛季未开始"
	seasonStatusDisplayOfEnd            = "赛季已结束"
	seasonStatusDisplayOfMatchInPlay    = "比赛进行中"
	seasonStatusDisplayOfMatchOutOfPlay = "比赛未在进行中"

	seasonStatusAlarmMsgTemplate = `赛季状态变更提示：<font color=\"info\">%v</font>，请相关同事注意。\n
>当前检查点状态:<font color=\"info\">%v</font> \n
>上一个检查点状态:<font color=\"info\">%v</font> \n
>上一次通知时间:<font color=\"info\">%v</font> \n
>已通知次数:<font color=\"warning\">%v</font> \n
>%v`
)

type SeasonNotifyStatus struct {
	seasonID       int64
	status         string
	notifyTimes    int
	lastNotifyTime time.Time
	lastStatusTime time.Time
}

var (
	seasonNotifyStatusM sync.Map
)

func SeasonNotifyStatusM(ctx *bm.Context) {
	m := make(map[string]interface{}, 0)

	f := func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if d, ok := value.(SeasonNotifyStatus); ok {
				subM := make(map[string]interface{}, 0)
				{
					subM["seasonID"] = d.seasonID
					subM["status"] = d.status
					subM["notifyTimes"] = d.notifyTimes
					subM["lastNotifyTime"] = d.lastNotifyTime.Format("2006-01-02 15:04:05")
					subM["lastStatusTime"] = d.lastStatusTime.Format("2006-01-02 15:04:05")
				}

				m[keyStr] = subM
			}
		}

		return true
	}
	seasonNotifyStatusM.Range(f)

	ctx.JSON(m, nil)
}

func (s *Service) WatchSeasonStatus(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			wg := new(sync.WaitGroup)

			notifies := conf.SeasonNotifies.Load().(map[string]conf.SeasonNotify)
			for k, notify := range notifies {
				wg.Add(1)
				go func(seasonID string, seasonNotify conf.SeasonNotify) {
					defer func() {
						wg.Done()
					}()

					if !isNeedWatch(seasonNotify) {
						seasonNotifyStatusM.Delete(seasonID)
						return
					}

					if d, err := strconv.ParseInt(seasonID, 10, 64); err == nil {
						if season, err := s.dao.InPlaySeasonByID(ctx, d); err == nil {
							status := s.seasonStatus(ctx, season)
							updateSeasonNotifyStatusM(seasonNotify, season, status)
						}
					}
				}(k, notify)
			}

			wg.Wait()
		}
	}
}

func isNeedWatch(seasonNotify conf.SeasonNotify) bool {
	if seasonNotify.StartTime == "" || seasonNotify.EndTime == "" {
		return false
	}

	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", seasonNotify.StartTime, time.Local)
	if err != nil {
		return false
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", seasonNotify.EndTime, time.Local)
	if err != nil {
		return false
	}

	now := time.Now()
	if now.Before(startTime) || now.After(endTime) {
		return false
	}

	return true
}

func updateSeasonNotifyStatusM(seasonNotify conf.SeasonNotify, season model.Season, status string) {
	seasonID := strconv.FormatInt(season.ID, 10)
	seasonNotifyStatus := SeasonNotifyStatus{}
	oldStatus := seasonStatusOfNotStart
	lastStatusTime := time.Now()
	notifyTimes := 0

	if d, ok := seasonNotifyStatusM.Load(seasonID); ok {
		if tmpSeasonNotifyStatus, ok := d.(SeasonNotifyStatus); ok {
			seasonNotifyStatus = tmpSeasonNotifyStatus
			lastStatusTime = seasonNotifyStatus.lastStatusTime
			now := time.Now()
			seasonNotifyStatus.lastStatusTime = now
			oldStatus = seasonNotifyStatus.status
			if seasonNotifyStatus.status != status {
				seasonNotifyStatus.status = status
				seasonNotifyStatus.notifyTimes = 0
			}

			notifyTimes = seasonNotifyStatus.notifyTimes
		}
	} else {
		{
			seasonNotifyStatus.status = status
			seasonNotifyStatus.lastStatusTime = time.Now()
			seasonNotifyStatus.seasonID = season.ID
			seasonNotifyStatus.notifyTimes = 0
		}
	}

	if seasonNotifyStatus.seasonID > 0 {
		if canNotify(seasonNotifyStatus, seasonNotify) {
			seasonNotifyStatus.lastNotifyTime = time.Now()
			_ = notify(seasonNotify, oldStatus, seasonNotifyStatus.status, lastStatusTime, notifyTimes)
			seasonNotifyStatus.notifyTimes++
		}

		seasonNotifyStatusM.Store(seasonID, seasonNotifyStatus)
	}
}

func canNotify(seasonNotifyStatus SeasonNotifyStatus, seasonNotify conf.SeasonNotify) bool {
	return seasonNotifyStatus.notifyTimes < seasonNotify.NotifyTimes ||
		(seasonNotify.NotifyInterval > 0 &&
			time.Now().Unix()-seasonNotifyStatus.lastStatusTime.Unix() >= seasonNotify.NotifyInterval)
}

func notify(seasonNotify conf.SeasonNotify, oldStatus, newStatus string, lastStatusTime time.Time, notifyTimes int) (err error) {
	m := make(map[string]interface{}, 0)
	{
		m["biz"] = seasonNotify.UniqID
		m["status"] = newStatus
	}
	bs, _ := json.Marshal(m)
	for _, url := range seasonNotify.HttpNotifies {
		if resp, reqErr := http.Post(url, tool.ContentTypeOfJson, strings.NewReader(string(bs))); reqErr != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("SeasonStatus >>> notify url failed", url, resp, reqErr)
			errMsg := fmt.Sprintf("SeasonStatus >>> notify url(%v) failed, now(%v)", url, time.Now())
			err = errors.New(errMsg)
		}
	}

	if seasonNotify.WebhookNotify && seasonNotify.WebhookUrl != "" &&
		len(seasonNotify.WebhookReceivers) > 0 {
		robot := tool.CorpWeChat{
			MentionUserIDs:  seasonNotify.WebhookReceivers,
			MentionUserTels: seasonNotify.WebhookTels,
			WebhookUrl:      seasonNotify.WebhookUrl,
		}

		var lastStatusTimeOfNew interface{}
		lastStatusTimeOfNew = lastStatusTime.Format("2006-01-02 15:04:05")
		if oldStatus == seasonStatusOfNotStart {
			lastStatusTimeOfNew = "-"
		}

		if err == nil {
			notifyTimes++
		}

		content := fmt.Sprintf(
			seasonStatusAlarmMsgTemplate,
			seasonNotify.UniqID,
			genSeasonStatusDisplay(newStatus),
			genSeasonStatusDisplay(oldStatus),
			lastStatusTimeOfNew,
			notifyTimes,
			tool.MentionUserIDs(robot, tool.AlarmMsgTypeOfMarkdown))
		if newBs, err := tool.GenAlarmMsgDataByTypeByRobot(robot, tool.AlarmMsgTypeOfMarkdown, content); err == nil {
			_ = tool.SendCorpWeChatRobotAlarmByRobot(robot, newBs)
		}
	}

	return
}

func genSeasonStatusDisplay(status string) (display string) {
	switch status {
	case seasonStatusOfNotStart:
		display = seasonStatusDisplayOfNotStart
	case seasonStatusOfEnd:
		display = seasonStatusDisplayOfEnd
	case seasonStatusOfMatchInPlay:
		display = seasonStatusDisplayOfMatchInPlay
	case seasonStatusOfMatchOutOfPlay:
		display = seasonStatusDisplayOfMatchOutOfPlay
	}

	return
}

func (s *Service) seasonStatus(ctx context.Context, season model.Season) (status string) {
	unixNow := time.Now().Unix()
	if unixNow < season.Stime {
		status = seasonStatusOfNotStart
	} else if unixNow >= season.Etime {
		status = seasonStatusOfEnd
	} else {
		status = seasonStatusOfMatchOutOfPlay
		if s.dao.HasInPlayContest(ctx, season.ID) {
			status = seasonStatusOfMatchInPlay
		}
	}

	return
}
