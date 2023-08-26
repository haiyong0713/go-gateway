package creative

import (
	"context"
	"fmt"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-view/interface/conf"

	upApi "git.bilibili.co/bapis/bapis-go/archive/service/up"
	"github.com/pkg/errors"
)

const (
	_special = "/x/internal/uper/special"
	_follow  = "/x/internal/uper/switch"
	_bgm     = "/x/internal/creative/archive/bgm"
	_points  = "/x/internal/creative/video/viewpoints"
)

// Dao is space dao
type Dao struct {
	client  *httpx.Client
	special string
	follow  string
	bgm     string
	points  string
	// grpc
	upClient upApi.UpClient
}

// New initial space dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:  httpx.NewClient(c.HTTPClient),
		special: c.Host.APICo + _special,
		follow:  c.Host.APICo + _follow,
		bgm:     c.Host.APICo + _bgm,
		points:  c.Host.APICo + _points,
	}
	var err error
	if d.upClient, err = upApi.NewClient(c.UpClient); err != nil {
		panic(fmt.Sprintf("UpClient not found err(%v)", err))
	}
	return
}

// Special is
func (d *Dao) Special(c context.Context, gid int64) (midsM map[int64]struct{}, err error) {
	var (
		rep *upApi.UpGroupMidsReply
		req = &upApi.UpGroupMidsReq{GroupID: gid, Pn: 1, Ps: 1000}
	)
	midsM = make(map[int64]struct{})
	if rep, err = d.upClient.UpGroupMids(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if rep != nil {
		for _, mid := range rep.Mids {
			midsM[mid] = struct{}{}
		}
	}
	return
}

// FollowSwitch .
func (d *Dao) FollowSwitch(c context.Context, vmid int64) (upSwitch *upApi.UpSwitchReply, err error) {
	if upSwitch, err = d.upClient.UpSwitch(c, &upApi.UpSwitchReq{Mid: vmid, From: 0}); err != nil {
		log.Error("d.upClient.UpSwitch mid(%d) err(%+v)", vmid, err)
		return
	}
	return
}

func (d *Dao) UpDownloadSwitch(ctx context.Context, vmid int64) (*upApi.UpSwitchReply, error) {
	upSwitch, err := d.upClient.UpSwitch(ctx, &upApi.UpSwitchReq{Mid: vmid, From: 9})
	if err != nil {
		return nil, err
	}
	return upSwitch, nil
}
