package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_top_upload    = "topphotoupload_%d_%d"
	_top_upload_id = "topphotoupload_id_%d"
	_topphoto      = "topphoto_%d"
)

func keyTopPhoto(mid int64) string {
	return fmt.Sprintf(_topphoto, mid)
}

func keyTopphotoByID(id int64) string {
	return fmt.Sprintf(_top_upload_id, id)
}

func keyTopPhotoUpload(mid int64, platFrom int) string {
	return fmt.Sprintf(_top_upload, mid, platFrom)
}

func (d *Dao) cacheSFMemberUploadTopphoto(mid int64, platFrom int) string {
	return keyTopPhotoUpload(mid, platFrom)
}

// cacheSFMemberTopphoto .
func (d *Dao) cacheSFMemberTopphoto(mid int64) string {
	return keyTopPhoto(mid)
}

// cacheSFMemberUploadTopphotoByID .
func (d *Dao) cacheSFMemberUploadTopphotoByID(id int64) string {
	return keyTopphotoByID(id)
}

// AddCacheMemberUploadTopphoto Set data
func (d *Dao) AddCacheMemberUploadTopphoto(c context.Context, mid int64, val *model.MemberPhotoUpload, platFrom int) (err error) {
	if val == nil {
		return
	}
	key := keyTopPhotoUpload(mid, platFrom)
	conn := d.redis.Get(c)
	defer conn.Close()
	data, err := json.Marshal(val)
	if err != nil {
		log.Error("dao.AddCacheMemberUploadTopphoto json marshal error:%v, key:%s, val:%+v", err, key, val)
		return
	}
	if _, err = conn.Do("SET", key, data); err != nil {
		log.Errorv(c, log.KV("AddCacheMemberUploadTopphoto", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// CacheMemberUploadTopphoto get data
func (d *Dao) CacheMemberUploadTopphoto(c context.Context, mid int64, platFrom int) (res *model.MemberPhotoUpload, err error) {
	key := keyTopPhotoUpload(mid, platFrom)
	conn := d.redis.Get(c)
	defer conn.Close()
	var data []byte
	if data, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorv(c, log.KV("CacheMemberUploadTopphoto", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	res = &model.MemberPhotoUpload{}
	if err = json.Unmarshal(data, res); err != nil {
		log.Error("dao.CacheMemberUploadTopphoto json unmarshal error:%v, key:%s, data:%s", err, key, string(data))
		return
	}
	return
}

// DelCacheMemberUploadTopphoto delete data
func (d *Dao) DelCacheMemberUploadTopphoto(c context.Context, mid int64, platFrom int) (err error) {
	key := keyTopPhotoUpload(mid, platFrom)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorv(c, log.KV("DelCacheMemberUploadTopphoto", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// AddCacheMemberTopphoto Set data
func (d *Dao) AddCacheMemberTopphoto(c context.Context, mid int64, val *model.MemberTopphoto) (err error) {
	if val == nil {
		return
	}
	key := keyTopPhoto(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	data, err := json.Marshal(val)
	if err != nil {
		log.Error("dao.AddCacheMemberTopphoto json marshal error:%v, key:%s, val:%+v", err, key, val)
		return
	}
	if _, err = conn.Do("SET", key, data); err != nil {
		log.Errorv(c, log.KV("AddCacheMemberTopphoto", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// CacheMemberTopphoto get data
func (d *Dao) CacheMemberTopphoto(c context.Context, mid int64) (res *model.MemberTopphoto, err error) {
	key := keyTopPhoto(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	var data []byte
	if data, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorv(c, log.KV("CacheMemberTopphoto", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	res = &model.MemberTopphoto{}
	if err = json.Unmarshal(data, res); err != nil {
		log.Error("dao.CacheMemberTopphoto json unmarshal error:%v, key:%s, data:%s", err, key, string(data))
		return
	}
	return
}

// DelCacheMemTopphotoCache delete data
func (d *Dao) DelCacheMemTopphotoCache(c context.Context, mid int64) (err error) {
	key := keyTopPhoto(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Errorv(c, log.KV("DelCacheMemTopphotoCache", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// AddCacheMemberUploadTopphotoByID Set data
func (d *Dao) AddCacheMemberUploadTopphotoByID(c context.Context, id int64, val *model.MemberPhotoUpload) (err error) {
	if val == nil {
		return
	}
	key := keyTopphotoByID(id)
	conn := d.redis.Get(c)
	defer conn.Close()
	data, err := json.Marshal(val)
	if err != nil {
		log.Error("dao.AddCacheMemberUploadTopphotoByID json marshal error:%v, key:%s, val:%+v", err, key, val)
		return
	}
	if _, err = conn.Do("SET", key, data); err != nil {
		log.Errorv(c, log.KV("AddCacheMemberUploadTopphotoByID", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// CacheMemberUploadTopphotoByID get data
func (d *Dao) CacheMemberUploadTopphotoByID(c context.Context, id int64) (res *model.MemberPhotoUpload, err error) {
	key := keyTopphotoByID(id)
	conn := d.redis.Get(c)
	defer conn.Close()
	var data []byte
	if data, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorv(c, log.KV("CacheMemberUploadTopphotoByID", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	res = &model.MemberPhotoUpload{}
	if err = json.Unmarshal(data, res); err != nil {
		log.Error("dao.CacheMemberUploadTopphotoByID json unmarshal error:%v, key:%s, data:%s", err, key, string(data))
		return
	}
	return
}
