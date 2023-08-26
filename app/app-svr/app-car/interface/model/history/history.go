package history

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

// HisParam fro history
type HisParam struct {
	model.DeviceInfo
	Max      int64  `form:"max"`
	MaxTP    int32  `form:"max_tp"`
	FromType string `form:"from_type"`
	ParamStr string `form:"param"`
}

type ReportParam struct {
	model.DeviceInfo
	Aid        int64        `form:"aid"`
	Cid        int64        `form:"cid"`
	SeasonId   int64        `form:"season_id"`
	EpId       int64        `form:"ep_id"`
	Otype      string       `form:"type"`
	Progress   int64        `form:"progress"`
	SeasonType int          `form:"season_type"`
	Timestamp  int64        `form:"timestamp"`
	Source     string       `form:"source"`
	FmType     fm_v2.FmType `form:"fm_type"`
	FmId       int64        `form:"fm_id"`
	PlayEvent  int          `form:"play_event"` // 1：自动上报（间隔25s） 2：其他行为（包含切集、完播、暂停）
	ItemType   string       `form:"item_type"`  // 物料类型: ugc_single-UGC单P、 ugc_multi-UGC多P、 ogv-OGV、 video_serial-视频合集、 video_channel-视频频道、 fm_serial-FM合集、 fm_channel-FM频道
	ItemID     int64        `form:"item_id"`    // 物料ID：video_serial-视频合集ID、 video_channel-视频频道ID、 fm_serial-FM合集ID、 fm_channel-FM频道ID
}
