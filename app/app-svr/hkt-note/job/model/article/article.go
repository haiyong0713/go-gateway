package article

import (
	"fmt"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
)

const (
	TpArtDetailNoteId = "note_id"
	TpArtDetailCvid   = "cvid"
	ArtTpNote         = 2 //  专栏内的笔记类型
	ArtNoChange       = 1 // 无需更新
	ArtCanView        = 2 // 笔记专栏状态变为可浏览
	ArtCantView       = 3 // 笔记专栏状态变为不可浏览,即该cvid被删除/锁定
	ArtDBNeedRm       = 2 // 专栏同步删除
	ArtDBNeedPub      = 3 // 专栏同步审核状态/原因
	ArtActionInvalid  = 0 // 专栏操作类型错误
	ArtActionAdd      = 1 // 新建专栏
	ArtActionEdit     = 2 // 编辑专栏
	AutoComment       = 1 // 过审时自动发评论

	_deleted        = 1
	ArtTempPicNone  = 1 // 无图模式
	ArtTempPicOne   = 4 // 单图模式
	ArtTempPicThree = 3 // 3图模式

	PubStatusPending = 1 // 审核中(-2 2 5  -8 -9 -12 -13)
	PubStatusPassed  = 2 // 审核通过(0 4 7 6 9 13)
	PubStatusFail    = 3 // 被打回，下次发布用老的cvid(-3 3)
	PubStatusLock    = 4 // 被锁定，下次发布需申请新的cvid(-11 -10)
	PubStatusBreak   = 6 // 发布流程出错，审核失败

	// Article state
	_stateTimingReReject           = -14 //定时待发重复编辑审核不通过
	_stateTimingRePass             = -13 // 定时发布重新编辑审核通过待发布
	_stateTimingRePending          = -12 // 定时待发重复编辑待审
	_stateAutoLock                 = -11 // 自动锁定
	_stateLock                     = -10 // 锁定
	_stateTiming                   = -8  // 定时发布待审
	_stateTimingPass               = -9  // 定时发布审核通过待发布
	_stateReject                   = -3  //打回
	_statePending                  = -2  //提交待审
	_stateOpen                     = 0   //开放浏览
	_stateOpenPending              = 2   //发布后打回修改再提交
	_stateOpenReject               = 3   //发布后打回
	_stateAutoPass                 = 4   //先发后审,不走人工审核无法编辑
	_stateRePending                = 5   // 重复编辑待审
	_stateReReject                 = 6   // 重复编辑未通过
	_stateRePass                   = 7   // 重复编辑通过
	_stateTimingPublished          = 9   //定时发布通过已发布
	_stateTimingRePendingPublished = 12  ////定时待发重复编辑审核审核中已发布
	_stateTimingRePublished        = 13  //定时发布重新编辑通过已发布
	_stateTimingReRejectPublished  = 14  //定时待发重复编辑审核不通过已发布

)

type ArtDetailBlog struct {
	Old *ArtDetailDB `json:"old"`
	New *ArtDetailDB `json:"new"`
}

