package model

type LoadReply struct {
	Url string `json:"url"`
	Img string `json:"img"`
	Qn  int    `json:"qn"`
}

type FramesCache struct {
	Aid       int    `json:"aid"`
	Cid       int    `json:"cid"`
	Url       string `json:"url"`
	KeyFrames string `json:"key_frames"`
}
