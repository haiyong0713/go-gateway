package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	_gotoAV = "av"
)

func (s *Service) loadPopularCard() {
	var (
		tabCards map[int][]*show.CardListAI
		err      error
	)
	if tabCards, err = s.loadHotHeTongTabCard(context.Background()); err != nil {
		log.Error("loadPopularCard loadHotHeTongTabCard err(%+v)", err)
		return
	}
	for k, item := range tabCards {
		var (
			cards    []*show.PopularCardAI
			cacheLen int
		)
		for _, hcard := range item {
			cards = append(cards, hcard.CardListChange())
		}
		if cacheLen, err = s.dao.TotalPopularCardTenCache(context.Background(), k); err != nil { // 获取当前缓存里面所有值的数量用于比较是不是数量变了
			log.Error("loadPopularCard TotalPopularCardTenCacheFail err(%+v) key(%d)", err, k)
			continue // 单独每个key更新
		}
		if err = s.dao.AddPopularCardTenCache(context.Background(), k, cards); err != nil {
			log.Error("loadPopularCard AddPopularCardTenCacheFail err(%+v) key(%d)", err, k)
			continue // 单独每个key更新
		}
		if cacheLen != len(cards) { // 不等于的时候加个报警日志
			log.Info("PopularCardTenCacheCountChange before(%d) after(%d)", cacheLen, len(cards))
		}
		if cacheLen > len(cards) { // 如果新的数量小于缓存里面的，需要删除掉多余的
			if err = s.dao.DelPopularCardTenCache(context.Background(), k, len(cards), cacheLen); err != nil {
				log.Error("loadPopularCard DelPopularCardTenCache err(%+v) key(%d)", err, k)
			}
		}
	}
	log.Info("loadPopularCard success len(%d)", len(tabCards))
}

func (s *Service) loadHotHeTongTabCard(c context.Context) (tmpList map[int][]*show.CardListAI, err error) {
	tmpList = make(map[int][]*show.CardListAI)
	for i := 0; i < 11; i++ {
		var (
			err        error
			hottabAids []*show.CardListAI
			flowResp   *creativeAPI.FlowJudgesReply
			oids       []int64
			forbidAids = make(map[int64]struct{})
		)
		if hottabAids, err = s.dao.HotHeTongTabCard(c, i); err != nil {
			log.Error("HotHeTongTabCardFail %+v", err)
			return nil, err
		}
		for _, hot := range hottabAids {
			if hot.Goto == _gotoAV {
				oids = append(oids, hot.ID)
			}
		}
		if flowResp, err = s.creativeClient.FlowJudges(context.Background(), &creativeAPI.FlowJudgesReq{
			Oids:     oids,
			Business: 4,
			Gid:      24,
		}); err != nil {
			log.Error("s.creativeClient.FlowJudge error(%v)", err)
			tmpList[i] = hottabAids
		} else {
			for _, oid := range flowResp.Oids {
				forbidAids[oid] = struct{}{}
			}
			for _, list := range hottabAids {
				if list.Goto == _gotoAV {
					if _, ok := forbidAids[list.ID]; ok {
						log.Info("aid(%d) is flowJundged", list.ID)
						continue
					}
				}
				tmpList[i] = append(tmpList[i], list)
			}
		}
		log.Info("buildHotSuccess(%d) len(%d)", i, len(tmpList[i]))
	}
	log.Info("HotHeTongTabCardSuccess len(%d)", len(tmpList))
	return
}
