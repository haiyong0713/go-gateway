package config

type Navigation struct {
	BaseCfgManager

	SelectedFontColor     string //选中态字体色
	SelectedBgColor       string //选中态背景色
	UnselectedFontColor   string //未选中态字体色
	UnselectedBgColor     string //未选中态背景色
	NtSelectedFontColor   string //夜间-选中态字体色
	NtSelectedBgColor     string //夜间-选中态背景色
	NtUnselectedFontColor string //夜间-未选中态字体色
	NtUnselectedBgColor   string //夜间-未选中态背景色
}
