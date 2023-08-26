package model

import "time"

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

// Contest contest.
type Contest struct {
	ID             int64  `json:"id"`
	Stime          int64  `json:"stime"`
	Etime          int64  `json:"etime"`
	LiveRoom       int64  `json:"live_room"`
	HomeID         int64  `json:"home_id"`
	AwayID         int64  `json:"away_id"`
	SuccessTeam    int64  `json:"success_team"`
	SeasonTitle    string `json:"season_title"`
	SeasonSubTitle string `json:"season_sub_title"`
	Special        int    `json:"special"`
	SpecialName    string `json:"special_name"`
	SpecialTips    string `json:"special_tips"`
	DataType       int64  `json:"data_type"`
	MatchID        int64  `json:"match_id"`
	SeasonID       int64  `json:"season_id"`
	ContestStatus  int64  `json:"contest_status"`
	MessageSendUid int64  `json:"message_send_uid"`
}

// ContestSeriesTable Contest Series table .
type ContestSeriesTable struct {
	ID       int64 `json:"id"`
	SeasonID int64 `json:"season_id"`
}

// Contest contest.
type Contest2Tab struct {
	ID             int64  `json:"id"`
	StimeDate      int64  `jsob:"date"`
	Stime          int64  `json:"stime"`
	Etime          int64  `json:"etime"`
	CollectionUrl  string `json:"collection_url"`
	LiveRoom       int64  `json:"live_room"`
	PlayBack       string `json:"play_back"`
	HomeID         int64  `json:"home_id"`
	HomeScore      int64  `json:"home_score"`
	AwayID         int64  `json:"away_id"`
	AwayScore      int64  `json:"away_score"`
	SuccessTeam    int64  `json:"success_team"`
	SeasonTitle    string `json:"season_title"`
	SeasonSubTitle string `json:"season_sub_title"`
	Special        int    `json:"special"`
	SpecialName    string `json:"special_name"`
	SpecialTips    string `json:"special_tips"`
	DataType       int64  `json:"data_type"`
	MatchID        int64  `json:"match_id"`
	SeasonID       int64  `json:"season_id"`
	GameStage      string `json:"stage"`
	SeriesID       int64  `json:"series_id"`
	GuessType      int64  `json:"guess_type"`
}

type ContestCard struct {
	Contest Contest4Frontend `json:"contest"`
	More    []*ContestMore   `json:"more"`
}

type ContestMore struct {
	Status  string `json:"status"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	OnClick string `json:"on_click"`
}

type Contest4Frontend struct {
	ID        int64          `json:"id"`
	StartTime int64          `json:"start_time"`
	EndTime   int64          `json:"end_time"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Home      Team4Frontend  `json:"home"`
	Away      Team4Frontend  `json:"away"`
	Series    *ContestSeries `json:"series"`
	SeriesID  int64          `json:"series_id"`
}

type Team4Frontend struct {
	ID       int64  `json:"id"`
	Icon     string `json:"icon"`
	Name     string `json:"name"`
	Wins     int64  `json:"wins"`
	Region   string `json:"region"`
	RegionID int    `json:"region_id"`
}

// Team team.
type Team struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
}

// Team team.
type Team2Tab struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	SubTitle    string `json:"sub_title"`
	Logo        string `json:"logo"`
	RegionID    int    `json:"region_id"`
	Region      string `json:"region"`
	ScoreTeamID int64  `json:"score_team_id"`
}

// ContestData .
type ContestData2Tab struct {
	CID       int64  `json:"cid" gorm:"column:cid"`
	URL       string `json:"url"`
	PointData int    `json:"point_data"`
	AvCID     int64  `json:"av_cid" gorm:"column:av_cid"`
}

type ContestSeries struct {
	ID          int64  `json:"id"`
	ParentTitle string `json:"parent_title" validate:"required"`
	ChildTitle  string `json:"child_title" validate:"required"`
	StartTime   int64  `json:"start_time" validate:"min=1"`
	EndTime     int64  `json:"end_time" validate:"min=1"`
	ScoreID     string `json:"score_id" validate:"required"`
}

// Arc  arc.
type Arc struct {
	ID        int64 `json:"id"`
	Aid       int64 `json:"aid"`
	Score     int64 `json:"score"`
	IsDeleted int   `json:"is_deleted"`
}

// Oid team or player id.
type Oid struct {
	ID int64 `json:"id"`
}

// LdInfo leida info.
type LdInfo struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	ID       int    `json:"id"`
}

// PlayerInfo leida player info.
type PlayerInfo struct {
	LdInfo
	Role string `json:"role"`
}

