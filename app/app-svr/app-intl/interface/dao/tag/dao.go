package tag

import (
	"context"

	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// Dao struct
type Dao struct {
	mInfo   string
	client  *httpx.Client
	tagGRPC taggrpc.TagRPCClient
}

// New a dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPTag),
		mInfo:  c.Host.APICo + _mInfo,
	}
	var err error
	if d.tagGRPC, err = taggrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	return
}

// InfoByIDs is.
func (d *Dao) Tags(c context.Context, mid int64, tids []int64) (map[int64]*taggrpc.Tag, error) {
	res, err := d.tagGRPC.Tags(c, &taggrpc.TagsReq{Mid: mid, Tids: tids})
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}
