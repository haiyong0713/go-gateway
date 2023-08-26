package image

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

const (
	_keyImg = "_img"
)

func (d *Dao) AddCacheImg(c context.Context, mid, imageId int64, data *note.ImgInfo) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := imgKey(mid, imageId)
	bs, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "addCacheImg key(%s) data(%+v)", key, data)
	}
	if _, err := conn.Do("SETEX", key, d.imgExpire, bs); err != nil {
		return errors.Wrapf(err, "addCacheImg key(%s) data(%+v)", key, data)
	}
	return nil
}

func (d *Dao) cacheImg(c context.Context, mid, imageId int64) (*note.ImgInfo, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := imgKey(mid, imageId)
	item, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	cache := &note.ImgInfo{}
	if err = json.Unmarshal(item, &cache); err != nil {
		return nil, err
	}
	return cache, nil
}

func (d *Dao) cacheImgs(c context.Context, mid int64, imageIds []int64) (cached map[int64]*api.PublishImgInfo, missed []int64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var (
		args    = redis.Args{}
		keysMap = make(map[int64]struct{})
	)
	for _, imgId := range imageIds {
		if _, ok := keysMap[imgId]; ok {
			continue
		}
		args = args.Add(imgKey(mid, imgId))
		keysMap[imgId] = struct{}{}
	}
	var items [][]byte
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		err = errors.Wrapf(err, "cacheImgs args(%+v)", args)
		return
	}
	cached = make(map[int64]*api.PublishImgInfo)
	for _, bs := range items {
		if bs == nil {
			continue
		}
		tmp := &api.PublishImgInfo{}
		if e := json.Unmarshal(bs, tmp); e != nil {
			log.Warn("noteWarn cacheImgs Unmarshal bs(%s) error(%v)", bs, e)
			continue
		}
		cached[tmp.ImageId] = tmp
		delete(keysMap, tmp.ImageId)
	}
	for aid := range keysMap {
		missed = append(missed, aid)
	}
	return
}

func imgKey(mid, imageId int64) string {
	return fmt.Sprintf("%d_%d%s", mid, imageId, _keyImg)
}
