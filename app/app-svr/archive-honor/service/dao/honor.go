package dao

import (
	"context"
	"encoding/json"

	"go-common/library/log"

	"go-gateway/app/app-svr/archive-honor/service/api"

	"github.com/pkg/errors"
)

// HonorsByAid is
func (d *Dao) HonorsByAid(c context.Context, aid int64) (res map[int32]*api.Honor, err error) {
	var noCache bool
	if res, noCache, err = d.HonorsCacheByAid(c, aid); err != nil { //qps太大不回源了
		log.Error("d.HonorsCache aid(%d) err(%+v)", aid, err)
		return
	}
	if !noCache {
		return
	}
	//没有空缓存的可以回源
	if res, err = d.honorsByAid(c, aid); err != nil {
		log.Error("d.honorsByAid aid(%d) err(%+v)", aid, err)
		return
	}
	if len(res) == 0 { //没数据的记一个空缓存防止再次回源
		if err = d.AddHonorCache(c, aid, &api.Honor{Type: 0}); err != nil {
			log.Error("d.AddHonorCache aid(%d) v(%+v) err(%+v)", aid, &api.Honor{Type: 0}, err)
			err = nil
		}
	} else {
		if err = d.AddHonorsCache(c, aid, res); err != nil {
			log.Error("d.AddHonorCache aid(%d) err(%+v)", aid, err)
			err = nil
		}
	}
	return
}

// HonorUpdate is
func (d *Dao) HonorUpdate(c context.Context, aid int64, typ int32, url, desc, naUrl string) (rows int64, err error) {
	//更新db
	if rows, err = d.UpHonor(c, aid, typ, url, desc, naUrl); err != nil {
		err = errors.Wrapf(err, "d.TxUpHonor err aid(%d) type(%d) url(%s) desc(%s) naUrl(%s)", aid, typ, url, desc, naUrl)
		return
	}
	//更新缓存
	if err = d.AddHonorCache(c, aid, &api.Honor{Aid: aid, Type: typ, Url: url, Desc: desc, NaUrl: naUrl}); err != nil {
		err = errors.Wrapf(err, "d.AddHonorCache err aid(%d) type(%d) url(%s) desc(%s) naUrl(%s)", aid, typ, url, desc, naUrl)
		return
	}
	return
}

// HonorDel is
func (d *Dao) HonorDel(c context.Context, aid int64, typ int32) (err error) {
	if _, err = d.delHonor(c, aid, typ); err != nil {
		err = errors.Wrapf(err, "aid(%d) type(%d)", aid, typ)
		return
	}
	if err = d.DelHonorCache(c, aid, typ); err != nil {
		err = errors.Wrapf(err, "aid(%d) type(%d)", aid, typ)
		return
	}
	return
}

// HonorsByAids is
func (d *Dao) HonorsByAids(c context.Context, aids []int64) (res map[int64]map[int32]*api.Honor, err error) {
	var noCacheAids []int64
	if res, noCacheAids, err = d.HonorsCacheByAids(c, aids); err != nil { //qps太大不回源了
		log.Error("d.HonorsCache aids(%v) err(%+v)", aids, err)
		return
	}
	log.Info("HonorsByAids noCacheAids %v", noCacheAids)
	for _, aid := range noCacheAids { //没有空缓存的可以回源
		var honor map[int32]*api.Honor
		if honor, err = d.honorsByAid(c, aid); err != nil {
			log.Error("d.honorsByAid aid(%d) err(%+v)", aid, err)
			continue
		}
		if len(honor) == 0 { //没数据的记一个空缓存防止再次回源
			if err = d.AddHonorCache(c, aid, &api.Honor{Type: 0}); err != nil {
				log.Error("d.AddHonorCache aid(%d) v(%+v) err(%+v)", aid, &api.Honor{Type: 0}, err)
				err = nil
			}
			log.Info("HonorsByAids add cache empty")
		} else {
			if err = d.AddHonorsCache(c, aid, honor); err != nil {
				log.Error("d.AddHonorCache aid(%d) err(%+v)", aid, err)
				err = nil
			}
			res[aid] = honor
			honorstr, _ := json.Marshal(honor)
			log.Info("HonorsByAids add cache %s", honorstr)
		}
	}
	return
}
