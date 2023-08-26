package like

import (
	"encoding/json"

	"go-common/library/time"
	"go-gateway/app/app-svr/archive/service/api"
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

// ParamTeams add follow param teams
type ParamTeams struct {
	Teams []string `form:"teams,split" validate:"gt=0,dive,gt=0"`
}

// ParamSid  sid param
type ParamSid struct {
	Sid int64 `form:"sid" validate:"required,min=1"`
}

// ParamAddGuess add guess param
type ParamAddGuess struct {
	ObjID  int64 `form:"obj_id" validate:"required,min=1"`
	Result int64 `form:"result" validate:"required,min=1"`
	Stake  int64 `form:"stake"  validate:"gt=0"`
}

// ParamObject unstart  object param
type ParamObject struct {
	Sid int64 `form:"sid" validate:"required,min=1"`
	Pn  int   `form:"pn" validate:"gt=0"`
	Ps  int   `form:"ps" validate:"gt=0,lte=50"`
}

// ParamAddLikeAct add likeAct param
type ParamAddLikeAct struct {
	Sid      int64  `form:"sid" validate:"required,min=1"`
	Lid      int64  `form:"lid" validate:"required,min=1"`
	Score    int64  `form:"score" validate:"min=1,max=5"`
	Buvid    string `form:"buvid"`
	Origin   string `form:"origin"`
	UA       string `form:"ua"`
	Referer  string `form:"referer"`
	IP       string `form:"ip"`
	Build    string `form:"build"`
	Platform string `form:"platform"`
	Device   string `form:"device"`
	MobiApp  string `form:"mobi_app"`
	API      string `form:"-"`
}

// 支持sid+cvid进行点赞
type ParamAddLikeActWithSidCVId struct {
	Sid   int64  `form:"sid" validate:"required,min=1"`
	Score int64  `form:"score" validate:"min=1,max=5"`
	CVId  string `form:"cvid" validate:"required"`
}

// ParamMissionLikeAct add missionAct param
type ParamMissionLikeAct struct {
	Sid int64 `form:"sid" validate:"min=1"`
	Lid int64 `form:"lid" validate:"min=1"`
}

// ParamMissionFriends get mission friends list
type ParamMissionFriends struct {
	Sid  int64 `form:"sid"  validate:"min=1"`
	Lid  int64 `form:"lid"  validate:"min=1"`
	Size int   `form:"size" validate:"min=1,max=50"`
}

// ParamStoryKingAct .
type ParamStoryKingAct struct {
	Sid      int64  `form:"sid" validate:"required,min=1"`
	Lid      int64  `form:"lid" validate:"required,min=1"`
	Score    int64  `form:"score" validate:"min=1,max=10"`
	Token    string `form:"token"`
	Buvid    string `form:"buvid"`
	Origin   string `form:"origin"`
	UA       string `form:"ua"`
	Referer  string `form:"referer"`
	IP       string `form:"ip"`
	Build    string `form:"build"`
	Platform string `form:"platform"`
	Device   string `form:"device"`
	MobiApp  string `form:"mobi_app"`
	UserName string `form:"-"`
}

// ParamList .
type ParamList struct {
	Sid     int64  `form:"sid" validate:"min=1"`
	Type    string `form:"type" default:"like"`
	Pn      int    `form:"pn" default:"1" validate:"min=1"`
	Ps      int    `form:"ps" default:"30" validate:"min=1"`
	Zone    int64  `form:"zone" default:"0"`
	Version int    `form:"version"`
}

// ParamText .
type ParamText struct {
	Sid        int64  `form:"sid" validate:"min=1"`
	Wid        int64  `form:"wid"`
	Type       int64  `form:"type"`
	Message    string `form:"message"`
	Plat       int64  `form:"plat"`
	RefererURI string `form:"referer_uri"`
}

// ParamOther .
type ParamOther struct {
	Sid        int64  `form:"sid" validate:"min=1"`
	Wid        int64  `form:"wid"`
	Type       int64  `form:"type"`
	Message    string `form:"message" validate:"min=1,max=600"`
	Device     int64  `form:"device"`
	Plat       int64  `form:"plat"`
	Image      []byte `form:"-"`
	FileType   string `form:"-"`
	RefererURI string `form:"referer_uri"`
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

// UpSpecial .
type UpSpecial struct {
	GroupIDs []int64 `json:"group_ids"`
}

// SortModule .
type SortModule struct {
	IDs     []int64 `json:"ids"`
	HasMore int32   `json:"has_more"`
	Offset  int64   `json:"offset"`
}

type ArcListData struct {
	List []*ArcData `json:"list"`
}

type ArcData struct {
	ID   string    `json:"id"`
	Data *AidsData `json:"data"`
}

type AidsData struct {
	Aids string `json:"aids"`
}

type TaafWebData struct {
	List []*struct {
		ID   string    `json:"id"`
		Name string    `json:"name"`
		Data *TaafData `json:"data"`
	} `json:"list"`
}

type TaafData struct {
	Stime        string `json:"stime"`
	Animate      string `json:"animate"`
	ArtDirector  string `json:"art_director"`
	Director     string `json:"director"`
	Role         string `json:"role"`
	Title        string `json:"title"`
	OriginalName string `json:"original_name"`
	Lidnew       string `json:"lidnew"`
	Type         string `json:"type"`
	Bianju       string `json:"bianju"`
}

type EntData struct {
	Lid   int64  `json:"lid"`
	Mid   int64  `json:"mid"`
	TagID int64  `json:"tag_id"`
	Aid   string `json:"aid"`
}

type EntDataV2 struct {
	Lid int64       `json:"lid"`
	Mid interface{} `json:"mid"`
	Aid string      `json:"aid"`
}

type EntRes struct {
	Lid     int64        `json:"lid"`
	Content *LikeContent `json:"content"`
	Mid     int64        `json:"mid"`
	TagID   int64        `json:"tag_id"`
	Arcs    []*SimpleArc `json:"arcs"`
}

type EntResV2 struct {
	Lid     int64        `json:"lid"`
	Content *LikeContent `json:"content"`
	Mid     interface{}  `json:"mid"`
	Arcs    []*SimpleArc `json:"arcs"`
}

type BdfData struct {
	Aids string `json:"aids"`
}

type SimpleArc struct {
	Aid       int64      `json:"aid"`
	Videos    int64      `json:"videos"`
	TypeID    int32      `json:"tid"`
	TypeName  string     `json:"tname"`
	Copyright int32      `json:"copyright"`
	Pic       string     `json:"pic"`
	Title     string     `json:"title"`
	PubDate   time.Time  `json:"pub_date"`
	Ctime     time.Time  `json:"ctime"`
	State     int32      `json:"state"`
	Duration  int64      `json:"duration"`
	MissionID int64      `json:"mission_id,omitempty"`
	Author    api.Author `json:"owner"`
	Stat      SimpleStat `json:"stat"`
	FirstCid  int64      `json:"cid,omitempty"`
}

type SimpleStat struct {
	Aid        int64  `json:"aid"`
	View       int32  `json:"view"`
	Danmaku    int32  `json:"danmaku"`
	Reply      int32  `json:"reply"`
	Fav        int32  `json:"favorite"`
	Coin       int32  `json:"coin"`
	Share      int32  `json:"share"`
	NowRank    int32  `json:"now_rank"`
	HisRank    int32  `json:"his_rank"`
	Like       int32  `json:"like"`
	Evaluation string `json:"evaluation"`
	Mark       int64  `json:"mark"`
}

func CopyFromArc(in *api.Arc) (out *SimpleArc) {
	if in == nil {
		return
	}
	out = &SimpleArc{
		Aid:       in.Aid,
		Videos:    in.Videos,
		TypeID:    in.TypeID,
		TypeName:  in.TypeName,
		Copyright: in.Copyright,
		Pic:       in.Pic,
		Title:     in.Title,
		PubDate:   in.PubDate,
		Ctime:     in.Ctime,
		State:     in.State,
		Duration:  in.Duration,
		MissionID: in.MissionID,
		Author:    in.Author,
		Stat: SimpleStat{
			Aid:     in.Stat.Aid,
			View:    in.Stat.View,
			Danmaku: in.Stat.Danmaku,
			Reply:   in.Stat.Reply,
			Fav:     in.Stat.Fav,
			Coin:    in.Stat.Coin,
			Share:   in.Stat.Share,
			NowRank: in.Stat.NowRank,
			HisRank: in.Stat.HisRank,
			Like:    in.Stat.Like,
		},
		FirstCid: in.FirstCid,
	}
	return
}

type SpecialArcList struct {
	Name    string `json:"name"`
	Subject []*Sub `json:"subject"`
}

type Sub struct {
	Name string `json:"name"`
	ID   int64  `json:"id"`
	Aids string `json:"aids"`
}

type SpecialArcListReply struct {
	List []*api.Arc `json:"list"`
	Page *Page      `json:"page"`
}
