package poll

import (
	"context"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	model "go-gateway/app/web-svr/activity/interface/model/poll"
)

// Dao is
type Dao struct {
	c          *conf.Config
	db         *sql.DB
	mc         *memcache.Memcache
	localcache localcache
	cache      *fanout.Fanout
}

// New is
func New(conf *conf.Config) *Dao {
	dao := &Dao{
		c:     conf,
		db:    sql.NewMySQL(conf.MySQL.Like),
		mc:    memcache.New(conf.Memcache.Like),
		cache: fanout.New("cache", fanout.Worker(1), fanout.Buffer(10240)),
	}
	dao.initCache()
	go dao.cacheloadproc()
	return dao
}

// Transact is
func (d *Dao) Transact(ctx context.Context, txFunc func(*sql.Tx) error) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			// panic(p) // re-throw panic after Rollback
			log.Error("Panic in Transact: %+v", p)
			return
		}
		if err != nil {
			tx.Rollback() // err is non-nil; don't change it
			return
		}
		err = tx.Commit() // err is nil; if Commit returns error update err
	}()
	err = txFunc(tx)
	return err
}

// AllPollMeta is
func (d *Dao) AllPollMeta(ctx context.Context) ([]*model.PollMeta, error) {
	return d.localcache.GetAllPollMeta(), nil
}

// PollMeta is
func (d *Dao) PollMeta(ctx context.Context, id int64) (*model.PollMeta, bool) {
	allPollMeta := d.localcache.GetAllPollMetaMap()
	pm, ok := allPollMeta[id]
	return pm, ok
}

// ListPollOption is
func (d *Dao) ListPollOption(ctx context.Context, pollID int64) ([]*model.PollOption, bool) {
	optionByPollID := d.localcache.GetAllPollOptionByPollID()
	options, ok := optionByPollID[pollID]
	return options, ok
}
