package poll

import (
	"context"
	"sync/atomic"
	"time"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/poll"
)

type localcache interface {
	GetAllPollMeta() []*model.PollMeta
	GetAllPollMetaMap() map[int64]*model.PollMeta
	GetAllPollOption() []*model.PollOption
	GetAllPollOptionMap() map[int64]*model.PollOption
	GetAllPollOptionByPollID() map[int64][]*model.PollOption

	SetAllPollMeta(in []*model.PollMeta)
	SetAllPollOption(in []*model.PollOption)
}

type pollMetaCache struct {
	storeList []*model.PollMeta
	storeMap  map[int64]*model.PollMeta
}

type pollOptionCache struct {
	storeList   []*model.PollOption
	storeMap    map[int64]*model.PollOption
	byPollIDMap map[int64][]*model.PollOption
}

type cacheImpl struct {
	allPollMeta   atomic.Value
	allPollOption atomic.Value
}

var _ localcache = &cacheImpl{}

// GetAllPollMeta is
func (d *Dao) GetAllPollMeta(ctx context.Context) ([]*model.PollMeta, error) {
	rows, err := d.db.Query(ctx, `SELECT id,title,unique_table,repeatable,daily_chance,vote_maximum,end_at FROM act_poll_meta ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.PollMeta{}
	for rows.Next() {
		item := &model.PollMeta{}
		if err := rows.Scan(&item.Id, &item.Title, &item.UniqueTable, &item.Repeatable, &item.DailyChance, &item.VoteMaximum, &item.EndAt); err != nil {
			log.Warn("Failed to poll meta: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// GetAllPollOption is
func (d *Dao) GetAllPollOption(ctx context.Context) ([]*model.PollOption, error) {
	rows, err := d.db.Query(ctx, "SELECT id,poll_id,title,image,`group` FROM act_poll_option WHERE is_deleted=0 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.PollOption{}
	for rows.Next() {
		item := &model.PollOption{}
		if err := rows.Scan(&item.Id, &item.PollId, &item.Title, &item.Image, &item.Group); err != nil {
			log.Warn("Failed to poll option: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (c *cacheImpl) GetAllPollMeta() []*model.PollMeta {
	allPollMetaCache := c.allPollMeta.Load().(*pollMetaCache)
	return allPollMetaCache.storeList
}

func (c *cacheImpl) GetAllPollMetaMap() map[int64]*model.PollMeta {
	allPollMetaCache := c.allPollMeta.Load().(*pollMetaCache)
	return allPollMetaCache.storeMap
}

func (c *cacheImpl) GetAllPollOption() []*model.PollOption {
	allPollOptionCache := c.allPollOption.Load().(*pollOptionCache)
	return allPollOptionCache.storeList
}

func (c *cacheImpl) GetAllPollOptionMap() map[int64]*model.PollOption {
	allPollOptionCache := c.allPollOption.Load().(*pollOptionCache)
	return allPollOptionCache.storeMap
}

func (c *cacheImpl) GetAllPollOptionByPollID() map[int64][]*model.PollOption {
	allPollOptionCache := c.allPollOption.Load().(*pollOptionCache)
	return allPollOptionCache.byPollIDMap
}

func (c *cacheImpl) SetAllPollMeta(in []*model.PollMeta) {
	storeMap := make(map[int64]*model.PollMeta, len(in))
	for _, item := range in {
		storeMap[item.Id] = item
	}
	c.allPollMeta.Store(&pollMetaCache{
		storeList: in,
		storeMap:  storeMap,
	})
}

func (c *cacheImpl) SetAllPollOption(in []*model.PollOption) {
	storeMap := make(map[int64]*model.PollOption, len(in))
	for _, item := range in {
		storeMap[item.Id] = item
	}
	byPollIDMap := make(map[int64][]*model.PollOption)
	for _, item := range in {
		_, ok := byPollIDMap[item.PollId]
		if !ok {
			byPollIDMap[item.PollId] = []*model.PollOption{}
		}
		byPollIDMap[item.PollId] = append(byPollIDMap[item.PollId], item)
	}
	c.allPollOption.Store(&pollOptionCache{
		storeList:   in,
		storeMap:    storeMap,
		byPollIDMap: byPollIDMap,
	})
}

func (d *Dao) cacheloadproc() {
	for {
		log.Info("Load poll cache at: %+v", time.Now())

		func() {
			allPollMeta, err := d.GetAllPollMeta(context.Background())
			if err != nil {
				log.Warn("Failed to load all poll meta: %+v", err)
				return
			}
			d.localcache.SetAllPollMeta(allPollMeta)
		}()

		func() {
			allPollOption, err := d.GetAllPollOption(context.Background())
			if err != nil {
				log.Warn("Failed to load all poll option: %+v", err)
				return
			}
			d.localcache.SetAllPollOption(allPollOption)
		}()

		time.Sleep(time.Second * 60)
	}
}

func (d *Dao) initCache() {
	cache := &cacheImpl{}

	allPollMeta, err := d.GetAllPollMeta(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllPollMeta(allPollMeta)

	allPollOption, err := d.GetAllPollOption(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllPollOption(allPollOption)

	d.localcache = cache
}
