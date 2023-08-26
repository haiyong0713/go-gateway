package model

// ParamCale  calendar params.
type ParamCale struct {
	Stime int64 `form:"stime" validate:"required"`
	Etime int64 `form:"etime" validate:"required"`
}

// ParamContest matchs params.
type ParamContest struct {
	Mid     int64   `form:"mid" validate:"gte=0"`
	Gid     int64   `form:"gid" validate:"gte=0"`
	Tid     int64   `form:"tid" validate:"gte=0"`
	Stime   string  `form:"stime"`
	Etime   string  `form:"etime"`
	GState  string  `form:"g_state"`
	Sids    []int64 `form:"sids,split"`
	Forbid  int     `form:"forbid"`
	Roomids []int64 `form:"roomids,split"`
	Sort    int     `form:"sort"`
	Pn      int     `form:"pn"  validate:"gt=0"`
	Ps      int     `form:"ps"  validate:"gt=0,lte=50"`
	Cids    []int64
	GsType  int
	GsRecT  int64
}

// ParamEsGuess .
type ParamEsGuess struct {
	CID    int64 `form:"cid" validate:"required"`
	HomeID int64 `form:"home_id" validate:"required"`
	AwayID int64 `form:"away_id" validate:"required"`
	Ps     int   `form:"ps" default:"5" validate:"lte=10"`
}

// ParamVideo video params
type ParamVideo struct {
	Mid  int64 `form:"mid"   validate:"gte=0"`
	Gid  int64 `form:"gid"   validate:"gte=0"`
	Tid  int64 `form:"tid"   validate:"gte=0"`
	Year int64 `form:"year"  validate:"gte=0"`
	Tag  int64 `form:"tag"   validate:"gte=0"`
	Sort int64 `form:"sort"  validate:"gte=0"`
	Pn   int   `form:"pn"    validate:"gt=0"`
	Ps   int   `form:"ps"    validate:"gt=0,lte=50"`
}

// ParamSearch search video params
type ParamSearch struct {
	Pn      int    `form:"pn"    validate:"gt=0"`
	Ps      int    `form:"ps"    validate:"gt=0"`
	Keyword string `form:"keyword" validate:"required"`
	Sort    int64  `form:"sort"  validate:"gte=0"`
}

// ParamSeason season params.
type ParamSeason struct {
	VMID int64 `form:"vmid"`
	Sort int64 `form:"sort"`
	Pn   int   `form:"pn"  validate:"gt=0"`
	Ps   int   `form:"ps"  validate:"gt=0,lte=50"`
}

// ParamFilter  filter video params
type ParamFilter struct {
	Mid   int64   `form:"mid"   validate:"gte=0"`
	Gid   int64   `form:"gid"   validate:"gte=0"`
	Tid   int64   `form:"tid"   validate:"gte=0"`
	Sids  []int64 `form:"sids,split"`
	Year  int64   `form:"year"  validate:"gte=0"`
	Tag   int64   `form:"tag"   validate:"gte=0"`
	Stime string  `form:"stime" `
	Etime string  `form:"etime" `
}

// ParamActPoint matchs params.
type ParamActPoint struct {
	Aid  int64 `form:"aid" validate:"gt=0"`
	MdID int64 `form:"md_id" validate:"gt=0"`
	Sort int   `form:"sort"`
	Tp   int64 `form:"tp"`
	Pn   int   `form:"pn"  validate:"gt=0"`
	Ps   int   `form:"ps"  validate:"gt=0,lte=50"`
}

// ParamActTop matchs params.
type ParamActTop struct {
	Aid   int64  `form:"aid" validate:"gt=0"`
	Sort  int    `form:"sort"`
	Stime string `form:"stime" `
	Etime string `form:"etime" `
	Tp    int64  `form:"tp"`
	Pn    int    `form:"pn" default:"1" validate:"min=1"`
	Ps    int    `form:"ps" default:"50" validate:"gt=0,lte=50"`
}

// ParamFav app fav list.
type ParamFav struct {
	VMID  int64   `form:"vmid"`
	Sids  []int64 `form:"sids,split"`
	Stime string  `form:"stime"`
	Etime string  `form:"etime"`
	Sort  int     `form:"sort"`
	Pn    int     `form:"pn" default:"1" validate:"min=1"`
	Ps    int     `form:"ps" default:"50" validate:"min=1"`
}

// ParamLd leidata param
type ParamLd struct {
	Route string `form:"route"`
}

// ParamCDRecent contest recently match
type ParamCDRecent struct {
	HomeID int64 `form:"home_id" validate:"gt=0"`
	AwayID int64 `form:"away_id" validate:"gt=0"`
	CID    int64 `form:"cid" validate:"gt=0"`
	Ps     int64 `form:"ps" default:"8" validate:"lte=10"`
}

// ParamGame game
type ParamGame struct {
	MatchID int64   `form:"match_id" validate:"required"`
	GameIDs []int64 `form:"game_ids,split" validate:"required"`
	Tp      int64   `form:"tp" default:"1" validate:"min=1"`
}

// ParamLeidas .
type ParamLeidas struct {
	IDs []int64 `form:"ids,split" validate:"required"`
	Tp  int64   `form:"tp" default:"1" validate:"min=1"`
}

// StatsBig .
type StatsBig struct {
	Sid       int64  `form:"sid" validate:"required"`
	Tp        int64  `form:"tp" default:"1" validate:"min=1"`
	SortValue int    `form:"sort_value"`
	SortType  string `form:"sort_type"`
	Role      string `form:"role"`
	Pn        int    `form:"pn" default:"1" validate:"min=1"`
	Ps        int    `form:"ps" default:"10" validate:"min=1,lte=100"`
}

