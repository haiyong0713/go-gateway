package model

import (
	xtime "time"

	"go-common/library/time"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	actpb "go-gateway/app/web-svr/activity/interface/api"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

// Filter filter struct
type Filter struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	SubTitle  string `json:"sub_title"`
	Logo      string `json:"logo"`
	Rank      int    `json:"rank"`
	URL       string `json:"url"`
	DataFocus string `json:"data_focus"`
	FocusURL  string `json:"focus_url"`
}

// Year year struct
type Year struct {
	ID   int64 `json:"id"`
	Year int64 `json:"year"`
	Aid  int64 `json:"aid"`
}

// Calendar calendar struct
type Calendar struct {
	Stime string `json:"stime"`
	Count int64  `json:"count"`
}

// Season season struct
type Season struct {
	ID           int64     `json:"id"`
	Mid          int64     `json:"mid"`
	Title        string    `json:"title"`
	SubTitle     string    `json:"sub_title"`
	Stime        int64     `json:"stime"`
	Etime        int64     `json:"etime"`
	Sponsor      string    `json:"sponsor"`
	Logo         string    `json:"logo"`
	Dic          string    `json:"dic"`
	Status       int64     `json:"status"`
	Ctime        time.Time `json:"ctime"`
	Mtime        time.Time `json:"mtime"`
	Rank         int64     `json:"rank"`
	IsApp        int64     `json:"is_app"`
	URL          string    `json:"url"`
	DataFocus    string    `json:"data_focus"`
	FocusURL     string    `json:"focus_url"`
	LeidaSID     int64     `json:"leida_sid"`
	GameType     int64     `json:"game_type"`
	SearchImage  string    `json:"search_image"`
	SyncPlatform int64     `json:"sync_platform"`
}

// Contest contest struct
type Contest struct {
	ID              int64       `json:"id"`
	GameStage       string      `json:"game_stage"`
	Stime           int64       `json:"stime"`
	Etime           int64       `json:"etime"`
	HomeID          int64       `json:"home_id"`
	AwayID          int64       `json:"away_id"`
	HomeScore       int64       `json:"home_score"`
	AwayScore       int64       `json:"away_score"`
	LiveRoom        int64       `json:"live_room"`
	Aid             int64       `json:"aid"`
	Collection      int64       `json:"collection"`
	CollectionBvid  string      `json:"collection_bvid"`
	GameState       int64       `json:"game_state"`
	Dic             string      `json:"dic"`
	Ctime           string      `json:"ctime"`
	Mtime           string      `json:"mtime"`
	Status          int64       `json:"status"`
	Sid             int64       `json:"sid"`
	Mid             int64       `json:"mid"`
	Season          interface{} `json:"season"`
	HomeTeam        interface{} `json:"home_team"`
	AwayTeam        interface{} `json:"away_team"`
	Special         int         `json:"special"`
	SuccessTeam     int64       `json:"success_team"`
	SuccessTeaminfo interface{} `json:"success_teaminfo"`
	SpecialName     string      `json:"special_name"`
	SpecialTips     string      `json:"special_tips"`
	SpecialImage    string      `json:"special_image"`
	Playback        string      `json:"playback"`
	CollectionURL   string      `json:"collection_url"`
	LiveURL         string      `json:"live_url"`
	DataType        int64       `json:"data_type"`
	MatchID         int64       `json:"match_id"`
	LiveSeason      *Season     `json:"-"`
	GuessType       int         `json:"guess_type"`
	GuessShow       int         `json:"guess_show"`
	Bvid            string      `json:"bvid"`
	GameStage1      string      `json:"game_stage1"`
	GameStage2      string      `json:"game_stage2"`
	// 0-未开播 1-直播中 2-轮播中
	LiveStatus  int64  `json:"live_status"`
	LivePopular int64  `json:"live_popular"`
	LiveCover   string `json:"live_cover"`
	PushSwitch  int64  `json:"push_switch"`
	LiveTitle   string `json:"live_title"`
	//关联阶段ID
	SeriesID int64 `json:"series_id"`
	// 比赛状态: 1未开始，2进行中，3已结束
	ContestStatus int64 `json:"contest_status"`
	ContestFreeze int64 `json:"contest_freeze"`

	// 赛程开始时间
	StartTime int64 `json:"start_time"`
	// 赛程结束时间
	EndTime int64 `json:"end_time"`
	// 赛程比赛阶段
	Title string `json:"title"`
	// 回播房间号url
	PlayBackV2 string `json:"play_back"`
	// 赛季id
	SeasonID int64 `json:"season_id"`
	// 是否订阅赛程
	IsSub int64 `json:"is_sub"`
	// 是否竞猜赛程
	IsGuess int64 `json:"is_guess"`
	// 主队
	Home *v1.Team4FrontendComponent `json:"home"`
	// 客队
	Away *v1.Team4FrontendComponent `json:"away"`
	// 系列赛阶段
	Series *v1.ContestSeriesComponent `json:"series"`
}

