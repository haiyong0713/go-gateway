package result

type VideoUpInfo struct {
	Table  string
	Action string
	Nw     *Video
	Old    *Video
}

type Video struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	Cid         int64  `json:"cid"`
	Aid         int64  `json:"aid"`
	Title       string `json:"eptitle"`
	Desc        string `json:"description"`
	SrcType     string `json:"src_type"`
	Duration    int64  `json:"duration"`
	Filesize    int64  `json:"filesize"`
	Resolutions string `json:"resolutions"`
	Playurl     string `json:"playurl"`
	FailCode    int8   `json:"failinfo"`
	Index       int    `json:"index_order"`
	Attribute   int32  `json:"attribute"`
	XcodeState  int8   `json:"xcode_state"`
	Status      int16  `json:"status"`
	CTime       string `json:"ctime"`
	MTime       string `json:"mtime"`
}
