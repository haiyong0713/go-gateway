package school

import (
	"context"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	api "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
)

type Dao struct {
	school api.CampusSvrClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	school, err := api.NewClient(c.MemClient)
	if err != nil {
		panic(err)
	}
	d.school = school
	return d
}

func (d *Dao) CampusInfo(ctx context.Context, req *api.CampusInfoReq) (*api.CampusInfoReply, error) {
	return d.school.CampusInfo(ctx, req)
}
