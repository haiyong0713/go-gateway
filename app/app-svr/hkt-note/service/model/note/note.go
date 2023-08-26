package note

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/article"
	"go-gateway/pkg/idsafe/bvid"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

const (
	_arcStateOK         = 0
	_cheeseStatusReturn = -2
	_cheeseStatusLock   = -3
	_arcSimpleValid     = 0
	_arcSimpleInvalid   = 1
	_webUrlFrom         = "fullpage"
	OidTypeUGC          = 0
	OidTypeCheese       = 1
)

type DtlCache struct {
	NoteId      int64      `json:"note_id"`
	Oid         int64      `json:"aid"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	NoteSize    int64      `json:"note_size"`
	Deleted     int        `json:"deleted"`
	AuditStatus int        `json:"audit_status"`
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

type NtList struct {
	NoteId int64      `json:"note_id"`
	Oid    int64      `json:"aid"`
	Mtime  xtime.Time `json:"mtime"`
}

func ToPage(pn int64, ps int64) (min int64, max int64) {
	min = (pn - 1) * ps
	max = pn*ps - 1
	return
}

func ToVideoIds(dtl map[int64]*DtlCache) (aids []int64, sids []int32) {
	for _, d := range dtl {
		switch d.OidType {
		case OidTypeUGC:
			aids = append(aids, d.Oid)
		case OidTypeCheese:
			sids = append(sids, int32(d.Oid))
		default:
			log.Warn("noteInfo ToVideoIds noteDetail(%+v) oidType invalid,skip", d)
		}
	}
	return
}

func ToNtKeys(mid int64, listKeys []string) (ntList []*NtList, noteIds []int64) {
	for _, key := range listKeys {
		tmp := toNtList(key)
		if tmp == nil {
			log.Warn("noteWarn NoteList mid(%d) ToNtList key(%s) invalid,skip", mid, key)
			continue
		}
		ntList = append(ntList, tmp)
		noteIds = append(noteIds, tmp.NoteId)
	}
	return
}

func toNtList(key string) (list *NtList) {
	arr := strings.Split(key, "-")
	if len(arr) != 2 { // nolint:gomnd
		return nil
	}
	noteId, _ := strconv.ParseInt(arr[0], 10, 64)
	if noteId == 0 {
		return nil
	}
	oid, _ := strconv.ParseInt(arr[1], 10, 64)
	if oid == 0 {
		return nil
	}
	return &NtList{NoteId: noteId, Oid: oid}
}

func toNoteSimple(nt *NtList, dtlCache *DtlCache, arc *arcapi.Arc, ssn *cssngrpc.SeasonCard, url string, art *article.ArtDtlCache, forbidEntrance bool) *api.NoteSimple {
	if dtlCache == nil {
		log.Warn("NoteWarn Notelist noteID(%d) no details", nt.NoteId)
		return nil
	}
	if art == nil {
		art = &article.ArtDtlCache{}
	}
	res := &api.NoteSimple{
		NoteId:             nt.NoteId,
		Title:              dtlCache.Title,
		Summary:            dtlCache.Summary,
		AuditStatus:        int64(dtlCache.AuditStatus),
		Mtime:              dtlCache.Mtime.Time().Format("2006-01-02 15:04"),
		WebUrl:             fmt.Sprintf(url, nt.Oid, dtlCache.OidType, _webUrlFrom),
		NoteIdStr:          strconv.FormatInt(nt.NoteId, 10),
		PubStatus:          int64(art.PubStatus),
		Message:            dtlCache.toMessage(art),
		ForbidNoteEntrance: forbidEntrance && dtlCache.OidType == OidTypeUGC,
	}
	switch dtlCache.OidType {
	case OidTypeUGC:
		res.Arc = toUGCArcSimple(nt.Oid, arc)
	case OidTypeCheese:
		res.Arc = toCheeseArcSimple(nt.Oid, ssn)
	default:
		res.Arc = &api.ArcSimple{Oid: nt.Oid}
		log.Warn("NoteWarn Notelist nt(%+v) oidType invalid", nt)
	}
	return res
}

func toUGCArcSimple(oid int64, arc *arcapi.Arc) *api.ArcSimple {
	res := &api.ArcSimple{Oid: oid, OidType: OidTypeUGC, Aid: oid}
	if arc == nil || arc.State < _arcStateOK {
		res.Status = _arcSimpleInvalid
		return res
	}
	res.Bvid, _ = bvid.AvToBv(arc.Aid)
	res.Pic = arc.Pic
	res.Desc = arc.Desc
	res.Status = _arcSimpleValid
	return res
}

func toCheeseArcSimple(oid int64, ssn *cssngrpc.SeasonCard) *api.ArcSimple {
	res := &api.ArcSimple{Oid: oid, OidType: OidTypeCheese, Aid: oid}
	if ssn == nil || ssn.Status == _cheeseStatusReturn || ssn.Status == _cheeseStatusLock {
		res.Status = _arcSimpleInvalid
		return res
	}
	res.Pic = ssn.Cover
	res.Desc = ssn.UpdateInfo1
	res.Status = _arcSimpleValid
	return res
}

func (v *DtlCache) ToSimpleCard(art *article.ArtDtlCache) *api.SimpleNoteCard {
	if v == nil || v.Deleted == 1 || v.NoteId == -1 {
		return nil
	}
	if art == nil || art.Cvid == -1 || art.NoteId == -1 || art.Deleted == 1 {
		log.Warn("noteWarn ToSimpleCard v(%+v) art(%+v) nil", v, art)
		art = &article.ArtDtlCache{}
	}
	return &api.SimpleNoteCard{
		NoteId:    v.NoteId,
		Oid:       v.Oid,
		Mid:       v.Mid,
		OidType:   int64(v.OidType),
		PubStatus: int64(art.PubStatus),
		PubReason: art.PubReason,
	}
}

func toArtSimple(dtlCache *article.ArtDtlCache, arc *arcapi.Arc, ssn *cssngrpc.SeasonCard, webUrl string) *api.NoteSimple {
	tm := dtlCache.Pubtime
	if tm <= 0 { // 先发后审未过审时，用mtime代替发布时间
		tm = dtlCache.Mtime
	}
	res := &api.NoteSimple{
		Cvid:    dtlCache.Cvid,
		Title:   dtlCache.Title,
		Summary: dtlCache.Summary,
		Mtime:   tm.Time().Format("2006-01-02 15:04"),
		WebUrl:  fmt.Sprintf(webUrl, dtlCache.Cvid),
		NoteId:  dtlCache.NoteId,
		Message: fmt.Sprintf("更新于 %s", tm.Time().Format("2006-01-02 15:04")),
	}
	switch dtlCache.OidType {
	case OidTypeUGC:
		res.Arc = toUGCArcSimple(dtlCache.Oid, arc)
	case OidTypeCheese:
		res.Arc = toCheeseArcSimple(dtlCache.Oid, ssn)
	default:
		res.Arc = &api.ArcSimple{Oid: dtlCache.Oid}
		log.Warn("ArtWarn ToArtSimple nt(%+v) oidType invalid", dtlCache)
	}
	return res
}

func (v *DtlCache) toMessage(art *article.ArtDtlCache) string {
	if art == nil || art.Cvid == -1 || art.PubStatus == 0 || art.Deleted == 1 {
		return fmt.Sprintf("更新于 %s", v.Mtime.Time().Format("2006-01-02 15:04"))
	}
	if art.Mtime > v.Mtime {
		switch art.PubStatus {
		case article.PubStatusPending:
			return "审核中"
		case article.PubStatusPassed:
			return "审核通过"
		case article.PubStatusFail:
			return "已打回"
		case article.PubStatusLock:
			return "已锁定"
		case article.PubStatusWaiting:
			return "待审核"
		case article.PubStatusBreak:
			return "发布失败"
		default:
		}
	}
	return fmt.Sprintf("更新于 %s", v.Mtime.Time().Format("2006-01-02 15:04"))
}

func DealNoteListItem(page *api.Page, ntList []*NtList, details map[int64]*DtlCache, arcs map[int64]*arcapi.Arc, ssns map[int32]*cssngrpc.SeasonCard, arts map[int64]*article.ArtDtlCache, webUrl string, arcsForbid map[int64]bool) *api.NoteListReply {
	list := make([]*api.NoteSimple, 0)
	if arcs == nil {
		arcs = make(map[int64]*arcapi.Arc)
	}
	if ssns == nil {
		ssns = make(map[int32]*cssngrpc.SeasonCard)
	}
	if arts == nil {
		arts = make(map[int64]*article.ArtDtlCache)
	}
	for _, nt := range ntList {
		if val := toNoteSimple(nt, details[nt.NoteId], arcs[nt.Oid], ssns[int32(nt.Oid)], webUrl, arts[nt.NoteId], arcsForbid[nt.Oid]); val != nil {
			list = append(list, val)
		}
	}
	return &api.NoteListReply{
		List: list,
		Page: page,
	}
}

func DealArtListItem(page *api.Page, cvids []int64, details map[int64]*article.ArtDtlCache, arcs map[int64]*arcapi.Arc, ssns map[int32]*cssngrpc.SeasonCard, webUrl string) *api.NoteListReply {
	list := make([]*api.NoteSimple, 0)
	if arcs == nil {
		arcs = make(map[int64]*arcapi.Arc)
	}
	if ssns == nil {
		ssns = make(map[int32]*cssngrpc.SeasonCard)
	}
	for _, cvid := range cvids {
		dtl, ok := details[cvid]
		if !ok {
			log.Warn("ArtWarn PublishListInUser cvid(%d) no details", cvid)
			continue
		}
		if dtl.PubStatus != article.PubStatusPassed {
			log.Warn("ArtInfo PublishListInUser cvid(%d) details(%+v) isn't pass,skip", cvid, dtl)
			continue
		}
		if val := toArtSimple(dtl, arcs[dtl.Oid], ssns[int32(dtl.Oid)], webUrl); val != nil {
			list = append(list, val)
		}
	}
	return &api.NoteListReply{
		List: list,
		Page: page,
	}
}
