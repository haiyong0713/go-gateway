package topic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	fav "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/topic"

	dyntopicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"

	"github.com/pkg/errors"
)

const _topic = "/x/internal/v2/fav/topic"

// Dao is topic dao
type Dao struct {
	client       *httpx.Client
	topic        string
	favRPC       fav.FavoriteClient
	dynTopicGRPC dyntopicapi.TopicClient
}

// New initial topic dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient),
		topic:  c.Host.APICo + _topic,
	}
	var err error
	d.favRPC, err = fav.NewClient(c.FavClient)
	if err != nil {
		panic(fmt.Sprintf("fav NewClient error(%v)", err))
	}
	d.dynTopicGRPC, err = dyntopicapi.NewClient(c.DynTopicGRPC)
	if err != nil {
		panic(fmt.Sprintf("dyntopicapi NewClient error(%v)", err))
	}
	return
}

func (d Dao) SubTopics(ctx context.Context, args *dyntopicapi.SubTopicsReq) (*dyntopicapi.SubTopicsRsp, error) {
	return d.dynTopicGRPC.SubTopics(ctx, args)
}

// Topic get topic list from UGC api.
func (d *Dao) Topic(c context.Context, accessKey, actionKey, device, mobiApp, platform string, build, ps, pn int, mid int64) (t *topic.Topic, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", accessKey)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("actionKey", actionKey)
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("ps", strconv.Itoa(ps))
	params.Set("pn", strconv.Itoa(pn))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int          `json:"code"`
		Data *topic.Topic `json:"data"`
	}
	if err = d.client.Get(c, d.topic, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.topic+"?"+params.Encode())
		return
	}
	t = res.Data
	return
}

// UserFolder is
func (d *Dao) UserFolder(c context.Context, mid int64, typ int32) (userFolder *fav.UserFolderReply, err error) {
	if userFolder, err = d.favRPC.UserFolder(c, &fav.UserFolderReq{Typ: typ, Mid: mid}); err != nil {
		log.Error("d.favRPC.UserFolder error(%+v)", err)
		return
	}
	return
}
