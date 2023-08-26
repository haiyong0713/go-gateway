package search

type DynamicTopics struct {
	DynamicID   int64          `json:"dynamic_id,omitempty"`
	FromContent []*FromContent `json:"from_content,omitempty"`
}

type FromContent struct {
	TopicID    int64  `json:"topic_id,omitempty"`
	TopicName  string `json:"topic_name,omitempty"`
	IsActivity int    `json:"is_activity,omitempty"`
	TopicLink  string `json:"topic_link,omitempty"`
}

// Detail struct
type Detail struct {
	ID              int64  `json:"dynamic_id,omitempty"`
	Mid             int64  `json:"mid,omitempty"`
	FaceImg         string `json:"face_img,omitempty"`
	NickName        string `json:"nick_name,omitempty"`
	PublishTimeText string `json:"publish_time_text,omitempty"`
	ImgCount        int    `json:"img_count,omitempty"`
	ViewCount       int    `json:"view_count,omitempty"`
	CommentCount    int    `json:"comment_count,omitempty"`
	LikeCount       int    `json:"like_count,omitempty"`
}
