package archive

import "time"

// const is
const (
	UpFromPGC       = 1
	UpFromPGCSecret = 5

	//archive_biz.type
	BizTypeArchive    = 14 //稿件
	BizTypeArchivePay = 17 //付费稿件

	//archive_biz.state
	BizStateOk = 0

	//首映稿件属性位
	AttrPremiere = uint(12)
	//付费稿件属性位
	AttrPay = uint(13)

	InnerAttrYes = int64(1)
	InnerAttrNo  = int64(0)
)

// Addit addit struct
type Addit struct {
	ID          int64
	Aid         int64
	Desc        string
	Source      string
	RedirectURL string
	MissionID   int64
	UpFrom      int32
	OrderID     int
	Dynamic     string
	InnerAttr   int64
	Ipv6        []byte
}

type Biz struct {
	Aid     int64
	Data    string
	SubType int
}

// AttrVal get attribute value.
func (ad *Addit) AttrVal(bit uint) int64 {
	return (ad.InnerAttr >> bit) & int64(1)
}

func (sep *SeasonEpisode) AttrVal(bit uint) int32 {
	return int32((sep.Attribute >> bit) & int64(1))
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
