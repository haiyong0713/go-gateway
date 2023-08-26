package config

import (
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type CarouselImg struct {
	BaseCfgManager

	ContentStyle      int64  //内容样式
	BgColor           string //背景色
	IndicatorColor    string //指示符颜色
	IsAutoCarousel    bool   //是否自动轮播
	IsTopTabFollowImg bool   //是否首页顶栏跟随图片变化
	IsTopTabFadeAway  bool   //轮播组件滑出屏幕后顶栏配置样式消失
	ImageTitle        string //图片标题
	Images            []*CarouselImgItem
}

type CarouselImgItem = model.MixExtCarouselImg
