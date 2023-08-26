package dynamic

type DrawDetailRes struct {
	Item *DrawItem `json:"item"`
	User *User     `json:"user"`
}

type DrawItem struct {
	ID            int64          `json:"id"`
	Pictures      []*DrawPicture `json:"pictures"`
	PicturesCount int            `json:"pictures_count"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Reply         int            `json:"reply"`
	UploadTime    int64          `json:"upload_time"`
	AtControl     string         `json:"at_control"`
}

type DrawPicture struct {
	ImgSrc    string        `json:"img_src"`
	ImgHeight int64         `json:"img_height"`
	ImgWidth  int64         `json:"img_width"`
	ImgSize   float32       `json:"img_size"`
	ImgTags   []*DrawImgTag `json:"img_tags"`
}

type DrawImgTag struct {
	Text        string `json:"text"`
	Type        int32  `json:"type"`
	Url         string `json:"url"`
	X           int64  `json:"x"`
	Y           int64  `json:"y"`
	Orientation int32  `json:"orientation"`
	SchemaURL   string `json:"schema_url"`
	ItemID      int64  `json:"item_id"`
	Source      int32  `json:"source_type"`
	Mid         int64  `json:"mid"`
	Tid         int64  `json:"tid"`
	Poi         string `json:"poi"`
}

type User struct {
	UID     int64  `json:"uid"`
	HeadURL string `json:"head_url"`
	Name    string `json:"name"`
}
