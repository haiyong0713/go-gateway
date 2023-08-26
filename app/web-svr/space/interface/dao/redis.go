package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
)

func keyUpArt(mid int64) string {
	return fmt.Sprintf("%s_%d", "uat", mid)
}

func keyUpArc(mid int64) string {
	return fmt.Sprintf("%s_%d", "uar", mid)
}

func keyOfficial(mid int64) string {
	return fmt.Sprintf("official_%d", mid)
}

func keyUserTab(mid int64) string {
	return fmt.Sprintf("usertab_%d", mid)
}

func keyWhitelist(mid int64) string {
	return fmt.Sprintf("whitelist_%d", mid)
}

func keyWhitelistValidTime(mid int64) string {
	return fmt.Sprintf("whitelistValidTime_%d", mid)
}

func keyCreativeViewData(mid int64) string {
	return fmt.Sprintf("creativeView_%d", mid)
}

// UpArtCache get up article cache.
func (d *Dao) UpArtCache(c context.Context, mid int64) (data *model.UpArtStat, err error) {
	var (
		value []byte
		key   = keyUpArt(mid)
		conn  = d.redis.Get(c)
	)
	defer conn.Close()
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	data = new(model.UpArtStat)
	if err = json.Unmarshal(value, &data); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

// SetUpArtCache set up article cache.
func (d *Dao) SetUpArtCache(c context.Context, mid int64, data *model.UpArtStat) (err error) {
	var (
		bs   []byte
		key  = keyUpArt(mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal(%v) error (%v)", data, err)
		return
	}
	return setKvCache(conn, key, bs, d.getCacheExpire(d.redisMinExpire, d.redisMaxExpire))
}

// UpArcCache get up archive cache.
func (d *Dao) UpArcCache(c context.Context, mid int64) (data *model.UpArcStat, err error) {
	var (
		value []byte
		key   = keyUpArc(mid)
		conn  = d.redis.Get(c)
	)
	defer conn.Close()
	if value, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET, %s) error(%v)", key, err)
		}
		return
	}
	data = new(model.UpArcStat)
	if err = json.Unmarshal(value, &data); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", value, err)
	}
	return
}

