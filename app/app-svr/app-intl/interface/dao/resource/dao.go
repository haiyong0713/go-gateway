package resource

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	rscrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	// rpc
	rscRPC *rscrpc.Service
	// grpc
	rpcClient resourcegrpc.ResourceClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// rpc
		rscRPC: rscrpc.New(c.ResourceRPC),
	}
	var err error
	if d.rpcClient, err = resourcegrpc.NewClient(nil); err != nil {
		panic(fmt.Sprintf("resourcegrpc NewClientt error (%+v)", err))
	}
	return
}

// PlayerIcon is.
func (d *Dao) PlayerIcon(c context.Context, aid int64, tagIds []int64, typeId int32) (res *model.PlayerIcon, err error) {
	if res, err = d.rscRPC.PlayerIcon2(c, &model.ArgPlayIcon{Aid: aid, TagIds: tagIds, TypeId: typeId}); err != nil {
		if ecode.Cause(err) == ecode.NothingFound {
			res, err = nil, nil
		}
	}
	return
}

// PasterCID get all paster cid.
func (d *Dao) PasterCID(c context.Context) (cids map[int64]int64, err error) {
	return d.rscRPC.PasterCID(c)
}

// Banner resource banner list
func (d *Dao) Banner(c context.Context, plat int8, build int, mid int64, resIDs, channel, buvid, network, mobiApp, device string, isAd bool, openEvent, adExtra, hash string) (res map[int][]*model.Banner, version string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &model.ArgBanner{Plat: plat, ResIDs: resIDs, Build: build, MID: mid, Channel: channel, IP: ip, Buvid: buvid, Network: network, MobiApp: mobiApp, Device: device, IsAd: isAd, OpenEvent: openEvent, AdExtra: adExtra, Version: hash}
	bs, err := d.rscRPC.Banners(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if bs != nil {
		res = bs.Banner
		version = bs.Version
	}
	return
}

func (d *Dao) AppAudit(ctx context.Context) (res map[string]map[int]struct{}, err error) {
	res = map[string]map[int]struct{}{}
	audit, err := d.rpcClient.AppAudit(ctx, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if audit == nil {
		return
	}
	for _, v := range audit.List {
		build := int(v.Build)
		if plat, ok := res[v.MobiApp]; ok {
			plat[build] = struct{}{}
		} else {
			res[v.MobiApp] = map[int]struct{}{
				build: {},
			}
		}
	}
	return
}

func (d *Dao) Menus(c context.Context) (res []*operate.Menu, err error) {
	menu, err := d.rpcClient.Menu(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if menu == nil {
		return
	}
	for _, v := range menu.List {
		m := &operate.Menu{
			TabID:       v.TabId,
			Plat:        int(v.Plat),
			Name:        v.Name,
			CType:       int(v.CType),
			CValue:      v.CValue,
			PlatVersion: v.PlatVersion,
			STime:       v.STime,
			ETime:       v.ETime,
			Status:      int(v.Status),
			Color:       v.Color,
			Badge:       v.Badge,
		}
		if m.CValue != "" {
			m.Change()
			res = append(res, m)
		}
	}
	return
}

func (d *Dao) Actives(c context.Context) (res []*operate.Active, err error) {
	active, err := d.rpcClient.Active(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if active == nil {
		return
	}
	for _, v := range active.List {
		ac := &operate.Active{
			ID:         v.Id,
			ParentID:   v.ParentID,
			Name:       v.Name,
			Background: v.Background,
			Type:       v.Type,
			Content:    v.Content,
		}
		ac.Change()
		res = append(res, ac)
	}
	return
}

func (d *Dao) ConvergeCards(c context.Context) (res map[int64]*operate.Converge, err error) {
	cards, err := d.rpcClient.Converge(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if cards == nil {
		return
	}
	res = map[int64]*operate.Converge{}
	for _, v := range cards.List {
		c := &operate.Converge{
			ID:      v.Id,
			ReType:  int(v.ReType),
			ReValue: v.ReValue,
			Title:   v.Title,
			Cover:   v.Cover,
			Content: v.Content,
		}
		c.Change()
		res[c.ID] = c
	}
	return
}

func (d *Dao) SpecialCards(c context.Context) (res map[int64]*operate.Special, err error) {
	cards, err := d.rpcClient.Special(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if cards == nil {
		return
	}
	res = map[int64]*operate.Special{}
	for _, v := range cards.List {
		sc := &operate.Special{
			ID:             v.Id,
			Title:          v.Title,
			Desc:           v.Desc,
			Cover:          v.Cover,
			SingleCover:    v.SingleCover,
			GifCover:       v.GifCover,
			BgCover:        v.BgCover,
			Reason:         v.Reason,
			TabURI:         v.TabUri,
			ReType:         int(v.ReType),
			ReValue:        v.ReValue,
			Badge:          v.Badge,
			Size:           v.Size_,
			PowerPicSun:    v.PowerPicSun,
			PowerPicNight:  v.PowerPicNight,
			PowerPicWidth:  v.PowerPicWidth,
			PowerPicHeight: v.PowerPicHeight,
		}
		sc.Change()
		res[sc.ID] = sc
	}
	return
}
