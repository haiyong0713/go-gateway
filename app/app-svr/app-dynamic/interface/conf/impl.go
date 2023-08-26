// 定义 conf 实现的外部接口
// 避免其他代码直接引用conf包造成误触init函数的问题
package conf

// 实现 AppDynamicConfig 接口
// 用于 app-dynamic model层
func (c *Config) GetResDynMixTopicSquareMore() (string, string) {
	return c.Resource.Icon.DynMixTopicSquareMore, c.Resource.Text.DynMixTopicSquareMore
}

func (c *Config) GetIconModuleExtendNewTopic() string {
	return c.Resource.Icon.ModuleExtendNewTopic
}

func (c *Config) GetResModuleTitleForCampusTopic() (title, moreBtnText, moreBtnIcon string) {
	return c.Resource.Text.CampusRcmdTopicTitle, c.Resource.Text.DynMixTopicSquareMore, c.Resource.Icon.DynMixTopicSquareMore
}

func (c *Config) GetPlusMarkIcon() string {
	return c.Resource.Icon.ModuleButtonPlusMark
}
