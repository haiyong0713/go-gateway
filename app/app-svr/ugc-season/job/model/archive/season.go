package archive

// is
const (
	ShowYes = 1
	ShowNo  = 0
)

// Season is
type Season struct {
	ID        int64  `json:"id"`
	SeasonID  int64  `json:"season_id"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Cover     string `json:"cover"`
	Mid       int64  `json:"mid"`
	Attribute int64  `json:"attribute"`
	SignState int32  `json:"sign_state"`
	Show      int32  `json:"show"`
	State     int32  `json:"state"`
	CTime     string `json:"ctime"`
	MTime     string `json:"mtime"`
	EpNum     int64  `json:"ep_num"`
}

type SeasonSection struct {
	ID        int64  `json:"id"`
	SeasonID  int64  `json:"season_id"`
	SectionID int64  `json:"section_id"`
	Type      int32  `json:"type"`
	Title     string `json:"title"`
	Order     int64  `json:"order"`
	Show      int32  `json:"show"`
	State     int32  `json:"state"`
	CTime     string `json:"ctime"`
	MTime     string `json:"mtime"`
}

type SeasonEp struct {
	ID        int64  `json:"id"`
	SeasonID  int64  `json:"season_id"`
	SectionID int64  `json:"section_id"`
	EpID      int64  `json:"episode_id"`
	Title     string `json:"title"`
	AID       int64  `json:"aid"`
	CID       int64  `json:"cid"`
	Order     int64  `json:"order"`
	Attribute int64  `json:"attribute"`
	Show      int32  `json:"show"`
	State     int32  `json:"state"`
	CTime     string `json:"ctime"`
	MTime     string `json:"mtime"`
}
