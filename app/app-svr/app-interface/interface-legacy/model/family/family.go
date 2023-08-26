package family

import (
	pushmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/push"
)

type AggregationReq struct {
	MobiApp     string `form:"mobi_app"`
	DeviceToken string `form:"device_token"`
}

type AggregationRly struct {
	TeenagerStatus bool `json:"teenager_status"`
	LessonStatus   bool `json:"lesson_status"`
	FamilyStatus   bool `json:"family_status"`
}

type TeenGuardRly struct {
	Url          string `json:"url"`
	RelationType int64  `json:"relation_type"`
}

type IdentityRly struct {
	Identity string `json:"identity"`
}

type CreateFamilyQrcodeRly struct {
	Ticket string `json:"ticket"`
	Url    string `json:"url"`
}

type QrcodeInfoReq struct {
	Ticket string `form:"ticket" validate:"required"`
}

type QrcodeInfoRly struct {
	Mid      int64  `json:"mid"`
	Name     string `json:"name"`
	Face     string `json:"face"`
	IsBinded bool   `json:"is_binded"`
}

type QrcodeStatusReq struct {
	Ticket string `form:"ticket" validate:"required"`
}

type QrcodeStatusRly struct {
	IsBinded       bool `json:"is_binded"`
	TeenagerStatus bool `json:"teenager_status"`
}

type ParentIndexRly struct {
	MaxBind    int64        `json:"max_bind"`
	ChildInfos []*ChildInfo `json:"child_infos"`
}

type ChildInfo struct {
	Mid            int64  `json:"mid"`
	Name           string `json:"name"`
	Face           string `json:"face"`
	TeenagerStatus bool   `json:"teenager_status"`
	TimelockStatus bool   `json:"timelock_status"`
}

type ParentUnbindReq struct {
	ChildMid int64 `form:"child_mid" validate:"required"`
}

type ParentUpdateTeenagerReq struct {
	Action   string `form:"action" validate:"required"`
	ChildMid int64  `form:"child_mid" validate:"required"`
}

type ChildIndexRly struct {
	ParentName     string `json:"parent_name"`
	ParentMid      int64  `json:"parent_mid"`
	ParentFace     string `json:"parent_face"`
	TeenagerStatus bool   `json:"teenager_status"`
	TimelockStatus bool   `json:"timelock_status"`
}

type ChildBindReq struct {
	Ticket string `form:"ticket" validate:"required"`
}

type TimelockInfoReq struct {
	ChildMid int64 `form:"child_mid" validate:"required"`
}

type TimelockInfoRly struct {
	TimelockStatus bool  `json:"timelock_status"`
	DailyDuration  int64 `json:"daily_duration"`
}

type UpdateTimelockReq struct {
	ChildMid      int64 `form:"child_mid" validate:"required"`
	Status        int64 `form:"status"`
	DailyDuration int64 `form:"daily_duration"`
}

type TimelockPwdReq struct {
	ChildMid int64 `form:"child_mid" validate:"required"`
}

type TimelockPwdRly struct {
	Pwd string `json:"pwd"`
}

type VerifyTimelockPwdReq struct {
	Pwd string `form:"pwd" validate:"required"`
}

type VerifyTimelockPwdRly struct {
	IsPassed bool `json:"is_passed"`
}

type Timelock struct {
	Switch        bool             `json:"switch"`
	DailyDuration int64            `json:"daily_duration"`
	PushTime      int64            `json:"push_time"`
	Push          *pushmdl.Message `json:"push"`
}
