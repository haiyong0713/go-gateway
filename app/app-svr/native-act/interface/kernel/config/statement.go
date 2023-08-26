package config

type Statement struct {
	BaseCfgManager

	Content             string //文本内容
	FontColor           string //文字色
	BgColor             string //背景色
	DisplayUnfoldButton bool   //是否展开收起按钮
}
