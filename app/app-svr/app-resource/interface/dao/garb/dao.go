package garb

import (
	"context"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	garb "git.bilibili.co/bapis/bapis-go/garb/service"
)

type Dao struct {
	garbClient garb.GarbClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.garbClient, err = garb.NewClient(c.GarbClient); err != nil {
		panic(err)
	}
	return
}

// SkinList .
func (d *Dao) SkinList(c context.Context, ids []int64) (*garb.SkinListReply, error) {
	return d.garbClient.SkinList(c, &garb.SkinListReq{IDs: ids})
}

// SkinUserEquip .
func (d *Dao) SkinUserEquip(c context.Context, mid int64) (*garb.SkinUserEquipReply, error) {
	return d.garbClient.SkinUserEquip(c, &garb.SkinUserEquipReq{Mid: mid})
}

// SkinColorUserList .
func (d *Dao) SkinColorUserList(c context.Context, mid, build int64, mobiApp string) (*garb.SkinColorUserListReply, error) {
	return d.garbClient.SkinColorUserList(c, &garb.SkinColorUserListReq{Mid: mid, MobiAPP: mobiApp, Build: build})
}

// LoadingUserEquip .
func (d *Dao) LoadingUserEquip(c context.Context, mid int64) (*garb.LoadingUserEquipReply, error) {
	return d.garbClient.LoadingUserEquip(c, &garb.LoadingUserEquipReq{Mid: mid})
}
