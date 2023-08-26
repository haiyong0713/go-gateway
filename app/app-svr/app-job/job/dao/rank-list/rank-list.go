package ranklist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-job/job/conf"
	model "go-gateway/app/app-svr/app-job/job/model/rank-list"
	showmodel "go-gateway/app/app-svr/app-show/interface/model/rank-list"
)

var (
	_rankListMeta = "/x/admin/feed/open/rank/list"
)

func keyRankLinkMeta(id int64) string {
	return fmt.Sprintf("rank_list_meta_%d", id)
}

// Dao is
type Dao struct {
	c      *conf.Config
	redis  *redis.Redis
	client *httpx.Client
}

// New rank dao.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		client: httpx.NewClient(c.HTTPClient),
		redis:  redis.NewRedis(c.Redis.Recommend.Config),
	}
	return d
}

// ScanRankMeta is
func (d *Dao) ScanRankMeta(ctx context.Context, size, page int64) (*model.ListPagination, error) {
	reply := struct {
		Code int                  `json:"code"`
		Data model.ListPagination `json:"data"`
		Msg  string               `json:"msg"`
	}{}
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("size", strconv.FormatInt(size, 10))
	params.Set("page", strconv.FormatInt(page, 10))
	if err := d.client.Get(ctx, d.c.Host.Manager+_rankListMeta, ip, params, &reply); err != nil {
		return nil, err
	}
	return &reply.Data, nil
}

// SetCacheRankMeta is
func (d *Dao) SetCacheRankMeta(ctx context.Context, metas ...*showmodel.Meta) error {
	pipe := d.redis.Pipeline()
	for _, meta := range metas {
		b, err := json.Marshal(meta)
		if err != nil {
			continue
		}
		pipe.Send("SET", keyRankLinkMeta(meta.ID), b)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}
