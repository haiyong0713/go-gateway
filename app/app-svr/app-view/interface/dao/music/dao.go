package music

import (
	"context"
	"fmt"
	url2 "net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"

	api "git.bilibili.co/bapis/bapis-go/crm/service/music-publicity-interface/toplist"
	"go-gateway/app/app-svr/app-view/interface/conf"
	musicmdl "go-gateway/app/app-svr/app-view/interface/model/music"

	"github.com/pkg/errors"
)

var (
	_entranceURI = "/x/copyright-music-publicity/bgm/entrance"
)

type Dao struct {
	// http client
	client *bm.Client
	c      *conf.Config
	// grpc client
	musicClient api.ToplistClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		c:      c,
		client: bm.NewClient(c.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
	}
	var err error
	if d.musicClient, err = api.NewClientToplist(c.ReplyClient); err != nil {
		panic(fmt.Sprintf("reply NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) BgmEntrance(c context.Context, aid, cid int64, platform string) (*musicmdl.Entrance, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int                `json:"code"`
		Data *musicmdl.Entrance `json:"data"`
	}
	params := url2.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("cid", strconv.FormatInt(cid, 10))
	params.Set("platform", platform)
	url := fmt.Sprintf("%s%s", d.c.Host.APICo, _entranceURI)
	if err := d.client.Get(c, url, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), url+"?"+params.Encode())
	}
	return res.Data, nil
}

func (d *Dao) ToplistEntrance(c context.Context, aid int64, musicId string) (*api.ToplistEntranceReply, error) {
	req := &api.ToplistEntranceReq{
		Aid:     aid,
		MusicID: musicId,
	}
	res, err := d.musicClient.ToplistEntrance(c, req)
	if err != nil {
		log.Error("ToplistEntrance fail: aid:%d,err:%+v", aid, err)
		return nil, err
	}
	if res == nil {
		log.Error("ToplistEntrance resp is nil: aid:%d", aid)
		return nil, ecode.NothingFound
	}

	return res, nil
}
