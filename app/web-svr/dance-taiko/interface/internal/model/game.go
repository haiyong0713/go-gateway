package model

import accClient "git.bilibili.co/bapis/bapis-go/account/service"

// id,game_id,aid,status
type Game struct {
	ID     int64
	GameID int64
	AID    int64
	Status string
	Stime  int64 // 开始时间，毫秒
}

const (
	GameJoining  = "joining"
	GamePlaying  = "playing"
	GameFinished = "finished"
)

type Player struct {
	Mid  int64  `json:"mid"`
	Face string `json:"face"`
	Name string `json:"name"`
}

func (p *Player) CopyFromGRPC(gp *accClient.Card) {
	p.Mid = gp.Mid
	p.Face = gp.Face
	p.Name = gp.Name
}

type OttGame struct {
	GameId int64  `json:"id"`
	Aid    int64  `json:"aid"`
	Cid    int64  `json:"cid"`
	Status string `json:"status"`
	Stime  int64  `json:"stime"`
}

type GameCreateReply struct {
	GameId int64 `json:"game_id"`
}

type GameStatusReply struct {
	GameStatus   string        `json:"game_status"`
	PlayerStatus []*PlayerInfo `json:"player_status"`
}

type PlayerInfo struct {
	*Player
	LastComment string `json:"last_comment"`
	Points      int    `json:"points"`
	ComboTimes  int    `json:"combo_times"`
	GlobalRank  int    `json:"global_rank"`
}

type QRCodeReply struct {
	QRCode string `json:"qrcode"`
	Msg    string `json:"msg"`
}

type GameJoinReply struct {
	ServerTime int64 `json:"server_time"`
}

type BwsPlayInfo struct {
	Mid    int64
	Valid  bool
	Energy int
	Star   int
}

type BwsPlayResult struct {
	Mid   int64 `json:"mid"`
	Score int64 `json:"score"`
	Star  int   `json:"star"`
}
