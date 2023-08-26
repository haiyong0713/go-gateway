package fm_v2

import (
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-car/interface/model/common"
)

type Scene string // 合集外露场景（FM页、视频页）

const (
	SceneFm    = Scene("fm")
	SceneVideo = Scene("video")
)

type FmSeasonInfoPo struct {
	Id       int64      `json:"id"`
	FmType   string     `json:"fm_type"`
	FmId     int64      `json:"fm_id"`
	Title    string     `json:"title"`
	Cover    string     `json:"cover"`
	Subtitle string     `json:"subtitle"`
	FmState  int        `json:"fm_state"`
	Ctime    xtime.Time `json:"ctime"`
	Mtime    xtime.Time `json:"mtime"`
	Count    int        `json:"-"` // 稿件数量查询走缓存
}

type FmSeasonOidPo struct {
	Id     int64      `json:"id"`
	FmType string     `json:"fm_type"`
	FmId   int64      `json:"fm_id"`
	Oid    int64      `json:"oid"`
	Seq    int        `json:"seq"`
	Ctime  xtime.Time `json:"ctime"`
	Mtime  xtime.Time `json:"mtime"`
}

type VideoSeasonInfoPo struct {
	Id          int64      `json:"id"`
	SeasonId    int64      `json:"fm_id"`
	Title       string     `json:"title"`
	Cover       string     `json:"cover"`
	Subtitle    string     `json:"subtitle"`
	SeasonState int        `json:"season_state"`
	Ctime       xtime.Time `json:"ctime"`
	Mtime       xtime.Time `json:"mtime"`
	Count       int        `json:"-"` // 稿件数量查询走缓存
}

type VideoSeasonOidPo struct {
	Id       int64      `json:"id"`
	SeasonId int64      `json:"season_id"`
	Oid      int64      `json:"oid"`
	Seq      int        `json:"seq"`
	Ctime    xtime.Time `json:"ctime"`
	Mtime    xtime.Time `json:"mtime"`
}

// SeasonInfoReq 合集基本信息查找请求
type SeasonInfoReq struct {
	Scene    Scene
	FmType   FmType // 仅FM场景下携带
	SeasonId int64
}

type SeasonInfoResp struct {
	Scene Scene
	Fm    *FmSeasonInfoPo
	Video *VideoSeasonInfoPo
}

// SeasonOidReq 合集内稿件分页查询
type SeasonOidReq struct {
	Scene       Scene
	FmType      FmType // 仅FM场景下携带
	SeasonId    int64
	Cursor      int64 // 游标，从哪个oid开始查
	Upward      bool  // 是否向上查找
	WithCurrent bool  // 返回结果中，是否包含当前游标
	Ps          int   // 分页大小
}

type SeasonOidResp struct {
	Scene Scene
	Aids  []int64   // 合集内分页稿件
	Page  *PageResp // 分页信息
}

func (r *SeasonInfoResp) ToSerialInfo() *common.SerialInfo {
	if r == nil {
		return nil
	}
	if r.Scene == SceneFm {
		if r.Fm == nil {
			return nil
		}
		return &common.SerialInfo{
			Title: r.Fm.Title,
			Cover: r.Fm.Cover,
			Count: r.Fm.Count,
		}
	}
	if r.Scene == SceneVideo {
		if r.Video == nil {
			return nil
		}
		return &common.SerialInfo{
			Title: r.Video.Title,
			Cover: r.Video.Cover,
			Count: r.Video.Count,
		}
	}
	return nil
}
