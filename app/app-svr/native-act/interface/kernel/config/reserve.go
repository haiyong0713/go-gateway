package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type Reserve struct {
	BaseCfgManager

	ImageTitle        string  //图片标题
	TextTitle         string  //文字标题
	BgColor           string  //组件背景色
	FontColor         string  //文字色
	CardBgColor       string  //卡片背景色
	DisplayUpFaceName bool    //是否展示UP主头像昵称
	UpRsvIDs          []int64 //UP主预约id
	UpRsvIDsReqID     kernel.RequestID
}
