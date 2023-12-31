package dao

import (
	"context"

	gmc "go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/dynamic/service/model"
)

const (
	_keyRegionArcs         = "dyra"  // key of region archives
	_keyRegionTagArcs      = "dyrta" // key of tag archives
	_keyRegionBusinessArcs = "rba"
)

// pingMC ping memcache.
func (d *Dao) pingMC(c context.Context) (err error) {
	conn := d.mc.Get(c)
	if err = conn.Set(&gmc.Item{Key: "ping", Value: []byte{1}, Expiration: d.mcExpire}); err != nil {
		log.Error("conn.Store(set, ping, 1) error(%v)", err)
	}
	conn.Close()
	return
}

// SetRegionCache set region archive to cache.
func (d *Dao) SetRegionCache(c context.Context, regionArcs map[int32][]int64) (err error) {
	conn := d.mc.Get(c)
	defer conn.Close()
	tmp := make(map[int32]*model.Aids)
	for k, v := range regionArcs {
		tmp[k] = &model.Aids{IDs: v}
	}
	item := &gmc.Item{Key: _keyRegionArcs, Object: &model.Region{Aids: tmp}, Expiration: d.mcExpire, Flags: gmc.FlagProtobuf}
	if err = conn.Set(item); err != nil {
		log.Error("SetRegionCache error(%v)", err)
	}
	return
}

// RegionCache get region archive from cache.
func (d *Dao) RegionCache(c context.Context) (rs map[int32][]int64) {
	conn := d.mc.Get(c)
	defer conn.Close()
	res, err := conn.Get(_keyRegionArcs)
	if err != nil {
		if err != gmc.ErrNotFound {
			log.Error("conn.Get(%s) error(%v)", _keyRegionArcs, err)
		}
		return
	}
	rc := &model.Region{}
	if err = conn.Scan(res, rc); err != nil {
		log.Error("conn.Scan error(%v)", err)
		return
	}
	rs = make(map[int32][]int64)
	for k, v := range rc.Aids {
		rs[k] = v.IDs
	}
	return
}

// SetTagCache set region tag archvie to cache.
func (d *Dao) SetTagCache(c context.Context, regionTagArcs map[string][]int64) (err error) {
	conn := d.mc.Get(c)
	defer conn.Close()
	tmp := make(map[string]*model.Aids)
	for k, v := range regionTagArcs {
		tmp[k] = &model.Aids{IDs: v}
	}
	item := &gmc.Item{Key: _keyRegionTagArcs, Object: &model.Tag{Aids: tmp}, Expiration: d.mcExpire, Flags: gmc.FlagProtobuf}
	if err = conn.Set(item); err != nil {
		log.Error("SetRegionCache error(%v)", err)
	}
	return
}

// TagCache get region tag archive from cache.
func (d *Dao) TagCache(c context.Context) (rs map[string][]int64) {
	conn := d.mc.Get(c)
	defer conn.Close()
	res, err := conn.Get(_keyRegionTagArcs)
	if err != nil {
		if err != gmc.ErrNotFound {
			log.Error("conn.Get(%s) error(%v)", _keyRegionTagArcs, err)
		}
		return
	}
	tc := &model.Tag{}
	if err = conn.Scan(res, tc); err != nil {
		log.Error("conn.Scan error(%v)", err)
		return
	}
	rs = make(map[string][]int64)
	for k, v := range tc.Aids {
		rs[k] = v.IDs
	}
	return
}

func (d *Dao) SetRegionBusinessCache(ctx context.Context, data map[string][]int64) error {
	conn := d.mc.Get(ctx)
	defer conn.Close()
	item := &gmc.Item{Key: _keyRegionBusinessArcs, Object: data, Expiration: d.mcExpire, Flags: gmc.FlagProtobuf}
	if err := conn.Set(item); err != nil {
		log.Error("SetRegionBusinessCache error(%v)", err)
		return err
	}
	return nil
}

func (d *Dao) RegionBusinessCache(ctx context.Context) map[string][]int64 {
	conn := d.mc.Get(ctx)
	defer conn.Close()
	res, err := conn.Get(_keyRegionBusinessArcs)
	if err != nil {
		if err != gmc.ErrNotFound {
			log.Error("RegionBusinessCache conn.Get(%s) error(%v)", _keyRegionBusinessArcs, err)
		}
		return nil
	}
	data := make(map[string][]int64)
	if err = conn.Scan(res, data); err != nil {
		log.Error("RegionBusinessCache conn.Scan error(%v)", err)
		return nil
	}
	return data
}
