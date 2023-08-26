package model

import (
	"encoding/json"

	xtime "go-common/library/time"
)

const (
	AwardStep1 = "meme"
	AwardStep2 = "allMeme"
	AwardStep3 = "shortTermSkin"
	AwardStep4 = "longTermSkin"
	AwardBadge = "badge"
)

type SeriesConfig struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Type          string     `json:"type"`
	Number        int64      `json:"number"`
	Hint          string     `json:"hint"`
	Subject       string     `json:"subject"`
	Color         int64      `json:"color"`
	Cover         string     `json:"cover"`
	ShareTitle    string     `json:"share_title"`
	ShareSubtitle string     `json:"share_subtitle"`
	PushTitle     string     `json:"push_title"`
	PushSubtitle  string     `json:"push_subtitle"`
	Status        int64      `json:"status"`
	TaskStatus    int64      `json:"task_status"`
	MediaID       int64      `json:"media_id"`
	Stime         xtime.Time `json:"stime"`
	Etime         xtime.Time `json:"etime"`
}

type SeriesList struct {
	Goto            string `json:"goto"`
	Param           int64  `json:"param"`
	Cover           string `json:"cover"`
	Title           string `json:"title"`
	CoverRightText1 string `json:"cover_right_text_1"`
	RightDesc1      string `json:"right_desc_1"`
	RightDesc2      string `json:"right_desc_2"`
	RcmdReason      string `json:"rcmd_reason"`
}

type SeriesRes struct {
	Number  int64  `json:"number"`
	Subject string `json:"subject"`
	Status  int64  `json:"status"`
	Name    string `json:"name"`
}

type SeriesOneConfig struct {
	ID            int64      `json:"id"`
	Type          string     `json:"type"`
	Number        int64      `json:"number"`
	Subject       string     `json:"subject"`
	Stime         xtime.Time `json:"stime"`
	Etime         xtime.Time `json:"etime"`
	Status        int64      `json:"status"`
	Name          string     `json:"name"`
	Label         string     `json:"label"`
	Hint          string     `json:"hint"`
	Color         int64      `json:"color"`
	Cover         string     `json:"cover"`
	ShareTitle    string     `json:"share_title"`
	ShareSubtitle string     `json:"share_subtitle"`
	MediaID       int64      `json:"media_id"` // 播单ID
}

type SeriesOne struct {
	Config   *SeriesOneConfig `json:"config"`
	Reminder string           `json:"reminder"`
	List     []*SeriesArc     `json:"list"`
}

type SeriesArc struct {
	*BvArc
	RcmdReason string `json:"rcmd_reason"`
}

type PreciousArc struct {
	*BvArc
	Achievement string `json:"achievement"`
}

type PreciousRes struct {
	Title   string         `json:"title"`
	MediaID int64          `json:"media_id"`
	Explain string         `json:"explain"`
	List    []*PreciousArc `json:"list"`
}

type HotItem struct {
	ID         int64           `json:"id"`
	Goto       string          `json:"goto"`
	FromType   string          `json:"from_type"`
	Source     string          `json:"source"`
	RcmdReason *HotRcmdReason  `json:"rcmd_reason"`
	AvFeature  json.RawMessage `json:"av_feature"`
	Sticky     int             `json:"sticky"`
	IsGif      int             `json:"is_gif"`
	HotwordID  int64           `json:"hotword_id"`
	TrackID    string          `json:"trackid"`
	GifCover   string          `json:"gif_cover"`
}

type PopularCard struct {
	Type       string `json:"type"`
	Value      int64  `json:"value"`
	Reason     string `json:"reason"`
	CornerMark int    `json:"corner_mark"`
}

type HotRcmdReason struct {
	Content    string `json:"content"`
	CornerMark int    `json:"corner_mark"`
}

type PopularArc struct {
	*BvArc
	RcmdReason *HotRcmdReason `json:"rcmd_reason"`
}

type PopularInfoc struct {
	MobiApp     string
	Device      string
	Build       string
	Time        string
	LoginEvent  int64
	Mid         int64
	Buvid       string
	Feed        string
	Page        int64
	Spmid       string
	URL         string
	Env         string
	Trackid     string
	IsRec       int64
	ReturnCode  string
	UserFeature string
	Flush       string
}

type PopularFeedInfoc struct {
	Goto         string          `json:"goto"`
	Param        string          `json:"param"`
	URI          string          `json:"uri"`
	AvFeature    json.RawMessage `json:"av_feature"`
	Source       string          `json:"source"`
	RPos         int             `json:"r_pos"`
	FromType     string          `json:"from_type"`
	CornerMark   int             `json:"corner_mark"`
	RcmdContent  string          `json:"rcmd_content"`
	CoverType    string          `json:"cover_type"`
	CardStyle    int             `json:"card_style"`
	HotAggreID   int64           `json:"hot_aggre_id"`
	ChannelOrder int             `json:"channel_order"`
	ChannelName  string          `json:"channel_name"`
	ChannelID    int             `json:"channel_id"`
}

type PopularActivityReply struct {
	Role           int8              `json:"role"`
	Rank           int64             `json:"rank"`
	HonorMeta      *HonorMeta        `json:"honor_meta"`
	ActAwardStatus []*ActAwardStatus `json:"act_award_status"`
}

type ActAwardStatus struct {
	AwardName string `json:"award_name"`
	State     int8   `json:"state"`
}

type HonorMeta struct {
	FirstLoginTime      int32      `json:"first_login_time"`
	FirstWatchTime      xtime.Time `json:"first_watch_time"`
	CurrentHonorGetTime xtime.Time `json:"current_honor_get_time,omitempty"`
	ThumbupCount        int64      `json:"thumbup_count"`
	CoinCount           int64      `json:"coin_count"`
	ViewCount           int64      `json:"view_count"`
}

type PopularActivityArchiveList struct {
	List []*ArchiveMeta `json:"list"`
}

type ArchiveMeta struct {
	Aid        int64  `json:"aid"`
	Pic        string `json:"pic"`
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
	View       int32  `json:"view"`
	Danmaku    int32  `json:"danmaku"`
}

type PopularAwardReply struct {
	ID int64 `json:"id"`
}
