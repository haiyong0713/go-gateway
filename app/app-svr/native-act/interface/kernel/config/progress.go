package config

type Progress struct {
	BaseCfgManager

	// 进度条样式
	Style              int64  //样式
	BgColor            string //背景色
	SlotType           string //未完成态（进度槽）
	BarType            string //达成态（进度条）
	BarColor           string //进度条颜色
	TextureType        int64  //进度条纹理类型
	DisplayProgressNum bool   //是否展示当前进度
	// 数据源配置
	Sid     int64 //数据源id
	GroupID int64 //节点组id
	// 节点设置
	DisplayNodeNum  bool //是否展示节点数值
	DisplayNodeDesc bool //是否展示节点描述
}
