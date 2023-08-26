package config

import (
	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

type CarouselWord struct {
	BaseCfgManager

	ContentStyle int64  //内容样式
	BgColor      string //背景色
	FontColor    string //文字色
	CardBgColor  string //卡片背景色
	ScrollType   int64  //滚动方向
	Words        []*CarouselWordItem
}

type CarouselWordItem = model.MixExtCarouselWord
