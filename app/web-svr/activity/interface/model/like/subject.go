package like

import (
	xtime "go-common/library/time"
)

// Subject type.
const (
	VIDEO             = 1
	PICTURE           = 2
	DRAWYOO           = 3
	VIDEOLIKE         = 4
	PICTURELIKE       = 5
	DRAWYOOLIKE       = 6
	TEXT              = 7
	TEXTLIKE          = 8
	ONLINEVOTE        = 9
	QUESTION          = 10
	LOTTERY           = 11
	ARTICLE           = 12
	VIDEO2            = 13
	MUSIC             = 15
	PHONEVIDEO        = 16
	SMALLVIDEO        = 17
	RESERVATION       = 18
	MISSIONGROUP      = 19
	STORYKING         = 20
	PREDICTION        = 21
	CLOCKIN           = 22
	USERACTIONSTAT    = 23
	UPRESERVATIONARC  = 24
	UPRESERVATIONLIVE = 25
	FLAGFIRST         = 1
	FLAGSPY           = 2
	FLAGUSTIME        = 4
	FLAGUETIME        = 8
	FLAGLEVEL         = 16
	FLAGIP            = 32
	FLAGRANKCLOSE     = 64
	FLAGPHONEBIND     = 128
	FLAGFANLIMIT      = 256
	FLAGVIPLIMIT      = 512
	FLAGYEARVIPLIMIT  = 1024
	FLAGUPUSTIME      = 2048
	FLAGUPUETIME      = 4096
	FLAGUPLEVEL       = 8192
	FLAGUPPHONEBIND   = 16384
	FLAGUPSPY         = 32768
	FLAGMONTHSCORE    = 65536
	FLAGYEARSCORE     = 131072
	FLAGDAILYLIKETYPE = 262144
	// 数据源开启分表
	FLAGATTRISNEW       = uint(19)
	FLATFORBIDHOTLIST   = uint(20)
	FLATFORBIDCANCEL    = uint(22)
	FLAGQUESTIONNAIRE   = uint(27)
	FLAGDUPLICATESUBMIT = uint(28)
	LIKETYPECOMMON      = "common"
	LIKETYPEUP          = "up_type"
	LIKETYPENO          = "no_like"
	// award
	AwardOnline        = 1
	AwardNotAllow      = 0
	AwardAllowed       = 1
	AwardReward        = 2
	AwardSidTypeSingle = 0
	AwardSidTypeMulti  = 1
)

const (
	// ShieldNoRank 禁止排行
	ShieldNoRank = uint(0)
	// ShieldNoDynamic 禁止首页动态
	ShieldNoDynamic = uint(1)
	// ShieldNoRecommend 禁止推荐
	ShieldNoRecommend = uint(2)
	// ShieldNoHot 禁止热门
	ShieldNoHot = uint(3)
	// ShieldNoFansDynamic 禁止粉丝动态
	ShieldNoFansDynamic = uint(4)
	// ShieldNoSearch 禁止搜索
	ShieldNoSearch = uint(5)
	// ShieldNoOversea 禁止海外
	ShieldNoOversea = uint(6)
	// ArchiveNoRank 稿件禁止排行
	ArchiveNoRank = "norank"
	// ArchiveNoDynamic 禁止首页动态
	ArchiveNoDynamic = "noindex"
	// ArchiveNoRecommend 禁止推荐
	ArchiveNoRecommend = "norecommend"
	// ArchiveNoHot 禁止热门
	ArchiveNoHot = "nohot"
	// ArchiveNoFansDynamic 禁止粉丝动态
	ArchiveNoFansDynamic = "push_blog"
	// ArchiveNoSearch 禁止搜索
	ArchiveNoSearch = "nosearch"
	// ArchiveNoOversea 禁止海外
	ArchiveNoOversea = "oversea_block"
	// FlowControlYes 禁止
	FlowControlYes = 1
	// FlowControlNo 不禁止
	FlowControlNo = 0
)

