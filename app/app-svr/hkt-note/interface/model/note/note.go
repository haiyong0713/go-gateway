package note

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
	"unicode/utf8"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

const (
	_platWeb            = 1
	_platAndroid        = 2
	_platIOS            = 3
	_arcStatusOK        = 0
	_cheeseStatusReturn = -2
	_cheeseStatusLock   = -3

	ActionAdd       = 1
	ActionDel       = 2
	ActionAuditFail = 3
	ActionView      = 4
	ActionPub       = 5
	AuditPass       = 0
	NeedAudit       = 1
	OidTypeUgc      = 0
	OidTypeCheese   = 1
	ArcStatusWrong  = 1
)

type NoteAddReq struct {
	Bvid        string `form:"bvid"`
	Aid         int64  `form:"aid"` // TODO 废弃
	Oid         int64  `form:"oid"`
	OidType     int    `form:"oid_type" validate:"lt=2"`
	NoteId      int64  `form:"note_id"`
	Title       string `form:"title" validate:"required"`
	Summary     string `form:"summary" validate:"required"`
	Content     string `form:"content"`
	Tags        string `form:"tags"`
	NeedAudit   int    `form:"cls"`
	ContLen     int64  `form:"cont_len"`
	Hash        string `form:"hash"`
	Publish     int    `form:"publish"`      // 是否发布
	Original    int    `form:"original"`     // 发布是否原创
	AutoComment int    `form:"auto_comment"` // 发布过审后是否自动发评论
	PubFrom     int    `form:"pub_from"`     // 发布来源，0-稿件页 1-评论区
	Device
	CommentFormat int32 `form:"comment_format"` //公开笔记在评论区样式，1表示旧样式 2 表示新样式；没传也表示旧样式
}

type NoteInfoReq struct {
	NoteId  int64  `form:"note_id" validate:"required"`
	Bvid    string `form:"bvid"`
	Aid     int64  `form:"aid"` // TODO 废弃
	Oid     int64  `form:"oid"`
	OidType int    `form:"oid_type" validate:"lt=2"`
	Device
}

type NoteDelReq struct {
	NoteId  int64   `form:"note_id"` // TODO 去掉单点删除
	NoteIds []int64 `form:"note_ids,split"`
	Device
}

type NoteAddRes struct {
	NoteId string `json:"note_id"`
}

type NoteInfoRes struct {
	Title              string     `json:"title"`
	Summary            string     `json:"summary"`
	Content            string     `json:"content"`
	CidCount           int64      `json:"cid_count"`
	AuditStatus        int64      `json:"audit_status"`
	PubStatus          int64      `json:"pub_status"`
	PubReason          string     `json:"pub_reason"`
	PubVersion         int64      `json:"pub_version"`
	ForbidNoteEntrance bool       `json:"forbid_note_entrance"`
	Tags               []*NoteTag `json:"tags"`
	Arc                *ArcCore   `json:"arc"`
}

type ArcCore struct {
	Oid       int64  `json:"oid"`
	Title     string `json:"title"`
	UpMid     int64  `json:"up_mid"`
	UpName    string `json:"-"`
	TypeId    int32  `json:"-"`
	TypeName  string `json:"-"`
	Status    int    `json:"status"` // 1-稿件状态不合法
	OidType   int    `json:"oid_type"`
	Pic       string `json:"pic"`
	Desc      string `json:"desc"`
	CvidCount int64  `json:"cvid_count"`
}

type NoteTag struct {
	Cid     int64 `json:"cid"`
	Status  int64 `json:"status"`
	Index   int64 `json:"index"`
	Seconds int64 `json:"seconds"`
}
type UserGray struct {
	IsGray bool `json:"is_gray"`
}

type NtContent struct {
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Tag     string `json:"tag"`
	Mid     int64  `json:"mid"`
	ContLen int64  `json:"cont_len"`
}

type NtInfoc struct {
	Mid      int64  `json:"mid"`
	Aid      int64  `json:"aid"`
	Title    string `json:"title"`
	UpMid    int64  `json:"up_mid"`
	UpName   string `json:"up_name"`
	TypeId   int32  `json:"type_id"`
	TypeName string `json:"type_name"`
	Ctime    string `json:"ctime"`
	Action   int    `json:"action"`
	Plat     int64  `json:"plat"`
	NoteId   int64  `json:"note_id"`
}

