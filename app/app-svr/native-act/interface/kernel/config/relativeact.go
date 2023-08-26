package config

import (
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Relativeact struct {
	BaseCfgManager

	ImageTitle         string //图片标题
	BgColor            string //背景色
	CardTitleFontColor string //卡片标题文字色
	Acts               []*natpagegrpc.NativePage
}
