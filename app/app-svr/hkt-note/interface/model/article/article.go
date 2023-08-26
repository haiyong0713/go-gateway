package article

import (
	"fmt"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accountRelationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-gateway/app/app-svr/hkt-note/common"
)

const (
	TpArtDetailNoteId = "note_id" // 笔记专栏缓存，note_id维度
	TpArtDetailCvid   = "cvid"    // 笔记专栏类型，cvid维度
	PubStatusPending  = 1         // 审核中
	PubStatusWaiting  = 5         // 待审核
)

type PubListInArcReq struct {
	Oid     int64 `form:"oid" validate:"required"`
	OidType int   `form:"oid_type"`
	Pn      int64 `form:"pn" validate:"required"`
	Ps      int64 `form:"ps" validate:"required"`
	UperMid int64 `form:"uper_mid"`
}

type PubNoteInfoReq struct {
	Cvid int64 `form:"cvid" validate:"required"`
	note.Device
}

type PubNoteInfoRes struct {
	Cvid               int64           `json:"cvid"`
	NoteId             int64           `json:"note_id"`
	Title              string          `json:"title"`
	Summary            string          `json:"summary"`
	Content            string          `json:"content"`
	CidCount           int             `json:"cid_count"`
	PubStatus          int             `json:"pub_status"`
	Tags               []*note.NoteTag `json:"tags"`
	Arc                *note.ArcCore   `json:"arc"`
	Author             *Author         `json:"author"`
	ForbidNoteEntrance bool            `json:"forbid_note_entrance"`
	LastTimeText       string          `json:"last_mtime_text"`
}

type Author struct {
	Mid            int64                `json:"mid"`
	Name           string               `json:"name"`
	Face           string               `json:"face"`
	Level          int32                `json:"level"`
	IsSeniorMember int32                `json:"is_senior_member"`
	VipInfo        accgrpc.VipInfo      `json:"vip_info"`
	Pendant        accgrpc.PendantInfo  `json:"pendant"`
	Official       accgrpc.OfficialInfo `json:"official"`
	//粉丝数
	Follower int64 `json:"follower"`
}

type PubListInArcRes struct {
	List           []*PubItemInArc `json:"list"`
	Page           *notegrpc.Page  `json:"page"`
	ShowPublicNote bool            `json:"show_public_note"`
	Message        string          `json:"message"`
}

type PubItemInArc struct {
	Cvid    int64   `json:"cvid"`
	Title   string  `json:"title"`
	Summary string  `json:"summary"`
	Pubtime string  `json:"pubtime"`
	WebUrl  string  `json:"web_url"`
	Message string  `json:"message"`
	Author  *Author `json:"author"`
	Likes   int64   `json:"likes"`
	HasLike bool    `json:"has_like"`
}

type ArtDtlCache struct {
	Cvid       int64      `json:"cvid"`
	NoteId     int64      `json:"note_id"`
	Mid        int64      `json:"mid"`
	PubStatus  int        `json:"pub_status"`
	PubReason  string     `json:"pub_reason"`
	PubVersion int        `json:"pub_version"`
	Oid        int64      `json:"oid"`
	OidType    int        `json:"oid_type"`
	Pubtime    xtime.Time `json:"pubtime"`
	Mtime      xtime.Time `json:"mtime"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	Deleted    int        `json:"deleted"`
}

func FromAcc(val *accgrpc.Card) *Author {
	return &Author{
		Mid:            val.Mid,
		Name:           val.Name,
		Face:           val.Face,
		Level:          val.Level,
		IsSeniorMember: val.IsSeniorMember,
		VipInfo:        val.Vip,
		Pendant:        val.Pendant,
		Official:       val.Official,
	}
}

func ToPubListInArcRes(pubList *notegrpc.NoteListReply, acc map[int64]*accgrpc.Card) *PubListInArcRes {
	if acc == nil {
		acc = make(map[int64]*accgrpc.Card)
	}
	list := make([]*PubItemInArc, 0, len(pubList.List))
	for _, l := range pubList.List {
		if _, ok := acc[l.Mid]; !ok {
			log.Warn("artWarn ToPubListInArcRes l(%+v) account not found,skip", l)
			continue
		}
		tmp := &PubItemInArc{
			Cvid:    l.Cvid,
			Title:   l.Title,
			Summary: l.Summary,
			Pubtime: l.Pubtime,
			WebUrl:  l.WebUrl,
			Message: l.Message,
			Author:  FromAcc(acc[l.Mid]),
			Likes:   l.Likes,
			HasLike: l.HasLike,
		}
		list = append(list, tmp)
	}
	return &PubListInArcRes{
		List:           list,
		Page:           pubList.Page,
		ShowPublicNote: true,
	}
}

func ToPubNoteInfoRes(info *notegrpc.PublishNoteInfoReply, author *Author, statReply *accountRelationGrpc.StatReply, arc *note.ArcCore, isForbid bool) *PubNoteInfoRes {
	tags := make([]*note.NoteTag, 0)
	for _, t := range info.Tags {
		tmp := &note.NoteTag{
			Cid:     t.Cid,
			Status:  t.Status,
			Index:   t.Index,
			Seconds: t.Seconds,
		}
		tags = append(tags, tmp)
	}
	arc.CvidCount = info.ArcCvidCnt
	author.Follower = statReply.Follower
	var lastMTimeText string
	lastMTime := xtime.Time(info.PubTime)
	if info.HasPubSuccessBefore {
		lastMTimeText = fmt.Sprintf("%s%s", common.Publish_Info_Last_Time_Text_Prefix, lastMTime.Time().Format("2006-01-02 15:04"))
	} else {
		lastMTimeText = lastMTime.Time().Format("2006-01-02 15:04")
	}
	return &PubNoteInfoRes{
		Cvid:               info.Cvid,
		NoteId:             info.NoteId,
		Title:              info.Title,
		Summary:            info.Summary,
		Content:            info.Content,
		CidCount:           int(info.CidCount),
		PubStatus:          int(info.PubStatus),
		Tags:               tags,
		Arc:                arc,
		Author:             author,
		ForbidNoteEntrance: isForbid,
		LastTimeText:       lastMTimeText,
	}
}
