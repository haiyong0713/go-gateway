package config

type Game struct {
	BaseCfgManager
	//图片标题
	ImageTitle string
	//文字标题
	TextTitle string
	//ids
	IDs []*GameID
	// 背景色
	BgColor string
	// 卡片标题文字色
	TitleColor string
}

type GameID struct {
	ID     int64
	Remark string
}
