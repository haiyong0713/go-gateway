package article

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
)

const (
	PubStatusPending  = 1 // state = -2 2 5  -8 -9 -12 -13
	PubStatusPassed   = 2 // state = 0 4 7 6 9 13
	PubStatusFail     = 3 // state = -3 3
	PubStatusLock     = 4 // state = -10 -11
	PubStatusWaiting  = 5 // 待审核
	PubStatusBreak    = 6 // 发布失败
	OidTypeUGC        = 0
	OidTypeCheese     = 1
	TpArtDetailNoteId = "note_id" // 笔记专栏缓存，note_id维度
	TpArtDetailCvid   = "cvid"    // 笔记专栏类型，cvid维度
)

type ArtDtlCache struct {
	Cvid       int64      `json:"cvid"`
	NoteId     int64      `json:"note_id"`
	Mid        int64      `json:"mid"`
	PubStatus  int        `json:"pub_status"`
	PubReason  string     `json:"pub_reason"`
	PubVersion int64      `json:"pub_version"`
	Oid        int64      `json:"oid"`
	OidType    int        `json:"oid_type"`
	Pubtime    xtime.Time `json:"pubtime"`
	Mtime      xtime.Time `json:"mtime"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	Deleted    int        `json:"deleted"`
}

type ArtContCache struct {
	Cvid       int64  `json:"cvid"`
	NoteId     int64  `json:"note_id"`
	Content    string `json:"content"`
	Tag        string `json:"tag"`
	Deleted    int    `json:"deleted"`
	Mid        int64  `json:"mid"`
	PubVersion int64  `json:"pub_version"`
	ContLen    int64  `json:"cont_len"`
	ImgCnt     int64  `json:"img_cnt"`
}

type ArtList struct {
	Cvid       int64      `json:"cvid"`
	NoteId     int64      `json:"note_id"`
	Pubtime    xtime.Time `json:"pubtime"`
	Mtime      xtime.Time `json:"mtime"`
	PubVersion int64      `json:"pub_version"`
	PubStatus  int        `json:"pub_status"`
}

func ToVideoIds(dtl map[int64]*ArtDtlCache) (aids []int64, sids []int32) {
	for _, d := range dtl {
		switch d.OidType {
		case OidTypeUGC:
			aids = append(aids, d.Oid)
		case OidTypeCheese:
			sids = append(sids, int32(d.Oid))
		default:
			log.Warn("artInfo ToVideoIds artDetail(%+v) oidType invalid,skip", d)
		}
	}
	return
}

func ToArtListVal(cvid, noteId int64) string {
	return fmt.Sprintf("%d-%d", cvid, noteId)
}

func ToArtKeys(listKeys []string) (cvids []int64, noteIds []int64) {
	for _, key := range listKeys {
		ids := strings.Split(key, "-")
		if len(ids) != 2 { // nolint:gomnd
			log.Warn("noteInfo ToArtKeys key(%s) invalid", key)
			continue
		}
		cvid, err := strconv.ParseInt(ids[0], 10, 64)
		if err != nil || cvid == 0 {
			log.Warn("noteInfo ToArtKeys key(%s) invalid", key)
			continue
		}
		noteId, err := strconv.ParseInt(ids[1], 10, 64)
		if err != nil || noteId == 0 {
			log.Warn("noteInfo ToArtKeys key(%s) invalid", key)
			continue
		}
		cvids = append(cvids, cvid)
		noteIds = append(noteIds, noteId)
	}
	return
}

func (v *ArtDtlCache) ToCard(webUrl string, likes int64, hasLike bool) *notegrpc.NoteSimple {
	tm := v.Pubtime
	if tm <= 0 { // 先发后审未过审时，用mtime代替发布时间
		tm = v.Mtime
	}
	return &notegrpc.NoteSimple{
		Cvid:    v.Cvid,
		Title:   v.Title,
		Summary: v.Summary,
		Pubtime: tm.Time().Format("2006-01-02 15:04"),
		Mid:     v.Mid,
		WebUrl:  webUrl,
		NoteId:  v.NoteId,
		Message: fmt.Sprintf("更新于 %s", tm.Time().Format("2006-01-02 15:04")),
		Likes:   likes,
		HasLike: hasLike,
	}
}

func (v *ArtDtlCache) ToSimpleCard() *notegrpc.SimpleArticleCard {
	return &notegrpc.SimpleArticleCard{
		Cvid:    v.Cvid,
		Oid:     v.Oid,
		OidType: int64(v.OidType),
		Mid:     v.Mid,
		NoteId:  v.NoteId,
	}
}

func (v *ArtDtlCache) ToId(tp string) int64 {
	switch tp {
	case TpArtDetailNoteId:
		return v.NoteId
	case TpArtDetailCvid:
		return v.Cvid
	default:
		return 0
	}
}

func (v *ArtDtlCache) ToPubInfo(auditReason string) (pubStatus, pubVersion int64, pubReason string) {
	if v.Deleted == 1 {
		return 0, 0, ""
	}
	pubStatus = int64(v.PubStatus)
	pubVersion = v.PubVersion
	pubReason = v.PubReason
	if auditReason != "" { // auditReason为每个版本的审核理由，v.PubReason为该cvid最后一次过审的审核理由
		pubReason = auditReason
	}
	return
}