type ArtDetailDB struct {
	Cvid        int64  `json:"cvid"`
	NoteId      int64  `json:"note_id"`
	Mid         int64  `json:"mid"`
	Oid         int64  `json:"oid"`
	OidType     int    `json:"oid_type"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	PubStatus   int    `json:"pub_status"`
	PubReason   string `json:"pub_reason"`
	PubVersion  int64  `json:"pub_version"`
	AutoComment int    `json:"auto_comment"`
	InjectTime  string `json:"inject_time"`
	Pubtime     string `json:"pubtime"`
	PubFrom     int    `json:"pub_from"`
	CommentInfo string `json:"comment_info"`
	Deleted     int    `json:"deleted"`
	Ctime       string `json:"ctime"`
	Mtime       string `json:"mtime"`
}

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
	Cvid    int64  `json:"cvid"`
	NoteId  int64  `json:"note_id"`
	Content string `json:"content"`
	Tag     string `json:"tag"`
	Mid     int64  `json:"mid"`
	Deleted int    `json:"deleted"`
	// retry use
	Mtime      xtime.Time `json:"mtime"`
	PubVersion int64      `json:"pub_version"`
	ContLen    int64      `json:"cont_len"`
	ImgCnt     int64      `json:"img_cnt"`
}

type PublicCont struct {
	HtmlCont string `json:"html_cont"`
	JsonCont string `json:"json_cont"`
}

// 专栏服务db
type ArtOriginalBlog struct {
	Old *ArtOriginalDB `json:"old"`
	New *ArtOriginalDB `json:"new"`
}

type ArtOriginalDB struct {
	Id          int64  `json:"id"`
	CategoryId  int64  `json:"category_id"`
	State       int    `json:"state"`
	Reason      string `json:"reason"`
	Mid         int64  `json:"mid"`
	DeletedTime int64  `json:"deleted_time"`
	CheckTime   string `json:"check_time"`
	Mtime       string `json:"mtime"`
	Type        int    `json:"type"`
}

func (v *ArtDetailDB) ToDtlCache() *ArtDtlCache {
	pubtime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Pubtime, time.Local)
	mtime, _ := time.ParseInLocation("2006-01-02 15:04:05", v.Mtime, time.Local)
	return &ArtDtlCache{
		Cvid:       v.Cvid,
		NoteId:     v.NoteId,
		Mid:        v.Mid,
		PubStatus:  v.PubStatus,
		PubReason:  v.PubReason,
		PubVersion: v.PubVersion,
		Oid:        v.Oid,
		OidType:    v.OidType,
		Pubtime:    xtime.Time(pubtime.Unix()),
		Title:      v.Title,
		Summary:    v.Summary,
		Deleted:    v.Deleted,
		Mtime:      xtime.Time(mtime.Unix()),
	}
}

func ToPubStatus(artStatus int) int {
	switch artStatus {
	case _stateReject, _stateOpenReject, _stateTimingReReject, _stateTimingReRejectPublished, _stateReReject:
		return PubStatusFail
	case _stateAutoPass, _stateOpen, _stateRePass, _stateTimingPublished, _stateTimingRePublished:
		return PubStatusPassed
	case _stateTimingPass, _stateTimingRePass, _statePending, _stateOpenPending, _stateRePending, _stateTiming, _stateTimingRePending, _stateTimingRePendingPublished:
		return PubStatusPending
	case _stateAutoLock, _stateLock:
		return PubStatusLock
	default:
		log.Warn("artWarn ToPubStatus artStatus(%d) unrecognized", artStatus)
		return 0
	}
}

// 客态笔记状态:
// pass:使用新版本,更新缓存
// fail:使用老版本，不更新缓存
// deleted/lock:解除cvid-note_id绑定关系，更新缓存
func (v *ArtDetailBlog) ToPublicStatus() (publicStat int, commentOperation bool) {
	if v.Old == nil {
		if v.New.Deleted != _deleted && v.New.PubStatus == PubStatusPassed {
			return ArtCanView, true
		}
		return ArtNoChange, false
	}
	if v.Old.Deleted != v.New.Deleted && v.New.Deleted == _deleted {
		return ArtCantView, false
	}
	if v.Old.PubStatus != v.New.PubStatus {
		if v.New.PubStatus == PubStatusPassed {
			return ArtCanView, true
		}
		if v.New.PubStatus == PubStatusLock {
			return ArtCantView, false
		}
	}
	if v.Old.PubStatus == v.New.PubStatus && v.New.PubStatus == PubStatusPassed && v.Old.Pubtime != v.New.Pubtime { // 先发后审之人工过审后
		return ArtCanView, false
	}
	return ArtNoChange, false
}

func (v *ArtDtlCache) ToPublicStatus(artHost *ArtDtlCache) int {
	if artHost.PubStatus == PubStatusLock { // 存在已锁定未删除的版本，客态不可见
		return ArtCantView
	}
	if v.Cvid <= 0 {
		return ArtCantView
	}
	return ArtCanView
}

func (v *ArtOriginalBlog) ArtNeed() (int, string) {
	reason := v.New.Reason
	if v.Old == nil {
		return ArtDBNeedPub, reason
	}
	if v.Old.State == 0 && v.New.State == 5 {
		return ArtNoChange, reason
	}
	if v.Old.DeletedTime == 0 && v.New.DeletedTime > 0 {
		return ArtDBNeedRm, reason
	}
	if v.Old.Reason != v.New.Reason {
		return ArtDBNeedPub, reason
	}
	if v.Old.State != v.New.State {
		if ToPubStatus(v.New.State) == PubStatusPending { // 专栏侧审核中的reason仍为上一次审核理由，不适用，用空字符串覆盖
			reason = ""
		}
		return ArtDBNeedPub, reason
	}
	return ArtNoChange, reason
}

func (v *ArtDtlCache) ToActionType() int {
	if v.Cvid <= 0 {
		return ArtActionAdd
	}
	switch v.PubStatus {
	case PubStatusPending:
		return ArtActionInvalid
	case PubStatusPassed, PubStatusFail:
		return ArtActionEdit
	case PubStatusLock:
		return ArtActionAdd
	default:
		return ArtActionInvalid
	}
}

func ToArtListVal(cvid, noteId int64) string {
	return fmt.Sprintf("%d-%d", cvid, noteId)
}

func (v *ArtDtlCache) ToDetailDB(noteId int64) *ArtDetailDB {
	return &ArtDetailDB{
		Cvid:       v.Cvid,
		NoteId:     noteId, // 重试用，因此需要真实id
		Mid:        v.Mid,
		Oid:        v.Oid,
		OidType:    v.OidType,
		Title:      v.Title,
		Summary:    v.Summary,
		PubStatus:  v.PubStatus,
		PubReason:  v.PubReason,
		PubVersion: v.PubVersion,
		Deleted:    v.Deleted,
	}
}