// ContestsData contest data struct
type ContestsData struct {
	ID         int64  `json:"id"`
	Cid        int64  `json:"cid"`
	URL        string `json:"url"`
	PointData  int64  `json:"point_data"`
	GameStatus int64  `json:"game_status"`
	DataType   int64  `json:"-"`
	Aid        int64  `json:"aid"`
	Pic        string `json:"pic"`
	View       int32  `json:"view"`
	Danmaku    int32  `json:"danmaku"`
	Duration   int64  `json:"duration"`
}

// ContestDataPage contest data pager
type ContestDataPage struct {
	Contest *Contest        `json:"contest"`
	Detail  []*ContestsData `json:"detail"`
}

func NewContestDataPage() *ContestDataPage {
	return &ContestDataPage{
		Contest: &Contest{},
		Detail:  make([]*ContestsData, 0),
	}
}

// ContestDataPageWithMatchRecord contest data pager
type ContestDataPageWithMatchRecord struct {
	Contest *Contest                `json:"contest"`
	Detail  []*ContestsData         `json:"detail"`
	Guess   []*actpb.GuessUserGroup `json:"guess"`
}

func NewContestDataPageWithMatchRecord() *ContestDataPageWithMatchRecord {
	return &ContestDataPageWithMatchRecord{
		Contest: nil,
		Detail:  make([]*ContestsData, 0),
		Guess:   make([]*actpb.GuessUserGroup, 0),
	}
}

// ElaSub elasticsearch sub contest.
type ElaSub struct {
	SeasonStime int64 `json:"season_stime"`
	Mid         int64 `json:"mid"`
	Stime       int64 `json:"stime"`
	Oid         int64 `json:"oid"`
	State       int64 `json:"state"`
	Sid         int64 `json:"sid"`
}

// Tree match Active
type Tree struct {
	ID        int64 `json:"id" form:"id"`
	MaID      int64 `json:"ma_id,omitempty" form:"ma_id" validate:"required"`
	MadID     int64 `json:"mad_id,omitempty" form:"mad_id" validate:"required"`
	Pid       int64 `json:"pid" form:"pid"`
	RootID    int64 `json:"root_id" form:"root_id"`
	GameRank  int64 `json:"game_rank,omitempty" form:"game_rank" validate:"required"`
	Mid       int64 `json:"mid" form:"mid"`
	IsDeleted int   `json:"is_deleted,omitempty" form:"is_deleted"`
}

// Team .
type Team struct {
	ID         int64  `json:"id" form:"id"`
	Title      string `json:"title" form:"title" validate:"required"`
	SubTitle   string `json:"sub_title" form:"sub_title"`
	ETitle     string `json:"e_title" form:"e_title"`
	CreateTime int64  `json:"create_time" form:"create_time"`
	Area       string `json:"area" form:"area"`
	Logo       string `json:"logo" form:"logo" validate:"required"`
	UID        int64  `json:"uid" form:"uid" gorm:"column:uid"`
	Members    string `json:"members" form:"members"`
	Dic        string `json:"dic" form:"dic"`
	IsDeleted  int    `json:"is_deleted" form:"is_deleted"`
	VideoURL   string `json:"video_url"`
	Profile    string `json:"profile"`
	LeidaTID   int64  `json:"leida_tid"`
	ReplyID    int64  `json:"reply_id"`
	TeamType   int64  `json:"team_type"`
	RegionID   int64  `json:"region_id"`
}

// ContestInfo .
type ContestInfo struct {
	*Contest
	HomeName    string `json:"home_name"`
	AwayName    string `json:"away_name"`
	SuccessName string `json:"success_name" form:"success_name"`
}

// TreeList .
type TreeList struct {
	*Tree
	*ContestInfo
}

