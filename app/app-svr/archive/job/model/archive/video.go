package archive

const (
	VideoStatusOpen   = int16(0)
	VideoStatusAccess = int16(10000)
	VideoStatusSubmit = int16(-30)
	VideoStatusDelete = int16(-100)
	VideoRelationBind = int16(0)
)

type VideoUpInfo struct {
	Nw  *Video
	Old *Video
}

type Video struct {
	ID          int64  `json:"id"`
	Aid         int64  `json:"aid"`
	Title       string `json:"eptitle"`
	Desc        string `json:"description"`
	Filename    string `json:"filename"`
	SrcType     string `json:"src_type"`
	Cid         int64  `json:"cid"`
	Duration    int64  `json:"duration"`
	Filesize    int64  `json:"filesize"`
	Resolutions string `json:"resolutions"`
	Index       int    `json:"index_order"`
	CTime       string `json:"ctime"`
	MTime       string `json:"mtime"`
	Status      int16  `json:"status"`
	State       int16  `json:"state"`
	Playurl     string `json:"playurl"`
	Attribute   int32  `json:"attribute"`
	FailCode    int8   `json:"failinfo"`
	XcodeState  int8   `json:"xcode_state"`
	WebLink     string `json:"weblink"`
	Dimensions  string `json:"dimensions"`
}

// SteinsCid gives the real first cid of the steins-gate video
type SteinsCid struct {
	Route string `json:"route"`
	Aid   int64  `json:"aid"`
	Cid   int64  `json:"cid"`
}

// VideoFF is
type VideoFF struct {
	Cid        int64
	FirstFrame string
}
