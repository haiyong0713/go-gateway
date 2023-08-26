package retry

import (
	"go-common/library/time"
	"go-gateway/app/app-svr/ugc-season/job/model/databus"
)

// list
const (
	FailList                 = "season_job_fail_list"
	FailSeasonAdd            = "season_add"
	FailUpSeasonCache        = "up_season_cache"
	FailDelSeasonCache       = "del_season_cache"
	FailForPubArchiveDatabus = "pub_archive_databus"
	FailUpSeasonStat         = "up_season_stat"
	ActionUp                 = "update"
	ActionDel                = "delete"
)

// Info retry data
type Info struct {
	Action string `json:"action"`
	Data   struct {
		SeasonID          int64                      `json:"season_id"`
		Mid               int64                      `json:"mid"`
		SeasonWithArchive *databus.SeasonWithArchive `json:"season_with_archive"`
		Ptime             time.Time                  `json:"ptime"`
	} `json:"data"`
}
