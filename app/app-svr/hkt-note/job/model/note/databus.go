package note

import (
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/job/model/article"

	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	frontgrpc "git.bilibili.co/bapis/bapis-go/frontend/bilinote/v1"
)

type NtNotify struct {
	Topic   string       `json:"topic"`
	Content *NtNotifyMsg `json:"content"`
}

type NtNotifyMsg struct {
	NtDelMsg *NtDelMsg `json:"nt_del_msg,omitempty"`
	NtAddMsg *NtAddMsg `json:"nt_add_msg,omitempty"`
	NtPubMsg *NtPubMsg `json:"nt_pub_msg,omitempty"`
	ReplyMsg *ReplyMsg `json:"reply_msg"`
}

type ReplyMsg struct {
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Oid     int64  `json:"oid"`
	Mid     int64  `json:"mid"`
}

type NtAuditNotify struct {
	Topic   string    `json:"topic"`
	Content *NtAddMsg `json:"content"`
}

type NtPubMsg struct {
	Mid      int64  `json:"mid"`
	NoteId   int64  `json:"note_id"`
	ContLen  int64  `json:"cont_len"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Oid      int64  `json:"oid"`
	OidType  int    `json:"oid_type"`
	ArcCover string `json:"arc_cover"`
	// retry use
	Cvid        int64 `json:"cvid"`
	Mtime       int64 `json:"mtime"`
	PubVersion  int64 `json:"pub_version"`
	Original    int32 `json:"original"`
	AutoComment int32 `json:"auto_comment"`
	PubFrom     int   `json:"pub_from"`
}

type NtDelMsg struct {
	NoteId  int64   `json:"note_id"` // TODO del
	NoteIds []int64 `json:"note_ids"`
	Mid     int64   `json:"mid"`
}

type NtAddMsg struct {
	Aid         int64      `json:"aid,omitempty"` // TODO del
	Oid         int64      `json:"oid,omitempty"`
	OidType     int        `json:"oid_type,omitempty"`
	Mid         int64      `json:"mid,omitempty"`
	NoteId      int64      `json:"note_id,omitempty"`
	Title       string     `json:"title,omitempty"`
	Summary     string     `json:"summary,omitempty"`
	NoteSize    int64      `json:"note_size,omitempty"`
	AuditStatus int        `json:"audit_status,omitempty"`
	Content     string     `json:"content,omitempty"`
	Mtime       xtime.Time `json:"mtime,omitempty"`
}

func (v *NtPubMsg) ToArgArticle(cvid int64, htmlCont string, imageUrls []string, cate int64) *artgrpc.ArgArticle {
	if cvid < 0 {
		cvid = 0
	}
	outImgUrls, tempId := toImageInfo(imageUrls)
	return &artgrpc.ArgArticle{
		Aid:        cvid,
		Category:   cate,
		Title:      v.Title,
		Summary:    v.Summary,
		BannerURL:  v.ArcCover,
		Mid:        v.Mid,
		ImageURLs:  outImgUrls,
		Content:    htmlCont,
		Words:      v.ContLen,
		Original:   v.Original,
		TemplateID: tempId,
		CoverAvid:  v.Oid,
	}
}

func toImageInfo(imageUrls []string) ([]string, int32) {
	switch len(imageUrls) {
	case 0:
		return imageUrls, article.ArtTempPicNone
	case 1:
		return imageUrls, article.ArtTempPicOne
	case 2: // nolint:gomnd
		return imageUrls[:1], article.ArtTempPicOne
	default:
		return imageUrls[:3], article.ArtTempPicThree
	}
}

func (v *NtPubMsg) ToCont(tag string, biliRes *frontgrpc.NoteReply) *article.ArtContCache {
	return &article.ArtContCache{
		Cvid:       v.Cvid,
		NoteId:     v.NoteId,
		Content:    biliRes.BiliJson,
		Tag:        tag,
		Mid:        v.Mid,
		Mtime:      xtime.Time(time.Now().Unix()),
		PubVersion: v.PubVersion,
		ContLen:    v.ContLen,
		ImgCnt:     int64(len(biliRes.ImgUrls)),
	}
}
