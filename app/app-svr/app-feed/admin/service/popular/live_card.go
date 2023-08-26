package popular

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	liveCardType  = "live_card"
	_addLiveCard  = "add"
	_editLiveCard = "up"
	_timeFormat   = "2006-01-02 15:04:05"
)

func (s *Service) PopLiveCardSave(ctx context.Context, param *show.PopLiveCard, uname string, uid int64) (err error) {
	obj := map[string]interface{}{
		"rid": param.RID,
	}
	if param.ID == 0 {
		if param.ID, err = s.showDao.PopLiveCardAdd(ctx, &show.PopLiveCardAD{
			CardType: liveCardType,
			RID:      param.RID,
			Cover:    param.Cover,
			CreateBy: param.CreateBy,
		}); err != nil {
			log.Error("[PopLiveCardSave] s.PopLiveCardAdd() error(%v)", err)
			return
		}
		if err = util.AddLogs(common.LogLiveCard, uname, uid, param.ID, _addLiveCard, obj); err != nil {
			log.Error("[PopLiveCardAdd] AddLogs error(%v)", err)
		}
		return
	}
	if err = s.showDao.PopLiveCardUpdate(ctx, &show.PopLiveCardUP{
		ID:    param.ID,
		RID:   param.RID,
		Cover: param.Cover,
	}); err != nil {
		log.Error("[PopLiveCardSave] s.PopLiveCardUpdate() error(%v)", err)
		return
	}
	if err = util.AddLogs(common.LogLiveCard, uname, uid, param.ID, _editLiveCard, obj); err != nil {
		log.Error("[PopLiveCardEdit] AddLogs error(%v)", err)
	}
	return
}

func (s *Service) PopLiveCardList(ctx context.Context, id int64, state int, createBy string, pn, ps int) (res *show.PopLiveCardRes, err error) {
	if res, err = s.showDao.PopLiveCardList(ctx, id, state, createBy, pn, ps); err != nil {
		log.Error("[PopLiveCardList] s.PopLiveCardList() error(%v)", err)
		return
	}
	for _, item := range res.Items {
		item.MtimeStr = item.Mtime.Format(_timeFormat)
	}
	return
}

func (s *Service) PopLiveCardOperate(ctx context.Context, id int64, state int, uname string, uid int64) (err error) {
	if err = s.showDao.PopLargeCardOperate(ctx, id, state); err != nil {
		log.Error("[PopLiveCardOperate] s.PopLiveCardOperate() id(%d) error(%v)", id, err)
	}
	obj := map[string]interface{}{
		"state": state,
	}
	if err = util.AddLogs(common.LogLiveCard, uname, uid, id, _editLiveCard, obj); err != nil {
		log.Error("[PopLiveCardOperate] AddLogs error(%v)", err)
	}
	return
}