// ScorePlayer.
type ScorePlayer struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	ID       int    `json:"id"`
	Role     string `json:"role"`
}

// LolPlayer lol player big data.
type LolPlayer struct {
	ID                  int64   `json:"id"`
	PlayerID            int64   `json:"player_id"`
	TeamID              int64   `json:"team_id"`
	TeamAcronym         string  `json:"team_acronym"`
	TeamImage           string  `json:"team_image"`
	LeidaTeamImage      string  `json:"leida_team_image"`
	LeidaSID            int64   `json:"leida_sid"`
	Name                string  `json:"name"`
	ImageURL            string  `json:"image_url"`
	LeidaImage          string  `json:"leida_image"`
	ChampionsImage      string  `json:"champions_image"`
	LeidaChampionsImage string  `json:"leida_champions_image"`
	Role                string  `json:"role"`
	Win                 float64 `json:"win"`
	KDA                 float64 `json:"kda"`
	Kills               float64 `json:"kills"`
	Deaths              float64 `json:"deaths"`
	Assists             float64 `json:"assists"`
	MinionsKilled       float64 `json:"minions_killed"`
	WardsPlaced         float64 `json:"wards_placed"`
	GamesCount          int64   `json:"games_count"`
	MVP                 int64   `json:"mvp"`
	PositionID          int64   `json:"position_id"`
	Position            string  `json:"position"`
	Ctime               string  `json:"ctime"`
	Mtime               string  `json:"mtime"`
}

// LolPlayerStats lol player stats.
type LolPlayerStats struct {
	Stats struct {
		Totals struct {
			MatchesWon    int64 `json:"matches_won"`
			MatchesPlayed int64 `json:"matches_played"`
			MatchesLost   int64 `json:"matches_lost"`
			GamesWon      int64 `json:"games_won"`
			GamesPlayed   int64 `json:"games_played"`
			GamesLost     int64 `json:"games_lost"`
		} `json:"totals"`
		GamesCount int64 `json:"games_count"`
		Averages   struct {
			WardsPlaced   float64 `json:"wards_placed"`
			MinionsKilled float64 `json:"minions_killed"`
			Kills         float64 `json:"kills"`
			Deaths        float64 `json:"deaths"`
			Assists       float64 `json:"assists"`
		} `json:"averages"`
	} `json:"stats"`
	Role              string       `json:"role"`
	Name              string       `json:"name"`
	ImageURL          string       `json:"image_url"`
	ID                int64        `json:"id"`
	FavoriteChampions []*FavChamps `json:"favorite_champions"`
	CurrentTeam       struct {
		Name     string `json:"name"`
		ImageURL string `json:"image_url"`
		ID       int64  `json:"id"`
		Acronym  string `json:"acronym"`
	} `json:"current_team"`
}

type LolDataPlayer struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []*LolDataPlayerData `json:"list"`
	} `json:"data"`
}

type LolDataPlayerData struct {
	ID                string `json:"id"`
	TournamentID      string `json:"tournament_id"`
	PlayerID          string `json:"player_id"`
	PlayerName        string `json:"player_name"`
	PlayerImage       string `json:"player_image"`
	TeamID            string `json:"team_id"`
	TeamName          string `json:"team_name"`
	TeamImage         string `json:"team_image"`
	Position          string `json:"position"`
	KDA               string `json:"KDA"`
	PLAYSTIMES        string `json:"PLAYS_TIMES"`
	OFFEREDRATE       string `json:"OFFERED_RATE"`
	AVERAGEKILLS      string `json:"AVERAGE_KILLS"`
	AVERAGEASSISTS    string `json:"AVERAGE_ASSISTS"`
	AVERAGEDEATHS     string `json:"AVERAGE_DEATHS"`
	MINUTEECONOMIC    string `json:"MINUTE_ECONOMIC"`
	MINUTEHITS        string `json:"MINUTE_HITS"`
	MINUTEDAMAGEDEALT string `json:"MINUTE_DAMAGEDEALT"`
	DAMAGEDEALTRATE   string `json:"DAMAGEDEALT_RATE"`
	MINUTEDAMAGETAKEN string `json:"MINUTE_DAMAGETAKEN"`
	DAMAGETAKENRATE   string `json:"DAMAGETAKEN_RATE"`
	MINUTEWARDSPLACED string `json:"MINUTE_WARDSPLACED"`
	MINUTEWARDKILLED  string `json:"MINUTE_WARDKILLED"`
	UpdateTime        string `json:"update_time"`
	MVP               string `json:"MVP"`
	PlayerChineseName string `json:"player_chinese_name"`
	Win               string `json:"win"`
	Los               string `json:"los"`
	VICTORYRATE       string `json:"VICTORY_RATE"`
	CountryID         string `json:"country_id"`
	CountryImage      string `json:"country_image"`
	FScore            string `json:"f_score"`
	PositionID        string `json:"position_id"`
	TotalKills        string `json:"total_kills"`
	TotalDeaths       string `json:"total_deaths"`
	TotalAssists      string `json:"total_assists"`
}

