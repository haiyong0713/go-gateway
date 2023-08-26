package fm

import (
	xtime "go-common/library/time"
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
