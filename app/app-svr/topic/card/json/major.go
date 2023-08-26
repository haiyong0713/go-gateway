package jsonwebcard

type DynMajor interface {
	GetMajorType() MajorType
}

type MajorArchive struct {
	MajorType MajorType `json:"type,omitempty"`
	Archive   *Archive  `json:"archive,omitempty"`
}

func (major MajorArchive) GetMajorType() MajorType {
	return major.MajorType
}

type Archive struct {
	MediaType      MediaType `json:"media_type,omitempty"`
	Bvid           string    `json:"bvid,omitempty"`
	Aid            int64     `json:"aid,omitempty"`
	Cover          string    `json:"cover,omitempty"`
	JumpUrl        string    `json:"jump_url,omitempty"`
	VideoStat      VideoStat `json:"stat,omitempty"`
	DurationText   string    `json:"duration_text,omitempty"`
	Title          string    `json:"title,omitempty"`
	Desc           string    `json:"desc,omitempty"`
	Badge          Badge     `json:"badge,omitempty"`
	DisablePreview bool      `json:"disable_preview,omitempty"`
}

type VideoStat struct {
	Danmaku string `json:"danmaku"`
	Play    string `json:"play"`
}

type Badge struct {
	Text    string `json:"text,omitempty"`
	BgColor string `json:"bg_color,omitempty"`
	Color   string `json:"color,omitempty"`
}

type MajorDraw struct {
	MajorType MajorType `json:"type,omitempty"`
	Draw      *Draw     `json:"draw,omitempty"`
}

func (major MajorDraw) GetMajorType() MajorType {
	return major.MajorType
}

type Draw struct {
	Id    int64       `json:"id,omitempty"`
	Items []*DrawItem `json:"items,omitempty"`
}

type DrawItem struct {
	// 图片链接
	Src string `json:"src,omitempty"`
	// 图片宽度
	Width int64 `json:"width,omitempty"`
	// 图片高度
	Height      int64          `json:"height,omitempty"`
	DrawItemTag []*DrawItemTag `json:"draw_item_tags,omitempty"`
}

type DrawItemTag struct {
	DrawTagType DrawTagType `json:"draw_tag_type,omitempty"`
	JumpUrl     string      `json:"jump_url,omitempty"`
	Text        string      `json:"text,omitempty"`
	X           int64       `json:"x,omitempty"`
	Y           int64       `json:"y,omitempty"`
	Orientation int32       `json:"orientation,omitempty"`
	Source      int32       `json:"source,omitempty"`
	Tid         int64       `json:"tid,omitempty"`
	Mid         int64       `json:"mid,omitempty"`
	Poi         string      `json:"poi,omitempty"`
	SchemaUrl   string      `json:"schema_url,omitempty"`
}

type MajorArticle struct {
	MajorType MajorType `json:"type,omitempty"`
	Article   *Article  `json:"article,omitempty"`
}

func (major MajorArticle) GetMajorType() MajorType {
	return major.MajorType
}

type Article struct {
	ID      int64    `json:"id,omitempty"`
	Title   string   `json:"title,omitempty"`
	Desc    string   `json:"desc,omitempty"`
	JumpUrl string   `json:"jump_url,omitempty"`
	Covers  []string `json:"covers,omitempty"`
	Label   string   `json:"label,omitempty"`
}

type MajorCommon struct {
	MajorType MajorType `json:"type,omitempty"`
	Common    *Common   `json:"common,omitempty"`
}

func (major MajorCommon) GetMajorType() MajorType {
	return major.MajorType
}

type Common struct {
	ID       int64  `json:"id,omitempty"`
	JumpUrl  string `json:"jump_url,omitempty"`
	Cover    string `json:"cover,omitempty"`
	Title    string `json:"title,omitempty"`
	Desc     string `json:"desc,omitempty"`
	Label    string `json:"label,omitempty"`
	Biz      int64  `json:"biz,omitempty"`
	SketchId int64  `json:"sketch_id,omitempty"`
	Badge    Badge  `json:"badge,omitempty"`
	Style    int64  `json:"style,omitempty"`
}

type MajorPGC struct {
	MajorType MajorType `json:"type,omitempty"`
	PGC       *PGC      `json:"pgc,omitempty"`
}

func (major MajorPGC) GetMajorType() MajorType {
	return major.MajorType
}

type PGC struct {
	Type      int64     `json:"type,omitempty"`
	SubType   int64     `json:"sub_type,omitempty"`
	SeasonId  int64     `json:"season_id,omitempty"`
	EpId      int64     `json:"ep_id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Cover     string    `json:"cover,omitempty"`
	Badge     Badge     `json:"badge,omitempty"`
	JumpUrl   string    `json:"jump_url,omitempty"`
	VideoStat VideoStat `json:"stat,omitempty"`
}
