package service

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/steins-gate/job/internal/model"

	"go-common/library/log"
)

func (s *Service) evalConsumeproc() {
	var err error
	defer s.waiter.Done()
	for {
		msg, ok := <-s.evaluationSub.Messages()
		if !ok || s.daoClosed {
			log.Info("arc databus Consumer exit")
			break
		}
		//nolint:errcheck
		msg.Commit()
		var ms = &model.EvaluationMsg{}
		log.Info("EvaluationMsg New message: %s", msg.Value)
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			log.Error("EvaluationMsg json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if ms.AID == 0 {
			log.Error("EvaluationMsg aid 0!")
			continue
		}
		s.dao.AddEval(context.Background(), ms.AID, ms.Score)
	}

}
