package service

import (
	"context"
	"encoding/json"

	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/model"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
	"strconv"

	"fmt"

	"go-common/library/log"
	"go-common/library/queue/databus"
)

const (
	_esContests             = "es_contests"
	_dbRecordUpdate         = "update"
	_contestStatusOver      = 3
	_contestStatusIng       = 2
	_contestStatusWaiting   = 1
	_contestStatusInit      = 0
	_contestFreezeFalse     = 0
	_contestConsumerLockKey = "esports_job:contest:consumer:lock:id:%d"
	_lockKeyTtl             = 86400
	_lockValue              = 1
)

type BinlogDataBusMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

type ContestValidInfo struct {
	ID            int64 `json:"id"`
	Sid           int64 `json:"sid"`
	ContestStatus int64 `json:"contest_status"`
	Status        int   `json:"status"`
}

func (s *Service) esportsBinlogConsumer(ctx context.Context) (err error) {
	var (
		msg *databus.Message
		ok  bool
	)
	if s.esportsBinlogSub == nil {
		return
	}
	log.Info("[esportsBinlogConsumer][Init]")
	msgs := s.esportsBinlogSub.Messages()
	for {
		if msg, ok = <-msgs; !ok {
			log.Error("[esportsBinlogConsumer][Exit], msg chan not ok")
			break
		}
		var ms = &BinlogDataBusMsg{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			msg.Commit()
			log.Errorc(ctx, "[esportsBinlogConsumer] json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		switch ms.Table {
		case _esContests:
			s.singleContestHandler(ctx, ms.Action, ms.New, ms.Old)
			s.contestScheduleHandler(ctx, ms.New)
		}
		msg.Commit()
	}
	return
}

func (s *Service) contestScheduleHandler(ctx context.Context, new []byte) {
	newContest := &model.ContestModel{}
	if err := json.Unmarshal(new, newContest); err != nil {
		log.Errorc(ctx, "[esportsBinlogConsumer][contestScheduleHandler][Unmarshal][Error], err:%+v", err)
		return
	}
	contestId := newContest.ID

	contestSchedule := &model.ContestSchedule{
		ID:            contestId,
		GameStage:     newContest.GameStage,
		Stime:         newContest.Stime,
		Etime:         newContest.Etime,
		HomeID:        newContest.HomeID,
		AwayID:        newContest.AwayID,
		HomeScore:     newContest.HomeScore,
		AwayScore:     newContest.AwayScore,
		GameId:        0,
		Sid:           newContest.Sid,
		Season:        nil,
		Mid:           newContest.Mid,
		SeriesId:      newContest.SeriesId,
		LiveRoom:      newContest.LiveRoom,
		Aid:           newContest.Aid,
		Collection:    newContest.Collection,
		Dic:           newContest.Dic,
		Special:       newContest.Special,
		SuccessTeam:   newContest.SuccessTeam,
		SpecialName:   newContest.SpecialName,
		SpecialTips:   newContest.SpecialTips,
		SpecialImage:  newContest.SpecialImage,
		Playback:      newContest.Playback,
		CollectionURL: newContest.CollectionURL,
		LiveURL:       newContest.LiveURL,
		DataType:      newContest.DataType,
		ContestFrozen: newContest.Status,
		ContestStatus: newContest.ContestStatus,
	}
	gameDetail, err := component.EspServiceClient.GetContestGameDetail(ctx, &v1.GetContestGameReq{
		ID: contestId,
	})

	if err != nil {
		log.Errorc(ctx, "[esportsBinlogConsumer][contestScheduleHandler][GetContestGame][Error], err:%+v", err)
	} else {
		contestSchedule.GameId = gameDetail.ID
		contestSchedule.Game = gameDetail
	}
	seasonDetail, err := component.EspServiceClient.GetSeasonDetail(ctx, &v1.GetSeasonModelReq{
		SeasonId: newContest.Sid,
	})
	if err != nil {
		log.Errorc(ctx, "[esportsBinlogConsumer][contestScheduleHandler][GetSeasonDetail][Error], err:%+v", err)
	} else {
		contestSchedule.Season = seasonDetail
	}
	log.Infoc(ctx, "[esportsBinlogConsumer][ContestScheduleHandler][push][Begin], contest: %+v", contestSchedule)
	s.pushContestSchedule(ctx, contestSchedule)
}

func (s *Service) singleContestHandler(ctx context.Context, action string, new []byte, old []byte) {
	push := false
	newContest := &ContestValidInfo{}
	switch action {
	case _insertAct:
		if err := json.Unmarshal(new, newContest); err != nil {
			log.Errorc(ctx, "[esportsBinlogConsumer][InsertRecord][Unmarshal][Error], err:%+v", err)
			return
		}
		if newContest.Status == _contestFreezeFalse && newContest.ContestStatus == _contestStatusIng {
			push = true
		}
	case _dbRecordUpdate:
		if err := json.Unmarshal(new, newContest); err != nil {
			log.Errorc(ctx, "[esportsBinlogConsumer][InsertRecord][Unmarshal][Error], err:%+v", err)
			return
		}
		oldContest := &ContestValidInfo{}
		if err := json.Unmarshal(old, oldContest); err != nil {
			log.Errorc(ctx, "[esportsBinlogConsumer][InsertRecord][Unmarshal][Error], err:%+v", err)
			return
		}
		if (oldContest.ContestStatus == _contestStatusInit || oldContest.ContestStatus == _contestStatusWaiting) &&
			newContest.ContestStatus == _contestStatusIng &&
			newContest.Status == _contestFreezeFalse {
			push = true
		}
	}
	if !push {
		return
	}
	log.Infoc(ctx, "[esportsBinlogConsumer][SingleContestHandler][push][Begin], contest: %+v", newContest)
	s.pushAndMessageSend(newContest)
}

func (s *Service) pushAndMessageSend(contest *ContestValidInfo) {
	if contest == nil {
		return
	}
	conn := component.GlobalAutoSubCache.Get(ctx)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Error("[Consumer][Lock][Redis][Close][Error], err: %+v", err)
		}
	}()
	key := fmt.Sprintf(_contestConsumerLockKey, contest.ID)
	if _, err := conn.Do("set", key, _lockValue, "EX", _lockKeyTtl, "NX"); err != nil {
		log.Error("[Consumer][Lock][Error], err: %+v", err)
		return
	}
	contestInfo, err := s.dao.ContestById(context.Background(), contest.ID)
	if err != nil {
		log.Error("[esportsBinlogConsumer][pushAndMessageSend][ContestById][Error], contest:%+v, err:%+v", contest, err)
		return
	}
	if contestInfo.LiveRoom <= 0 {
		log.Warn("[esportsBinlogConsumer][pushAndMessageSend][LiveRoom][Empty], contest:%+v", contestInfo)
		return
	}
	s.goroutineRegister(func(ctx context.Context) {
		s.pubContests(contestInfo)
	})
	if contestInfo.MessageSendUid == 0 {
		s.goroutineRegister(func(ctx context.Context) {
			s.sendContests(contestInfo)
		})
	}
}

func (s *Service) pushContestSchedule(ctx context.Context, contestSchedule *model.ContestSchedule) {
	key := strconv.FormatInt(contestSchedule.ID, 10)
	if err := retry.WithAttempts(ctx, "job_contest_ContestSchedule_pub", 3, netutil.DefaultBackoffConfig,
		func(c context.Context) error {
			buf, _ := json.Marshal(contestSchedule)
			return component.ContestSchedulePub.Send(ctx, key, buf)
		}); err != nil {
		log.Errorc(ctx, "[ContestScheduleHandler][Pub][Error] pushContestSchedule contest:%+v, error(%+v)", contestSchedule, err)
	}
	log.Infoc(ctx, "[ContestScheduleHandler][Pub][OK] pushContestSchedule contest:%+v", contestSchedule)
}
