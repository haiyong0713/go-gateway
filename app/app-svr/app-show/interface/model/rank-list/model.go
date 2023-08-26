package ranklist

import (
	"go-gateway/pkg/idsafe/bvid"
)

const (
	FloatTypeWatching = "watching"
	FloatTypeOwner    = "owner"

	// 0-未开始 1-进行中 2-已结束 3-已结榜
	StatePending      = int64(0)
	StateVoting       = int64(1)
	StateStopped      = int64(2)
	StateRankFinished = int64(3)

	PayloadModeArchive = int64(1)
	PayloadModeAccount = int64(2)

	DescriptionContentImage = "image"
	DescriptionContentText  = "text"
)

// IndexReq is
type IndexReq struct {
	Mid int64 `form:"-"`

	ID             int64  `form:"id" validate:"required"`
	FromViewBVID   string `form:"from_view_bvid"`
	RawFromViewAid string `form:"from_view_aid"`
	FromViewAid    int64  `form:"-"`
	MobiApp        string `form:"mobi_app"`
	Device         string `form:"device"`
}

// ResolveFromViewArchive is
func (r IndexReq) ResolveFromViewArchive() int64 {
	if r.FromViewBVID != "" {
		aid, err := bvid.BvToAv(r.FromViewBVID)
		if err == nil {
			return aid
		}
	}
	return r.FromViewAid
}

// IndexReply is
type IndexReply struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	Cover       string         `json:"cover"`
	State       int64          `json:"state"`
	Tids        []int64        `json:"tids"`
	ActIDs      []int64        `json:"act_ids"`
	Tags        []string       `json:"tags"`
	Description []*Description `json:"description"`
	RankArchive []*RankArchive `json:"rank_archive,omitempty"`
	RankPayload []*RankPayload `json:"rank_payload,omitempty"`
	Floating    []*Floating    `json:"floating,omitempty"`
}

// Description is
type Description struct {
	Title       string `json:"title"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
}

// AuthorSchema is
type AuthorSchema struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

// ArchiveSchema is
type ArchiveSchema struct {
	Aid       int64        `json:"aid"`
	Title     string       `json:"title"`
	Pic       string       `json:"pic"`
	MissionID int64        `json:"mission_id"`
	Author    AuthorSchema `json:"author"`
}

// RankArchive is
type RankArchive struct {
	Ranking int64 `json:"ranking"`
	ArchiveSchema
}

// RankAuthor is
type RankAuthor struct {
	Ranking int64 `json:"ranking"`
	AuthorSchema
}

// RankPayload is
type RankPayload struct {
	Mode        int64          `json:"mode"`
	Title       string         `json:"title"`
	ArchiveList []*RankArchive `json:"archive_list,omitempty"`
	AccountList []*RankAuthor  `json:"account_list,omitempty"`
}

// Floating is
type Floating struct {
	Type       string        `json:"type"`
	Archive    ArchiveSchema `json:"archive"`
	RankStatus struct {
		Ranked  bool  `json:"ranked"`
		Ranking int64 `json:"ranking"`
	} `json:"rank_status"`
	Text string `json:"text"`
}
