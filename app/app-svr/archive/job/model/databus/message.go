package databus

import "encoding/json"

// Message databus
type Message struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

// const is
const (
	RouteFirstRoundForbid = "first_round_forbid"
	RouteSecondRound      = "second_round"
	RouteAutoOpen         = "auto_open"
	RouteDelayOpen        = "delay_open"
	RouteDeleteArchive    = "delete_archive"
	RouteForceSync        = "force_sync"
	RouteVideoShotChanged = "videoshot_changed"
	RouteVideoFF          = "video_first_frame"
	RoutePremierePass     = "premiere_pass_audit"
	// season with archive
	SeasonRouteForUpdate = "season_update"
	SeasonRouteForRemove = "season_remove"
)

// Videoup message for videoup2BVC
type Videoup struct {
	Route     string  `json:"route"`
	Timestamp int64   `json:"timestamp"`
	Aid       int64   `json:"aid"`
	CIDs      []int64 `json:"cids"`
	Cid       int64   `json:"cid"`
	UpFrom    int64   `json:"up_from"`
}

// Rebuild is
type Rebuild struct {
	Aid int64 `json:"aid"`
}

// SeasonWithArchive is
type SeasonWithArchive struct {
	Route    string  `json:"route"`
	SeasonID int64   `json:"season_id"`
	Aids     []int64 `json:"aids"`
}

type ForbidArc struct {
	Aid          int64 `json:"aid"`
	OverseaBlock bool  `json:"oversea_block"`
}

type InternalMessage struct {
	Router string `json:"router"`
	Data   struct {
		Oid int64 `json:"oid"`
	} `json:"data"`
}
