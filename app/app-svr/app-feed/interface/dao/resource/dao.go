package resource

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	rscrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	feedMgr "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"

	"github.com/pkg/errors"
)

type Dao struct {
	c *conf.Config
	// rpc
	rscRPC *rscrpc.Service
	// grpc
	rpcClient    resourcegrpc.ResourceClient
	rscV2Client  resourceV2grpc.ResourceClient
	tunnelClient feedMgr.FeedMgrServiceClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// rpc
		rscRPC: rscrpc.New(c.ResourceRPC),
	}
	var err error
	if d.rpcClient, err = resourcegrpc.NewClient(c.ResourceGRPC); err != nil {
		panic(fmt.Sprintf("resourcegrpc NewClientt error (%+v)", err))
	}
	if d.rscV2Client, err = resourceV2grpc.NewClient(c.ResourceV2GRPC); err != nil {
		panic(err)
	}
	if d.tunnelClient, err = feedMgr.NewClientFeedMgrService(c.TunnelV2Client); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Banner(c context.Context, plat int8, build int, mid int64, resIDs, channel, buvid, network, mobiApp, device string, isAd bool, openEvent, adExtra, hash string, splashID int64) (res map[int][]*model.Banner, version string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &model.ArgBanner{
		Plat:      plat,
		ResIDs:    resIDs,
		Build:     build,
		MID:       mid,
		Channel:   channel,
		IP:        ip,
		Buvid:     buvid,
		Network:   network,
		MobiApp:   mobiApp,
		Device:    device,
		IsAd:      isAd,
		OpenEvent: openEvent,
		AdExtra:   adExtra,
		Version:   hash,
		SplashID:  splashID,
	}
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

// FeedBanners is
func (d *Dao) FeedBanners(ctx context.Context, req *resourcegrpc.FeedBannersRequest) (*resourcegrpc.FeedBannersReply, error) {
	return d.rpcClient.FeedBanners(ctx, req)
}

// AbTest resource abtest
func (d *Dao) AbTest(ctx context.Context, groups string) (res map[string]*model.AbTest, err error) {
	arg := &model.ArgAbTest{
		Groups: groups,
	}
	if res, err = d.rscRPC.AbTest(ctx, arg); err != nil {
		log.Error("resource d.resRpc.AbTest error(%v)", err)
		return
	}
	return
}

func (d *Dao) AppAudit(ctx context.Context) (res map[string]map[int]struct{}, err error) {
	audit, err := d.rpcClient.AppAudit(ctx, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if audit == nil {
		return
	}
	res = map[string]map[int]struct{}{}
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

func (d *Dao) Follow(c context.Context) (res map[int64]*operate.Follow, err error) {
	follow, err := d.rpcClient.CardFollow(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if follow == nil {
		return
	}
	res = map[int64]*operate.Follow{}
	for _, v := range follow.List {
		c := &operate.Follow{
			ID:      v.Id,
			Type:    v.Type,
			Title:   v.Title,
			Content: v.Content,
		}
		c.Change()
		res[c.ID] = c
	}
	return
}

func (d *Dao) Menus(c context.Context, plat int8, build int) ([]*operate.Menu, error) {
	menu, err := d.rpcClient.AppMenu(c, &resourcegrpc.AppMenusRequest{Plat: int32(plat), Build: int32(build)})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if menu == nil {
		return nil, ecode.NothingFound
	}
	var res []*operate.Menu
	for _, v := range menu.List {
		m := &operate.Menu{
			TabID: v.TabId,
			Name:  v.Name,
			Img:   v.Img,
			Icon:  v.Icon,
			Color: v.Color,
			ID:    v.Id,
		}
		res = append(res, m)
	}
	return res, nil
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

func (d *Dao) DownLoadCards(c context.Context) (res map[int64]*operate.Download, err error) {
	cards, err := d.rpcClient.DownLoad(c, &resourcegrpc.NoArgRequest{})
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if cards == nil {
		return
	}
	res = map[int64]*operate.Download{}
	for _, v := range cards.List {
		d := &operate.Download{
			ID:          v.Id,
			Title:       v.Title,
			Desc:        v.Desc,
			Icon:        v.Icon,
			Cover:       v.Cover,
			URLType:     int(v.UrlType),
			URLValue:    v.UrlValue,
			BtnTxt:      int(v.BtnTxt),
			ReType:      int(v.ReType),
			ReValue:     v.ReValue,
			Number:      v.Number,
			DoubleCover: v.DoubleCover,
		}
		d.Change()
		res[d.ID] = d
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

func (d *Dao) CardPosRecs(c context.Context, ids []int64) (map[int64]*resourcegrpc.CardPosRec, error) {
	reply, err := d.rpcClient.CardPosRecs(c, &resourcegrpc.CardPosRecReplyRequest{CardIds: ids})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.Card, nil
}

func (d *Dao) AppActive(c context.Context, id int64) ([]*operate.Active, string, error) {
	reply, err := d.rpcClient.AppActive(c, &resourcegrpc.AppActiveRequest{Id: id})
	if err != nil {
		log.Error("%+v", err)
		return nil, "", err
	}
	acs := []*operate.Active{}
	for _, v := range reply.List {
		ac := &operate.Active{
			ID:         v.Id,
			ParentID:   v.ParentID,
			Name:       v.Name,
			Background: v.Background,
			Type:       v.Type,
			Content:    v.Content,
		}
		ac.Change()
		acs = append(acs, ac)
	}
	return acs, reply.Cover, nil
}

func (d *Dao) MultiMaterials(ctx context.Context, ids []int64) (map[int64]*feedMgr.Material, error) {
	reply, err := d.tunnelClient.GetMaterial(ctx, &feedMgr.MaterialReq{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[int64]*feedMgr.Material, len(reply.Material))
	for _, material := range reply.Material {
		out[material.Id] = material
	}
	return out, nil
}

func (d *Dao) SpecialV2(ctx context.Context, ids []int64) (map[int64]*resourceV2grpc.AppSpecialCard, error) {
	reply, err := d.rscV2Client.GetSpecialCard(ctx, &resourceV2grpc.SpecialCardReq{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}
	return reply.GetSpecialCard(), nil
}
