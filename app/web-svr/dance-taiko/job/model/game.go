package model

const (
	DanceGame    = "dance_game"
	DancePlayers = "dance_players"

	GameJoining  = "joining"
	GamePlaying  = "playing"
	GameFinished = "finished"
)

type DatabusMsg struct {
	Action string `json:"action"`
	Table  string `json:"table"`
}

type GameDatabus struct {
	New *OttGame `json:"new"`
	Old *OttGame `json:"old"`
}

type OttGame struct {
	GameId int64  `json:"id"`
	Aid    int64  `json:"aid"`
	Cid    int64  `json:"cid"`
	Status string `json:"status"`
	Stime  int64  `json:"stime"`
}

type PlayersDatabus struct {
	New *GamePlayer `json:"new"`
	Old *GamePlayer `json:"old"`
}

type GamePlayer struct {
	Id        int64  `json:"id"`
	Mid       int64  `json:"mid"`
	GameId    int64  `json:"game_id"`
	Score     int64  `json:"score"`
	IsDeleted int8   `json:"is_deleted"`
	Ctime     string `json:"ctime"`
	Mtime     string `json:"mtime"`
}

type PlayerHonor struct {
	Mid   int64
	Score int64
}

type PlayerComment struct {
	Mid     int64
	Comment string
}

type PlayerCombo struct {
	Mid   int64
	Combo int64 // combo次数
}
