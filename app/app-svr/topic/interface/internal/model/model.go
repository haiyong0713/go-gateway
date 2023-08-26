package model

import (
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

type CommonDetailsParams struct {
	TopicId int64
	SortBy  int64
	Offset  string
	Source  string
}

// 动态列表资源
type DynListRes struct {
	Dynamics   []*dynmdlV2.Dynamic                  // 动态核心
	DynCmtMode map[int64]*topiccardmodel.DynCmtMeta // 动态评论模式
}

// 新话题topic通用结构
type TopicItem struct {
	Id                 int64  `json:"id"`                            //话题id
	Name               string `json:"name"`                          //话题名称
	View               int64  `json:"view,omitempty"`                //浏览量
	Discuss            int64  `json:"discuss,omitempty"`             //讨论量
	Fav                int64  `json:"fav,omitempty"`                 //收藏数
	Dynamics           int64  `json:"dynamics,omitempty"`            //动态数
	State              int32  `json:"state,omitempty"`               //话题状态(0:已经上线 1:审核中 -1:已驳回 -2:已下线)
	JumpUrl            string `json:"jump_url,omitempty"`            //跳转链接
	StatDesc           string `json:"stat_desc,omitempty"`           //话题描述字段
	BackColor          string `json:"back_color,omitempty"`          //话题背景色
	IsFav              bool   `json:"is_fav,omitempty"`              //请求用户是否收藏
	Description        string `json:"description,omitempty"`         //话题描述
	CreateSource       int32  `json:"create_source,omitempty"`       //话题创建来源 0: 线上用户 1: 后台运营
	SharePic           string `json:"share_pic,omitempty"`           //分享图
	Share              int64  `json:"share,omitempty"`               //分享数
	Like               int64  `json:"like,omitempty"`                //点赞数
	ShareUrl           string `json:"share_url,omitempty"`           //分享链接
	IsLike             bool   `json:"is_like,omitempty"`             //是否点赞
	RcmdText           string `json:"rcmd_text,omitempty"`           //引导文案
	RcmdIconUrl        string `json:"rcmd_icon_url,omitempty"`       //推荐图片url
	TopicRcmdType      int32  `json:"topic_rcmd_type,omitempty"`     //话题推荐类型
	LancerInfo         string `json:"lancer_info,omitempty"`         //服务端透传上报信息
	ServerInfo         string `json:"server_info,omitempty"`         //服务端透传算法上报信息
	Rid                int64  `json:"rid,omitempty"`                 //动态id
	DescriptionSubject string `json:"description_subject,omitempty"` //描述主体
	DescriptionContent string `json:"description_content,omitempty"` //描述内容文案
	UpId               int64  `json:"-"`                             //创建者mid
}

// TagItem 老话题tag结构
type TagItem struct {
	Id       int64  `json:"id"`                  //tag id
	Name     string `json:"name"`                //tag名称
	View     int64  `json:"view,omitempty"`      //浏览量
	Discuss  int64  `json:"discuss,omitempty"`   //讨论量
	JumpUrl  string `json:"jump_url,omitempty"`  //跳转链接
	StatDesc string `json:"stat_desc,omitempty"` //话题描述字段
}