type LolDataHero2 struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []*LolDataHero2Data `json:"list"`
	} `json:"data"`
}

type LolDataHero2Data struct {
	HeroID        string `json:"hero_id"`
	HeroName      string `json:"hero_name"`
	HeroImage     string `json:"hero_image"`
	AppearCount   string `json:"appear_count"`
	ProhibitCount string `json:"prohibit_count"`
	VictoryCount  string `json:"victory_count"`
	GameCount     string `json:"game_count"`
}

// FavChamps  lol Favorite Champions.
type FavChamps struct {
	Champion struct {
		ImageURL string `json:"image_url"`
	} `json:"champion"`
}

// LolTeam lol team big data.
type LolTeam struct {
	ID                 int64   `json:"id"`
	TeamID             int64   `json:"team_id"`
	Acronym            string  `json:"acronym"`
	LeidaSID           int64   `json:"leida_sid"`
	Name               string  `json:"name"`
	ImageURL           string  `json:"image_url"`
	LeidaImage         string  `json:"leida_image"`
	Win                float64 `json:"win"`
	KDA                float64 `json:"kda"`
	Kills              float64 `json:"kills"`
	Deaths             float64 `json:"deaths"`
	Assists            float64 `json:"assists"`
	TowerKills         float64 `json:"tower_kills"`
	TotalMinionsKilled float64 `json:"total_minions_killed"`
	FirstTower         float64 `json:"first_tower"`
	FirstInhibitor     float64 `json:"first_inhibitor"`
	FirstDragon        float64 `json:"first_dragon"`
	FirstBaron         float64 `json:"first_baron"`
	FirstBlood         float64 `json:"first_blood"`
	WardsPlaced        float64 `json:"wards_placed"`
	InhibitorKills     float64 `json:"inhibitor_kills"`
	BaronKills         float64 `json:"baron_kills"`
	GoldEarned         float64 `json:"gold_earned"`
	GamesCount         int64   `json:"games_count"`
	Players            string  `json:"players"`
	Ctime              string  `json:"ctime"`
	Mtime              string  `json:"mtime"`
	BaronRate          float64 `json:"baron_rate"`
	DragonRate         float64 `json:"dragon_rate"`
	Hits               float64 `json:"hits"`
	LoseNum            float64 `json:"lose_num"`
	Money              float64 `json:"money"`
	TotalDamage        float64 `json:"total_damage"`
	WinNum             float64 `json:"win_num"`
	ImageThumb         string  `json:"image_thumb"`
	NewData            int64   `json:"new_data"`
}

// LolTeamStats lol team stats.
type LolTeamStats struct {
	Stats struct {
		GamesCount int64 `json:"games_count"`
		Averages   struct {
			WardsPlaced        float64 `json:"wards_placed"`
			TowerKills         float64 `json:"tower_kills"`
			TotalMinionsKilled float64 `json:"total_minions_killed"`
			Ratios             struct {
				Win            float64 `json:"win"`
				FirstTower     float64 `json:"first_tower"`
				FirstInhibitor float64 `json:"first_inhibitor"`
				FirstDragon    float64 `json:"first_dragon"`
				FirstBlood     float64 `json:"first_blood"`
				FirstBaron     float64 `json:"first_baron"`
			} `json:"ratios"`
			Kills          float64 `json:"kills"`
			InhibitorKills float64 `json:"inhibitor_kills"`
			HeraldKill     float64 `json:"herald_kill"`
			GoldEarned     float64 `json:"gold_earned"`
			DragonKills    float64 `json:"dragon_kills"`
			Deaths         float64 `json:"deaths"`
			BaronKills     float64 `json:"baron_kills"`
			Assists        float64 `json:"assists"`
		} `json:"averages"`
	} `json:"stats"`
	Name     string        `json:"name"`
	ImageURL string        `json:"image_url"`
	ID       int64         `json:"id"`
	Acronym  string        `json:"acronym"`
	Players  []*PlayerInfo `json:"players"`
}

type ScoreTeamInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Team struct {
			TeamID     string `json:"teamID"`
			Name       string `json:"name"`
			EnName     string `json:"en_name"`
			ShortName  string `json:"short_name"`
			Image      string `json:"image"`
			ImageThumb string `json:"image_thumb"`
			Player     []*struct {
				PlayerID     string `json:"playerID"`
				Nickname     string `json:"nickname"`
				ImageThumb   string `json:"image_thumb"`
				PositionID   string `json:"positionID"`
				StatusID     string `json:"statusID"`
				CountryImage string `json:"country_image"`
				Heroes       []struct {
					PlayerID    string `json:"playerID"`
					HeroID      string `json:"heroID"`
					HeroName    string `json:"hero_name"`
					HeroImage   string `json:"hero_image"`
					Count       string `json:"count"`
					Win         string `json:"win"`
					VictoryRate string `json:"victory_rate"`
				} `json:"heroes"`
				PositionName string `json:"position_name"`
			} `json:"player"`
		} `json:"team"`
		Data struct {
			KDA                         string `json:"KDA"`
			Assists                     string `json:"assists"`
			Baron                       string `json:"baron"`
			BaronRate                   string `json:"baron_rate"`
			Count                       string `json:"count"`
			Deaths                      string `json:"deaths"`
			Dragon                      string `json:"dragon"`
			DragonRate                  string `json:"dragon_rate"`
			FScore                      string `json:"f_score"`
			FirstBloodKill              string `json:"firstBloodKill"`
			FirstBloodRate              string `json:"first_blood_rate"`
			AvgWardsPlaced              string `json:"avg_wardsPlaced"`
			GameTime                    string `json:"game_time"`
			Hits                        string `json:"hits"`
			Kills                       string `json:"kills"`
			Lose                        string `json:"lose"`
			Money                       string `json:"money"`
			SumBaron                    string `json:"sum_baron"`
			SumDragon                   string `json:"sum_dragon"`
			SumTower                    string `json:"sum_tower"`
			TotalDamageDealtToChampions string `json:"totalDamageDealtToChampions"`
			Tournaments                 []struct {
				TournamentID int    `json:"tournamentID"`
				ShortName    string `json:"short_name"`
			} `json:"tournaments"`
			Towers         string `json:"towers"`
			VictoryRate    string `json:"victory_rate"`
			Win            string `json:"win"`
			FirstTowerRate string `json:"first_tower_rate"`
		} `json:"data"`
	} `json:"data"`
}

// DotaTeam dota team big data.
type DotaTeam struct {
	ID           int64   `json:"id"`
	TeamID       int64   `json:"team_id"`
	Acronym      string  `json:"acronym"`
	LeidaSID     int64   `json:"leida_sid"`
	Name         string  `json:"name"`
	ImageURL     string  `json:"image_url"`
	LeidaImage   string  `json:"leida_image"`
	Win          float64 `json:"win"`
	KDA          float64 `json:"kda"`
	Kills        float64 `json:"kills"`
	Deaths       float64 `json:"deaths"`
	Assists      float64 `json:"assists"`
	TowerKills   float64 `json:"tower_kills"`
	LastHits     float64 `json:"last_hits"`
	ObserverUsed float64 `json:"observer_used"`
	SentryUsed   float64 `json:"sentry_used"`
	XpPerMinute  float64 `json:"xp_per_minute"`
	FirstBlood   float64 `json:"first_blood"`
	Heal         float64 `json:"heal"`
	GoldSpent    float64 `json:"gold_spent"`
	GoldPerMin   float64 `json:"gold_per_min"`
	Denies       float64 `json:"denies"`
	DamageTaken  float64 `json:"damage_taken"`
	CampsStacked float64 `json:"camps_stacked"`
	GamesCount   int64   `json:"games_count"`
	Players      string  `json:"players"`
	Ctime        string  `json:"ctime"`
	Mtime        string  `json:"mtime"`
}

// DotaTeamStats lol team stats.
type DotaTeamStats struct {
	Stats struct {
		GamesCount int64 `json:"games_count"`
		Averages   struct {
			XpPerMin   float64 `json:"xp_per_min"`
			TowerKills float64 `json:"tower_kills"`
			SentryUsed float64 `json:"sentry_used"`
			Ratios     struct {
				Win        float64 `json:"win"`
				FirstBlood float64 `json:"first_blood"`
			} `json:"ratios"`
			ObserverUsed float64 `json:"observer_used"`
			LastHits     float64 `json:"last_hits"`
			Kills        float64 `json:"kills"`
			Heal         float64 `json:"heal"`
			GoldSpent    float64 `json:"gold_spent"`
			GoldPerMin   float64 `json:"gold_per_min"`
			Denies       float64 `json:"denies"`
			Deaths       float64 `json:"deaths"`
			DamageTaken  float64 `json:"damage_taken"`
			CampsStacked float64 `json:"camps_stacked"`
			Assists      float64 `json:"assists"`
		} `json:"averages"`
	} `json:"stats"`
	Name     string        `json:"name"`
	ImageURL string        `json:"image_url"`
	ID       int64         `json:"id"`
	Acronym  string        `json:"acronym"`
	Players  []*PlayerInfo `json:"players"`
}

