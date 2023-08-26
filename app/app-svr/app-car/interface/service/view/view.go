package view

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"
	"go-gateway/app/app-svr/app-car/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"
	thumbup "go-main/app/community/thumbup/service/model"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

const _overseaBlockKey = "oversea_block"

func (s *Service) View(c context.Context, plat int8, mid int64, buvid string, param *view.ViewParam) (*view.View, error) {
	if param.SeasonID > 0 {
		return s.viewPGC(c, mid, buvid, param)
	}
	return s.viewArc(c, plat, mid, buvid, param)
}

func (s *Service) viewArc(c context.Context, plat int8, mid int64, buvid string, param *view.ViewParam) (*view.View, error) { //nolint:gocognit
	var (
		vp                        *api.ViewReply
		his                       *hisApi.ModelHistory
		stat                      *relationgrpc.StatReply
		authorRelations           map[int64]*relationgrpc.InterrelationReply
		isFavored, overseaBlocked bool
		isLike                    int
		asDesc                    string
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		if vp, err = s.arc.View(ctx, param.Aid); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if mid > 0 || buvid != "" {
		group.Go(func(ctx context.Context) (err error) {
			if his, err = s.his.Progress(ctx, param.Aid, mid, buvid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			likeState, err := s.thumbupDao.HasLike(ctx, mid, _businessLike, buvid, param.Aid)
			if err != nil {
				log.Error("%+v", err)
			}
			if likeState == thumbup.StateLike {
				isLike = 1
			}
			return nil
		})
		if mid > 0 {
			group.Go(func(ctx context.Context) error {
				isFavored = s.fav.IsFavored(ctx, mid, param.Aid)
				return nil
			})
		}
	}
	group.Go(func(ctx context.Context) (err error) {
		desc, err := s.arc.Description(ctx, param.Aid)
		if err != nil {
			return nil
		}
		asDesc = desc
		return nil
	})
	// 稿件特殊属性位
	group.Go(func(ctx context.Context) error {
		flowCtrlData, flowErr := s.arc.FlowControlInfoV2(ctx, param.Aid, s.c.FlowControl)
		if flowErr != nil {
			log.Error("%+v", flowErr)
			return nil
		}
		for _, v := range flowCtrlData {
			if v != nil && v.Key == _overseaBlockKey && v.Value == 1 {
				overseaBlocked = true
				break
			}
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		// 如果接口错误统一返回视频不存在
		return nil, xecode.AppNotVedio
	}
	if vp == nil || vp.Arc == nil || len(vp.Pages) == 0 || !vp.Arc.IsNormal() || vp.Arc.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes {
		return nil, xecode.AppNotVedio
	}
	if vp.AttrVal(api.AttrBitSteinsGate) == api.AttrYes { // 互动视频返回error不播放
		return nil, xecode.AppAttrBitSteinsGate
	}
	if overseaBlocked {
		return nil, xecode.AppAreaLimit
	}
	vs := &view.View{Arc: vp.Arc, VideoType: view.BiliType, PagesStyle: view.HorizontalStyle}
	// 获取当前UP主粉丝数
	group = errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		if stat, err = s.reldao.StatGRPC(ctx, vp.Author.Mid); err != nil {
			log.Error("%+v", err)
		}
		return nil
	})
	if mid > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if authorRelations, err = s.reldao.RelationsInterrelations(ctx, mid, []int64{vp.Author.Mid}); err != nil {
				log.Error("s.accd.Relations2 error(%v)", err)
			}
			return nil
		})
	}
	// 有error无需return
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// 用户头像信息
	vs.FromViewOwner(plat, param.Build, stat.GetFollower())
	// 关注按钮
	if mid > 0 && vp.GetAuthor().Mid != mid {
		vs.FromButtonArc(vp.GetAuthor().Mid, model.GotoUp, authorRelations)
	}
	// 按钮组件v2
	var isFav int
	if isFavored {
		// 已收藏
		isFav = favorite.IsFav
	}
	vs.FromButtonV2(model.GotoUp, int8(isFav), authorRelations)
	if his != nil {
		vs.History = &view.History{
			Cid:      his.Cid,
			Progress: his.Pro,
			ViewAt:   his.Unix,
		}
	}
	pages := []*view.Page{}
	for _, v := range vp.Pages {
		page := &view.Page{}
		page.FromPageArc(v)
		if vp.AttrVal(api.AttrBitIsBangumi) == api.AttrYes {
			page.From = "bangumi"
		}
		page.Aid = vp.Aid
		// bvid
		if bvid, err := model.GetBvID(vp.Aid); err == nil {
			page.ShareURL = model.FillURI(model.GotoWebBV, plat, param.Build, bvid, model.SuffixHandler(fmt.Sprintf("p=%d", v.Page)))
		}
		pages = append(pages, page)
	}
	vs.ReqUser = &view.ReqUser{Like: isLike}
	vs.Pages = pages
	vs.QRCode = s.qrCode(model.GotoAv, param.Aid)
	// 简介
	vs.Introduction = &view.Introduction{}
	vs.Introduction.FromIntroductionArc(vp.Arc, asDesc)
	// 长简介
	if asDesc != "" {
		vs.Arc.Desc = asDesc
	}
	return vs, nil
}

