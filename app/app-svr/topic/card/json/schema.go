package jsonwebcard

type TopicCard interface {
	GetTopicCardType() TopicCardType
	GetModules() *Modules
}

type Base struct {
	IdStr    string   `json:"id_str"`
	CardType CardType `json:"type"`
	Visible  bool     `json:"visible"`
	TopicId  int64    `json:"topic_id,omitempty"`
}

type Basic struct {
	RidStr       string    `json:"rid_str"`
	CommentType  int64     `json:"comment_type"`
	CommentIdStr string    `json:"comment_id_str"`
	LikeShowIcon *LikeIcon `json:"like_show_icon,omitempty"`
}

type LikeIcon struct {
	NewIconId int64  `json:"new_icon_id,omitempty"`
	StartUrl  string `json:"start_url,omitempty"`
	ActionUrl string `json:"action_url,omitempty"`
	EndUrl    string `json:"end_url,omitempty"`
}

type Fold struct {
	FoldType   FoldType  `json:"fold_type"`
	FoldUser   *UserInfo `json:"fold_user"`
	Statement  string    `json:"statement"`
	DynamicIDs []int64   `json:"dynamic_ids"`
}

type UserInfo struct {
	Face    string `json:"face"`
	Name    string `json:"name"`
	Mid     int64  `json:"mid"`
	FaceNft int32  `json:"face_nft"`
}
