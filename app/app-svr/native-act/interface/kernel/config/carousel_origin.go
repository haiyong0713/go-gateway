package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type CarouselOrigin struct {
	BaseCfgManager

	ContentStyle   int64  //内容样式
	BgColor        string //背景色
	IndicatorColor string //指示符颜色
	IsAutoCarousel bool   //是否自动轮播
	ImageTitle     string //图片标题
	ImgHeight      int64  //图片高度
	ImgWidth       int64  //图片宽度
	UpListReqID    kernel.RequestID
}
