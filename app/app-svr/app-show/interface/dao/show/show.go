package show

import (
	"context"

	"go-common/library/cache/redis"
	jobApi "go-gateway/app/app-svr/app-job/job/api"
	"go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

const (
	_loadShowKey    = "loadShowCache"
	_loadShowTmpKey = "loadShowTempCache"
)

// Heads get show head data.
func (d *Dao) Heads(ctx context.Context) (map[int8][]*show.Head, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey(_loadShowKey, "ShowHdmReply")))
	if err != nil {
		return nil, err
	}
	raw := &jobApi.ShowHdmReply{}
	if err = raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	res := map[int8][]*show.Head{}
	for _, hdm := range raw.Hdm {
		for _, hd := range hdm.Heads {
			head := show.Head{}
			head.FromJobPBHead(hd)
			res[int8(hdm.Key)] = append(res[int8(hdm.Key)], &head)
		}
	}
	return res, nil
}

// Items get item data.
func (d *Dao) Items(ctx context.Context) (map[int][]*show.Item, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey(_loadShowKey, "ShowItmReply")))
	if err != nil {
		return nil, err
	}
	raw := &jobApi.ShowItmReply{}
	if err = raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	res := map[int][]*show.Item{}
	for _, itm := range raw.Itm {
		for _, it := range itm.Items {
			item := show.Item{}
			item.FromJobPBItem(it)
			res[int(itm.Key)] = append(res[int(itm.Key)], &item)
		}
	}
	return res, nil
}

// TempHeads get show temp head data.
func (d *Dao) TempHeads(ctx context.Context) (map[int8][]*show.Head, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey(_loadShowTmpKey, "ShowHdmReply")))
	if err != nil {
		return nil, err
	}
	raw := &jobApi.ShowHdmReply{}
	if err := raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	res := map[int8][]*show.Head{}
	for _, hdm := range raw.Hdm {
		for _, hd := range hdm.Heads {
			head := show.Head{}
			head.FromJobPBHead(hd)
			res[int8(hdm.Key)] = append(res[int8(hdm.Key)], &head)
		}
	}
	return res, nil
}

// TempItems get temp item data.
func (d *Dao) TempItems(ctx context.Context) (map[int][]*show.Item, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", showActionKey(_loadShowTmpKey, "ShowItmReply")))
	if err != nil {
		return nil, err
	}
	raw := &jobApi.ShowItmReply{}
	if err := raw.Unmarshal(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	res := map[int][]*show.Item{}
	for _, itm := range raw.Itm {
		for _, it := range itm.Items {
			item := show.Item{}
			item.FromJobPBItem(it)
			res[int(itm.Key)] = append(res[int(itm.Key)], &item)
		}
	}
	return res, nil
}