// Active match Active
type Active struct {
	ID           int64  `json:"id"`
	Mid          int64  `json:"mid"`
	Sid          int64  `json:"sid"`
	Background   string `json:"background"`
	Liveid       int64  `json:"live_id"`
	Intr         string `json:"intr"`
	Focus        string `json:"focus"`
	URL          string `json:"url"`
	BackColor    string `json:"back_color"`
	ColorStep    string `json:"color_step"`
	H5Background string `json:"h5_background"`
	H5BackColor  string `json:"h5_back_color"`
	IntrLogo     string `json:"intr_logo"`
	IntrTitle    string `json:"intr_title"`
	IntrText     string `json:"intr_text"`
	H5Focus      string `json:"h5_focus"`
	H5Url        string `json:"h5_url"`
	Sids         string `json:"sids"`
	IsLive       string `json:"is_live"`
}

// Module match module
type Module struct {
	ID   int64  `json:"id"`
	MAid int64  `json:"ma_id"`
	Name string `json:"name"`
	Oids string `json:"oids"`
}

// ActiveDetail active detail.
type ActiveDetail struct {
	ID           int64  `json:"id"`
	Maid         int64  `json:"ma_id"`
	GameType     int    `json:"game_type"`
	STime        int64  `json:"stime"`
	ETime        int64  `json:"etime"`
	ScoreID      int64  `json:"score_id"`
	GameStage    string `json:"game_stage"`
	KnockoutType int    `json:"knockout_type"`
	WinnerType   int    `json:"winner_type"`
	Online       int    `json:"online"`
}

// ActiveLives active lives.
type ActiveLives struct {
	ID     int64  `json:"id"`
	Maid   int64  `json:"ma_id"`
	LiveID int64  `json:"live_id"`
	Title  string `json:"title"`
}

// ActivePage  active page.
type ActivePage struct {
	Active       *Active         `json:"active"`
	Videos       []*Video        `json:"video_first"`
	Modules      []*Module       `json:"video_module"`
	ActiveDetail []*ActiveDetail `json:"active_detail"`
	ActiveLives  []*ActiveLives  `json:"active_lives"`
}

// VideoList .
type Video struct {
	*arcmdl.Arc
	Bvid string `json:"bvid"`
}

// SearchMain search card main.
type SearchMain struct {
	ID        int64  `json:"id"`
	QueryName string `json:"query_name"`
	Stime     int64  `json:"stime"`
	Etime     int64  `json:"etime"`
}

// Search card main detail.
type SearchMD struct {
	SearchMain
	Cid int64 `json:"cid"`
}

// SearchRes search card res.
type SearchRes struct {
	*SearchMain
	ContestIDs []int64 `json:"contest_ids"`
}

// GameRank.
type GameRank struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	Rank     int    `json:"rank"`
}

// SeasonRank.
type SeasonRank struct {
	ID       int64  `json:"id"`
	Sid      int64  `json:"sid"`
	Rank     int    `json:"rank"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
}

type MoreContestCard struct {
	Timestamp    int64          `json:"timestamp"`
	ContestCards []*ContestCard `json:"contests"`
}

type DataInContestArea struct {
	TabCovers TabCovers      `json:"tab_covers"`
	Live      *Live4Frontend `json:"live"`
	LPL       []*ContestCard `json:"lpl"`
	Recent    []*ContestCard `json:"recent"`
	ShowLPL   bool           `json:"show_lpl"`
}

type TabCovers struct {
	Top    string `json:"top"`
	Middle string `json:"middle"`
	Bottom string `json:"bottom"`
}

type ContestCard struct {
	Contest   *Contest4Frontend `json:"contest"`
	More      []*ContestMore    `json:"more"`
	Timestamp int64             `json:"timestamp"`
}

type ContestCardList4Live struct {
	CardList  []*ContestCard `json:"card_list"`
	Timestamp int64          `json:"timestamp"`
	IsLocated bool           `json:"is_located"`
}

type ContestMore struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	OnClick string `json:"on_click"`
}

type Contest4Frontend struct {
	ID        int64         `json:"id"`
	StartTime int64         `json:"start_time"`
	EndTime   int64         `json:"end_time"`
	Title     string        `json:"title"`
	Status    string        `json:"status"`
	Home      Team4Frontend `json:"home"`
	Away      Team4Frontend `json:"away"`
	Series    ContestSeries `json:"series"`
	SeriesID  int64         `json:"series_id"`
}

type ContestSeries struct {
	ID          int64  `json:"id"`
	ParentTitle string `json:"parent_title"`
	ChildTitle  string `json:"child_title"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	ScoreID     string `json:"score_id"`
	InTheSeries bool   `json:"in_the_series"`

	Detail []*ContestCardList4Live `json:"detail"`
}

