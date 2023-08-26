package config

type Participation struct {
	BaseCfgManager

	Items []*ParticipationItem
}

type ParticipationItem struct {
	Type          int64 //投稿类型
	Sid           int64
	ButtonContent string
	UploadType    int64
	NewTid        int64 //新话题id
}
