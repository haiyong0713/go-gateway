package model

import (
	xtime "go-common/library/time"
)

// bcurd -dsn='test:test@tcp(172.16.33.205:3308)/bilibili_lottery?parseTime=true'  -schema=bilibili_lottery -table=act_page -tmpl=bilibili_log.tmpl > act_page.go

// ActPage represents a row from 'act_page'.
type ActPage struct {
	ID       int64      `json:"id"`        // 自增ID, 无意义
	State    int8       `json:"state"`     // 活动状态 0-正常，1-关闭评论
	Stime    xtime.Time `json:"stime"`     // 开始时间
	Etime    xtime.Time `json:"etime"`     // 结束时间
	Ctime    xtime.Time `json:"ctime"`     // record create timestamp
	Mtime    xtime.Time `json:"mtime"`     // record update/modify timestamp
	Name     string     `json:"name"`      // 活动名称
	Author   string     `json:"author"`    // 活动作者
	PcURL    string     `json:"pc_url"`    // 活动地址
	Rank     uint32     `json:"rank"`      // 排序接口
	H5URL    string     `json:"h5_url"`    // h5地址
	PcCover  string     `json:"pc_cover"`  // pc封面
	H5Cover  string     `json:"h5_cover"`  // h5封面
	PageName string     `json:"page_name"` // 自定义上传名
	Plat     int8       `json:"plat"`      // 平台 1,web,2app,3web and app
	Desc     string     `json:"desc"`      // 活动描述
	Click    uint64     `json:"click"`     // 点击量
	Type     int32      `json:"type"`      // 分区id
	Mold     uint8      `json:"mold"`      // 模式
	Series   uint32     `json:"series"`    // 系列
	Dept     uint32     `json:"dept"`      // 部门 0默认
	ReplyID  int32      `json:"reply_id"`  // 评论id
	TpID     int32      `json:"tp_id"`     // 模板id
	Ptime    xtime.Time `json:"ptime"`     // 发布时间
	Catalog  int32      `json:"catalog"`   // 目录id
	Creator  string     `json:"creator"`   // 创建人姓名
	SpmID    string     `json:"spm_id"`    // spm id
}
