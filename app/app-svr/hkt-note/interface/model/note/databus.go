package note

import (
	xtime "go-common/library/time"
	"time"
)

const (
	TopicNtNotify    = "NoteNotify-T"
	TopicAuditNotify = "NoteAuditNotify-T"
	NeedPublish      = 1
)

type NtAuditNotify struct {
	Topic   string    `json:"topic"`
	Content *NtAddMsg `json:"content"`
}

type NtNotify struct {
	Topic   string       `json:"topic"`
	Content *NtNotifyMsg `json:"content"`
}

type NtAddMsg struct {
	Oid         int64      `json:"oid"`
	OidType     int        `json:"oid_type"`
	Mid         int64      `json:"mid"`
	NoteId      int64      `json:"note_id"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	NoteSize    int64      `json:"note_size"`
	AuditStatus int        `json:"audit_status"`
	Mtime       xtime.Time `json:"mtime"`
}

type NtDelMsg struct {
	NoteIds []int64 `json:"note_ids"`
	Mid     int64   `json:"mid"`
}

type NtPubMsg struct {
	Mid         int64  `json:"mid"`
	NoteId      int64  `json:"note_id"`
	ContLen     int64  `json:"cont_len"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Oid         int64  `json:"oid"`
	OidType     int    `json:"oid_type"`
	ArcCover    string `json:"arc_cover"`
	Original    int    `json:"original"`
	AutoComment int    `json:"auto_comment"`
	PubFrom     int    `json:"pub_from"`
}

type NtNotifyMsg struct {
	NtAddMsg *NtAddMsg `json:"nt_add_msg,omitempty"`
	NtDelMsg *NtDelMsg `json:"nt_del_msg,omitempty"`
	NtPubMsg *NtPubMsg `json:"nt_pub_msg,omitempty"`
	ReplyMsg *ReplyMsg `json:"reply_msg"`
}

type ReplyMsg struct {
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Oid     int64  `json:"oid"`
	Mid     int64  `json:"mid"`
}

func (v *NoteAddReq) ToNtNotifyMsg(mid int64, noteSize int64) *NtAddMsg {
	return &NtAddMsg{
		Mid:      mid,
		Oid:      v.Oid,
		OidType:  v.OidType,
		NoteId:   v.NoteId,
		Title:    v.Title,
		Summary:  v.Summary,
		NoteSize: noteSize,
		Mtime:    xtime.Time(time.Now().Unix()),
	}
}

func (v *NoteAddReq) ToNtPubMsg(mid int64, arc *ArcCore) *NtPubMsg {
	var arcCover string
	if arc != nil {
		arcCover = arc.Pic
	}
	return &NtPubMsg{
		Mid:         mid,
		NoteId:      v.NoteId,
		ContLen:     v.ContLen,
		Title:       v.Title,
		Summary:     v.Summary,
		Oid:         v.Oid,
		OidType:     v.OidType,
		ArcCover:    arcCover,
		Original:    v.Original,
		AutoComment: v.AutoComment,
		PubFrom:     v.PubFrom,
	}
}

func ToDelNotifyMsg(mid int64, noteIds []int64) *NtNotifyMsg {
	return &NtNotifyMsg{
		NtDelMsg: &NtDelMsg{
			NoteIds: noteIds,
			Mid:     mid,
		},
	}
}
