package service

import (
	"encoding/json"

	"go-gateway/app/web-svr/activity/job/model/match"
	"go-gateway/app/web-svr/activity/job/model/s10"

	"go-common/library/log"
)

func (s *Service) parseNewUserCostRecord(msg *match.Message) (new *s10.UserCostRecord, err error) {
	new = &s10.UserCostRecord{}
	if err = json.Unmarshal(msg.New, new); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", msg.New, err)
	}
	return
}

func (s *Service) parseNewUserLotteryRecord(msg *match.Message) (new *s10.UserLotteryRecord, err error) {
	new = &s10.UserLotteryRecord{}
	if err = json.Unmarshal(msg.New, new); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", msg.New, err)
	}
	return
}
