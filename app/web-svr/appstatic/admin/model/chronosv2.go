package model

import (
	"fmt"
	"strconv"
)

type AppInfo struct {
	ID       int64  `form:"id" gorm:"column:id" json:"id"`
	Name     string `form:"name" gorm:"column:app_name" json:"name"`
	Desc     string `form:"desc" gorm:"column:app_desc" json:"desc"`
	Icon     string `form:"icon" gorm:"column:icon" json:"icon"`
	AppID    string `form:"app_id" gorm:"column:app_id" json:"app_id"`
	Platform string `form:"platform" gorm:"column:platform" json:"platform"`
	AppKey   string `form:"app_key" validate:"required" gorm:"column:app_key" json:"app_key"`
	MobiApp  string `form:"mobi_app" gorm:"column:mobi_app" json:"mobi_app"`
}

func (a AppInfo) TableName() string {
	return "chronos_app"
}

type ServiceInfo struct {
	ID         int64  `form:"id" gorm:"column:id" json:"id"`
	Name       string `form:"name" gorm:"column:service_name" json:"name"`
	Desc       string `form:"desc" gorm:"column:service_desc" json:"desc"`
	ServiceKey string `form:"service_key" validate:"required" gorm:"column:service_key" json:"service_key"`
}

func (s ServiceInfo) TableName() string {
	return "chronos_service"
}

type PackageInfo struct {
	ID            int64  `form:"id" gorm:"column:id" json:"id"`
	UUID          string `form:"uuid" validate:"required" gorm:"column:uuid" json:"uuid,omitempty"`
	Name          string `form:"name" gorm:"column:package_name" json:"name,omitempty"`
	Desc          string `form:"desc" gorm:"column:package_desc" json:"desc,omitempty"`
	Rank          int64  `form:"rank" gorm:"column:rank" json:"rank,omitempty"`
	AppKey        string `form:"app_key" validate:"required" gorm:"column:app_key" json:"app_key,omitempty"`
	ServiceKey    string `form:"service_key" validate:"required" gorm:"column:service_key" json:"service_key,omitempty"`
	ResourceUrl   string `form:"resource_url" gorm:"column:resource_url" json:"resource_url,omitempty"`
	Gray          int64  `form:"gray" gorm:"column:gray" json:"gray,omitempty"`
	BlackList     string `form:"black_list" gorm:"column:black_list" json:"black_list,omitempty"`
	WhiteList     string `form:"white_list" gorm:"column:white_list" json:"white_list,omitempty"`
	VideoList     string `form:"video_list" gorm:"column:video_list" json:"video_list,omitempty"`
	RomVersion    string `form:"rom_version" gorm:"column:rom_version" json:"rom_version,omitempty"`
	NetType       string `form:"net_type" gorm:"column:net_type" json:"net_type,omitempty"`
	DeviceType    string `form:"device_type" gorm:"column:device_type" json:"device_type,omitempty"`
	EngineVersion string `form:"engine_version" gorm:"column:engine_version" json:"engine_version,omitempty"`
	BuildLimitExp string `form:"build_limit_exp" gorm:"column:buildlimit_exp" json:"build_limit_exp,omitempty"`
	Version       int64  `form:"version" gorm:"column:version" json:"version,omitempty"`
	Sign          string `form:"sign" gorm:"column:sign" json:"sign,omitempty"`
	MD5           string `form:"md5" gorm:"column:md5" json:"md5,omitempty"`
	IsDeleted     int64  `gorm:"column:is_deleted" json:"-"`
}

func (p PackageInfo) TableName() string {
	return "chronos_package_show"
}

func (p *PackageInfo) VersionAdder(addend int64) {
	p.Version += addend
}

func (p *PackageInfo) PackageInfoMatchKey() string {
	return fmt.Sprintf("%s:%s", p.AppKey, p.ServiceKey)
}

type PackageAudit struct {
	ID          uint   `json:"id" gorm:"column:id"`
	AuditStatus int64  `json:"audit_status" gorm:"column:audit_status"`
	ServiceKey  string `json:"service_key" gorm:"column:service_key"`
	AppKey      string `json:"app_key" gorm:"column:app_key"`
	Operator    string `json:"operator" gorm:"column:operator"`
	Behavior    string `json:"behavior" gorm:"column:behavior"`
}

func (pa PackageAudit) TableName() string {
	return "chronos_package_audit"
}

type PackageAuditBehaviorList struct {
	PackageBehavior map[string]*BehaviorDetail `json:"package_behavior"`
}

type BehaviorDetail struct {
	Action     string       `json:"action"`
	Version    int64        `json:"version"`
	Update     *PackageInfo `json:"update"`
	RankResult int64        `json:"rank_result"`
}

type BehaviorListHandler interface {
	GenBehaviorList(action string) *PackageAuditBehaviorList
}

type RankInfos struct {
	Infos []*RankInfo `json:"infos"`
}

type RankInfo struct {
	ID      int64 `json:"id"`
	Version int64 `json:"version"`
	Rank    int64 `json:"rank"`
}

type DeleteInfo struct {
	ID      int64 `json:"id"`
	Version int64 `json:"version"`
}

func (pa *PackageInfo) GenBehaviorList(action string) *PackageAuditBehaviorList {
	if pa == nil {
		return nil
	}
	key := pa.UUID
	if action == "update" {
		key = strconv.FormatInt(pa.ID, 10)
	}
	return &PackageAuditBehaviorList{
		PackageBehavior: map[string]*BehaviorDetail{
			key: {
				Action:  action,
				Version: pa.Version,
				Update:  pa,
			},
		},
	}
}

func (r *RankInfos) GenBehaviorList(action string) *PackageAuditBehaviorList {
	if r == nil {
		return nil
	}
	packageOpDetail := make(map[string]*BehaviorDetail, len(r.Infos))
	for _, v := range r.Infos {
		packageOpDetail[strconv.FormatInt(v.ID, 10)] = &BehaviorDetail{
			Action:     action,
			Version:    v.Version,
			RankResult: v.Rank,
		}
	}
	return &PackageAuditBehaviorList{
		PackageBehavior: packageOpDetail,
	}
}

func (d *DeleteInfo) GenBehaviorList(action string) *PackageAuditBehaviorList {
	if d == nil {
		return nil
	}
	return &PackageAuditBehaviorList{
		PackageBehavior: map[string]*BehaviorDetail{
			strconv.FormatInt(d.ID, 10): {
				Action:  action,
				Version: d.Version,
			},
		},
	}
}

type PrePackageInfo struct {
	ID         int64  `gorm:"column:id"`
	UUID       string `gorm:"column:uuid"`
	Version    int64  `gorm:"column:version"`
	AppKey     string `gorm:"column:app_key"`
	ServiceKey string `gorm:"column:service_key"`
}

type PackageOpReply struct {
	AuditID  int64  `form:"audit_id" json:"audit_id"`
	Behavior string `json:"behavior"`
}
