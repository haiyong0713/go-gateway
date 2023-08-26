package retry

import (
	"go-gateway/app/app-svr/archive/job/model/result"
)

const (
	FailList          = "arc_job_fail_list"
	FailVideoshotList = "arc_job_fail_videoshot_list"
	FailVideoFFList   = "arc_job_fail_videoff_list"
	FailInternalList  = "arc_job_fail_internal_list"
	FailUpCache       = "up_cache"
	FailUpVideoCache  = "up_video_cache"
	FailDelVideoCache = "del_video_cache"
	FailDatabus       = "up_databus"
	FailResultAdd     = "result_add"
	FailVideoShot     = "videoshot_cache"
	FailVideoFF       = "video_ff"
	FailUpInternal    = "up_internal_ff"
	FailInternalCache = "up_internal_cache"
)

// Info retry data
type Info struct {
	Action string `json:"action"`
	Data   struct {
		Aid                 int64                 `json:"aid"`
		ArcAction           int                   `json:"arc_action"`
		State               int32                 `json:"state"`
		DatabusMsg          *result.ArchiveUpInfo `json:"dbus_msg"`
		Cids                []int64               `json:"cids"`
		SeasonID            int64                 `json:"season_id"`
		SeasonDatabusAction string                `json:"season_databus_action"`
		WithState           bool                  `json:"with_state"`
		Cid                 int64                 `json:"cid"`
	} `json:"data"`
}
