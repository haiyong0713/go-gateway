package model

type ReplyWallModel struct {
	ID           int64  `json:"id"`
	ContestID    int64  `json:"contest_id"`
	Mid          int64  `json:"mid"`
	ReplyDetails string `json:"reply_details"`
	IsDeleted    int64  `json:"is_deleted"`
}

func (r *ReplyWallModel) TableName() string {
	return "es_reply_wall"
}