func (s *Service) viewPGC(c context.Context, mid int64, buvid string, param *view.ViewParam) (*view.View, error) {
	vp, err := s.bgm.View(c, mid, param.SeasonID, param.AccessKey, "", param.MobiApp, param.Platform, buvid, "", param.Build)
	if err != nil {
		log.Error("%+v", err)
		// 如果接口错误统一返回视频不存在
		return nil, xecode.AppNotVedio
	}
	if vp == nil {
		return nil, xecode.AppNotVedio
	}
	// 0=不可观看 1=可观看
	if vp.Rights != nil && vp.Rights.CanWatch != 1 {
		return nil, xecode.AppNotVedio
	}
	vs := &view.View{VideoType: view.BangumiType, PagesStyle: view.HorizontalStyle}
	vs.FromViewPGC(vp)
	if vp.UserStatus != nil {
		vs.FromButtonPgc(mid, int8(vp.UserStatus.Follow), model.GotoPGC)
	}
	// 按钮组件v2
	vs.FromButtonV2(model.GotoPGC, int8(vp.UserStatus.Follow), nil)
	positivepage := []*view.Page{}
	sectionPage := []*view.Page{}
	for _, m := range vp.Modules {
		// positive正片、section其他
		switch m.Style {
		case "positive":
			for _, ep := range m.Data.Episodes {
				// 互动视频不展示
				if ep.Interaction != nil {
					continue
				}
				p := &view.Page{}
				p.FromPagePgc(ep)
				positivepage = append(positivepage, p)
			}
		case "section":
			for _, ep := range m.Data.Episodes {
				// 互动视频不展示
				if ep.Interaction != nil {
					continue
				}
				p := &view.Page{}
				p.FromPagePgc(ep)
				sectionPage = append(sectionPage, p)
			}
		default:
			continue
		}
	}
	// positive正片、section其他，优先正片，没有正片就展示非正片
	vs.Pages = positivepage
	if len(positivepage) == 0 {
		vs.Pages = sectionPage
	}
	if len(vs.Pages) == 0 {
		// 一条ep信息都没有直接不可以播放
		return nil, xecode.AppNotVedio
	}
	vs.QRCode = s.qrCode(model.GotoPGC, param.SeasonID)
	// 简介
	vs.Introduction = &view.Introduction{}
	vs.Introduction.FromIntroductionPGC(vp)
	return vs, nil
}

func (s *Service) qrCode(gt string, id int64) string {
	url, ok := s.c.Custom.ViewQRCode[fmt.Sprintf("%s_%d", gt, id)]
	if !ok {
		return ""
	}
	return url
}
