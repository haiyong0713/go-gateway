package archive

import (
	"go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"

	batch "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcbatch"
)

// 各属性地址见 http://syncsvn.bilibili.co/platform/doc/blob/master/archive/field/state.md

// all const
const (
	// open state
	StateOpen = 0
	// attribute yes and no
	AttrYes = int32(1)
	AttrNo  = int32(0)
	// attribute bit
	AttrBitIsPGC         = uint(9)
	AttrBitIsCooperation = uint(24)
	//mob
	MobileAppIphone    = "iphone"
	MobileAppAndroid   = "android"
	RedirectTypeLegacy = "legacy_url"
	RedirectTypeUrl    = "url"
)

// AidPubTime aid's pubdate and copyright
type AidPubTime struct {
	Aid       int64     `json:"aid"`
	PubDate   time.Time `json:"pubdate"`
	Copyright int8      `json:"copyright"`
}

// ArgPlayer ArgPlayer
type ArgPlayer struct {
	Aids     []int64
	Qn       int64
	Platform string
	RealIP   string
	Fnval    int64
	Fnver    int64
	Build    int64
	// 非必传
	Session            string
	ForceHost          int64
	Mid                int64
	AidsWithoutPlayurl []int64
	Buvid              string
}

// ArchiveWithPlayer with first player info
type ArchiveWithPlayer struct {
	*api.Arc
	PlayerInfo *PlayerInfo `json:"player_info,omitempty"`
}

// PlayerInfo player info
type PlayerInfo struct {
	Cid                int64                     `json:"cid"`
	ExpireTime         int64                     `json:"expire_time,omitempty"`
	FileInfo           map[int][]*PlayerFileInfo `json:"file_info"`
	SupportQuality     []int                     `json:"support_quality"`
	SupportFormats     []string                  `json:"support_formats"`
	SupportDescription []string                  `json:"support_description"`
	Quality            int                       `json:"quality"`
	URL                string                    `json:"url,omitempty"`
	VideoCodecid       uint32                    `json:"video_codecid"`
	VideoProject       bool                      `json:"video_project"`
	Fnver              int                       `json:"fnver"`
	Fnval              int                       `json:"fnval"`
	Dash               *api.ResponseDash         `json:"dash,omitempty"`
	NoRexcode          int32                     `json:"no_rexcode,omitempty"`
}

// PlayerFileInfo is
type PlayerFileInfo struct {
	TimeLength int64  `json:"timelength"`
	FileSize   int64  `json:"filesize"`
	Ahead      string `json:"ahead,omitempty"`
	Vhead      string `json:"vhead,omitempty"`
	URL        string `json:"url,omitempty"`
	Order      int64  `json:"order,omitempty"`
}

// ArcType arctype
type ArcType struct {
	ID   int16  `json:"id"`
	Pid  int16  `json:"pid"`
	Name string `json:"name"`
}

// Videoshot videoshot
type Videoshot struct {
	// 定位文件
	PvData string `json:"pvdata"`
	// 一行多少小图
	XLen int `json:"img_x_len"`
	// 一列多少小图
	YLen int `json:"img_y_len"`
	// 缩略图宽
	XSize int `json:"img_x_size"`
	// 缩略图高
	YSize int      `json:"img_y_size"`
	Image []string `json:"image"`
	Attr  int32    `json:"-"`
}

type PGCPlayer struct {
	PlayerInfo *PlayerInfo `json:"player_info"`
	Aid        int64       `json:"aid"`
}

// PGCPlayurl is
type PGCPlayurl struct {
	PlayerInfo *api.BvcVideoItem `json:"player_info"`
	Aid        int64             `json:"aid"`
	IsPreview  int32             `json:"is_preview"`
	EpisodeId  int64             `json:"episode_id"`
	SeasonID   int64             `json:"season_id"`
	SeasonType int32             `json:"season_type"`
}

func FromDash(in *batch.ResponseDash) (out *api.ResponseDash) {
	out = new(api.ResponseDash)
	for _, v := range in.Video {
		if v == nil {
			continue
		}
		videoItem := &api.DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
		}
		out.Video = append(out.Video, videoItem)
	}
	for _, v := range in.Audio {
		if v == nil {
			continue
		}
		audioItem := &api.DashItem{
			Id:        v.Id,
			BaseUrl:   v.BaseUrl,
			Bandwidth: v.Bandwidth,
			Codecid:   v.Codecid,
			Size_:     v.Size_,
		}
		out.Audio = append(out.Audio, audioItem)
	}
	return
}

type ArcRedirect struct {
	Aid            int64                  `json:"aid"`
	RedirectType   api.RedirectType       `json:"redirect_type"`
	RedirectTarget string                 `json:"redirect_target"`
	PolicyType     api.RedirectPolicyType `json:"policy_type"`
	PolicyId       int64                  `json:"policy_id"`
}

type ArcExpand struct {
	Aid          int64     `json:"aid"`
	Mid          int64     `json:"mid"`
	ArcType      int64     `json:"arc_type"`
	RoomId       int64     `json:"room_id"`
	PremiereTime time.Time `json:"premiere_time"`
}

type SeasonEpisode struct {
	SeasonId  int64 `json:"season_id"`
	SectionId int64 `json:"section_id"`
	EpisodeId int64 `json:"episode_id"`
	Aid       int64 `json:"aid"`
	Attribute int64 `json:"attribute"`
}

func (sep *SeasonEpisode) AttrVal(bit uint) int32 {
	return int32((sep.Attribute >> bit) & int64(1))
}
