package model

import (
	"fmt"
)

func PopUpsKey(pop_id, mid int64, buvid string) string {
	return fmt.Sprintf("popups_%d_%v_%d", pop_id, buvid, mid)
}

type PopUps struct {
	ID             int64  `json:"id" form:"id"`
	Pic            string `json:"pic" form:"pic"`
	PicIpad        string `json:"pic_ipad" form:"pic_ipad"`
	Description    string `json:"description" form:"description"`
	LinkType       int32  `json:"link_type" form:"link_type"`
	Link           string `json:"link" form:"link"`
	TeenagePush    int    `json:"teenage_push" form:"teenage_push"`
	AutoHideStatus int    `json:"auto_hide_status" form:"auto_hide_status"`
	CloseTime      int64  `json:"close_time" form:"close_time" default:"5"`
	IsPop          bool   `json:"is_pop" form:"is_pop"`
	CrowdType      int    `json:"crow_type" form:"crowd_type"`
	CrowdBase      int    `json:"crowd_base" form:"crowd_base"`
	CrowdValue     string `json:"crowd_value" form:"crowd_value"`
	Builds         string `json:"builds" form:"builds"`
}
