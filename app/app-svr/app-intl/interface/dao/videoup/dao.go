package videoup

import (
	"context"
	"fmt"

	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	"github.com/pkg/errors"
)

// Dao is videoup dao
type Dao struct {
	// grpc
	videoupGRPC vuapi.VideoUpOpenClient
}

// New videoup dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.videoupGRPC, err = vuapi.NewClient(c.VideoupClient)
	if err != nil {
		panic(fmt.Sprintf("videoup NewClient error(%v)", err))
	}
	return
}

func (d *Dao) ArcViewAddit(c context.Context, aid int64) (res *vuapi.ArcViewAdditReply, err error) {
	if res, err = d.videoupGRPC.ArcViewAddit(c, &vuapi.ArcViewAdditReq{Aid: aid}); err != nil {
		err = errors.Wrapf(err, "d.videoupGRPC.ArcViewAddit err aid(%d)", aid)
		return
	}
	return
}
