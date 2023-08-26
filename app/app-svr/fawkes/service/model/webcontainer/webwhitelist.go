package webcontainer

import "time"

// WebWhiteList web容器安全白名单表
type WebWhiteList struct {
	Id             int64     `json:"id"`               //自增ID
	AppKey         string    `json:"app_key"`          //APP在平台内唯一标识
	Title          string    `json:"title"`            //标题
	Domain         string    `json:"domain"`           //域名
	Reason         string    `json:"reason"`           //原因
	IsThirdParty   bool      `json:"is_third_party"`   //是否第三方域名
	Feature        string    `json:"feature"`          //1-调用js-bridge、2-扫码、3-高亮链接
	Effective      time.Time `json:"effective"`        //生效时间
	Expires        time.Time `json:"expires"`          //过期时间
	IsDomainActive bool      `json:"is_domain_active"` //域名是否有效
	Mtime          time.Time `json:"mtime"`            //修改时间
	Ctime          time.Time `json:"ctime"`            //创建时间
	CometId        string    `json:"comet_id"`         //comet工单ID
}