type ContentBody struct {
	Insert     interface{} `json:"insert"` // 只有string类型的正文
	Attributes interface{} `json:"attributes,omitempty"`
}

type NoteCount struct {
	Total       int64 `json:"total"`
	FromArchive int64 `json:"from_archive"`
	FromCheese  int64 `json:"from_cheese"`
}

func (v *ArcCore) FromUGC(data *arcgrpc.Arc) {
	v.Oid = data.Aid
	v.Title = data.Title
	v.TypeId = data.TypeID
	v.TypeName = data.TypeName
	v.UpMid = data.Author.Mid
	v.UpName = data.Author.Name
	v.OidType = OidTypeUgc
	v.Pic = data.Pic
	if data.State < _arcStatusOK {
		v.Status = ArcStatusWrong
	}
}

func (v *ArcCore) FromCheese(data *cssngrpc.SeasonCard) {
	v.Oid = int64(data.Id)
	v.Title = data.Title
	v.TypeName = "课堂"
	v.UpMid = data.UpId
	v.OidType = OidTypeCheese
	v.Pic = data.Cover
	if data.Status == _cheeseStatusLock || data.Status == _cheeseStatusReturn {
		v.Status = ArcStatusWrong
	}
}

func (v *NoteAddReq) ToNtContent(mid int64) *NtContent {
	return &NtContent{
		NoteId:  v.NoteId,
		Content: v.Content,
		Tag:     v.Tags,
		Mid:     mid,
		ContLen: v.ContLen,
	}
}

func (v *NoteInfoRes) From(val *notegrpc.NoteInfoReply, arc *ArcCore, isForbid bool) {
	v.Title = val.Title
	v.Content = val.Content
	v.Summary = val.Summary
	v.CidCount = val.CidCount
	v.AuditStatus = val.AuditStatus
	v.PubStatus = val.PubStatus
	v.PubReason = val.PubReason
	v.PubVersion = val.PubVersion
	v.ForbidNoteEntrance = isForbid
	v.Tags = make([]*NoteTag, 0)
	v.Arc = arc
	for _, t := range val.Tags {
		tmp := &NoteTag{
			Cid:     t.Cid,
			Status:  t.Status,
			Index:   t.Index,
			Seconds: t.Seconds,
		}
		v.Tags = append(v.Tags, tmp)
	}
}

func ToNtInfoc(arc *ArcCore, mid int64, action int, plat, noteId int64) *NtInfoc {
	return &NtInfoc{
		Mid:      mid,
		Aid:      arc.Oid,
		Title:    arc.Title,
		UpMid:    arc.UpMid,
		UpName:   arc.UpName,
		TypeId:   arc.TypeId,
		TypeName: arc.TypeName,
		Ctime:    time.Now().Format("2006-01-02 15:04:05"),
		Action:   action,
		Plat:     plat,
		NoteId:   noteId,
	}
}

var (
	FileTypeJPG  = "image/jpg"  // FileTypeJPG file type jpg.
	FileTypeJPEG = "image/jpeg" // FileTypeJPEG file type jpeg.
	FileTypePNG  = "image/png"  // FileTypePNG file type png.
)

func ToContentLen(data string) int64 { // json转正文
	bodyArr := make([]*ContentBody, 0)
	if err := json.Unmarshal([]byte(data), &bodyArr); err != nil {
		log.Error("noteWarn ToBody data(%s) error(%v)", data, err)
		return 0
	}
	var count int64
	for _, b := range bodyArr {
		if b == nil || b.Insert == nil {
			continue
		}
		if reflect.TypeOf(b.Insert).Name() == "string" {
			str := fmt.Sprintf("%v", b.Insert)
			count += int64(utf8.RuneCountInString(str))
		}
	}
	return count
}

type NoteListInArcReply struct {
	NoteIds []string `json:"noteIds"`
}

type IsForbidReply struct {
	ForbidNoteEntrance bool `json:"forbid_note_entrance"`
}
