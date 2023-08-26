package resource

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	resApi "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	rscrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	resApiV3 "git.bilibili.co/bapis/bapis-go/resource/service/v1"
	resApiV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	// rpc
	rscRPC      *rscrpc.Service
	resClient   resApi.ResourceClient
	resV2Client resApiV2.ResourceClient
	resV3Client resApiV3.ResourceClient
	c           *conf.Config
	manager     string
	client      *httpx.Client
}

const _manager = "/x/admin/manager/interface/blackwhite/list/scene/oid_list"

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// rpc
		rscRPC:  rscrpc.New(c.ResourceRPC),
		c:       c,
		manager: c.Host.ManagerHost + _manager,
		client:  httpx.NewClient(c.HTTPClient),
	}
	var err error
	if d.resClient, err = resApi.NewClient(c.ResClient); err != nil {
		panic(fmt.Sprintf("resApi NewClient not found err(%v)", err))
	}
	if d.resV2Client, err = resApiV2.NewClient(c.ResClient); err != nil {
		panic(fmt.Sprintf("resApiV2 NewClient not found err(%v)", err))
	}
	if d.resV3Client, err = resApiV3.NewClient(c.ResClient); err != nil {
		panic(fmt.Sprintf("resApiV3 NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) Paster(c context.Context, plat, adType int8, aid, typeID, buvid string) (res *model.Paster, err error) {
	arg := &model.ArgPaster{Platform: plat, AdType: adType, Aid: aid, TypeId: typeID, Buvid: buvid}
	if res, err = d.rscRPC.PasterAPP(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
	}
	return
}

func (d *Dao) PlayerIcon(c context.Context, aid, mid int64, tagIds []int64, typeId int32, showPlayicon bool, build int, mobiApp, device string) (res *model.PlayerIcon, err error) {
	if res, err = d.rscRPC.PlayerIcon2(c, &model.ArgPlayIcon{Aid: aid, Mid: mid, TagIds: tagIds, TypeId: typeId, ShowPlayicon: showPlayicon, Build: int32(build), MobiApp: mobiApp, Device: device}); err != nil {
		if ecode.Cause(err) == ecode.NothingFound {
			res, err = nil, nil
		}
	}
	return
}

// HasCustomConfig is
func (d *Dao) HasCustomConfig(ctx context.Context, tp int32, oid int64) bool {
	_, err := d.rscRPC.CustomConfig(ctx, &pb.CustomConfigRequest{
		TP:  tp,
		Oid: oid,
	})
	if err != nil {
		log.Error("Failed to get custom config: %d,%d: %+v", tp, oid, err)
		return false
	}
	return true
}

func (d *Dao) GetPlayerCustomizedPanel(ctx context.Context, tids []int64) (*resApi.GetPlayerCustomizedPanelV2Rep, error) {
	req := &resApi.GetPlayerCustomizedPanelReq{
		Tids: tids,
	}
	reply, err := d.resClient.GetPlayerCustomizedPanelV2(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) PlayerIconNew(c context.Context, aid, mid int64, tagIds []int64, typeId int32, showPlayicon bool, build int, mobiApp, device string) (*model.PlayerIconRly, error) {
	req := &resApiV3.PlayerIconRequest{
		Aid:          aid,
		TagIDs:       tagIds,
		TypeID:       typeId,
		Mid:          mid,
		ShowPlayIcon: showPlayicon,
		Build:        int32(build),
		MobiApp:      mobiApp,
		Device:       device,
	}
	res, err := d.resV3Client.PlayerIcon2NewV2(c, req)
	if err != nil {
		return nil, err
	}
	if res.GetItem() == nil {
		return &model.PlayerIconRly{}, nil
	}
	reply := res.GetItem()
	item := &model.PlayerIcon{
		URL1:         reply.GetUrl1(),
		Hash1:        reply.GetHash1(),
		URL2:         reply.GetUrl2(),
		Hash2:        reply.GetHash2(),
		CTime:        reply.GetCtime(),
		Type:         int8(reply.GetType()),
		TypeValue:    reply.GetTypeValue(),
		MTime:        reply.GetMtime(),
		DragLeftPng:  reply.GetDragLeftPng(),
		MiddlePng:    reply.GetMiddlePng(),
		DragRightPng: reply.GetDragRightPng(),
	}
	if reply.GetDragData() != nil {
		item.DragData = &resApi.IconData{
			MetaJson:  reply.GetDragData().MetaJson,
			SpritsImg: reply.GetDragData().SpritsImg,
		}
	}
	if reply.GetNodragData() != nil {
		item.NoDragData = &resApi.IconData{
			MetaJson:  reply.GetNodragData().MetaJson,
			SpritsImg: reply.GetNodragData().SpritsImg,
		}
	}
	return &model.PlayerIconRly{
		Item: item,
	}, nil
}

//nolint:gomnd
func (d *Dao) ViewTab(c context.Context, aid int64, tagIDs, upIDs []int64, typeId, plat, build int32) (*viewApi.Tab, error) {
	req := &resApi.UgcTabReq{
		Tid:   int64(typeId),
		Tag:   tagIDs,
		UpId:  upIDs, // 需求包括联合投稿人
		AvId:  aid,
		Plat:  plat,
		Build: build,
	}
	reply, err := d.resClient.UgcTabV2(c, req)
	if err != nil {
		return nil, err
	}
	if reply.GetItem() == nil {
		return nil, errors.Errorf("ViewTab no online view tab req(%+v)", req)
	}
	res := reply.GetItem()
	tab := &viewApi.Tab{
		Background:        res.Bg,
		Otype:             viewApi.TabOtype(res.LinkType),
		Style:             viewApi.TabStyle(res.TabType),
		TextColor:         res.Color,
		TextColorSelected: res.Selected,
		Id:                res.Id,
	}
	switch res.LinkType {
	case 1: // h5
		if res.Link == "" {
			return nil, errors.Errorf("物料告警 ViewTab h5 invalid link:%s", res.Link)
		}
		tab.Uri = res.Link
	case 2: // native
		oid, _ := strconv.ParseInt(res.Link, 10, 64)
		if oid <= 0 {
			return nil, errors.Errorf("物料告警 ViewTab native invalid link:%s", res.Link)
		}
		tab.Oid = oid
	default:
		return nil, errors.Errorf("物料告警 ViewTab unknown link_type:%d", res.LinkType)
	}
	if res.Tab == "" {
		return nil, errors.Errorf("物料告警 ViewTab invalid tab:%s", res.Tab)
	}
	switch res.TabType {
	case 1: // 文字
		tab.Text = res.Tab
	case 2: // 图片
		tab.Pic = res.Tab
	default:
		return nil, errors.Errorf("物料告警 ViewTab unknown tab_type:%d", res.TabType)
	}
	return tab, nil
}

func (d *Dao) BWList(ctx context.Context, aid int64) bool {
	req := &resApiV2.CheckCommonBWListReq{
		Token: d.c.Mng.EncyclopediaToken,
		Oid:   strconv.FormatInt(aid, 10),
	}
	reply, err := d.resV2Client.CheckCommonBWList(ctx, req)
	if err != nil {
		log.Error("d.resV2Client.CheckCommonBWList aid(%d) err(%+v)", aid, err)
		return false
	}
	return reply.GetIsInList()
}

func (d *Dao) GetSpecialCard(ctx context.Context) ([]*resApiV2.AppSpecialCard, error) {
	req := &resApiV2.NoArgRequest{}
	resp, err := d.resV2Client.GetAppSpecialCard(ctx, req)
	if err != nil {
		log.Error("GetAppSpecialCard err(%+v)", err)
		return nil, err
	}
	if resp == nil {
		return nil, ecode.NothingFound
	}
	return resp.Card, nil
}

func (d *Dao) FetchAllOnlineBlackList(c context.Context) (map[int64]struct{}, error) {
	if d.c.LegoToken == nil {
		return nil, errors.New("d.c.LegoToken.PlayOnlineToken is nil")
	}
	params := url.Values{}
	params.Set("token", d.c.LegoToken.PlayOnlineToken)
	var res struct {
		Code int `json:"code"`
		Data struct {
			Oids []string `json:"oids"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.manager, "", params, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.New(fmt.Sprintf("d.FetchAllOnlineBlackList res.Code(%d)", res.Code))
	}
	blackList := make(map[int64]struct{})
	for _, aidStr := range res.Data.Oids {
		aid, err := strconv.ParseInt(aidStr, 10, 64)
		if err != nil {
			log.Error("d.FetchAllOnlineBlackList strconv.ParseInt error(%+v)", err)
			continue
		}
		blackList[aid] = struct{}{}
	}
	return blackList, nil
}

func (d *Dao) MultiMaterials(ctx context.Context, ids []int64) (map[int64]*resApiV2.Material, error) {
	res := make(map[int64]*resApiV2.Material, len(ids))

	if len(ids) == 0 {
		return res, nil
	}

	reply, err := d.resV2Client.GetMaterial(ctx, &resApiV2.MaterialReq{
		Id: sets.NewInt64(ids...).List(),
	})
	if err != nil {
		log.Error("d.MultiMaterials ids(%+v), error (%+v)", ids, err)
		return nil, err
	}
	for _, material := range reply.Material {
		res[material.Id] = material
	}
	return res, nil
}
