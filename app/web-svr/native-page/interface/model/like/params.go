package like

import (
	"encoding/json"

	"go-common/library/time"
)

const (
	ActUpdate = "update"
	ActInsert = "insert"
	ActDelete = "delete"
)

// Message canal binlog message.
type Message struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

// ParamMsg notify param msg.
type ParamMsg struct {
	Msg string `form:"msg" validate:"required"`
}

// PageMsgPub .
type PageMsgPub struct {
	Category string      `json:"category"`
	Value    *DynamicMsg `json:"value,omitempty"`
}

// DynamicMsg .
type DynamicMsg struct {
	PageID       int64     `json:"page_id"`
	TopicID      int64     `json:"topic_id"`
	TopicName    string    `json:"topic_name"`
	Online       int       `json:"online"`
	TopicLink    string    `json:"topic_link"`
	Uid          int64     `json:"uid"`
	Stime        time.Time `json:"stime"`
	Etime        time.Time `json:"etime"`
	ActType      int32     `json:"act_type"`
	Hot          int64     `json:"hot"`
	DynamicID    int64     `json:"dynamic_id"`
	Attribute    int64     `json:"attribute"`
	PcURL        string    `json:"pc_url"`
	AnotherTitle string    `json:"another_title"`
	FromType     int32     `json:"from_type"`
	State        int64     `json:"state"`
}

// SortModule .
type SortModule struct {
	IDs     []int64 `json:"ids"`
	HasMore int32   `json:"has_more"`
	Offset  int64   `json:"offset"`
}