type PosterList4S10 struct {
	UpdateAt int64         `json:"updated_at"`
	List     []*Poster4S10 `json:"list"`
}

type Poster4S10 struct {
	BackGround string           `json:"back_ground"`
	InCenter   int64            `json:"in_center"`
	ContestID  int64            `json:"contest_id"`
	Contest    Contest4Frontend `json:"contest"`
	More       []*ContestMore   `json:"more"`
}

type Live4Frontend struct {
	IsLive    bool          `json:"is_live"`
	ContestID int64         `json:"contest_id"`
	Home      Team4Frontend `json:"home"`
	Away      Team4Frontend `json:"away"`
}

type Team4Frontend struct {
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Wins     int64  `json:"wins"`
	Region   string `json:"region"`
	RegionID int    `json:"region_id"`
}

const (
	secondsOf10Minutes = 600
)

// reset some fields by biz
func (series *ContestSeries) Rebuild() {
	now := xtime.Now().Unix()
	if series.StartTime <= now && series.EndTime > now {
		series.InTheSeries = true
	}
}

func (contest *Contest4Frontend) CalculateTimestampDiff() int64 {
	now := xtime.Now().Unix()
	diff4StartTime := now - contest.StartTime
	diff4EndTime := now - contest.EndTime
	if diff4StartTime < 0 {
		diff4StartTime = -diff4StartTime
	}

	if diff4EndTime < 0 {
		diff4EndTime = -diff4EndTime
	}

	if diff4StartTime < diff4EndTime {
		return diff4StartTime
	}

	return diff4EndTime
}

func (card *ContestCard) FromLPL() bool {
	if card.Contest.Home.RegionID == teamRegionIDOfChina ||
		card.Contest.Home.RegionID == teamRegionIDOfChinaTaiWan ||
		card.Contest.Away.RegionID == teamRegionIDOfChina ||
		card.Contest.Away.RegionID == teamRegionIDOfChinaTaiWan {
		return true
	}

	return false
}

func (card *ContestCard) ResetMore(subM map[int64]bool, guessM map[int64]bool) {
	tmpMoreList := make([]*ContestMore, 0)

	for _, v := range card.More {
		tmpMore := new(ContestMore)
		*tmpMore = *v
		switch card.Contest.Status {
		case ContestStatusOfNotStart:
			switch v.Status {
			case MoreStatusOfSubscribe:
				if d, ok := subM[card.Contest.ID]; ok && d {
					tmpMore.OnClick = ClickStatusOfDisabled
					tmpMore.Title = MoreDisplayOfSubscribed
				}

				tmpMoreList = append(tmpMoreList, tmpMore)
			case MoreStatusOfPrediction:
				if d, ok := guessM[card.Contest.ID]; ok && d {
					tmpMore.OnClick = ClickStatusOfDisabled
					tmpMore.Title = MoreDisplayOfPredicted

					tmpMoreList = append(tmpMoreList, tmpMore)
				} else {
					if card.Contest.StartTime-xtime.Now().Unix() >= secondsOf10Minutes {
						tmpMoreList = append(tmpMoreList, tmpMore)
					}
				}
			}
		default:
			switch v.Status {
			case MoreStatusOfSubscribe:
				if d, ok := subM[card.Contest.ID]; ok && d {
					tmpMore.OnClick = ClickStatusOfDisabled
					tmpMore.Title = MoreDisplayOfSubscribed
				}

				tmpMoreList = append(tmpMoreList, tmpMore)
			case MoreStatusOfPrediction:
				if d, ok := guessM[card.Contest.ID]; ok && d {
					tmpMore.OnClick = ClickStatusOfDisabled
					tmpMore.Title = MoreDisplayOfPredicted

					tmpMoreList = append(tmpMoreList, tmpMore)
				}
			default:
				tmpMoreList = append(tmpMoreList, tmpMore)
			}
		}
	}

	card.More = tmpMoreList
}

const (
	teamRegionIDOfNull = iota
	teamRegionIDOfChina
	teamRegionIDOfChinaTaiWan

	ContestStatusOfNotStart = "not_start"
	ContestStatusOfOngoing  = "ongoing"
	ContestStatusOfEnd      = "end"
)

