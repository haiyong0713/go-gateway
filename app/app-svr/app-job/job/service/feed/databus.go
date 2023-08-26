package feed

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
)

func (s *Service) cardConsumer() {
	defer s.waiter.Done()
	msgs := s.cardSub.Messages()
	for {
		msg, ok := <-msgs
		if !ok {
			log.Info("cardConsumerSub Cloesd ok(%v)", ok)
			return
		}
		_ = msg.Commit()
		var (
			card *operate.Card
			err  error
			c    = context.TODO()
		)
		if err = json.Unmarshal(msg.Value, &card); err != nil {
			log.Error("cardConsumer json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		cardID, _ := strconv.ParseInt(msg.Key, 10, 64)
		if err = s.dao.AddConvergeAiCache(c, cardID, card); err != nil {
			log.Error("cardConsumer s.dao.AddConvergeAiCache error(%v)", err)
			continue
		}
	}
}
