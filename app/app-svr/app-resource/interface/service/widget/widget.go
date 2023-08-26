package widget

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/dao/widget"
	model "go-gateway/app/app-svr/app-resource/interface/model/widget"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

var (
	widgetButtons = []*model.Button{
		{
			Text: "历史记录",
			Icon: "http://i0.hdslb.com/bfs/archive/852cfc4e06e58fb7746ec75b9b6840e6bcab0d5f.png",
			URL:  "bilibili://user_center/history",
		},
		{
			Text: "我的收藏",
			Icon: "http://i0.hdslb.com/bfs/archive/widgets/favorite.png",
			URL:  "bilibili://user_center/favourite",
		},
		{
			Text: "离线缓存",
			Icon: "http://i0.hdslb.com/bfs/archive/2ca672c6e05f4af6b0e9d9379d57fcaff9f622d7.png",
			URL:  "bilibili://user_center/download",
		},
		{
			Text: "稍后再看",
			Icon: "http://i0.hdslb.com/bfs/archive/46c4f7776ae63183a6700e0fdc58dbb3b4308b20.png",
			URL:  "bilibili://user_center/watch_later",
		},
	}
	androidWidgetButtons = []*model.Button{
		{
			Text: "热门",
			Icon: "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/x5TPoeDKb3.png",
			URL:  "bilibili://root?bottom_tab_id=home&tab_id=hottopic&blockInTeen=1",
		},
		{
			Text: "动态",
			Icon: "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/7IOH4ahdTq.png",
			URL:  "bilibili://root?bottom_tab_id=dynamic",
		},
		{
			Text: "我的收藏",
			Icon: "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/cX1M4A89LJ.png",
			URL:  "bilibili://main/favorite?blockInTeen=1",
		},
		{
			Text: "历史记录",
			Icon: "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/u1M9kTRAcy.png",
			URL:  "bilibili://history?blockInTeen=1",
		},
	}
)

type Service struct {
	c   *conf.Config
	dao *widget.Dao
}

func New(c *conf.Config) *Service {
	return &Service{
		c:   c,
		dao: widget.New(c),
	}
}

// WidgetMeta acquires widget meta for IOS user
func (s *Service) WidgetMeta(c context.Context, req *model.WidgetsMetaReq) (*model.WidgetsMeta, error) {
	res := &model.WidgetsMeta{
		WidgetButtons: widgetButtons,
		HotWord:       s.getSingleHotWord(c, req),
	}
	return res, nil
}

// WidgetAndroid acquires widget meta for Android user
func (s *Service) WidgetAndroid(ctx context.Context, mid int64) (*model.WidgetsAndroidMeta, error) {
	dev, _ := device.FromContext(ctx)
	params := &model.WidgetsMetaReq{
		Buvid:    dev.Buvid,
		MobiApp:  dev.MobiApp(),
		Device:   dev.Device,
		Platform: dev.RawPlatform,
		Build:    dev.Build,
	}
	res := &model.WidgetsAndroidMeta{
		UserInfo:      s.getWidgetUserInfo(ctx, mid),
		WidgetButtons: androidWidgetButtons,
		HotWord:       s.getSingleHotWord(ctx, params),
	}
	return res, nil
}

func (s *Service) getSingleHotWord(c context.Context, req *model.WidgetsMetaReq) string {
	hot, err := s.dao.HotSearch(c, req)
	if err != nil {
		log.Error("s.dao.HotSearch error(%+v)", err)
		return ""
	}
	if len(hot.Result) > 0 {
		return hot.Result[0].ShowName
	}
	return ""
}

func (s *Service) getWidgetUserInfo(ctx context.Context, mid int64) *model.UserInfo {
	if mid == 0 {
		return &model.UserInfo{Name: "未登录"}
	}
	user, err := s.dao.Card3(ctx, &accgrpc.MidReq{Mid: mid})
	if err != nil {
		log.Error("getWidgetUserInfo() s.dao.Card3 error(%+v)", err)
		return nil
	}
	return &model.UserInfo{
		Mid:  user.Mid,
		Face: user.Face,
		Name: user.Name,
	}
}