// SetUpArcCache set up archive cache.
func (d *Dao) SetUpArcCache(c context.Context, mid int64, data *model.UpArcStat) (err error) {
	var (
		bs   []byte
		key  = keyUpArc(mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal(%v) error (%v)", data, err)
		return
	}
	return setKvCache(conn, key, bs, d.getCacheExpire(d.redisMinExpire, d.redisMaxExpire))
}

// setKvCache .
func setKvCache(conn redis.Conn, key string, value []byte, expire int32) (err error) {
	if _, err = conn.Do("SETEX", key, expire, value); err != nil {
		log.Error("setKvCache SETEX key(%s) value(%s) expire(%d) error(%v)", key, string(value), expire, err)
		return
	}
	return
}

// CacheOfficial .
func (d *Dao) CacheOfficial(c context.Context, req *pb.OfficialRequest) (res *pb.OfficialReply, err error) {
	var (
		bs   []byte
		key  = keyOfficial(req.Mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheOfficial conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = &pb.OfficialReply{}
	if err = res.Unmarshal(bs); err != nil {
		log.Error("CacheOfficial json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// Cache SpaceUserTab
func (d *Dao) CacheUserTab(c context.Context, req *pb.UserTabReq) (res *model.UserTab, err error) {
	var (
		bs   []byte
		key  = keyUserTab(req.Mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheUserTab conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("CacheUserTab json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return res, nil
}

// Cache SpaceUserTab
func (d *Dao) CacheWhitelist(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistReply, err error) {
	var (
		bs   []byte
		key  = keyWhitelist(req.Mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheWhitelist conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = &pb.WhitelistReply{}
	if err = res.Unmarshal(bs); err != nil {
		log.Error("CacheWhitelist json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

func (d *Dao) CacheQueryWhitelistValid(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistValidTimeReply, err error) {
	var (
		bs   []byte
		key  = keyWhitelistValidTime(req.Mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("CacheWhitelist conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = &pb.WhitelistValidTimeReply{}
	if err = res.Unmarshal(bs); err != nil {
		log.Error("CacheWhitelist json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// cacheSFOfficial .
func (d *Dao) cacheSFOfficial(req *pb.OfficialRequest) string {
	return keyOfficial(req.Mid)
}

// cacheSFUserTab .
func (d *Dao) cacheSFUserTab(req *pb.UserTabReq) string {
	return keyUserTab(req.Mid)
}

// cacheSFWhitelist .
func (d *Dao) cacheSFWhitelist(req *pb.WhitelistReq) string {
	return keyWhitelist(req.Mid)
}

func (d *Dao) cacheSFQueryWhitelistValid(req *pb.WhitelistReq) string {
	return keyWhitelistValidTime(req.Mid)
}

// AddCacheOfficial .
func (d *Dao) AddCacheOfficial(c context.Context, req *pb.OfficialRequest, res *pb.OfficialReply) (err error) {
	var (
		key  = keyOfficial(res.Uid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = res.Marshal(); err != nil {
		log.Error("AddCacheOfficial json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisOfficialExpire, bs); err != nil {
		log.Error("AddCacheOfficial conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

// AddCacheUserTab
func (d *Dao) AddCacheUserTab(c context.Context, _ *pb.UserTabReq, res *model.UserTab) (err error) {
	var (
		key  = keyUserTab(res.Mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(res); err != nil {
		log.Error("AddCacheUserTab json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisUserTabExpire, bs); err != nil {
		log.Error("AddCacheUserTab conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

func (d *Dao) DelCacheUserTab(c context.Context, mid int64) error {
	key := keyUserTab(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("AddCacheUserTab conn.Do DEL key(%s) err(%+v)", key, err)
		return err
	}
	return nil
}

// AddCacheWhitelist
func (d *Dao) AddCacheWhitelist(c context.Context, req *pb.WhitelistReq, res *pb.WhitelistReply) (err error) {
	var (
		key  = keyWhitelist(req.Mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = res.Marshal(); err != nil {
		log.Error("AddCacheWhitelist json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisWhitelistExpire, bs); err != nil {
		log.Error("AddCacheWhitelist conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

func (d *Dao) DelCacheWhitelist(c context.Context, mid int64) error {
	key := keyWhitelist(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("AddCacheWhitelist conn.Do DEL key(%s) err(%+v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddCacheQueryWhitelistValid(c context.Context, req *pb.WhitelistReq, res *pb.WhitelistValidTimeReply) (err error) {
	var (
		key  = keyWhitelistValidTime(req.Mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = res.Marshal(); err != nil {
		log.Error("AddCacheQueryWhitelistValid json.Marshal req(%+v) error(%v)", res, err)
		return
	}
	if _, err = conn.Do("SETEX", key, d.redisWhitelistExpire, bs); err != nil {
		log.Error("AddCacheQueryWhitelistValid conn.Do SETEX key(%s) req(%+v) err(%+v)", key, string(bs), err)
		return
	}
	return
}

// DelOfficialCache .
func (d *Dao) DelOfficialCache(c context.Context, mid int64) (err error) {
	var (
		key  = keyOfficial(mid)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelOfficialCache conn.Do DEL key(%s) req(%+v) err(%v)", key, string(bs), err)
		return
	}
	return
}

func (d *Dao) getCacheExpire(min, max int) (res int32) {
	return int32(model.RandInt(d.rand, min, max))
}

func (d *Dao) CacheNotice(ctx context.Context, mid int64) (*model.Notice, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := noticeKey(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("CacheNotice conn.Do(GET,%s) error(%v)", key, err)
		return nil, err
	}
	data := new(model.Notice)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("CacheNotice json.Unmarshal(%s) error(%v)", string(bs), err)
		return nil, err
	}
	return data, nil
}

func (d *Dao) AddCacheNotice(ctx context.Context, mid int64, data *model.Notice) error {
	key := noticeKey(mid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheNotice json.Marshal mid:%d req(%+v) error(%v)", mid, data, err)
		return err
	}
	if _, err = conn.Do("SETEX", key, d.noticeExpire, bs); err != nil {
		log.Error("AddCacheNotice conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) DelCacheNotice(c context.Context, mid int64) error {
	key := noticeKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (d *Dao) CacheMasterpiece(ctx context.Context, mid int64) (*model.AidReasons, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := masterpieceKey(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("CacheMasterpiece conn.Do(GET,%s) error(%v)", key, err)
		return nil, err
	}
	data := new(model.AidReasons)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("CacheMasterpiece json.Unmarshal(%s) error(%v)", string(bs), err)
		return nil, err
	}
	return data, nil
}

func (d *Dao) AddCacheMasterpiece(ctx context.Context, mid int64, data *model.AidReasons) error {
	key := masterpieceKey(mid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheMasterpiece json.Marshal mid:%d req(%+v) error(%v)", mid, data, err)
		return err
	}
	if _, err = conn.Do("SETEX", key, d.mpExpire, bs); err != nil {
		log.Error("AddCacheMasterpiece conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) DelCacheMasterpiece(c context.Context, mid int64) error {
	key := masterpieceKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (d *Dao) CacheTopArc(ctx context.Context, mid int64) (*model.AidReason, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := topArcKey(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("CacheTopArc conn.Do(GET,%s) error(%v)", key, err)
		return nil, err
	}
	data := new(model.AidReason)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("CacheTopArc json.Unmarshal(%s) error(%v)", string(bs), err)
		return nil, err
	}
	return data, nil
}

func (d *Dao) AddCacheTopArc(ctx context.Context, mid int64, data *model.AidReason) error {
	key := topArcKey(mid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheTopArc json.Marshal mid:%d req(%+v) error(%v)", mid, data, err)
		return err
	}
	if _, err = conn.Do("SETEX", key, d.topArcExpire, bs); err != nil {
		log.Error("AddCacheTopArc conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) DelCacheTopArc(c context.Context, mid int64) error {
	key := topArcKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (d *Dao) CacheTheme(ctx context.Context, mid int64) (*model.ThemeDetails, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := themeKey(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("CacheTheme conn.Do(GET,%s) error(%v)", key, err)
		return nil, err
	}
	data := new(model.ThemeDetails)
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("CacheTheme json.Unmarshal(%s) error(%v)", string(bs), err)
		return nil, err
	}
	return data, nil
}

func (d *Dao) AddCacheTheme(ctx context.Context, mid int64, data *model.ThemeDetails) error {
	key := themeKey(mid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheTheme json.Marshal mid:%d req(%+v) error(%v)", mid, data, err)
		return err
	}
	if _, err = conn.Do("SETEX", key, d.themeExpire, bs); err != nil {
		log.Error("AddCacheTheme conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) DelCacheTheme(c context.Context, mid int64) error {
	key := themeKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (d *Dao) CacheTopDynamic(ctx context.Context, mid int64) (int64, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := topDyKey(mid)
	dyID, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Error("CacheTopDynamic conn.Do(GET,%s) error(%v)", key, err)
		return 0, err
	}
	return dyID, nil
}

func (d *Dao) AddCacheTopDynamic(ctx context.Context, mid int64, dyID int64) error {
	key := topDyKey(mid)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("SETEX", key, d.topDyExpire, dyID); err != nil {
		log.Error("AddCacheTheme conn.Do SETEX(%s) error(%v)", key, err)
		return err
	}
	return nil
}
