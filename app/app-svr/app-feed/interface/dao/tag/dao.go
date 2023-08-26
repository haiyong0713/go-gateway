package tag

import (
	"context"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/interface/conf"
)

type Dao struct {
	// conf
	// http client
	hot     string
	tags    string
	detail  string
	client  *httpx.Client
	tagGRPC taggrpc.TagRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPTag),
		hot:    c.Host.APICo + _hot,
		tags:   c.Host.APICo + _tags,
		detail: c.Host.APICo + _detail,
	}
	var err error
	if d.tagGRPC, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) TagsInfoByIDs(c context.Context, mid int64, tids []int64) (map[int64]*taggrpc.Tag, error) {
	res, err := d.tagGRPC.Tags(c, &taggrpc.TagsReq{Mid: mid, Tids: tids})
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}
