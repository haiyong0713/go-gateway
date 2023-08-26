package config

type HoverButton struct {
	BaseCfgManager

	Item       ClickItem
	MutexUkeys []string //当该组件划出屏幕后，悬浮按钮才会出现
}
