package popular

import (
	"context"

	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const largeCardType = "av_largecard"

// PopEntranceSave .
func (s *Service) PopLargeCardSave(ctx context.Context, param *show.PopLargeCard) (err error) {
	if param.ID == 0 {
		if err = s.showDao.PopLargeCardAdd(ctx, &show.PopLargeCardAD{
			Title:     param.Title,
			CardType:  largeCardType,
			RID:       param.RID,
			WhiteList: param.WhiteList,
			CreateBy:  param.CreateBy,
			Auto:      param.Auto,
			Deleted:   common.NotDeleted,
		}); err != nil {
			log.Error("[PopLargeCardSave] s.PopLargeCardAdd() error(%v)", err)
			return
		}
		return
	}
	if err = s.showDao.PopLargeCardUpdate(ctx, &show.PopLargeCardUP{
		ID:        param.ID,
		CardType:  largeCardType,
		RID:       param.RID,
		WhiteList: param.WhiteList,
		CreateBy:  param.CreateBy,
		Auto:      param.Auto,
		Title:     param.Title,
	}); err != nil {
		log.Error("[PopLargeCardSave] s.PopLargeCardUpdate() error(%v)", err)
		return
	}
	return
}

// PopularEntrance .
func (s *Service) PopLargeCardList(ctx context.Context, id int64, createBy string, rid int64, pn, ps int) (res *show.PopLargeCardRes, err error) {
	if res, err = s.showDao.PopLargeCardList(ctx, id, createBy, rid, pn, ps); err != nil {
		log.Error("[PopLargeCardList]s.PopLargeCardList() error(%v)", err)
		return
	}
	var (
		aids []int64
		arcs map[int64]*api.Arc
	)
	for _, item := range res.Items {
		aids = append(aids, item.RID)
		item.Bvid, _ = common.GetBvID(item.RID)
	}
	if arcs, err = s.arrDao.Arcs(ctx, aids); err != nil {
		return
	}
	for _, r := range res.Items {
		if item, ok := arcs[r.RID]; ok {
			r.VideoTitle = item.Title
			r.Author = item.Author.Name
		}
	}
	return
}

// PopEntranceOperate .
func (s *Service) PopLargeCardOperate(ctx context.Context, id int64, state int) (err error) {
	if state == common.NotDeleted {
		if err = s.showDao.PopLargeCardNotDelete(ctx, id); err != nil {
			log.Error("[PopEntranceOperate]s.PopEntranceOperate() id(%d) error(%v)", id, err)
		}
		return
	}
	if state == common.Deleted {
		if err = s.showDao.PopLargeCardDelete(ctx, id); err != nil {
			log.Error("[PopEntranceOperate]s.PopEntranceOperate() id(%d) error(%v)", id, err)
		}
	}
	return
}
