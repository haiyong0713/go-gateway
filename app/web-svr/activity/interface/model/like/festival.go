package like

// ImageList ...
type ImageList struct {
	Img string `json:"img"`
}

// FestivalProcessReply ...
type FestivalProcessReply struct {
	ImageList []ImageList `json:"img_list"`
}
