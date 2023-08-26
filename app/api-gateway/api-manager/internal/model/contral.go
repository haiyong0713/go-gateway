package model

import xtime "go-common/library/time"

type ContralListPage struct {
	Total    int64 `json:"total"`
	PageSize int64 `json:"page_size"`
	PageNum  int64 `json:"page_num"`
}

type ContralGroupSaveReq struct {
	ID        int64  `form:"id"`
	GroupName string `form:"group_name"`
	Manager   string `form:"manager"`
	Desc      string `form:"desc"`
}

type ContralGroupListReq struct {
	GroupName string `form:"group_name"`
	PageNum   int64  `form:"page_num" default:"1" validate:"min=1"`
	PageSize  int64  `form:"page_size" default:"1" validate:"min=20"`
}

type ContralGroupListReply struct {
	Page *ContralListPage `json:"page"`
	List []*ContralGroup  `json:"list"`
}

type ContralGroup struct {
	ID        int64      `json:"id"`
	GroupName string     `json:"group_name"`
	Creator   string     `json:"creator"`
	Modifier  string     `json:"modifier"`
	Manager   string     `json:"manager"`
	Desc      string     `json:"desc"`
	CTime     xtime.Time `json:"ctime"`
	MTime     xtime.Time `json:"mtime"`
}

type ContralGroupFollowActionPeq struct {
	Id    int64  `form:"id" validate:"required"`
	State string `form:"state" validate:"required"`
}

type ContralApiAddReq struct {
	Gid        int64  `form:"gid" validate:"required"`
	ApiName    string `form:"api_name" validate:"required"`
	ApiType    string `form:"api_type" validate:"required"` // 接口类型：http、grpc
	Domain     string `form:"domain"`
	Router     string `form:"router"`
	Handler    string `form:"handler"`
	Req        string `form:"req" validate:"required"`
	Reply      string `form:"reply" validate:"required"`
	DSLCode    string `form:"dsl_code" validate:"required"`
	DSLStruct  string `form:"dsl_struct" validate:"required"`
	CustomCode string `form:"custom_code"`
	Desc       string `form:"desc" validate:"required"`
}

type ContralApiEditReq struct {
	ID         int64  `form:"id" validate:"required"`
	ApiType    string `form:"api_type" validate:"required"` // 接口类型：http、grpc
	Domain     string `form:"domain"`
	Router     string `form:"router"`
	Handler    string `form:"handler"`
	Req        string `form:"req" validate:"required"`
	Reply      string `form:"reply" validate:"required"`
	DSLCode    string `form:"dsl_code" validate:"required"`
	DSLStruct  string `form:"dsl_struct" validate:"required"`
	CustomCode string `form:"custom_code"`
	Desc       string `form:"desc" validate:"required"`
}

type ContralApi struct {
	ID         int64      `json:"id"`
	Gid        int64      `json:"gid"`
	ApiName    string     `json:"api_name"`
	ApiType    string     `json:"api_type"` // 接口类型：http、grpc
	Domain     string     `json:"domain"`
	Router     string     `json:"router"`
	Handler    string     `json:"handler"`
	Req        string     `json:"req"`
	Reply      string     `json:"reply"`
	DSLCode    string     `json:"dsl_code"`
	DSLStruct  string     `json:"dsl_struct"`
	CustomCode string     `json:"custom_code"`
	Creator    string     `json:"creator"`
	Modifier   string     `json:"modifier"`
	Desc       string     `json:"desc"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
}

type ContralApiListReq struct {
	GID      int64  `form:"gid" default:"0" validate:"min=0"`
	ApiName  string `form:"api_name"`
	PageNum  int64  `form:"page_num" default:"1" validate:"min=1"`
	PageSize int64  `form:"page_size" default:"1" validate:"min=20"`
}

type ContralApiListReply struct {
	Page *ContralListPage `json:"page"`
	List []*ContralApi    `json:"list"`
}

type ContralApiConfigAddReq struct {
	ApiID   int64  `form:"api_id" validate:"required"`
	Version string `form:"version" validate:"required"`
	Desc    string `form:"desc" validate:"required"`
}

type ContralApiConfig struct {
	ID         int64      `json:"id"`
	ApiID      int64      `json:"api_id"`
	Version    string     `json:"version"`
	ApiType    string     `json:"api_type"` // 接口类型：http、grpc
	Domain     string     `json:"domain"`
	Router     string     `json:"router"`
	Handler    string     `json:"handler"`
	Req        string     `json:"req"`
	Reply      string     `json:"reply"`
	DSLCode    string     `json:"dsl_code"`
	DSLStruct  string     `json:"dsl_struct"`
	CustomCode string     `json:"custom_code"`
	Creator    string     `json:"creator"`
	Desc       string     `json:"desc"`
	CTime      xtime.Time `json:"ctime"`
	MTime      xtime.Time `json:"mtime"`
}

type ContralApiConfigRollbackReq struct {
	ApiConfigID int64 `form:"api_config_id" validate:"required"`
}

type ContralApiConfigListReq struct {
	ApiID    int64 `form:"api_id" validate:"required"`
	PageNum  int64 `form:"page_num" default:"1" validate:"min=1"`
	PageSize int64 `form:"page_size" default:"1" validate:"min=20"`
}

type ContralApiConfigListReply struct {
	Page *ContralListPage    `json:"page"`
	List []*ContralApiConfig `json:"list"`
}

type ContralApiPublishListReq struct {
	ApiID    int64 `form:"api_id" validate:"required"`
	PageNum  int64 `form:"page_num" default:"1" validate:"min=1"`
	PageSize int64 `form:"page_size" default:"1" validate:"min=20"`
}

type ContralApiPublishListReply struct {
	Page *ContralListPage     `json:"page"`
	List []*ContralApiPublish `json:"list"`
}

type ContralApiPublish struct {
	ID        int64      `json:"id"`
	ApiID     int64      `json:"api_id"`
	PublishID int64      `json:"publish_id"`
	Version   string     `json:"version"`
	State     string     `json:"state"`
	Operator  string     `json:"operator"`
	CTime     xtime.Time `json:"ctime"`
	MTime     xtime.Time `json:"mtime"`
}

type ContralapiPublishCallbackReq struct {
	ApiID     int64  `form:"api_id" validate:"required"`
	PublishID int64  `form:"publish_id" validate:"required"`
	Version   string `form:"version"`
	State     string `form:"state" validate:"required"`
}

type DynpathParam struct {
	Node          string `json:"node"`
	Gateway       string `json:"gateway"`
	Pattern       string `json:"pattern"`
	ClientInfo    string `json:"client_info"`
	Enable        int    `json:"enable"`
	ClientTimeout int    `json:"client_timeout"`
}
