package note

import (
	"strings"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/job/model/article"
)

const (
	NotDeleted    = 0
	Deleted       = 1
	OidTypeUGC    = 0
	OidTypeCheese = 1
)

type Binlog struct {
	Action string `json:"action"`
	Table  string `json:"table"`
}

type NoteDetailBlog struct {
	Old *NtDetailDB `json:"old"`
	New *NtDetailDB `json:"new"`
}

type NoteContentBlog struct {
	Old *NtContentDB `json:"old"`
	New *NtContentDB `json:"new"`
}

type NtDetailDB struct {
	NoteId      int64  `json:"note_id"`
	Mid         int64  `json:"mid"`
	Aid         int64  `json:"aid"`
	NoteIdx     int64  `json:"note_idx"`
	NoteSize    int64  `json:"note_size"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	AuditStatus int    `json:"audit_status"`
	Deleted     int    `json:"deleted"`
	Ctime       string `json:"ctime"`
	Mtime       string `json:"mtime"`
	OidType     int    `json:"oid_type"`
}

type NtContentDB struct {
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Tag     string `json:"tag"`
	Deleted int    `json:"deleted"`
	Ctime   string `json:"ctime"`
	Mtime   string `json:"mtime"`
}

type DtlCache struct {
	NoteId      int64      `json:"note_id"`
	Aid         int64      `json:"aid"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	NoteSize    int64      `json:"note_size"`
	AuditStatus int        `json:"audit_status"`
	Deleted     int        `json:"deleted"`
	Mid         int64      `json:"mid"`
	Mtime       xtime.Time `json:"mtime"`
	OidType     int        `json:"oid_type"`
}

type ContCache struct {
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Tag     string `json:"tag"`
	Deleted int    `json:"deleted"`
}

type UserCache struct {
	Mid       int64 `json:"mid"`
	NoteSize  int64 `json:"note_size"`
	NoteCount int64 `json:"note_count"`
}

type ArcCore struct {
	Oid      int64  `json:"oid"`
	Title    string `json:"title"`
	UpMid    int64  `json:"-"`
	UpName   string `json:"-"`
	TypeId   int32  `json:"-"`
	TypeName string `json:"-"`
	Status   int    `json:"status"` // 1-稿件状态不合法
	OidType  int    `json:"oid_type"`
	Pic      string `json:"pic"`
	Desc     string `json:"desc"`
}

func (v *Binlog) ToTableName() string {
	if strings.Contains(v.Table, "note_detail") {
		return "note_detail"
	}
	return v.Table
}

func (v *NtDetailDB) ToDtlCache() *DtlCache {
	mtime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Mtime, time.Local)
	return &DtlCache{
		NoteId:      v.NoteId,
		Aid:         v.Aid,
		Title:       v.Title,
		Summary:     v.Summary,
		NoteSize:    v.NoteSize,
		AuditStatus: v.AuditStatus,
		Deleted:     v.Deleted,
		Mid:         v.Mid,
		Mtime:       xtime.Time(mtime.Unix()),
		OidType:     v.OidType,
	}
}

func (v *NtContentDB) ToContCache(cont, tag string) *ContCache {
	return &ContCache{
		NoteId:  v.NoteId,
		Content: cont,
		Tag:     tag,
		Deleted: v.Deleted,
	}
}

func (v *ContCache) ToArtCont(cvid int64) *article.ArtContCache {
	return &article.ArtContCache{
		Cvid:    cvid,
		NoteId:  v.NoteId,
		Content: v.Content,
		Tag:     v.Tag,
	}
}
