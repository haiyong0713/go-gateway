package image

import (
	"context"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

func (d *Dao) Image(c context.Context, mid, imageId int64) (*note.ImgInfo, error) {
	addCache := true
	res, err := d.cacheImg(c, mid, imageId)
	if err != nil {
		if err != redis.ErrNil {
			log.Warn("noteWarn image mid(%d) imageId(%d) err(%+v)", mid, imageId, err)
			addCache = false
		}
	}
	if res != nil {
		cache.MetricHits.Inc("bts:Img")
		return res, nil
	}
	cache.MetricMisses.Inc("bts:Img")
	res, err = d.rawImage(c, mid, imageId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	miss := res
	if miss == nil {
		miss = &note.ImgInfo{}
	}
	if !addCache {
		return miss, nil
	}
	if foErr := d.cache.Do(c, func(ctx context.Context) {
		if err := d.AddCacheImg(ctx, mid, imageId, miss); err != nil {
			log.Warn("noteWarn Image err(%+v)", err)
		}
	}); foErr != nil {
		log.Warn("noteInfo fanout err(%+v)", foErr)
	}
	return miss, nil
}

func (d *Dao) Images(c context.Context, keys []int64, mid int64) (map[int64]*api.PublishImgInfo, error) {
	if len(keys) == 0 {
		return make(map[int64]*api.PublishImgInfo), nil
	}
	addCache := true
	var miss []int64
	res, miss, err := d.cacheImgs(c, mid, keys)
	if err != nil {
		addCache = false
		res = nil
	}
	for _, key := range keys {
		if res == nil || res[key] == nil {
			miss = append(miss, key)
		}
	}
	if len(miss) == 0 {
		if res == nil {
			return nil, errors.Wrapf(ecode.NothingFound, "Images keys(%v) mid(%d)", keys, mid)
		}
		return res, nil
	}
	var missData map[int64]*api.PublishImgInfo
	missData, err = d.rawImages(c, miss, mid)
	if err != nil {
		log.Error("noteWarn Images miss(%v) err(%+v)", miss, err)
		if res == nil {
			return nil, err
		}
	}
	if res == nil {
		res = make(map[int64]*api.PublishImgInfo, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if missData == nil {
		missData = make(map[int64]*api.PublishImgInfo)
	}
	if !addCache {
		return res, nil
	}
	for _, m := range missData {
		curM := &note.ImgInfo{ImageId: m.ImageId, Location: m.Location}
		if foErr := d.cache.Do(c, func(ctx context.Context) {
			if err := d.AddCacheImg(ctx, mid, curM.ImageId, curM); err != nil {
				log.Warn("noteWarn Images err(%+v)", err)
			}
		}); foErr != nil {
			log.Warn("noteInfo fanout err(%+v)", foErr)
		}
	}
	return res, nil
}
