package channel

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	xhttp "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/conf"
	chmdl "go-gateway/app/web-svr/web/interface/model/channel"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"github.com/pkg/errors"
)

const (
	TypWeb        = 1
	SubVersionWeb = 1
	ActionStick   = 2

	_searchChannel    = "/x/admin/search"
	_searchChannelSQL = `deal SELECT * FROM link_channel WHERE name LIKE "%v" AND state=%d LIMIT %d,%d`
)

// Dao web channel dao.
type Dao struct {
	c            *conf.Config
	chClient     changrpc.ChannelRPCClient
	searchClient *xhttp.Client
}

// New new web channel  dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.chClient, err = changrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(fmt.Sprintf("New ChannelRPCClient error (%+v)", err))
	}
	d.searchClient = xhttp.NewClient(c.HTTPClient.Search)
	return
}

func (d *Dao) SubscribedChannel(c context.Context, mid int64) (*changrpc.SubscribeReply, error) {
	var (
		err   error
		req   = &changrpc.SubscribeReq{Mid: mid, SubVersion: SubVersionWeb, Typ: TypWeb}
		reply *changrpc.SubscribeReply
	)
	if reply, err = d.chClient.Subscribe(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.Subscribe(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) NewNotify(c context.Context, mid int64) (*changrpc.NewNotifyReply, error) {
	var (
		err   error
		req   = &changrpc.NewNotifyReq{Mid: mid, Typ: TypWeb}
		reply *changrpc.NewNotifyReply
	)
	if reply, err = d.chClient.NewNotify(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.NewNotify(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) Category(c context.Context) (*changrpc.CategoryReply, error) {
	var (
		err   error
		reply *changrpc.CategoryReply
	)
	if reply, err = d.chClient.Category(c, &changrpc.NoReq{}); err != nil {
		err = errors.Wrapf(err, "d.chClient.Category")
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ViewChannel(c context.Context, mid int64) (*changrpc.ViewChannelReply, error) {
	var (
		err   error
		req   = &changrpc.ViewChannelReq{Mid: mid}
		reply *changrpc.ViewChannelReply
	)
	if reply, err = d.chClient.ViewChannel(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.ViewChannel error(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ChannelList(c context.Context, req *changrpc.ChannelListReq) (*changrpc.ChannelListReply, error) {
	reply, err := d.chClient.ChannelList(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.chClient.ChannelList(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ChannelResourceList(c context.Context, req *changrpc.ChannelResourceListReq) (*changrpc.ChannelResourceListReply, error) {
	var (
		err   error
		reply *changrpc.ChannelResourceListReply
	)
	req.Typ = TypWeb
	if reply, err = d.chClient.ChannelResourceList(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.ChannelResourceList(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) UpdateSubscribe(c context.Context, mid int64, tops, cids string) (err error) {
	req := &changrpc.UpdateSubscribeReq{
		Mid:    mid,
		Action: ActionStick,
		Tops:   tops,
		Cids:   cids,
	}
	if _, err = d.chClient.UpdateSubscribe(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.UpdateSubscribe(%+v)", req)
		return err
	}
	return nil
}

func (d *Dao) ChannelDetail(c context.Context, mid, channelID int64) (*changrpc.ChannelDetailReply, error) {
	var (
		err   error
		req   = &changrpc.ChannelDetailReq{ChannelId: channelID, Mid: mid}
		reply *changrpc.ChannelDetailReply
	)
	if reply, err = d.chClient.ChannelDetail(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.ChannelDetail(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) HotChannel(c context.Context, mid int64, req *changrpc.HotChannelReq) (*changrpc.HotChannelReply, error) {
	var (
		err   error
		reply *changrpc.HotChannelReply
	)
	if reply, err = d.chClient.HotChannel(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.HotChannel(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ResourceList(c context.Context, req *changrpc.ResourceListReq) (*changrpc.ResourceListReply, error) {
	req.Typ = TypWeb
	reply, err := d.chClient.ResourceList(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.chClient.ResourceList(%+v)", req)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) SearchChannel(c context.Context, mid int64, channelIDs []int64) (*changrpc.SearchChannelReply, error) {
	req := &changrpc.SearchChannelReq{Mid: mid, Cids: channelIDs}
	reply, err := d.chClient.SearchChannel(c, req)
	if err != nil {
		err = errors.Wrapf(err, "d.chClient.SearchChannel error")
		return nil, err
	}
	return reply, err
}

func (d *Dao) SearchChannelsInfo(c context.Context, req *changrpc.SearchChannelsInfoReq) (*changrpc.SearchChannelsInfoReply, error) {
	var (
		err   error
		reply *changrpc.SearchChannelsInfoReply
	)
	if reply, err = d.chClient.SearchChannelsInfo(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.SearchChannelsInfo(%+v)", req)
		return nil, err
	}
	return reply, err
}

func (d *Dao) SearchEs(c context.Context, keyword string, pn, ps, state int32) (*chmdl.EsRes, []int64, error) {
	var (
		err     error
		req     *http.Request
		chanIDs []int64
		ip      = metadata.String(c, metadata.RemoteIP)
		sql     = fmt.Sprintf(_searchChannelSQL, keyword, state, (pn-1)*ps, ps)
		uri     = fmt.Sprintf("%s/%s", d.c.Host.API, _searchChannel)
		params  = url.Values{}
		res     = &chmdl.EsRes{}
	)
	params.Set("sql", sql)
	if req, err = d.searchClient.NewRequest("GET", uri, ip, params); err != nil {
		err = errors.Wrapf(err, "d.searchClient.NewRequest(%+v, %+v)", uri, params)
		return nil, nil, err
	}
	if err := d.searchClient.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "d.searchClient.Do(%+v)", req)
		return nil, nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrapf(ecode.Int(res.Code), uri+"?"+params.Encode())
		return nil, nil, err
	}
	if res == nil || res.Data == nil {
		err := errors.New("esResult || esResult.Data is nil")
		return nil, nil, err
	}
	for _, item := range res.Data.Result {
		chanIDs = append(chanIDs, item.ChannelId)
	}
	return res, chanIDs, nil
}

func (d *Dao) RelativeChannel(c context.Context, req *changrpc.RelativeChannelReq) (*changrpc.RelativeChannelReply, error) {
	var (
		err   error
		reply *changrpc.RelativeChannelReply
	)
	if reply, err = d.chClient.RelativeChannel(c, req); err != nil {
		err = errors.Wrapf(err, "d.chClient.RelativeChannel(%+v)", req)
		return nil, err
	}
	return reply, err
}
