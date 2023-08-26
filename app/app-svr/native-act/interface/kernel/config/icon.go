package config

type Icon struct {
	BaseCfgManager

	BgColor   string //背景色
	FontColor string //文字色
	Items     []*IconItem
}

type IconItem struct {
	Title string //图标名
	Image string //图片
	Uri   string //跳转链接
}
