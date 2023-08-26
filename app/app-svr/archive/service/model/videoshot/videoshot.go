package videoshot

// Videoshot is struct.
type Videoshot struct {
	Cid     int64  `json:"cid"`
	Count   int64  `json:"cnt"`
	HDImg   string `json:"hd_image"`
	HDCount int64  `json:"hd_count"`
	SdCount int64  `json:"sd_count"`
	SdImg   string `json:"sd_image"`
}
