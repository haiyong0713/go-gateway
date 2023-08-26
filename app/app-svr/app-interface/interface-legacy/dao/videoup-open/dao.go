package videoup_open

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	voApi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

// Dao is account dao.
type Dao struct {
	voGRPC voApi.VideoUpOpenClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.voGRPC, err = voApi.NewClient(c.VideoupOpenClient); err != nil {
		panic(fmt.Sprintf("voApi.NewClient error(%v)", err))
	}
	return
}

// AndroidCreative get android creative
func (d *Dao) AndroidCreative(c context.Context, mid int64, build int) (res *space.SectionV2, err error) {
	var (
		req   = &voApi.AppPreReqV2{Uid: mid, Build: int64(build)}
		reply *voApi.AppPreReplyV3
	)
	if reply, err = d.voGRPC.AndAppPreV3(c, req); err != nil {
		log.Error("d.voGRPC.AndAppPreV3 err(%v)", err)
		return
	}
	if reply != nil {
		res = &space.SectionV2{
			Items:     d.buildItems(reply.CreativeItems),
			UpTitle:   reply.UpTitle,
			BeUpTitle: reply.BeUPTitle,
			TipIcon:   reply.Icon,
			TipTitle:  reply.Title,
		}
	}
	return
}

// IOSCreative get android creative
func (d *Dao) IOSCreative(c context.Context, mid int64, build int) (res *space.SectionV2, err error) {
	var (
		req   = &voApi.AppPreReqV2{Uid: mid, Build: int64(build)}
		reply *voApi.AppPreReplyV3
	)
	if reply, err = d.voGRPC.IosAppPreV3(c, req); err != nil {
		log.Error("d.voGRPC.IosAppPreV3 err(%v)", err)
		return
	}
	if reply != nil {
		res = &space.SectionV2{
			Items:     d.buildItems(reply.CreativeItems),
			UpTitle:   reply.UpTitle,
			BeUpTitle: reply.BeUPTitle,
			TipTitle:  reply.Title,
			TipIcon:   reply.Icon,
		}
	}
	return
}

func (d *Dao) buildItems(items []*voApi.SectionItem) (res []*space.SectionItem) {
	if len(items) == 0 {
		return
	}
	for _, v := range items {
		tmp := &space.SectionItem{
			ID:           int64(v.Id),
			Title:        v.Title,
			URI:          v.Uri,
			Icon:         v.Icon,
			NeedLogin:    int8(v.NeedLogin),
			RedDot:       int8(v.RedDot),
			GlobalRedDot: int8(v.GlobalRedDot),
			Display:      v.Display,
		}
		res = append(res, tmp)
	}
	return
}

func (d *Dao) Creative(c context.Context, mid int64) (isUp, show int, err error) {
	var (
		req   = &voApi.AppPreReq{Mid: mid}
		reply *voApi.AppPreReply
	)
	if reply, err = d.voGRPC.AppPre(c, req); err != nil {
		log.Error("d.voGRPC.AppPre err(%v)", err)
		return
	}
	isUp = int(reply.GetIsUp())
	show = int(reply.GetShow())
	return
}
