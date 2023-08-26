package like

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

func keySelInfo(mid int64) string {
	return fmt.Sprintf("sel_info_%d", mid)
}

func keyProductRole(categoryID int64) string {
	return fmt.Sprintf("sel_pr_new_%d", categoryID)
}

func keyPrNotVote(categoryID int64) string {
	return fmt.Sprintf("sel_no_%d", categoryID)
}

func keyPrArc(id int64) string {
	return fmt.Sprintf("pr_a_%d", id)
}

func (d *Dao) CacheSelectionInfo(c context.Context, mid int64) ([]*like.SelectionQA, error) {
	key := keySelInfo(mid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheSelectionInfo(%s) return nil", key)
			return nil, nil
		}
		log.Error("CacheSelectionInfo conn.Do(GET key(%s)) error(%v)", key, err)
		return nil, err
	}
	var res []*like.SelectionQA
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("CacheSelectionInfo json.Unmarshal(%v) error(%v)", bs, err)
	}
	return res, nil
}

func (d *Dao) AddCacheSelectionInfo(c context.Context, mid int64, data []*like.SelectionQA) error {
	key := keySelInfo(mid)
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheSelectionInfo json.Marshal(%v) error (%v)", data, err)
		return err
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, 8640000, bs); err != nil {
		log.Error("AddCacheSelectionInfo conn.Do(SETEX, %s, %d, %d)", key, 8640000, bs)
	}
	return nil
}

func (d *Dao) DelCacheSelectionInfo(ctx context.Context, mid int64) error {
	key := keySelInfo(mid)
	_, err := component.GlobalRedis.Do(ctx, "DEL", key)
	if err != nil {
		log.Errorc(ctx, "DelCacheSelectionInfo Del mid(%d) error(%+v)", mid, err)
	}
	return err
}

