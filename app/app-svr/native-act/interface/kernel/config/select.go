package config

import (
	"encoding/json"

	"go-common/library/log"
)

type Select struct {
	BaseCfgManager
	Items []*SelectItem //tab选项
	Title string        //标题
	// 纯色
	BgColor                string     //背景色
	TopFontColor           string     //顶部文字色
	PanelBgColor           string     //展开面板背景色
	PanelSelectColor       string     //展开面板选中色
	PanelSelectFontColor   string     //展开面板选中背景色
	PanelNtSelectFontColor string     //展开面板未选中色
	PrimaryPageID          int64      //父页面id
	ShareInfo              *ShareInfo //分享信息
}

type ShareInfo struct {
	Image   string
	Title   string
	Caption string
}

type SelectItem struct {
	PageID int64 //子页面id
	SelectExt
}

type SelectExt struct {
	DefType     int64  `json:"def_type"`     //生效方式
	DStime      int64  `json:"d_stime"`      //生效开始时间
	DEtime      int64  `json:"d_etime"`      //生效结束时间
	Type        string `json:"type"`         //业务方：week 每周必看
	LocationKey string `json:"location_key"` //业务方唯一id
}

func UnmarshalSelectedExt(data string) (*SelectExt, error) {
	if data == "" {
		return &SelectExt{}, nil
	}
	ext := &SelectExt{}
	if err := json.Unmarshal([]byte(data), ext); err != nil {
		log.Error("Fail to unmarshal NativeMixtureExt.Reason of Selected, data=%+v error=%+v", data, err)
		return &SelectExt{}, err
	}
	return ext, nil
}
