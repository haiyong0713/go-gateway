package dynamic

import "encoding/json"

// Resources .
type Resources struct {
	Array []*RidInfo `json:"array"`
}

// RidInfo .
type RidInfo struct {
	Rid  int64 `json:"rid"`
	Type int64 `json:"type"`
}

// TopicCount .
type TopicCount struct {
	ViewCount    *int64 `json:"view_count"`
	DiscussCount *int64 `json:"discuss_count"`
}

// DyReply .
type DyReply struct {
	HasMore int       `json:"has_more"`
	Offset  string    `json:"offset"`
	Cards   []*DyCard `json:"cards"`
}

// DyResult
type DyResult struct {
	Cards map[int64]*DyCard `json:"cards"`
}

// DyCard dynamic data.
type DyCard struct {
	Card json.RawMessage `json:"card"`
	Desc struct {
		UID           int64       `json:"uid"`
		Type          int         `json:"type"`
		ACL           int         `json:"acl"`
		Rid           int64       `json:"rid"`
		View          int32       `json:"view"`
		Repost        int32       `json:"repost"`
		Comment       int32       `json:"comment"`
		Like          int32       `json:"like"`
		IsLiked       int32       `json:"is_liked"`
		DynamicID     int64       `json:"dynamic_id"`
		CommentID     *int64      `json:"comment_id,omitempty"`
		Timestamp     int64       `json:"timestamp"`
		PreDyID       int64       `json:"pre_dy_id"`
		OrigDyID      int64       `json:"orig_dy_id"`
		OrigType      int         `json:"orig_type"`
		RType         int         `json:"r_type"`
		InnerID       int64       `json:"inner_id"`
		SpecType      int         `json:"spec_type"`
		Status        int         `json:"status"`
		UIDType       int         `json:"uid_type"`
		DynamicIDStr  string      `json:"dynamic_id_str"`
		RecommendInfo interface{} `json:"recommend_info,omitempty"`
		UserProfile   interface{} `json:"user_profile"`
	} `json:"desc"`
	ExtendJSON    json.RawMessage `json:"extend_json"`
	NeedRefresh   *int            `json:"need_refresh,omitempty"`
	Display       interface{}     `json:"display,omitempty"`
	ActivityInfos interface{}     `json:"activity_infos,omitempty"`
}

type BriefReply struct {
	HasMore  int         `json:"has_more"`
	Offset   string      `json:"offset"`
	Dynamics []*Dynamics `json:"dynamics"`
}

type Dynamics struct {
	Rid  int64 `json:"rid"`
	Type int64 `json:"type"`
}

// ResourceDynReq .
type ResourceDynReq struct {
	TopicID  int64  `json:"topic_id"`
	Types    string `json:"types"`
	PageSize int64  `json:"page_size"`
	Offset   string `json:"offset"`
	Mid      int64
	MobiApp  string
	Device   string
}
