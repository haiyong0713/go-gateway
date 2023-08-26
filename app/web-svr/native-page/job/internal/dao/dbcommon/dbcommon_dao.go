package dbcommon

import (
	"context"
	"go-common/library/cache/credis"
	"time"

	actGRPC "go-gateway/app/web-svr/native-page/interface/api"
	natmdl "go-gateway/app/web-svr/native-page/job/internal/model"
	"go-gateway/app/web-svr/native-page/job/util"

	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
)

type Config struct {
	UpNatPagesExpire int64
}

type Dao struct {
	cfg   *Config
	db    *sql.DB
	cache *fanout.Fanout
	redis credis.Redis
}

func NewDao(cfg *Config, db *sql.DB, cache *fanout.Fanout, r credis.Redis) *Dao {
	return &Dao{
		cfg:   cfg,
		db:    db,
		cache: cache,
		redis: r,
	}
}

func (d *Dao) OnlinePage(c context.Context) {
	log.Warn("OnlinePage start")
	list, err := d.SearchPage(c)
	if err != nil {
		log.Error("OnlinePage d.SearchPage() error(%v)", err)
		return
	}
	if len(list) == 0 {
		log.Info("OnlinePage success nothing to do")
		return
	}
	var (
		ids []int64
	)
	for _, v := range list {
		tmpState := natmdl.PageOnLine
		if v.Type == natmdl.DynamicType {
			// topic只能同时对应一个上线话题活动页面
			if ids, err = d.ForeignFromIDs(c, v.ForeignID, v.Type); err != nil {
				log.Error("OnlinePage d.ForeignFromIDs(%d,%d) error(%v)", v.ForeignID, v.Type, err)
				continue
			}
			if len(ids) > 0 {
				tmpState = natmdl.PageOffLine
			}
		}
		if _, err = d.UpPage(c, v.ID, int64(tmpState)); err != nil {
			log.Error("OnlinePage d.UpPage(%d,%d,%d) online error(%v)", v.ID, v.Type, tmpState, err)
			return
		}
		log.Info("OnlinePage online (%d) upstate(%d)", v.ID, tmpState)
	}
	log.Info("OnlinePage success")
}

func (d *Dao) OfflinePage(c context.Context) ([]*actGRPC.NativePage, error) {
	list, err := d.EndList(c)
	if err != nil {
		log.Error("OfflinePage d.EndList() error(%v)", err)
		return nil, err
	}
	if len(list) == 0 {
		log.Info("OfflinePage success nothing to do")
		return list, nil
	}
	var ids []int64
	for _, v := range list {
		ids = append(ids, v.ID)
		log.Info("OfflinePage offline(%d,%d)", v.ID, v.Etime)
	}
	if _, err = d.OffLinePage(c, ids, "活动已过期"); err != nil {
		log.Error("OfflinePage s.nat.EndList() error(%v)", err)
		return nil, err
	}
	return list, nil
}

func timeStrToInt(timeStr string) (timeInt xtime.Time, err error) {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, timeStr, loc)
	if err = timeInt.Scan(theTime); err != nil {
		return
	}
	return
}

func (d *Dao) NewTopicPage(c context.Context) ([]*actGRPC.NativePage, map[string]struct{}, error) {
	log.Info("Start to NewTopicPage")
	var (
		id           int64 = 0
		err          error
		stime, etime xtime.Time
	)
	newStime := time.Now().AddDate(0, 0, -1).Format("2006-01-02") + " 19:00:00"
	if stime, err = timeStrToInt(newStime); err != nil {
		return nil, nil, err
	}
	newEtime := time.Now().Format("2006-01-02") + " 19:00:00"
	if etime, err = timeStrToInt(newEtime); err != nil {
		return nil, nil, err
	}
	newPages := make([]*actGRPC.NativePage, 0)
	creators := make(map[string]struct{})
	for {
		time.Sleep(10 * time.Millisecond)
		var pages []*actGRPC.NativePage
		if pages, err = d.AttemptNewNatPages(c, id, 1000); err != nil {
			//已经有重试逻辑，依然失败，则发出告警
			log.Error("NewTopicPage not get data id(%d)", id)
			break
		}
		if len(pages) == 0 {
			break
		}
		for _, v := range pages {
			if v == nil {
				continue
			}
			if v.ID > id {
				id = v.ID
			}
			//过滤非运营发起的话题活动
			if v.Type != actGRPC.TopicActType || v.FromType != actGRPC.PageFromSystem {
				continue
			}
			if v.Creator != "" {
				creators[v.Creator] = struct{}{}
			}
			//头一天19点至当天19点
			if v.Ctime >= stime && v.Ctime < etime {
				newPages = append(newPages, v)
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}
	log.Info("load NewTopicPage success, pagesLen=%+v", len(newPages))
	return newPages, creators, nil
}

func (d *Dao) AttemptNewNatPages(c context.Context, id, limit int64) ([]*actGRPC.NativePage, error) {
	rly, err := util.WithAttempts(3, netutil.BackoffConfig{
		MaxDelay:  2 * time.Second,
		BaseDelay: 100 * time.Millisecond,
		Factor:    1.6,
		Jitter:    0.2,
	}, func() (interface{}, error) {
		return d.pagingNewNatPages(c, id, limit)
	})
	if err != nil {
		log.Errorc(c, "Fail to attempt get attemptOnlineNatPages, error=%+v", err)
		return nil, err
	}
	pages, ok := rly.([]*actGRPC.NativePage)
	if !ok {
		log.Errorc(c, "Fail to type assertion on NativePage, rly=%+v", rly)
		return nil, errors.New("Fail to type assertion")
	}
	return pages, nil
}