// DotaPlayer dota player big data.
type DotaPlayer struct {
	ID                  int64   `json:"id"`
	PlayerID            int64   `json:"player_id"`
	TeamID              int64   `json:"team_id"`
	TeamAcronym         string  `json:"team_acronym"`
	TeamImage           string  `json:"team_image"`
	LeidaTeamImage      string  `json:"leida_team_image"`
	LeidaSID            int64   `json:"leida_sid"`
	Name                string  `json:"name"`
	ImageURL            string  `json:"image_url"`
	LeidaImage          string  `json:"leida_image"`
	HeroesImage         string  `json:"heroes_image"`
	LeidaHeroesImage    string  `json:"leda_heroes_image"`
	Role                string  `json:"role"`
	Win                 float64 `json:"win"`
	KDA                 float64 `json:"kda"`
	Kills               float64 `json:"kills"`
	Deaths              float64 `json:"deaths"`
	Assists             float64 `json:"assists"`
	WardsPlaced         float64 `json:"wards_placed"`
	LastHits            float64 `json:"last_hits"`
	ObserverWardsPlaced float64 `json:"observer_wards_placed"`
	SentryWardsPlaced   float64 `json:"sentry_wards_placed"`
	XpPerMinute         float64 `json:"xp_per_minute"`
	GoldPerMinute       float64 `json:"gold_per_minute"`
	GamesCount          int64   `json:"games_count"`
	Ctime               string  `json:"ctime"`
	Mtime               string  `json:"mtime"`
}

// DotaPlayerStats dota player stats.
type DotaPlayerStats struct {
	Stats struct {
		Totals struct {
			MatchesWon    int64 `json:"matches_won"`
			MatchesPlayed int64 `json:"matches_played"`
			MatchesLost   int64 `json:"matches_lost"`
			GamesWon      int64 `json:"games_won"`
			GamesPlayed   int64 `json:"games_played"`
			GamesLost     int64 `json:"games_lost"`
		} `json:"totals"`
		GamesCount int64 `json:"games_count"`
		Averages   struct {
			XpPerMinute         float64 `json:"xp_per_minute"`
			WardsPlaced         float64 `json:"wards_placed"`
			SentryWardsPlaced   float64 `json:"sentry_wards_placed"`
			ObserverWardsPlaced float64 `json:"observer_wards_placed"`
			LastHits            float64 `json:"last_hits"`
			Kills               float64 `json:"kills"`
			GoldPerMinute       float64 `json:"gold_per_minute"`
			Deaths              float64 `json:"deaths"`
			Assists             float64 `json:"assists"`
		} `json:"averages"`
	} `json:"stats"`
	Role           string       `json:"role"`
	Name           string       `json:"name"`
	ImageURL       string       `json:"image_url"`
	ID             int64        `json:"id"`
	FavoriteHeroes []*FavHeroes `json:"favorite_heroes"`
	CurrentTeam    struct {
		Name     string `json:"name"`
		ImageURL string `json:"image_url"`
		ID       int64  `json:"id"`
		Acronym  string `json:"acronym"`
	} `json:"current_team"`
}

// FavHeroes dota Favorite Heroes.
type FavHeroes struct {
	Hero struct {
		Name          string `json:"name"`
		LocalizedName string `json:"localized_name"`
		ImageURL      string `json:"image_url"`
		ID            int64  `json:"id"`
	} `json:"hero"`
}

// Season season struct.
type Season struct {
	ID         int64  `json:"id"`
	LeidaSid   int64  `json:"leida_sid"`
	Stime      int64  `json:"stime"`
	Etime      int64  `json:"etime"`
	SerieType  int64  `json:"serie_type"`
	Title      string `json:"title"`
	SeasonType int64  `json:"season_type"`
}

// BaseInfo struct.
type BaseInfo struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
}

func (contest *Contest2Tab) CalculateStatus() string {
	now := time.Now().Unix()
	if now >= contest.Etime {
		return ContestStatusOfEnd
	} else if now >= contest.Stime {
		return ContestStatusOfOngoing
	}

	return ContestStatusOfNotStart
}

const (
	ContestStatusOfNotStart = "not_start"
	ContestStatusOfOngoing  = "ongoing"
	ContestStatusOfEnd      = "end"
)
