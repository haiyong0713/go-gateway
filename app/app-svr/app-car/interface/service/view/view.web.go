package view

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func (s *Service) ViewWeb(c context.Context, mid int64, cookie, buvid, referer string, param *view.ViewWebParam) (*view.ViewWeb, error) {
	switch param.Otype {
	case model.GotoPGC:
		return s.viewPGCWeb(c, mid, cookie, buvid, referer, param)
	}
	return s.viewArcWeb(c, mid, buvid, param)
}

func (s *Service) viewArcWeb(c context.Context, mid int64, buvid string, param *view.ViewWebParam) (*view.ViewWeb, error) {
	var (
		vp                        *api.ViewReply
		his                       *hisApi.ModelHistory
		stat                      *relationgrpc.StatReply
		authorRelations           map[int64]*relationgrpc.InterrelationReply
		isLike                    int
		asDesc                    string
		isFavored, overseaBlocked bool
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		if vp, err = s.arc.View(ctx, param.Oid); err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if mid > 0 || buvid != "" {
		group.Go(func(ctx context.Context) (err error) {
			if his, err = s.his.Progress(ctx, param.Oid, mid, buvid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			likeState, err := s.thumbupDao.HasLike(ctx, mid, _businessLike, buvid, param.Oid)
			if err != nil {
				log.Error("%+v", err)
			}
			if likeState == thumbup.State_STATE_LIKE {
				isLike = 1
			}
			return nil
		})
		if mid > 0 {
			group.Go(func(ctx context.Context) error {
				isFavored = s.fav.IsFavored(ctx, mid, param.Oid)
				return nil
			})
		}
	}
	group.Go(func(ctx context.Context) (err error) {
		desc, err := s.arc.Description(ctx, param.Oid)
		if err != nil {
			return nil
		}
		asDesc = desc
		return nil
	})
	// 稿件特殊属性位
	group.Go(func(ctx context.Context) error {
		flowCtrlData, flowErr := s.arc.FlowControlInfoV2(ctx, param.Oid, s.c.FlowControl)
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
	vs := &view.ViewWeb{Otype: model.GotoAv, PagesStyle: view.HorizontalStyle}
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
	vs.FromViewUGC(vp, his)
	vs.FromViewOwnerWeb(stat.GetFollower(), vp, authorRelations)
	// 简介
	vs.Introduction = &view.Introduction{}
	vs.Introduction.FromIntroductionArc(vp.Arc, asDesc)
	vs.ReqUser = &view.ReqUser{Like: isLike, IsFav: isFavored}
	return vs, nil
}

func (s *Service) viewPGCWeb(c context.Context, mid int64, cookie, buvid, referer string, param *view.ViewWebParam) (*view.ViewWeb, error) {
	vp, err := s.bgm.View(c, mid, param.Oid, "", cookie, model.AndroidBilithings, "android", buvid, referer, 0)
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
	vs := &view.ViewWeb{Otype: model.GotoPGC, PagesStyle: view.HorizontalStyle}
	vs.FromViewPGC(vp)
	// 简介
	vs.Introduction = &view.Introduction{}
	vs.Introduction.FromIntroductionPGC(vp)
	vs.FromButton(model.GotoPGC, int8(vp.UserStatus.Follow))
	if len(vs.Pages) == 0 {
		// 一条ep信息都没有直接不可以播放
		return nil, xecode.AppNotVedio
	}
	return vs, nil
}
