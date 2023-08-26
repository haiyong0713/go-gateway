package model

import (
	"go-common/library/time"
)

// FtpMatchs .
type FtpMatchs struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	Logo     string `json:"logo"`
	Rank     int    `json:"rank"`
}

type FtpEsports struct {
	ID          int64  `json:"id"`
	Gname       string `json:"gname"`       //赛程（比赛）名称
	AliasSearch string `json:"aliasSearch"` //赛程别名
	StartTime   int64  `json:"start_time"`  //赛程开始时间
	EndTime     int64  `json:"end_time"`    //赛程结束时间
	IsBlock     int    `json:"is_block"`    //是否有官方战队
	Ischeck     int    `json:"ischeck"`     //字段过滤过滤	false		0-正常状态，1-无直播&视频回放&集锦
	Status      int    `json:"status"`      //赛程类型
	Spid        int64  `json:"spid"`        //主队id
	IBrandname  string `json:"i_brandname"` //主队名
	SBrandname  string `json:"s_brandname"` //主队别名
	SalerID     int64  `json:"saler_id"`    //客队id
	ICategory   string `json:"i_category"`  //客队名
	SCategory   string `json:"s_category"`  //客队别名
	TpID        int64  `json:"tp_id"`       //所属赛季id
	Title       string `json:"title"`       //赛季名称
	AliasTitle  string `json:"alias_title"` //赛事别名
	Pubtime     int64  `json:"pubtime"`     //赛季开始时间
	Lastupdate  int64  `json:"lastupdate"`  //赛季结束时间
	URL         string `json:"url"`         //赛季落地页地址
	SqURL       string `json:"sq_url"`      //赛季落地页地址
}

// FtpTeams .
type FtpTeams struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	SubTitle  string `json:"sub_title"`
	Logo      string `json:"logo"`
	Rank      int    `json:"rank"`
	URL       string `json:"url"`
	DataFocus string `json:"data_focus"`
	FocusURL  string `json:"focus_url"`
	TeamType  int    `json:"team_type"`
	ETitle    string `json:"e_title"`
}

// FtpContest .
type FtpContest struct {
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
}

// FtpSeason .
type FtpSeason struct {
	ID        int64     `json:"id"`
	Mid       int64     `json:"mid"`
	Title     string    `json:"title"`
	SubTitle  string    `json:"sub_title"`
	Stime     int64     `json:"stime"`
	Etime     int64     `json:"etime"`
	Sponsor   string    `json:"sponsor"`
	Logo      string    `json:"logo"`
	Dic       string    `json:"dic"`
	Status    int64     `json:"status"`
	Ctime     time.Time `json:"ctime"`
	Mtime     time.Time `json:"mtime"`
	Rank      int64     `json:"rank"`
	IsApp     int64     `json:"is_app"`
	URL       string    `json:"url"`
	DataFocus string    `json:"data_focus"`
	FocusURL  string    `json:"focus_url"`
	LeidaSID  int64     `json:"leida_sid"`
	GameType  int64     `json:"game_type"`
}
