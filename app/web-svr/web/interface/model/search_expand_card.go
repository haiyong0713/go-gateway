package model

import (
	"go-gateway/app/web-svr/web/interface/model/search"

	esportConfGRPC "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	esportGRPC "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
)

const (
	_tipsCover = "https://i0.hdslb.com/bfs/app/f1342cd72a4a042535cc8daa6c7e8c17eb8eb419.png"
)

type SearchESportsCard struct {
	ConfigInfo *esportConfGRPC.EsportConfigInfo `json:"config_info"`
	Contest    []*esportGRPC.ContestDetail      `json:"contest"`
}

type SearchGameCard struct {
	GameName    string        `json:"game_name"`
	GameIcon    string        `json:"game_icon"`
	Summary     string        `json:"summary"`
	GameStatus  int           `json:"game_status"`
	GameLink    string        `json:"game_link"`
	Grade       float64       `json:"grade"`
	BookNum     int           `json:"book_num"`
	DownloadNum int           `json:"download_num"`
	CommentNum  int           `json:"comment_num"`
	Platform    string        `json:"platform"`
	MediaScores []*mediaScore `json:"media_scores,omitempty"`
}

type SearchTipCard struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	HasBgImg int    `json:"has_bg_img"`
	JumpUrl  string `json:"jump_url"`
	Cover    string `json:"cover"`
}

func (c *SearchTipCard) FromTip(d *TipDetail) {
	c.ID = d.ID
	c.Title = d.Title
	c.SubTitle = d.SubTitle
	if d.HasBgImg == 1 {
		c.Cover = _tipsCover
	}
	c.HasBgImg = d.HasBgImg
	c.JumpUrl = d.JumpUrl
}

type SearchBiliUserCard struct {
	*search.SearchUser
	Expand *SearchBiliUserCardExpand `json:"expand"`
}

type SearchBiliUserCardExpand struct {
	IsPowerUp    bool          `json:"is_power_up"`
	SystemNotice *SystemNotice `json:"system_notice"`
}

type SearchUserCard struct {
	*search.SearchUser
	Expand *SearchBiliUserCardExpand `json:"expand"`
}
