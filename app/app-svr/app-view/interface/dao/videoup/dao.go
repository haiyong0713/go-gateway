package videoup

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	vuapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/pkg/errors"
)

// Dao is videoup dao
type Dao struct {
	//v2版本
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

func (d *Dao) MaterialView(c context.Context, params *view.MaterialParam) (res []*vuapi.MaterialViewRes, err error) {
	var (
		arg = &vuapi.MaterialViewReq{
			Aid:      params.AID,
			Cid:      params.CID,
			Platform: params.Platform,
			MobiApp:  params.MobiApp,
			Device:   params.Device,
			Build:    params.Build,
		}
		materialViewReply *vuapi.MaterialViewReply
	)
	if materialViewReply, err = d.videoupGRPC.AppMaterialView(c, arg); err != nil {
		log.Error("d.videoupGRPC.AppMaterialView err(%+v)", err)
		return
	}
	res = materialViewReply.Vals
	return
}

func (d *Dao) ArcCommercial(c context.Context, aid int64) (gameId int64, err error) {
	var (
		req = &vuapi.ArcCommercialReq{Aid: aid}
		acr *vuapi.ArcCommercialReply
	)
	if acr, err = d.videoupGRPC.ArcCommercial(c, req); err != nil {
		return
	}
	if acr == nil {
		err = ecode.NothingFound
		return 0, err
	}
	gameId = acr.GameID
	return
}

func (d *Dao) ArcBgmList(c context.Context, aid, cid int64) (res []*viewApi.Bgm, err error) {
	var (
		req = &vuapi.BgmListReq{
			Aid: aid,
			Cid: cid,
		}
		blr *vuapi.BgmListReply
	)
	if blr, err = d.videoupGRPC.ArcBgmList(c, req); err != nil {
		return
	}
	if blr == nil {
		err = ecode.NothingFound
		return nil, err
	}
	for _, v := range blr.Bgms {
		if v == nil {
			continue
		}
		res = append(res, &viewApi.Bgm{
			Sid:     v.Sid,
			Mid:     v.Mid,
			Title:   v.Title,
			Author:  v.Author,
			JumpUrl: v.JumpUrl,
			Cover:   v.Cover,
		})
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

func (d *Dao) GetVideoViewPoints(c context.Context, aid, cid int64) (*vuapi.VideoPointsReply, error) {
	req := vuapi.VideoPointsReq{
		Aid: aid,
		Cid: cid,
	}
	res, err := d.videoupGRPC.VideoViewPoints(c, &req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		log.Error("VideoViewPoints res is nil: aid:%d,cid:%d", aid, cid)
		return nil, ecode.NothingFound
	}
	return res, nil
}

func (d *Dao) GetMaterialList(c context.Context, aid, cid int64) (bgm []*viewApi.Bgm, sticker []*viewApi.ViewMaterial, videoSource []*viewApi.ViewMaterial, err error) {
	req := vuapi.MaterialListReq{
		Aid: aid,
		Cid: cid,
	}
	material, err := d.videoupGRPC.ArcMaterialList(c, &req)
	if err != nil {
		return nil, nil, nil, err
	}
	if material == nil {
		return nil, nil, nil, ecode.NothingFound
	}
	if len(material.Bgm) > 0 {
		for _, v := range material.Bgm {
			bgm = append(bgm, &viewApi.Bgm{
				Sid:     v.Oid,
				Mid:     v.Mid,
				Title:   v.Title,
				Author:  v.Author,
				JumpUrl: v.JumpUrl,
			})
		}
	}
	if len(material.Sticker) > 0 {
		for _, v := range material.Sticker {
			sticker = append(sticker, &viewApi.ViewMaterial{
				Oid:     v.Oid,
				Mid:     v.Mid,
				Title:   v.Title,
				Author:  v.Author,
				JumpUrl: v.JumpUrl,
			})
		}
	}
	if len(material.VideoSource) > 0 {
		//潮点视频
		for _, v := range material.VideoSource {
			videoSource = append(videoSource, &viewApi.ViewMaterial{
				Title: v.Title,
			})
		}
	}
	return
}

func (d *Dao) MultiArchiveArgument(ctx context.Context, req *vuapi.MultiArchiveArgumentReq) (*vuapi.MultiArchiveArgumentReply, error) {
	return d.videoupGRPC.MultiArchiveArgument(ctx, req)
}
