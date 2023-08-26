package pack

import (
	"time"
)

type ClearPackReq struct {
	Tstart  time.Time `json:"start_time" form:"start_time" time_format:"2006-01-02 15:04:05"`
	Tend    time.Time `json:"end_time" form:"end_time" time_format:"2006-01-02 15:04:05"`
	AppKey  string    `json:"app_key" form:"app_key"`
	PkgType []int64   `json:"pkg_type" form:"pkg_type" validate:"required"`
}

// BuildKey buildId and appKey
type BuildKey struct {
	AppKey  string `form:"app_key" json:"app_key"`
	BuildId int64  `form:"build_id" json:"build_id"`
}

// DeleteResp 删除文件  返回参数
type DeleteResp struct {
	BuildIdFail  []int64 `json:"failed_id_list"`
	AffectedRows int64   `json:"affected_rows"`
}
