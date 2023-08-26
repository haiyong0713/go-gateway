package thumbup

import (
	"context"

	"git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-channel/interface/conf"

	"github.com/pkg/errors"
)

const (
	_businessLike = "archive"
)

// Dao is tag dao
type Dao struct {
	thumbupClient api.ThumbupClient
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.thumbupClient, err = api.NewClient(c.ThumbupGRPC); err != nil {
		panic(err)
	}
	return
}

// HasLike user has like
func (d *Dao) HasLike(c context.Context, mid int64, messageIDs []int64) (res map[int64]int8, err error) {
	var reply *api.HasLikeReply
	ip := metadata.String(c, metadata.RemoteIP)
	// arg := &thumbup.ArgHasLike{Mid: mid, MessageIDs: messageIDs, Business: _businessLike, RealIP: ip}
	arg := &api.HasLikeReq{
		Business:   _businessLike,
		MessageIds: messageIDs,
		Mid:        mid,
		IP:         ip,
	}
	if reply, err = d.thumbupClient.HasLike(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = make(map[int64]int8)
	for k, v := range reply.States {
		res[k] = int8(v.State)
	}
	return
}
