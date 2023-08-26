package config

type Rcmd struct {
	BaseCfgManager

	RcmdCommon
	RcmdUsers []*RcmdUser //推荐理由
}

type RcmdUser struct {
	Mid    int64
	Reason string //推荐理由
	Uri    string //指定链接
}

type RcmdCommon struct {
	ImageTitle         string //图片标题
	BgColor            string //背景色
	CardTitleFontColor string //卡片标题文字色
}
