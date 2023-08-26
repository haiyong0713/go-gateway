package bplus

import (
	xtime "go-common/library/time"
)

// Picture struct
type Picture struct {
	DynamicID    int64      `json:"dynamic_id,omitempty"`
	PublishTime  xtime.Time `json:"publish_time,omitempty"`
	Mid          int64      `json:"mid,omitempty"`
	RidType      int8       `json:"rid_type,omitempty"`
	Rid          int64      `json:"rid,omitempty"`
	ImgCount     int        `json:"img_count,omitempty"`
	Imgs         []string   `json:"imgs,omitempty"`
	DynamicText  string     `json:"dynamic_text,omitempty"`
	ViewCount    int64      `json:"view_count,omitempty"`
	Topics       []string   `json:"topics,omitempty"`
	NickName     string     `json:"nick_name,omitempty"`
	FaceImg      string     `json:"face_img,omitempty"`
	CommentCount int64      `json:"comment_count,omitempty"`
	LikeCount    int32      `json:"like_count,omitempty"`
	TopicInfos   []struct {
		TopicID    int64  `json:"topic_id,omitempty"`
		TopicName  string `json:"topic_name,omitempty"`
		TopicLink  string `json:"topic_link,omitempty"`
		IsActivity int    `json:"is_activity,omitempty"`
	} `json:"topic_infos,omitempty"`
	JumpUrl      string `json:"jump_url,omitempty"`
	IsNewChannel bool   `json:"is_new_channel,omitempty"`
}
