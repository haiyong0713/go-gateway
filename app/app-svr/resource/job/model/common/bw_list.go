package common

type BWListItem struct {
	Oid          string `json:"oid"`           //业务对象唯一标识
	AreaID       int64  `json:"area_id"`       //区域 ID
	IsDeleted    int8   `json:"is_deleted"`    //删除状态 0-未删除，1-已删除
	Token        string `json:"token"`         // 场景识别码
	Status       int32  `json:"status"`        //业务场景状态 0-启用，1-不启用
	IsOnline     int32  `json:"is_online"`     //业务场景描述
	DefaultValue int32  `json:"default_value"` //未在名单中的oid的兜底展现结果 0-false，1-true
}