// MatchLive .
type MatchLive struct {
	GID   int64 `form:"g_id"`
	MID   int64 `form:"m_id"`
	SID   int64 `form:"s_id"`
	STime int64 `form:"stime"`
	Ps    int   `form:"ps" default:"20"`
	Pn    int   `form:"pn" default:"1"`
}

// ParamSpecTeams .
type ParamSpecTeams struct {
	Tp       int64 `form:"tp" default:"1" validate:"min=1"`
	LeidaSID int64 `form:"leida_sid"`
	Sort     int   `form:"sort"`
	Pn       int   `form:"pn" default:"1" validate:"min=1"`
	Ps       int   `form:"ps" default:"30" validate:"min=1,lte=50"`
}

// ParamSpecial .
type ParamSpecial struct {
	Tp       int64 `form:"tp" validate:"required"`
	ID       int64 `form:"id" validate:"required"`
	LeidaSID int64 `form:"leida_sid" validate:"required"`
	Recent   int64 `form:"recent"`
}

// ParamRecent .
type ParamRecent struct {
	LeidaSID int64 `form:"leida_sid" validate:"required"`
	LeidaTID int64 `form:"leida_tid" validate:"required"`
	Ps       int   `form:"ps" default:"20"`
	Pn       int   `form:"pn" default:"1"`
}

// ParamGQuess .
type ParamGQuess struct {
	Gid   int64  `form:"gid" validate:"required"`
	Sid   int64  `form:"sid" validate:"required"`
	Stime string `form:"stime"`
	Etime string `form:"etime"`
}

type ParamAllContest struct {
	Sid   int64 `form:"sid" validate:"min=1"`
	Sort  int   `form:"sort"`
	Stime int64 `form:"stime"`
	Etime int64 `form:"etime"`
	Pn    int   `form:"pn" default:"1" validate:"min=1"`
	Ps    int   `form:"ps" default:"5" validate:"min=1,lte=50"`
}

type ParamAllFold struct {
	Sid   int64 `form:"sid" validate:"min=1"`
	Front int   `form:"front"`
	Back  int   `form:"back"`
}

type ParamAbstract struct {
	Sid int64 `form:"sid" validate:"min=1"`
}

type ParamSeasonContests struct {
	Sid  int64 `form:"sid" validate:"min=1"`
	Ps   int   `form:"ps" default:"4" validate:"min=1,lte=50"`
	Prev int   `form:"prev"`
	Next int   `form:"next"`
}

type ParamMatchSeasons struct {
	MatchID   int64   `form:"match_id" validate:"min=1"`
	SeasonIDs []int64 `form:"season_ids,split" validate:"required,dive,gt=0"`
}

type ParamSeasonTeams struct {
	SeasonID int64 `form:"season_id" validate:"min=1"`
}

type ParamSeasonsInfo struct {
	SeasonIDs []int64 `form:"season_ids,split" validate:"required,dive,gt=0"`
}

type ParamContestBattle struct {
	Sid     int64 `form:"sid" validate:"min=1"`
	Sort    int   `form:"sort"`
	Stime   int64 `form:"stime"`
	Etime   int64 `form:"etime"`
	TeamTop int   `form:"team_top" default:"3"`
	Pn      int   `form:"pn" default:"1" validate:"min=1"`
	Ps      int   `form:"ps" default:"5" validate:"min=1,lte=50"`
}

type ParamBattleTeams struct {
	Sid       int64 `form:"sid" validate:"min=1"`
	ContestID int64 `form:"contest_id" validate:"min=1"`
}

type ParamTeamContest struct {
	Sid       int64   `form:"sid" validate:"min=1"`
	SeriesID  int64   `form:"series_id" validate:"min=1"`
	TeamIDs   []int64 `form:"team_ids,split"`
	GroupName string  `form:"group_name"`
	Sort      int     `form:"sort"`
	Pn        int     `form:"pn" default:"1" validate:"min=1"`
	Ps        int     `form:"ps" default:"5" validate:"min=1,lte=50"`
}

type ParamV2TeamContest struct {
	Sid       int64   `form:"sid" validate:"min=1"`
	SeriesID  int64   `form:"series_id" validate:"min=1"`
	TeamIDs   []int64 `form:"team_ids,split"`
	GroupName string  `form:"group_name"`
	Ps        int     `form:"ps" default:"5" validate:"min=1,lte=50"`
	Prev      int     `form:"prev"`
	Next      int     `form:"next"`
}

type ParamVideoList struct {
	ID int64 `form:"id" validate:"min=1"`
	Pn int   `form:"pn" default:"1" validate:"min=1"`
	Ps int   `form:"ps" default:"5" validate:"min=1,lte=50"`
}

// ParamMvpRank .
type ParamMvpRank struct {
	Top      int    `form:"top" default:"20" validate:"min=1"`
	SeasonID int64  `form:"season_id" validate:"required"`
	SortType string `form:"sort_type" default:"mvp" validate:"required"`
}

// ParamKdaRank .
type ParamKdaRank struct {
	SeasonID int64  `form:"season_id" validate:"required"`
	SortType string `form:"sort_type" default:"kda" validate:"required"`
}

// ParamHero2Rank .
type ParamHero2Rank struct {
	Top      int   `form:"top" default:"10" validate:"min=1"`
	SeasonID int64 `form:"season_id" validate:"required"`
}

type ParamWall struct {
	Sid    int64 `form:"sid" validate:"min=1"`
	RoomID int64 `form:"room_id" validate:"min=1"`
}
