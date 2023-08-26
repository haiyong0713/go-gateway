package show

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/show"
)

func (s *Service) IndexPopUp(ctx context.Context, mid int64, buvid, mobiApp, device string, build int32, teenagersMode int) *show.PopUp {
	var plat int32
	switch mobiApp {
	case "iphone":
		plat = 1
	case "android":
		// 和后台同步后 20210412 安卓的 palt 值应该是 0
		// 原来被写成 2 了
		plat = 0
	}
	data, err := s.rdao.PopUp(ctx, mid, buvid, plat, build)
	if err != nil {
		log.Error("IndexPopUp PopUp mid:%d buvid:%s plat:%d build:%d error:%+v", mid, buvid, plat, build, err)
		return nil
	}
	if data.IsPoped {
		return nil
	}
	if teenagersMode == 1 && data.TeenagePush != 1 {
		return nil
	}
	if model.Plat(mobiApp, device) == model.PlatIPad && data.LinkType == 6 { // 粉版ipad不支持专栏
		return nil
	}
	var gt string
	// -1为不跳转，1为URL，2为游戏小卡，3为稿件，4为PGC，5为直播，6为专栏，7为每日精选，8为歌单，9为歌曲，10为相簿，11为小视频
	//nolint:gomnd
	switch data.LinkType {
	case -1:
		gt = ""
	case 1:
		gt = model.GotoWeb
	case 2:
		gt = model.GotoPopGame
	case 3:
		gt = model.GotoAv
	case 4:
		gt = model.GotoBangumi
	case 5:
		gt = model.GotoLive
	case 6:
		gt = model.GotoArticle
	case 7:
		gt = model.GotoDaily
	case 8:
		gt = model.GotoAudio
	case 9:
		gt = model.GotoSong
	case 10:
		gt = model.GotoClip
	case 11:
		gt = model.GotoAlbum
	}
	res := &show.PopUp{
		PopUpReport: &show.PopUpReport{
			ID:     data.Id,
			Pic:    data.Pic,
			Detail: data.Description,
			Link:   model.FillURI(gt, data.Link, nil),
		},
		AutoClose:     s.c.Custom.PopUpAutoClose,     // 默认自动关闭
		AutoCloseTime: s.c.Custom.PopUpAutoCloseTime, // 自动关闭倒计时5s
	}
	if data.LinkType == -1 { // -1为不跳转
		res.PopUpReport.Link = ""
	}
	// ipad优先使用ipad图片
	if model.Plat(mobiApp, device) == model.PlatIPad && data.PicIpad != "" {
		res.PopUpReport.Pic = data.PicIpad
	}
	reportData, _ := json.Marshal(res.PopUpReport)
	res.ReportData = string(reportData)
	return res
}
