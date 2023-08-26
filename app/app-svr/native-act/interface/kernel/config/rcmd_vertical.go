package config

type RcmdVertical struct {
	BaseCfgManager

	RcmdCommon
	RcmdUsers []*RcmdUser //推荐理由
}
