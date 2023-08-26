package medialist

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	medialist "go-gateway/app/app-svr/app-feed/admin/model/media_list"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	mediaListUrl = "/x/admin/medialist/show"
)

// Dao .
type Dao struct {
	c      *conf.Config
	client *bm.Client
	host   string
}

// New .
func New(c *conf.Config) *Dao {
	return &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient.MediaList),
		host:   c.Host.Manager,
	}
}

// MediaListInfo  get medialist info by media_id
func (d *Dao) MediaListInfo(c context.Context, id int64) (ret *medialist.MediaListInfo, err error) {
	var (
		ip     = metadata.String(c, metadata.RemoteIP)
		uri    = d.host + mediaListUrl
		params = url.Values{}

		res struct {
			Code int                      `json:"code"`
			Data *medialist.MediaListInfo `json:"data"`
		}
	)

	params.Add("media_id", strconv.FormatInt(id, 10))

	if err = d.client.Get(c, uri, ip, params, &res); err != nil {
		log.Error("MediaListInfo Req(%v) error(%v) res(%+v)", id, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, uri+"?"+params.Encode(), err.Error())
	}

	if res.Code != ecode.OK.Code() || res.Data == nil {
		log.Error("MediaListInfo Req(%v) error(%v) res(%+v)", id, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.MediaList, uri+"?"+params.Encode())
	}

	return res.Data, nil
}
