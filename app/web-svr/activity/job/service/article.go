package service

import (
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
)

const (
	_filteredArticles = "filtered_articles"
)

func (s *Service) articlePassproc() {
	defer s.waiter.Done()
	if s.articlePassSub == nil {
		return
	}
	for {
		msg, ok := <-s.articlePassSub.Messages()
		if !ok {
			log.Info("articlePassproc databus exit!")
			return
		}
		msg.Commit()
		m := &match.Message{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("articlePassproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Action != match.ActInsert {
			continue
		}
		switch m.Table {
		case _filteredArticles:
			article := new(like.ArticleMsg)
			if err := json.Unmarshal(m.New, article); err != nil {
				log.Error("articlePassproc article json.Unmarshal(%s) error(%v)", msg.Value, err)
				continue
			}
		}
		log.Info("articlePassproc key:%s partition:%d offset:%d table:%s value:%b", msg.Key, msg.Partition, msg.Offset, m.Table, msg.Value)
	}
}
