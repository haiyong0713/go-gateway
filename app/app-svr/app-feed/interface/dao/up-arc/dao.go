package uparc

import (
	"context"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

type Dao struct {
	uparc api.UpArchiveClient
}

// New new a archive dao.
func New(c *conf.Config) *Dao {
	d := &Dao{}
	uparc, err := api.NewClient(c.AccountGRPC)
	if err != nil {
		panic(err)
	}
	d.uparc = uparc
	return d
}

// ArcPassedStory is
func (d *Dao) ArcPassedStory(ctx context.Context, in *api.ArcPassedStoryReq) (*api.ArcPassedStoryReply, error) {
	return d.uparc.ArcPassedStory(ctx, in)
}
