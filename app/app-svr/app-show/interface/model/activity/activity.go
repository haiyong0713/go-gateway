package activity

type Activity struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	H5URL   string `json:"h5_url"`
	H5Cover string `json:"h5_cover"`
	Desc    string `json:"desc"`
	Stime   string `json:"stime"`
	Etime   string `json:"etime"`
}

const (
	// Subject.Type
	VideoLike  = 4
	Article    = 12
	Video2     = 13
	PhoneVideo = 16
)
