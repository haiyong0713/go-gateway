package model

import (
	"encoding/json"
	"sync"
)

// LolGame lol and dota game.
type LolGame struct {
	Game
	Teams   json.RawMessage `json:"teams"`
	Players json.RawMessage `json:"players"`
}

// OwGame overwatch game.
type OwGame struct {
	Game
	WinTeam int64           `json:"win_team"`
	Teams   json.RawMessage `json:"teams"`
	Map     json.RawMessage `json:"map"`
}

// Game common game.
type Game struct {
	ID       int64  `json:"id"`
	GameID   int64  `json:"game_id"`
	Position int64  `json:"position"`
	MatchID  int64  `json:"match_id"`
	BeginAt  string `json:"begin_at"`
	EndAt    string `json:"end_at"`
	Finished int64  `json:"finished"`
}

// LdInfo .
type LdInfo struct {
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	ID       int64  `json:"id"`
}

// Header .
type Header struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// LolPlayer lol player big data.
type LolPlayer struct {
	ID             int64           `json:"id"`
	PlayerID       int64           `json:"player_id"`
	TeamID         int64           `json:"team_id"`
	TeamAcronym    string          `json:"team_acronym"`
	TeamImage      string          `json:"team_image"`
	LeidaSID       int64           `json:"leida_sid"`
	Name           string          `json:"name"`
	ImageURL       string          `json:"image_url"`
	ChampionsImage json.RawMessage `json:"champions_image"`
	Role           string          `json:"role"`
	Win            float64         `json:"win"`
	KDA            float64         `json:"kda"`
	Kills          float64         `json:"kills"`
	Deaths         float64         `json:"deaths"`
	Assists        float64         `json:"assists"`
	MinionsKilled  float64         `json:"minions_killed"`
	WardsPlaced    float64         `json:"wards_placed"`
	GamesCount     int64           `json:"games_count"`
	RoleName       string          `json:"role_name"`
	Ctime          string          `json:"ctime"`
	Mtime          string          `json:"mtime"`
	PositionID     int64           `json:"position_id"`
	Position       string          `json:"position"`
	MVP            int64           `json:"mvp"`
}

// LolTeam lol team big data.
type LolTeam struct {
	ID                 int64           `json:"id"`
	TeamID             int64           `json:"team_id"`
	Acronym            string          `json:"acronym"`
	LeidaSID           int64           `json:"leida_sid"`
	Name               string          `json:"name"`
	ImageURL           string          `json:"image_url"`
	Win                float64         `json:"win"`
	KDA                float64         `json:"kda"`
	Kills              float64         `json:"kills"`
	Deaths             float64         `json:"deaths"`
	Assists            float64         `json:"assists"`
	TowerKills         float64         `json:"tower_kills"`
	TotalMinionsKilled float64         `json:"total_minions_killed"`
	FirstTower         float64         `json:"first_tower"`
	FirstInhibitor     float64         `json:"first_inhibitor"`
	FirstDragon        float64         `json:"first_dragon"`
	FirstBaron         float64         `json:"first_baron"`
	FirstBlood         float64         `json:"first_blood"`
	WardsPlaced        float64         `json:"wards_placed"`
	InhibitorKills     float64         `json:"inhibitor_kills"`
	BaronKills         float64         `json:"baron_kills"`
	GoldEarned         float64         `json:"gold_earned"`
	GamesCount         int64           `json:"games_count"`
	Players            json.RawMessage `json:"players"`
	Ctime              string          `json:"ctime"`
	Mtime              string          `json:"mtime"`
	BaronRate          float64         `json:"baron_rate"`
	DragonRate         float64         `json:"dragon_rate"`
	Hits               float64         `json:"hits"`
	LoseNum            int64           `json:"lose_num"`
	Money              float64         `json:"money"`
	TotalDamage        float64         `json:"total_damage"`
	WinNum             int64           `json:"win_num"`
	ImageThumb         string          `json:"image_thumb"`
	NewData            int64           `json:"new_data"`
}