// Subject group type
var (
	VIDEOALL = []int64{VIDEO, VIDEOLIKE, VIDEO2, PHONEVIDEO, SMALLVIDEO}
	VIDEOUP  = []int64{VIDEO, VIDEOLIKE, VIDEO2, PHONEVIDEO, SMALLVIDEO, CLOCKIN}
	LIKETYPE = []int64{VIDEOLIKE, PICTURELIKE, DRAWYOOLIKE, TEXTLIKE}
	VIDEOS   = []int64{VIDEO, VIDEOLIKE, VIDEO2}
	PICS     = []int64{PICTURELIKE, PICTURE}
	DRAWYOOS = []int64{DRAWYOO, DRAWYOOLIKE}
	ARTICLES = []int64{ARTICLE}
	MUSICS   = []int64{MUSIC}
)

const (
	// RuleAidSourceTypeFile 文件
	RuleAidSourceTypeFile = 1
	// RuleAidSourceTypeSid 稿件数据源
	RuleAidSourceTypeSid = 2
	// RuleAidSourceTypeFav 收藏夹
	RuleAidSourceTypeFav = 3
)

const (
	DynamicArc  = 10
	DynamicLive = 11
	DanmakuArc  = 20
	DanmakuLive = 21
)

const (
	SensitiveLevelNormal      = 0
	SensitiveLevelTest        = 10
	SensitiveLevelPass        = 15
	SensitiveLevelAudit       = 16
	SensitiveLevelIntercept20 = 20
	SensitiveLevelIntercept30 = 30
	SensitiveLevelIntercept40 = 40
)

// Subject struct
type Subject struct {
	ID       int64      `json:"id"`
	Name     string     `json:"name"`
	Dic      string     `json:"dic"`
	Cover    string     `json:"cover"`
	Stime    xtime.Time `json:"stime"`
	Interval int32      `json:"interval"`
	Tlimit   int32      `json:"tlimit"`
	Ltime    int32      `json:"ltime"`
	List     []*Like    `json:"list"`
}

// SubItem .
type SubItem struct {
	ID    int64      `json:"id"`
	Ctime xtime.Time `json:"ctime"`
}

// SubjectStat .
type SubjectStat struct {
	Sid   int64 `json:"sid" form:"sid" validate:"min=1"`
	Count int64 `json:"count" form:"count"`
	View  int64 `form:"view" form:"view"`
	Like  int64 `form:"like" form:"like"`
	Fav   int64 `form:"fav" form:"fav"`
	Coin  int64 `form:"coin" form:"coin"`
}

// SubjectScore .
type SubjectScore struct {
	Score int64 `json:"score"`
}

// Page .
type Page struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// SubProtocol .
type SubProtocol struct {
	*SubjectItem
	*ActSubjectProtocol
	Rules []*SubjectRule
}

// SubAndProtocol .
type SubAndProtocol struct {
	*SubjectItem
	Protocol *ActSubjectProtocol `json:"protocol"`
}

type ProtocolReply struct {
	List map[int64]*PubicProto `json:"list"`
}

type PubicProto struct {
	Tags     string `json:"tags"`
	Sid      int64  `json:"sid"`
	Types    string `json:"types"`
	BgmID    int64  `json:"bgm_id"`
	PasterID int64  `json:"paster_id"`
	InstepID int64  `json:"instep_id"`
	Oids     string `json:"oids"`
	Award    string `json:"award"`
	AwardURL string `json:"award_url"`
}

type ActVideoSourceRelationReserve struct {
	Sid   int64
	Stime int64
	Etime int64
	Types string
	Tags  string
}

type CreateDynamicExtension struct {
	FlagCfg struct {
		Reserve struct {
			ReserveID     int64 `json:"reserve_id"`
			ReserveSource int64 `json:"reserve_source"`
		} `json:"reserve"`
	} `json:"flag_cfg"`
}

const (
	CreateDynamicFrom = "create.reserve.svr"
)
