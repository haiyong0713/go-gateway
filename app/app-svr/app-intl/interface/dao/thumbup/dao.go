package thumbup

import (
	"context"

	api "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/pkg/errors"
)

// Dao is tag dao
type Dao struct {
	thumbupClient api.ThumbupClient
}

// New initial tag dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.thumbupClient, err = api.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	return
}

// HasLike user has like
func (d *Dao) HasLike(c context.Context, mid int64, business string, messageIDs []int64) (res map[int64]api.State, err error) {
	var reply *api.HasLikeReply
	ip := metadata.String(c, metadata.RemoteIP)
	// arg := &thumbup.ArgHasLike{Mid: mid, MessageIDs: messageIDs, Business: _businessLike, RealIP: ip}
	arg := &api.HasLikeReq{
		Business:   business,
		MessageIds: messageIDs,
		Mid:        mid,
		IP:         ip,
	}
	if reply, err = d.thumbupClient.HasLike(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = make(map[int64]api.State)
	for k, v := range reply.States {
		res[k] = v.State
	}
	return
}