const (
	MoreStatusOfSubscribe  = "subscribe"
	MoreStatusOfPrediction = "prediction"
	MoreStatusOfLive       = "live"
	MoreStatusOfReplay     = "replay"
	MoreStatusOfCollection = "collection"
	MoreStatusOfEnd        = "end"

	MoreDisplayOfSubscribe  = "去订阅"
	MoreDisplayOfSubscribed = "已订阅"
	MoreDisplayOfPrediction = "预测"
	MoreDisplayOfPredicted  = "已预测"
	MoreDisplayOfLive       = "直播中"
	MoreDisplayOfReplay     = "回放"
	MoreDisplayOfCollection = "集锦"
	MoreDisplayOfEnd        = "已结束"

	ClickStatusOfEnabled  = "enabled"
	ClickStatusOfDisabled = "disabled"
)

// NOTE: do not include all fields!!!
func (d *Contest) DeepCopy() *Contest {
	tmp := new(Contest)
	{
		tmp.ID = d.ID
		tmp.GameStage = d.GameStage
		tmp.Stime = d.Stime
		tmp.Etime = d.Etime
		tmp.HomeID = d.HomeID
		tmp.AwayID = d.AwayID
		tmp.HomeScore = d.HomeScore
		tmp.AwayScore = d.AwayScore
		tmp.LiveRoom = d.LiveRoom
		tmp.Aid = d.Aid
		tmp.Collection = d.Collection
		tmp.GameState = d.GameState
		tmp.Dic = d.Dic
		tmp.Ctime = d.Ctime
		tmp.Mtime = d.Mtime
		tmp.Status = d.Status
		tmp.Sid = d.Sid
		tmp.Mid = d.Mid
		tmp.Special = d.Special
		tmp.SpecialName = d.SpecialName
		tmp.SpecialTips = d.SpecialTips
		tmp.SuccessTeam = d.SuccessTeam
		tmp.SpecialImage = d.SpecialImage
		tmp.Playback = d.Playback
		tmp.CollectionURL = d.CollectionURL
		tmp.LiveURL = d.LiveURL
		tmp.DataType = d.DataType
		tmp.MatchID = d.MatchID
		tmp.GuessType = d.GuessType
		tmp.GameStage1 = d.GameStage1
		tmp.GameStage2 = d.GameStage2
		tmp.PushSwitch = d.PushSwitch
		tmp.SeriesID = d.SeriesID
		tmp.ContestFreeze = d.Status
		tmp.ContestStatus = d.ContestStatus
	}

	return tmp
}

func (d *Season) DeepCopy() *Season {
	tmp := new(Season)
	{
		tmp.ID = d.ID
		tmp.Mid = d.Mid
		tmp.Title = d.Title
		tmp.SubTitle = d.SubTitle
		tmp.Stime = d.Stime
		tmp.Etime = d.Etime
		tmp.Sponsor = d.Sponsor
		tmp.Logo = d.Logo
		tmp.Dic = d.Dic
		tmp.Status = d.Status
		tmp.Ctime = d.Ctime
		tmp.Mtime = d.Mtime
		tmp.Rank = d.Rank
		tmp.IsApp = d.IsApp
		tmp.URL = d.URL
		tmp.DataFocus = d.DataFocus
		tmp.FocusURL = d.FocusURL
		tmp.LeidaSID = d.LeidaSID
		tmp.GameType = d.GameType
		tmp.SearchImage = d.SearchImage
		tmp.SyncPlatform = d.SyncPlatform
	}

	return tmp
}

func (d *Team) DeepCopy() *Team {
	tmp := new(Team)
	{
		tmp.ID = d.ID
		tmp.Title = d.Title
		tmp.SubTitle = d.SubTitle
		tmp.ETitle = d.ETitle
		tmp.CreateTime = d.CreateTime
		tmp.Area = d.Area
		tmp.Logo = d.Logo
		tmp.UID = d.UID
		tmp.Members = d.Members
		tmp.Dic = d.Dic
		tmp.IsDeleted = d.IsDeleted
		tmp.VideoURL = d.VideoURL
		tmp.Profile = d.Profile
		tmp.LeidaTID = d.LeidaTID
		tmp.ReplyID = d.ReplyID
		tmp.TeamType = d.TeamType
		tmp.RegionID = d.RegionID
	}

	return tmp
}

func (d *Team) Convert2SimplifyEdition() *Team4SimplifyEdition {
	tmp := new(Team4SimplifyEdition)
	{
		tmp.ID = d.ID
		tmp.Title = d.Title
		tmp.Logo = d.Logo
	}

	return tmp
}
