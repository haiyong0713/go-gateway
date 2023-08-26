package show

import (
	"context"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"
)

func (s *Service) loadCarPopularCard() {
	var (
		tabCards map[int][]*show.CardListAI
		err      error
	)
	if tabCards, err = s.loadCarHotHeTongTabCard(context.Background()); err != nil {
		log.Error("loadCarPopularCard loadHotHeTongTabCard err(%+v)", err)
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
		if cacheLen, err = s.dao.TotalCarPopularCardTenCache(context.Background(), k); err != nil { // 获取当前缓存里面所有值的数量用于比较是不是数量变了
			log.Error("loadCarPopularCard TotalCarPopularCardTenCacheFail err(%+v) key(%d)", err, k)
			continue // 单独每个key更新
		}
		if err = s.dao.AddCarPopularCardTenCache(context.Background(), k, cards); err != nil {
			log.Error("loadCarPopularCard AddCarPopularCardTenCacheFail err(%+v) key(%d)", err, k)
			continue // 单独每个key更新
		}
		if cacheLen != len(cards) { // 不等于的时候加个报警日志
			log.Info("PopularCardTenCacheCountChange before(%d) after(%d)", cacheLen, len(cards))
		}
		if cacheLen > len(cards) { // 如果新的数量小于缓存里面的，需要删除掉多余的
			if err = s.dao.DelCarPopularCardTenCache(context.Background(), k, len(cards), cacheLen); err != nil {
				log.Error("loadCarPopularCard DelCarPopularCardTenCache err(%+v) key(%d)", err, k)
			}
		}
	}
	log.Info("loadCarPopularCard success len(%d)", len(tabCards))
}

func (s *Service) loadCarHotHeTongTabCard(c context.Context) (tmpList map[int][]*show.CardListAI, err error) {
	tmpList = make(map[int][]*show.CardListAI)
	for i := 0; i < 11; i++ {
		var (
			flowResp   *creativeAPI.FlowJudgesReply
			oids       []int64
			forbidAids = make(map[int64]struct{})
			rcmdItems  []*show.CardListAI
		)
		hottabAids, err := s.dao.HotHeTongTabCard(c, i)
		if err != nil {
			log.Error("CarHotHeTongTabCardFail %+v", err)
			return nil, err
		}
		for _, hot := range hottabAids {
			// 车载只需要UGC普通卡片数据，其他数据都不要
			if hot.Goto != _gotoAV {
				continue
			}
			oids = append(oids, hot.ID)
			rcmdItems = append(rcmdItems, hot)
		}
		if flowResp, err = s.creativeClient.FlowJudges(context.Background(), &creativeAPI.FlowJudgesReq{
			Oids:     oids,
			Business: 4,
			Gid:      24,
		}); err != nil {
			log.Error("s.creativeClient.FlowJudge error(%v)", err)
			tmpList[i] = rcmdItems
		} else {
			for _, oid := range flowResp.Oids {
				forbidAids[oid] = struct{}{}
			}
			for _, list := range rcmdItems {
				if _, ok := forbidAids[list.ID]; ok {
					log.Info("aid(%d) is flowJundged", list.ID)
					continue
				}
				tmpList[i] = append(tmpList[i], list)
			}
		}
		log.Info("buildHotSuccess(%d) len(%d)", i, len(tmpList[i]))
	}
	log.Info("HotHeTongTabCardSuccess len(%d)", len(tmpList))
	return
}
