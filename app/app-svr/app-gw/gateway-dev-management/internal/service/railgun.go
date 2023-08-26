package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

var (
	schedulerURL = "http://hawkeye.bilibili.co/buzzer/scheduler_day"
	sendMsg      = "本周接口人oncall: %v\n" +
		"本周SRE oncall: %v\n" +
		"本群专用于研发同学，反馈需要了解的网关具体逻辑实现的问题，对接各类咨询与需求以及线上SRE相关事宜，我们oncall会响应并根据情况安排跟进。\n" +
		"紧急联系方式：zhoujiahui\n" +
		"备用联系方式：zhangxin, xialinjuan\n" +
		"终极联系方式：liuguodong\n"
)

func (s *Service) initBG() {
	cronConfig, _ := s.ac.Get("cronConfig").String()
	s.sendScheduleTask = railgun.NewRailGun("pushSchedule", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cronConfig}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			if err := s.PushScheduleMessage(); err != nil {
				return railgun.MsgPolicyFailure
			}
			return railgun.MsgPolicyNormal
		}))
	s.sendScheduleTask.Start()
}

func (s *Service) PushScheduleMessage() error {
	name, metions, err := s.GetOnCall()
	if err != nil {
		return err
	}
	metions = append(metions, "@all")
	data := &model.PushScheduleMessage{
		MsgType: "text",
		Text: struct {
			Content     string   `json:"content"`
			MentionList []string `json:"mentioned_list"`
		}{
			Content:     fmt.Sprintf(sendMsg, name[0], name[1]),
			MentionList: metions,
		},
	}
	botURL, err := s.ac.Get("scheduleMessageBot").String()
	if err != nil {
		return err
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	if _, err = httpPost(botURL, data, headers); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetSchedule(ctx context.Context, key string) ([]*model.Stuff, error) {
	scheduleJson, err := s.dao.SelectValueByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	var stuff []*model.Stuff
	if err := json.Unmarshal([]byte(scheduleJson), &stuff); err != nil {
		return nil, err
	}
	return stuff, nil
}

func (s *Service) GetArrange(ctx context.Context, key string) (int, error) {
	arrangeJson, err := s.dao.SelectValueByKey(ctx, key)
	if err != nil {
		return -1, err
	}
	var index int
	if err := json.Unmarshal([]byte(arrangeJson), &index); err != nil {
		return -1, err
	}
	return index, nil
}

func (s *Service) UpdateAndGetArrange(ctx context.Context, name string) (*model.Stuff, error) {
	curNum, err := s.GetArrange(ctx, fmt.Sprintf("%s_arrange", name))
	if err != nil {
		return nil, err
	}
	sreSchedule, err := s.GetSchedule(ctx, fmt.Sprintf("%s_schedule", name))
	if err != nil {
		return nil, err
	}
	newNum := (curNum + 1) % len(sreSchedule)
	stuff := sreSchedule[newNum]
	if err := s.dao.UpdateValueByKey(ctx, fmt.Sprintf("%s_arrange", name), strconv.Itoa(newNum)); err != nil {
		return nil, err
	}
	return stuff, nil
}

func (s *Service) GetCurrentSre(ctx context.Context) (string, error) {
	curNum, err := s.GetArrange(ctx, "sre_arrange")
	if err != nil {
		return "", err
	}
	sreSchedule, err := s.GetSchedule(ctx, "sre_schedule")
	if err != nil {
		return "", err
	}
	newNum := curNum % len(sreSchedule)
	stuff := sreSchedule[newNum]
	return stuff.Username, nil
}

func (s *Service) ScheduleArrange(ctx context.Context) ([]*model.Stuff, error) {
	var stuffs []*model.Stuff
	service, err := s.UpdateAndGetArrange(ctx, "service")
	if err != nil {
		return nil, err
	}
	sre, err := s.UpdateAndGetArrange(ctx, "sre")
	if err != nil {
		return nil, err
	}
	stuffs = append(stuffs, service, sre)
	//nolint:gomnd
	if len(stuffs) < 2 {
		return nil, errors.New("failed select data")
	}
	return stuffs, nil
}

func (s *Service) GetOnCall() ([]string, []string, error) {
	var nameList, idList []string
	ctx := context.Background()
	stuffs, err := s.ScheduleArrange(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, stuff := range stuffs {
		nameList = append(nameList, stuff.Username)
		idList = append(idList, stuff.Id)
	}
	if err = s.Scheduler(ctx, stuffs[1].Username); err != nil {
		return nil, nil, err
	}
	return nameList, idList, nil
}

func (s *Service) RailgunBotHandler() func(ctx *bm.Context) {
	return s.sendScheduleTask.BMHandler
}

func (s *Service) Scheduler(ctx context.Context, sre string) error {
	startDay, endDay := GetStartAndEndTime()
	schedulerId, err := s.ac.Get("schedulerId").Int()
	if err != nil {
		return err
	}
	schedulerReq := &model.SchedulerReq{
		StartDay:    startDay,
		EndDay:      endDay,
		Level:       1,
		Users:       []string{sre},
		SchedulerId: schedulerId,
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	token, err := s.ac.Get("scheduleToken").String()
	if err != nil {
		return err
	}
	headers["Authorization"] = token
	if _, err = httpPost(schedulerURL, schedulerReq, headers); err != nil {
		return err
	}
	return nil
}

func GetStartAndEndTime() (string, string) {
	sTimeObj := time.Now()
	eTimeObj := sTimeObj.Add(144 * time.Hour)
	sTime := sTimeObj.Format("2006-01-02")
	eTime := eTimeObj.Format("2006-01-02")
	return sTime, eTime
}
