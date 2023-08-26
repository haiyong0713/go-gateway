package aggregation_v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	showmdl "go-gateway/app/app-svr/app-feed/admin/model/show"
)

const (
	GotoArtificial  = "artificial"
	GotoWeibo       = "weibo"
	GotoDouyin      = "douyin"
	GotoZhihu       = "zhihu"
	GotoAcFun       = "acfun"
	GotoBilibili    = "bilibili"
	GotoBiliPopular = "bili_popular"
	GotoKuaishou    = "kuaishou"
	GotoXigua       = "xigua"

	StateWait   = 0
	StatePass   = 1
	StateForbid = 2
	StateDown   = 3
	StateDel    = 4
	StateOnline = 6

	ActiveStateNew     = 0
	ActiveStateOnline  = 1
	ActiveStateOffLine = 2

	MaterialStateOK    = 1
	MaterialStateFobid = 2

	TagStateOK  = 0
	TagStateDel = 1
)

var ListSort = []*Plat{

	{FormPlatName(GotoArtificial), GotoArtificial},
	{FormPlatName(GotoWeibo), GotoWeibo},
	{FormPlatName(GotoDouyin), GotoDouyin},
	{FormPlatName(GotoZhihu), GotoZhihu},
	{FormPlatName(GotoAcFun), GotoAcFun},
	{FormPlatName(GotoBilibili), GotoBilibili},
	{FormPlatName(GotoBiliPopular), GotoBiliPopular},
	{FormPlatName(GotoKuaishou), GotoKuaishou},
	{FormPlatName(GotoXigua), GotoXigua},
}

type List struct {
	Sort  []*Plat        `json:"sort"`
	Items []*Aggregation `json:"items"`
}

type Plat struct {
	Tille string `json:"title"`
	Value string `json:"value"`
}

type Aggregation struct {
	ID          int64  `json:"id"`
	PlatType    string `json:"plat_type"`
	Plat        string `json:"plat"`
	Idx         string `json:"idx"`
	RankType    int    `json:"rank_type"`
	RankValue   int    `json:"rank_value"`
	Hotword     string `json:"hot_word"`
	WordLabel   string `json:"word_label"`
	Hot         string `json:"hot"`
	ArcCnt      int    `json:"arc_cnt"`
	NewArcCnt   int    `json:"new_arc_cnt"`
	State       int    `json:"state"`
	ActiveState int    `json:"active_state"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Cover       string `json:"cover"`
	Update      string `json:"update"`
	CTime       int64  `json:"-"`
	MTime       int64  `json:"-"`
}

type Hotword struct {
	ID          int64  `json:"id"`
	Plat        string `json:"plat"`
	HotTitle    string `json:"hot_title"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	Image       string `json:"image"`
	State       int    `json:"state"`
	ActiveState int    `json:"active_state"`
	CTime       int64  `json:"ctime"`
	MTime       int64  `json:"mtime"`
}

func (a *Aggregation) FormAggregationMc(ai *showmdl.AggregationItem) {
	if a == nil {
		return
	}
	a.PlatType = ai.Platform
	a.Plat = FormPlatName(ai.Platform)
	a.Idx = strconv.Itoa(ai.Rank)
	a.RankType = ai.RankType
	a.RankValue = ai.RankValue
	a.Hotword = ai.Title
	a.WordLabel = ai.HotSearchType
	a.Hot = StatString(ai.HotFactor, "")
	a.Update = FormDatabusTime(ai.DatabusTime)
}

func (a *Aggregation) FormHot(h *Hotword) {
	if h == nil {
		return
	}
	a.ID = h.ID
	a.PlatType = h.Plat
	a.Plat = FormPlatName(h.Plat)
	a.Hotword = h.HotTitle
	a.Title = h.Title
	a.Subtitle = h.Subtitle
	a.Cover = h.Image
	a.State = h.State
	a.ActiveState = h.ActiveState
	a.CTime = h.CTime
	a.MTime = h.MTime
}

func FormPlatName(plat string) (name string) {
	switch plat {
	case GotoArtificial:
		name = "人工"
	case GotoWeibo:
		name = "微博"
	case GotoDouyin:
		name = "抖音"
	case GotoZhihu:
		name = "知乎"
	case GotoAcFun:
		name = "A站"
	case GotoBilibili:
		name = "B站热搜"
	case GotoBiliPopular:
		name = "B站热门"
	case GotoKuaishou:
		name = "快手"
	case GotoXigua:
		name = "西瓜视频"
	default:
		name = "未知来源"
	}
	return
}

// StatString Stat to string
func StatString(number int, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	//nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	//nolint:gomnd
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

func FormDatabusTime(databusTime int64) string {
	month := time.Unix(databusTime, 0).Format("01")
	day := time.Unix(databusTime, 0).Format("02")
	minite := time.Unix(databusTime, 0).Format("15:04")
	return fmt.Sprintf("%v月%v日%v更新", month, day, minite)
}

type MaterielsList struct {
	Tags      []*AggregationTag `json:"tags"`
	Hotword   string            `json:"hot_word"`
	ArcCnt    int               `json:"arc_cnt"`
	NewArcCnt int               `json:"new_arc_cnt"`
	Items     []*Materiel       `json:"items"`
}

type AggregationTag struct {
	ID    int64  `json:"id"`
	HotID int64  `json:"hot_id"`
	TagID int64  `json:"tag_id"`
	Title string `json:"title"`
	State int    `json:"state"`
}

type Materiel struct {
	ID           int64  `json:"id"`
	Source       string `json:"source"`
	HotID        int64  `json:"hot_id"`
	OID          int64  `json:"oid"`
	View         string `json:"view"`
	ViewInt      int32  `json:"-"`
	ViewSpeed    string `json:"view_speed"`
	ViewSpeedInt int32  `json:"-"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	State        int    `json:"state"`
}

func (m *Materiel) FormAI(ai *showmdl.CardList, arcm map[int64]*showmdl.ArcInfo) {
	if ai == nil {
		return
	}
	m.ID = ai.ID
	m.Source = ai.TagNames
	m.OID = ai.ID
	if arc, ok := arcm[ai.ID]; ok && arc != nil {
		m.ViewInt = arc.View
		m.View = StatString(int(arc.View), "")
		m.ViewSpeedInt = arc.ViewSpeed
		m.ViewSpeed = StatString(int(arc.ViewSpeed), "")
		m.Title = arc.Title
		m.Author = arc.Author
	}
}