// DotaPlayer dota player big data.
type DotaPlayer struct {
	ID                  int64           `json:"id"`
	PlayerID            int64           `json:"player_id"`
	TeamID              int64           `json:"team_id"`
	TeamAcronym         string          `json:"team_acronym"`
	TeamImage           string          `json:"team_image"`
	LeidaSID            int64           `json:"leida_sid"`
	Name                string          `json:"name"`
	ImageURL            string          `json:"image_url"`
	HeroesImage         json.RawMessage `json:"heroes_image"`
	Role                string          `json:"role"`
	Win                 float64         `json:"win"`
	KDA                 float64         `json:"kda"`
	Kills               float64         `json:"kills"`
	Deaths              float64         `json:"deaths"`
	Assists             float64         `json:"assists"`
	WardsPlaced         float64         `json:"wards_placed"`
	LastHits            float64         `json:"last_hits"`
	ObserverWardsPlaced float64         `json:"observer_wards_placed"`
	SentryWardsPlaced   float64         `json:"sentry_wards_placed"`
	XpPerMinute         float64         `json:"xp_per_minute"`
	GoldPerMinute       float64         `json:"gold_per_minute"`
	GamesCount          int64           `json:"games_count"`
	RoleName            string          `json:"role_name"`
	Ctime               string          `json:"ctime"`
	Mtime               string          `json:"mtime"`
}

// DotaTeam dota team big data.
type DotaTeam struct {
	ID           int64           `json:"id"`
	TeamID       int64           `json:"team_id"`
	Acronym      string          `json:"acronym"`
	LeidaSID     int64           `json:"leida_sid"`
	Name         string          `json:"name"`
	ImageURL     string          `json:"image_url"`
	Win          float64         `json:"win"`
	KDA          float64         `json:"kda"`
	Kills        float64         `json:"kills"`
	Deaths       float64         `json:"deaths"`
	Assists      float64         `json:"assists"`
	TowerKills   float64         `json:"tower_kills"`
	LastHits     float64         `json:"last_hits"`
	ObserverUsed float64         `json:"observer_used"`
	SentryUsed   float64         `json:"sentry_used"`
	XpPerMinute  float64         `json:"xp_per_minute"`
	FirstBlood   float64         `json:"first_blood"`
	Heal         float64         `json:"heal"`
	GoldSpent    float64         `json:"gold_spent"`
	GoldPerMin   float64         `json:"gold_per_min"`
	Denies       float64         `json:"denies"`
	DamageTaken  float64         `json:"damage_taken"`
	CampsStacked float64         `json:"camps_stacked"`
	GamesCount   int64           `json:"games_count"`
	Players      json.RawMessage `json:"players"`
	Ctime        string          `json:"ctime"`
	Mtime        string          `json:"mtime"`
}

// LdTeam leida team data.
type LdTeam struct {
	ID       int64  `json:"id"`
	TeamID   int64  `json:"team_id"`
	Acronym  string `json:"acronym"`
	LeidaSID int64  `json:"leida_sid"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	GameType int64  `json:"game_type"`
}

// SpecialTeam special team.
type SpecialTeam struct {
	Team   interface{} `json:"team"`
	Stats  interface{} `json:"stats"`
	Recent []*Contest  `json:"recent"`
	GID    int64       `json:"gid"`
}

// ActiveLive .
type ActiveLive struct {
	MaID   int64  `json:"ma_id"`
	LiveID int64  `json:"live_id"`
	Title  string `json:"title"`
}

// SyncGame store leida game list
type SyncGame struct {
	Data map[int64][]*LolGame
	sync.Mutex
}

// SyncOwGame store leida ow game list
type SyncOwGame struct {
	Data map[int64][]*OwGame
	sync.Mutex
}

// SyncItem store item list
type SyncItem struct {
	Data map[int64]*LdInfo
	sync.Mutex
}

// SyncInfo store leida base info list
type SyncInfo struct {
	Data map[int64]*LdInfo
	sync.Mutex
}

// SyncLolPlayers store leida lol players.
type SyncLolPlayers struct {
	Data map[int64][]*LolPlayer
	sync.Mutex
}

// SyncLolTeams store leida lol teams.
type SyncLolTeams struct {
	Data map[int64][]*LolTeam
	sync.Mutex
}

// SyncDotaPlayers store leida dota players.
type SyncDotaPlayers struct {
	Data map[int64][]*DotaPlayer
	sync.Mutex
}

// SyncDotaTeams store leida dota teams.
type SyncDotaTeams struct {
	Data map[int64][]*DotaTeam
	sync.Mutex
}

// SyncSeasonGame season game type.
type SyncSeasonGame struct {
	Data map[int64]int64
	sync.Mutex
}

type PlayerDataRank struct {
	ID         int64  `json:"id"`
	PlayerID   int64  `json:"player_id"`
	PlayerName string `json:"player_name"`
	ImageURL   string `json:"image_url"`
	TeamID     int64  `json:"team_id"`
	TeamName   string `json:"team_name"`
	PositionID int64  `json:"position_id"`
	Position   string `json:"position"`
}

type PlayerDataMvpRank struct {
	*PlayerDataRank
	Mvp  int64
	Rank int `json:"rank"`
}

type PlayerDataKdaRank struct {
	*PlayerDataRank
	Kda float64
}
