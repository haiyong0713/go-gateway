package like

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_prefixInfo = "m_"
)

func keyInfo(sid int64) string {
	return fmt.Sprintf("%s%d", _prefixInfo, sid)
}

// SetInfoCache Dao
func (dao *Dao) SetInfoCache(c context.Context, v *like.Subject, sid int64) (err error) {
	if v == nil {
		v = &like.Subject{}
	}
	var (
		mckey = keyInfo(sid)
	)
	if err = dao.mc.Set(c, &memcache.Item{Key: mckey, Object: v, Flags: memcache.FlagGOB, Expiration: dao.mcLikeExpire}); err != nil {
		log.Error("conn.Set error(%v)", err)
		return
	}
	return
}

// InfoCache Dao
func (dao *Dao) InfoCache(c context.Context, sid int64) (v *like.Subject, err error) {
	var (
		mckey = keyInfo(sid)
	)
	if err = dao.mc.Get(c, mckey).Scan(&v); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			v = nil
		} else {
			log.Error("conn.Get error(%v)", err)
		}
	}
	return
}

const _viewData = "view:data:%d"

func viewDataKey(sid int64) string {
	return fmt.Sprintf(_viewData, sid)
}

// ViewDataCache .
func (dao *Dao) ViewDataCache(ctx context.Context, sid int64) (res []*like.WebData, err error) {
	res = make([]*like.WebData, 0)
	if err = dao.mc.Get(ctx, viewDataKey(sid)).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return res, nil
		}
		log.Error("ViewData d.ViewDataCache(sid:%d) error(%v)", sid, err)
		return
	}
	return
}

// del GetUpActReserveRelationInfoKey
func (dao *Dao) DelUpActReserveRelationInfoCache(ctx context.Context, sid int64) (err error) {
	key := GetUpActReserveRelationInfoBySid(sid)
	if err = dao.mc.Delete(ctx, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Errorv(ctx, log.KV("DelUpActReserveRelationInfoCache", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// del GetUpActReserveRelationInfo4SpaceCardIDs
func (dao *Dao) DelUpActReserveRelationInfoReachCache(ctx context.Context, mid int64) (err error) {
	key := GetUpActReserveRelationInfo4SpaceCardIDs(mid)
	if err = dao.mc.Delete(ctx, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Errorv(ctx, log.KV("GetUpActReserveRelationInfo4SpaceCardIDs", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}

// del UpActReserveRelation4LiveCache
func (dao *Dao) UpActReserveRelation4LiveCache(ctx context.Context, upMid int64) (err error) {
	key := GetUpActReserveRelationInfo4Live(upMid)
	if err = dao.mc.Delete(ctx, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Errorv(ctx, log.KV("UpActReserveRelation4LiveCache", fmt.Sprintf("%+v", err)), log.KV("key", key))
		return
	}
	return
}
