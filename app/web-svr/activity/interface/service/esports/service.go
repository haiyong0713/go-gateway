package esports

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	l "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	"time"

	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/esports"
	gameModel "go-gateway/app/web-svr/activity/interface/model/esports"
)

// Service ...
type Service struct {
	c     *conf.Config
	dao   *esports.Dao
	cache *fanout.Fanout
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   esports.New(c),
		cache: fanout.New("esports", fanout.Worker(1), fanout.Buffer(1024)),
	}
	return s
}

func (s *Service) AddFavGame(ctx context.Context, mid, fav1Id, fav2Id, fav3Id int64) (dbId int64, err error) {
	if fav, err := s.dao.GetEsportsArenaFav(ctx, "fav", mid); err == nil && fav != nil && fav.FirstFavGameId > 0 {
		// 已经提交过了
		return 0, ecode.ActivityTaskHasFinish
	}
	dbId, err = s.dao.InsertEsportsArenaFav(ctx, mid, fav1Id, fav2Id, fav3Id)
	log.Infoc(ctx, "AddFavGame id:%v , err:%v", dbId, err)
	if (err == nil && dbId > 0) || xecode.EqualError(ecode.ActivityTaskHasFinish, err) {
		s.cache.SyncDo(ctx, func(ctx context.Context) {
			s.dao.CacheEsportsArenaFav(ctx, "fav", &gameModel.EsportsActFav{
				ID:              dbId,
				Mid:             mid,
				FirstFavGameId:  fav1Id,
				SecondFavGameId: fav2Id,
				ThirdFavGameId:  fav3Id,
				Ctime:           xtime.Time(time.Now().Unix()),
				Mtime:           xtime.Time(time.Now().Unix()),
			})
		})
	}
	return
}

func (s *Service) UserInfo(ctx context.Context, mid int64) (fav *gameModel.EsportsActFav, err error) {
	if fav, err = s.dao.GetEsportsArenaFav(ctx, "fav", mid); err == nil && fav != nil && fav.FirstFavGameId > 0 {
		return
	}

	favs, err := s.dao.EsportsArenaFavDB(ctx, mid)
	if err == nil && favs != nil && len(favs) > 0 {
		fav = favs[0]
	}
	return
}

func (s *Service) CheckLotteryTimeLimit(ctx context.Context, sid string, actionType int, cid int64, mid int64, orderNo string) (err error) {
	log.Infoc(ctx, "CheckLotteryTimeLimit sid:%v , actionType:%v , cid:%v , mid:%v", sid, actionType, cid, mid)
	if sid == conf.Conf.EsportsArena.Sid {
		if actionType == l.TimesArchiveType ||
			actionType == l.TimesFollowType ||
			(actionType == l.TimesFeType && cid == conf.Conf.EsportsArena.ZDCid) {

			//  特殊场景(关注UP主)校验一下是否重入
			if actionType == l.TimesFollowType {
				ok, err := s.dao.SetNxOrder(ctx, mid, orderNo)
				log.Infoc(ctx, "CheckLotteryTimeLimit check orderNo:%v , exist:%v , err:%v", orderNo, ok, err)
				// 重入校验成功（不存在当前orderNo），次数+1
				if !ok && err == nil {
					return ecode.ActivityRepeatSubmit
				}
			}

			date := time.Now().Format("2006-01-02")
			times, err := s.dao.IncrLotteryTimes(ctx, date, mid)
			log.Infoc(ctx, "CheckLotteryTimeLimit nowtimes:%v , LTimeLimit:%v , err:%v", times, s.c.EsportsArena.LTimeLimit, err)
			if err == nil {
				if times > s.c.EsportsArena.LTimeLimit {
					return ecode.ActivityLotteryAddTimesLimit
				}
			}
		}
	}
	return
}