// SetCacheProductRoles .
func (d *Dao) SetCacheProductRoles(c context.Context, categoryID int64, list []*like.ProductRoleDB) (err error) {
	var ok bool
	key := keyProductRole(categoryID)
	if ok, err = redis.Bool(component.GlobalRedis.Do(c, "EXPIRE", key, d.voteCategoryExpire)); err != nil {
		log.Error("conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	//无缓存时重新回源
	if !ok {
		err = d.AddCacheProductRoles(c, categoryID, list)
	}
	return
}

// AddCacheProductRoles .
func (d *Dao) AddCacheProductRoles(c context.Context, categoryID int64, list []*like.ProductRoleDB) (err error) {
	if len(list) == 0 {
		return
	}
	key := keyProductRole(categoryID)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	for _, object := range list {
		productRole := &like.ProductRole{
			ID:           object.ID,
			CategoryID:   object.CategoryID,
			CategoryType: object.CategoryType,
			Role:         object.Role,
			Product:      object.Product,
		}
		bs, _ := json.Marshal(productRole)
		if err = conn.Send("ZADD", key, mtimeCombine(object.VoteNum, object.Mtime.Time().Unix()), bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, d.voteCategoryExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.voteCategoryExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheProductRoles .
func (d *Dao) CacheProductRoles(c context.Context, categoryID int64) (res []*like.ProductRole, scoreMap map[int64]*like.ProductroleVote, maxVote int64, err error) {
	key := keyProductRole(categoryID)
	values, err := redis.Values(component.GlobalRedis.Do(c, "ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	scoreMap = make(map[int64]*like.ProductroleVote, 300)
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		pr := &like.ProductRole{}
		if err = json.Unmarshal(bs, pr); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		voteNum := voteNumFrom(num)
		mtime := mtimeFrom(num)
		scoreMap[pr.ID] = &like.ProductroleVote{
			VoteNum: voteNum,
			Mtime:   mtime,
		}
		if voteNum > maxVote {
			maxVote = voteNum
		}
		res = append(res, pr)
	}
	return
}

// SetNXValue Dao
func (d *Dao) SetNXValue(c context.Context, key string, productRoleID int64, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		rly  interface{}
	)
	if rly, err = component.GlobalRedis.Do(c, "SET", rkey, productRoleID, "EX", expire, "NX"); err != nil {
		log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		return
	}
	if rly != nil {
		res = true
	}
	return
}

// SetCachePrNotVote .
func (d *Dao) SetCachePrNotVote(c context.Context, categoryID int64, list []*like.ProductRoleDB, reSet bool) (err error) {
	var ok bool
	key := keyPrNotVote(categoryID)
	conn := d.redis.Get(c)
	defer conn.Close()
	if !reSet {
		if ok, err = redis.Bool(conn.Do("EXPIRE", key, d.voteCategoryExpire)); err != nil {
			log.Error("conn.Do(EXPIRE %s) error(%v)", key, err)
			return
		}
	}
	//无缓存时重新回源
	if !ok {
		var prNotVoteList []*like.ProductRole
		for _, pr := range list {
			prNotVoteList = append(prNotVoteList, &like.ProductRole{
				ID:           pr.ID,
				CategoryID:   pr.CategoryID,
				CategoryType: pr.CategoryType,
				Role:         pr.Role,
				Product:      pr.Product,
			})
		}
		err = d.AddCachePrNotVote(c, categoryID, prNotVoteList)
	}
	return
}

func (d *Dao) AddCachePrNotVote(c context.Context, categoryID int64, data []*like.ProductRole) error {
	key := keyPrNotVote(categoryID)
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCachePrNotVote json.Marshal(%v) error (%v)", data, err)
		return err
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, d.voteCategoryExpire, bs); err != nil {
		log.Error("AddCachePrNotVote conn.Do(SETEX, %s, %d, %d)", key, d.voteCategoryExpire, bs)
	}
	return nil
}

func (d *Dao) CachePrNotVote(c context.Context, categoryID int64) ([]*like.ProductRole, error) {
	key := keyPrNotVote(categoryID)
	bs, err := redis.Bytes(component.GlobalRedis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CachePrNotVote(%s) return nil", key)
			return nil, nil
		}
		log.Error("CachePrNotVote conn.Do(GET key(%s)) error(%v)", key, err)
		return nil, err
	}
	var res []*like.ProductRole
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("CachePrNotVote json.Unmarshal(%v) error(%v)", bs, err)
	}
	return res, nil
}

// AddCachePRVote.
func (d *Dao) AddCachePRVote(c context.Context, categoryID int64, productRole *like.ProductRole, currentVote *like.ProductroleVote) (err error) {
	key := keyProductRole(categoryID)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	bs, _ := json.Marshal(productRole)
	if err = conn.Send("ZADD", key, mtimeCombine(currentVote.VoteNum, currentVote.Mtime), bs); err != nil {
		log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.voteCategoryExpire); err != nil {
		log.Error("AddCachePRVote conn.Send(Expire, %s, %d) error(%v)", key, d.voteCategoryExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheAssistance conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCachePRVote conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheHotAssistance .
func (d *Dao) CacheHotAssistance(c context.Context, productroleID int64, start, end int) (res []*like.ProductRoleArc, total int, err error) {
	key := keyPrArc(productroleID)
	values, err := redis.Values(component.GlobalRedis.Do(c, "ZREVRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		pr := &like.ProductRoleArc{}
		if err = json.Unmarshal(bs, pr); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, pr)
	}
	total = from(num)
	return
}

// CacheTimeAssistance
func (d *Dao) CacheTimeAssistance(c context.Context, productroleID int64) (res []*like.ProductRoleArc, err error) {
	key := keyPrArc(productroleID)
	values, err := redis.Values(component.GlobalRedis.Do(c, "ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		pr := &like.ProductRoleArc{}
		if err = json.Unmarshal(bs, pr); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, pr)
	}
	return
}

func voteNumFrom(i int64) int64 {
	return i >> 32
}

func mtimeFrom(i int64) int64 {
	return i & 0xffffffff
}

func mtimeCombine(voteNum int64, mtime int64) int64 {
	return voteNum<<32 | mtime
}
