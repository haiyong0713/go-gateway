package show

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-job/job/model/show"
)

func (s *Service) dealGoodHis(c context.Context, goodHis *show.GoodHisDatabus) {
	var (
		changed = false
		action  string
	)
	if goodHis.Old == nil {
		if !goodHis.New.IsDeleted() { // 新增一个稿件
			action = _archiveHonorUpdate
			changed = true
		}
	} else { // old != nil && new != nil
		if goodHis.Old.IsDeleted() && !goodHis.New.IsDeleted() { // 软删除恢复
			action = _archiveHonorUpdate
			changed = true
		}
		if !goodHis.Old.IsDeleted() && goodHis.New.IsDeleted() { // 软删除
			action = _archiveHonorDelete
			changed = true
		}
	}
	if !changed {
		return
	}
	// 如果修改了，则 1.刷播单 2.发消息
	s.goodHisMedListAndHonor(c)
	if action == _archiveHonorDelete { // 删除数据需要额外发一次
		honorMsg := new(show.HonorMsg)
		honorMsg.FromGoodHistory(goodHis.New.Aid, action, 0, s.c.GoodHis.URL)
		_ = s.sendArcHonor(c, honorMsg)
	}
}

func (s *Service) goodHisMedListAndHonor(c context.Context) {
	var (
		res   []*show.GoodHisRes
		aids  []int64
		err   error
		count int
	)
	if res, err = s.dao.RawGoodHisRes(c); err != nil {
		log.Error("[GoodHistory] Pick Resources Err %v", err)
		return
	}
	count = len(res)
	for _, v := range res {
		aids = append(aids, v.Aid)
	}
	if err = s.favDao.ReplaceMedias(c, s.c.GoodHis.MID, s.c.GoodHis.FID, aids); err != nil {
		log.Error("[GoodHistory] ReplaceMedias Aids %v Err %v", aids, err)
		return
	}
	for _, v := range aids { // 发消息到稿件成就，需要更新所有描述
		honorMsg := new(show.HonorMsg)
		honorMsg.FromGoodHistory(v, _archiveHonorUpdate, int64(count), s.c.GoodHis.URL)
		_ = s.sendArcHonor(c, honorMsg)
	}
	log.Info("[GoodHistory] ReplaceMedias Succ Aids %v", aids)
}

func (s *Service) loadGoodHistory() error {
	cards, err := s.dao.RawGoodHisRes(context.Background())
	if err != nil {
		log.Error("s.dao.RawGoodHisRes error(%+v)", err)
		return err
	}
	if err = s.dao.AddCacheGoodHistory(context.Background(), cards); err != nil {
		log.Error("Failed to add cache good history: cards(%+v), err(%+v)", cards, err)
		return err
	}
	return nil
}

func (s *Service) loadMidTopPhoto() error {
	res, err := s.dao.MidTopPhoto(context.Background())
	if err != nil || len(res) == 0 {
		log.Error("s.dao.MidTopPhoto error(%+v)", err)
		return err
	}
	if err = s.dao.AddCacheMidTopPhoto(context.Background(), res); err != nil {
		log.Error("Failed to add cache mid top photo: %+v", err)
		return err
	}
	return nil
}
