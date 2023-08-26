package grpc

import (
	"context"
	"time"

	mauth "go-common/component/auth/middleware/grpc"
	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	mrestrict "go-common/component/restriction/middleware/grpc"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/search"
	"go-gateway/app/app-svr/app-interface/interface-legacy/http"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/search"
	searchsvr "go-gateway/app/app-svr/app-interface/interface-legacy/service/search"
)

const (
// header

)

type SearchServer struct {
	srcSvr *searchsvr.Service
}

//nolint:unparam
func newSearch(ws *warden.Server, svr *http.Server) (err error) {
	s := &SearchServer{
		srcSvr: svr.SrcSvr,
	}
	api.RegisterSearchServer(ws.Server(), s)
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.interface.v1.Search/Suggest3", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor(), svr.FeatureSvc.BuildLimitGRPC())
	ws.Add("/bilibili.app.interface.v1.Search/DefaultWords", auther.UnaryServerInterceptor(true), mrestrict.UnaryServerInterceptor(), svr.FeatureSvc.BuildLimitGRPC())
	return
}

func (s *SearchServer) DefaultWords(ctx context.Context, req *api.DefaultWordsReq) (*api.DefaultWordsReply, error) {
	var (
		mid int64
	)
	if req.TeenagersMode != 0 || req.LessonsMode != 0 {
		return nil, ecode.AccessDenied
	}
	// 获取鉴权mid
	if au, ok := auth.FromContext(ctx); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(ctx)
	if !ok {
		return nil, ecode.RequestErr
	}
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	var disableRcmdTmp int64
	if limit.DisableRcmd {
		disableRcmdTmp = 1
	}
	data, err := s.srcSvr.DefaultWords(ctx, mid, int(dev.Build), int(req.From), dev.Buvid, dev.RawPlatform, dev.MobiApp(), dev.Device, req.LoginEvent,
		&search.DefaultWordsExtParam{
			Tab:         req.Tab,
			EventId:     req.EventId,
			Avid:        req.Avid,
			Query:       req.Query,
			An:          req.An,
			IsFresh:     req.IsFresh,
			DisableRcmd: disableRcmdTmp,
		})
	if err != nil {
		return nil, err
	}
	return &api.DefaultWordsReply{
		Trackid:   data.Trackid,
		Param:     data.Param,
		Show:      data.Show,
		Word:      data.Word,
		ShowFront: int64(data.ShowFront),
		ExpStr:    data.ExpStr,
		Goto:      data.Goto,
		Value:     data.Value,
		Uri:       data.URI,
	}, nil
}

func (s *SearchServer) Suggest3(c context.Context, arg *api.SuggestionResult3Req) (res *api.SuggestionResult3Reply, err error) {
	var (
		data *search.SuggestionResult3
		mid  int64
	)
	// 获取鉴权mid
	if au, ok := auth.FromContext(c); ok {
		mid = au.Mid
	}
	// 获取设备信息
	dev, ok := device.FromContext(c)
	if !ok {
		err = ecode.RequestErr
		return
	}
	if arg == nil {
		err = ecode.RequestErr
		return
	}
	res = &api.SuggestionResult3Reply{}
	if arg.TeenagersMode != 0 {
		return
	}
	data = s.srcSvr.Suggest3(c, mid, dev.RawPlatform, dev.Buvid, arg.Keyword, dev.Device, int(dev.Build), int(arg.Highlight), dev.RawMobiApp, time.Now())
	if data == nil {
		return
	}
	res.Trackid = data.TrackID
	res.ExpStr = data.ExpStr
	for _, v := range data.List {
		var (
			officialVerify *api.OfficialVerify
			badges         []*api.ReasonStyle
		)
		if overify := v.OfficialVerify; overify != nil {
			officialVerify = &api.OfficialVerify{
				Type: int32(overify.Type),
				Desc: overify.Desc,
			}
		}
		for _, b := range v.Badges {
			tmpb := &api.ReasonStyle{
				Text:             b.Text,
				TextColor:        b.TextColor,
				TextColorNight:   b.TextColorNight,
				BgColor:          b.BgColor,
				BgColorNight:     b.BgColorNight,
				BorderColor:      b.BorderColor,
				BorderColorNight: b.BorderColorNight,
				BgStyle:          int32(b.BgStyle),
			}
			badges = append(badges, tmpb)
		}
		si := &api.ResultItem{
			From:           v.From,
			Title:          v.Title,
			Keyword:        v.KeyWord,
			Position:       int32(v.Position),
			Cover:          v.Cover,
			CoverSize:      v.CoverSize,
			SugType:        v.SugType,
			TermType:       int32(v.TermType),
			Goto:           v.Goto,
			Uri:            v.URI,
			Param:          v.Param,
			OfficialVerify: officialVerify,
			Mid:            v.Mid,
			Fans:           int32(v.Fans),
			Level:          int32(v.Level),
			Archives:       int32(v.Arcs),
			Ptime:          int64(v.PTime),
			SeasonTypeName: v.SeasonTypeName,
			Area:           v.Area,
			Style:          v.Style,
			Label:          v.Label,
			Rating:         v.Rating,
			Vote:           int32(v.Vote),
			Badges:         badges,
			Styles:         v.Styles,
			ModuleId:       v.ModuleID,
			LiveLink:       v.LiveLink,
			FaceNftNew:     v.FaceNftNew,
			IsSeniorMember: v.IsSeniorMember,
		}
		if v.NftFaceIcon != nil {
			si.NftFaceIcon = &api.NftFaceIcon{
				RegionType: v.NftFaceIcon.RegionType,
				Icon:       v.NftFaceIcon.Icon,
				ShowStatus: v.NftFaceIcon.ShowStatus,
			}
		}
		res.List = append(res.List, si)
	}
	return
}
