package appstore

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

// appstoreMIDIsRecievedKey .
func appstoreMIDIsRecievedKey(batchToken string, mid int64) string {
	return fmt.Sprintf("asmir_%s_%d", batchToken, mid)
}

// CacheAppstoreMIDIsRecieved .
func (dao *Dao) CacheAppstoreMIDIsRecieved(ctx context.Context, batchToken string, mid int64) (r int64, err error) {
	var (
		key  = appstoreMIDIsRecievedKey(batchToken, mid)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if r, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheAppstoreMIDIsRecieved(%s) return nil", key)
		} else {
			log.Error("CacheAppstoreMIDIsRecieved conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	return
}

// AddCacheAppstoreMIDIsRecieved .
func (dao *Dao) AddCacheAppstoreMIDIsRecieved(ctx context.Context, batchToken string, mid int64, val int64) (err error) {
	var (
		key  = appstoreMIDIsRecievedKey(batchToken, mid)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, dao.appstoreExpire, val); err != nil {
		log.Error("AddCacheAppstoreMIDIsRecieved conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.appstoreExpire, val, err)
	}
	return
}

// DelCacheAppstoreMIDIsRecieved .
func (dao *Dao) DelCacheAppstoreMIDIsRecieved(ctx context.Context, batchToken string, mid int64) (err error) {
	var (
		key  = appstoreMIDIsRecievedKey(batchToken, mid)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheAppstoreMIDIsRecieved conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// appstoreTelIsRecievedKey .
func appstoreTelIsRecievedKey(batchToken string, tel string) string {
	return fmt.Sprintf("astir_%s_%s", batchToken, tel)
}

// CacheAppstoreMIDIsRecieved .
func (dao *Dao) CacheAppstoreTelIsRecieved(ctx context.Context, batchToken string, tel string) (r int64, err error) {
	var (
		key  = appstoreTelIsRecievedKey(batchToken, tel)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if r, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheAppstoreTelIsRecieved(%s) return nil", key)
		} else {
			log.Error("CacheAppstoreTelIsRecieved conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	return
}

// AddCacheAppstoreTelIsRecieved .
func (dao *Dao) AddCacheAppstoreTelIsRecieved(ctx context.Context, batchToken string, tel string, val int64) (err error) {
	var (
		key  = appstoreTelIsRecievedKey(batchToken, tel)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, dao.appstoreExpire, val); err != nil {
		log.Error("AddCacheAppstoreTelIsRecieved conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.appstoreExpire, val, err)
	}
	return
}

// DelCacheAppstoreTelIsRecieved .
func (dao *Dao) DelCacheAppstoreTelIsRecieved(ctx context.Context, batchToken string, tel string) (err error) {
	var (
		key  = appstoreTelIsRecievedKey(batchToken, tel)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheAppstoreTelIsRecieved conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// appstoreIsRecievedKey .
func appstoreIsRecievedKey(batchToken string, matchLabel string, matchKind int64) string {
	return fmt.Sprintf("asis_%s_%s_%d", batchToken, matchLabel, matchKind)
}

// CacheAppstoreIsRecieved .
func (dao *Dao) CacheAppstoreIsRecieved(ctx context.Context, batchToken string, matchLabel string, matchKind int64) (r int64, err error) {
	var (
		key  = appstoreIsRecievedKey(batchToken, matchLabel, matchKind)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if r, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheAppstoreIsRecieved(%s) return nil", key)
		} else {
			log.Error("CacheAppstoreIsRecieved conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	return
}

// AddCacheAppstoreIsRecieved .
func (dao *Dao) AddCacheAppstoreIsRecieved(ctx context.Context, batchToken string, matchLabel string, matchKind int64, val int64) (err error) {
	var (
		key  = appstoreIsRecievedKey(batchToken, matchLabel, matchKind)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, dao.appstoreExpire, val); err != nil {
		log.Error("AddCacheAppstoreIsRecieved conn.Send(SETEX, %s, %v, %s) error(%v)", key, dao.appstoreExpire, val, err)
	}
	return
}

// DelCacheAppstoreIsRecieved .
func (dao *Dao) DelCacheAppstoreIsRecieved(ctx context.Context, batchToken string, matchLabel string, matchKind int64) (err error) {
	var (
		key  = appstoreIsRecievedKey(batchToken, matchLabel, matchKind)
		conn = dao.redis.Get(ctx)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheAppstoreIsRecieved conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}
