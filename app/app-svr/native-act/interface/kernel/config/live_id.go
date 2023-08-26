package config

type LiveID struct {
	BaseCfgManager
	//图片标题
	ImageTitle string
	//文字标题
	TextTitle string
	//房间id
	ID int64
	//开始时间
	Stime int64
	//直播封面
	Cover string
	//是否展示标题
	DisplayTitle bool
	//背景颜色
	BgColor string
	//文字色
	FontColor string
	//文字标题-文字颜色
	DisplayColor string
	// 直播卡类型 0:隐藏卡片 1:直播间
	LiveType int32
}
