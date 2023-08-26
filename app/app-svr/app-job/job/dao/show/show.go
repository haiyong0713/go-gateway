package show

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-job/job/api"

	"github.com/pkg/errors"
)

const (
	// real data
	_headSQL = "SELECT s.id,s.plat,s.title,s.type,s.param,s.style,s.rank,s.build,s.conditions,l.name FROM show_head AS s," +
		"language AS l WHERE l.id=s.lang_id ORDER BY rank DESC"
	_itemSQL = "SELECT sid,title,random,cover,param FROM show_item"
	// temp preview
	_headTmpSQL = "SELECT s.id,s.plat,s.title,s.type,s.param,s.style,s.rank,s.build,s.conditions,l.name FROM show_head_temp AS s," +
		"language AS l WHERE l.id=s.lang_id ORDER BY rank DESC"
	_itemTmpSQL = "SELECT sid,title,random,cover,param FROM show_item_temp"
	// redis key
	_loadShowKey    = "loadShowCache"
	_loadShowTmpKey = "loadShowTempCache"
)

// Heads get show head data.
func (d *Dao) Heads(ctx context.Context) (*v1.ShowHdmReply, error) {
	rows, err := d.getHead.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	heads := make(map[int32][]*v1.Head, 20)
	for rows.Next() {
		h := &v1.Head{}
		if err = rows.Scan(&h.Id, &h.Plat, &h.Title, &h.Type, &h.Param, &h.Style, &h.Rank, &h.Build, &h.Condition, &h.Language); err != nil {
			return nil, err
		}
		heads[h.Plat] = append(heads[h.Plat], h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.ShowHdmReply{}
	for k, v := range heads {
		res.Hdm = append(res.Hdm, &v1.ShowHdMap{
			Key:   k,
			Heads: v,
		})
	}
	return res, nil
}

// Items get item data.
func (d *Dao) Items(ctx context.Context) (*v1.ShowItmReply, error) {
	rows, err := d.getItem.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int32][]*v1.Item, 50)
	for rows.Next() {
		i := &v1.Item{}
		if err = rows.Scan(&i.Sid, &i.Title, &i.Random, &i.Cover, &i.Param); err != nil {
			return nil, err
		}
		items[i.Sid] = append(items[i.Sid], i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.ShowItmReply{}
	for k, v := range items {
		res.Itm = append(res.Itm, &v1.ShowItMap{
			Key:   k,
			Items: v,
		})
	}
	return res, nil
}

// TempHeads get show temp head data.
func (d *Dao) TempHeads(ctx context.Context) (*v1.ShowHdmReply, error) {
	rows, err := d.db.Query(ctx, _headTmpSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	heads := make(map[int32][]*v1.Head, 20)
	for rows.Next() {
		h := &v1.Head{}
		if err = rows.Scan(&h.Id, &h.Plat, &h.Title, &h.Type, &h.Param, &h.Style, &h.Rank, &h.Build, &h.Condition, &h.Language); err != nil {
			return nil, err
		}
		heads[h.Plat] = append(heads[h.Plat], h)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.ShowHdmReply{}
	for k, v := range heads {
		res.Hdm = append(res.Hdm, &v1.ShowHdMap{
			Key:   k,
			Heads: v,
		})
	}
	return res, nil
}

// TempItems get temp item data.
func (d *Dao) TempItems(ctx context.Context) (*v1.ShowItmReply, error) {
	rows, err := d.db.Query(ctx, _itemTmpSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make(map[int32][]*v1.Item, 50)
	for rows.Next() {
		i := &v1.Item{}
		if err = rows.Scan(&i.Sid, &i.Title, &i.Random, &i.Cover, &i.Param); err != nil {
			return nil, err
		}
		items[i.Sid] = append(items[i.Sid], i)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	res := &v1.ShowItmReply{}
	for k, v := range items {
		res.Itm = append(res.Itm, &v1.ShowItMap{
			Key:   k,
			Items: v,
		})
	}
	return res, nil
}

func (d *Dao) AddCacheShow(ctx context.Context, hdm *v1.ShowHdmReply, itm *v1.ShowItmReply) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string

	bs, err := hdm.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey(_loadShowKey, "ShowHdmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = itm.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key = showActionKey(_loadShowKey, "ShowItmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)
	if err := conn.Send("MSET", argsMDs...); err != nil {
		return err
	}
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, _showExpire); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) AddTempCacheShow(ctx context.Context, hdm *v1.ShowHdmReply, itm *v1.ShowItmReply) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string

	bs, err := hdm.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey(_loadShowTmpKey, "ShowHdmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)
	if _, err = conn.Do("SETEX", key, _showExpire, bs); err != nil {
		return err
	}

	bs, err = itm.Marshal()
	if err != nil {
		return errors.WithStack(err)
	}
	key = showActionKey(_loadShowTmpKey, "ShowItmReply")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)
	if err := conn.Send("MSET", argsMDs...); err != nil {
		return err
	}
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, _showExpire); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}
