package bangumi

import "go-gateway/app/app-svr/archive/service/api"

type Season struct {
	Aid         int64  `json:"aid,omitempty"`
	SeasonID    int64  `json:"season_id,omitempty"`
	EpisodeID   string `json:"episode_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Cover       string `json:"cover,omitempty"`
	PlayCount   int32  `json:"play_count,omitempty"`
	Favorites   int32  `json:"favorites,omitempty"`
	SeasonType  int8   `json:"season_type,omitempty"`
	TypeBadge   string `json:"type_badge,omitempty"`
	SeasonCover string `json:"season_cover,omitempty"`
	UpdateDesc  string `json:"update_desc,omitempty"`
}

type Update struct {
	SquareCover string `json:"square_cover"`
	Title       string `json:"title"`
	Updates     int    `json:"updates"`
}

type Moe struct {
	ID     int64  `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
	Cover  string `json:"cover,omitempty"`
	Link   string `json:"link,omitempty"`
	Desc   string `json:"desc,omitempty"`
	Badge  string `json:"badge,omitempty"`
	Square string `json:"square,omitempty"`
}

type Remind struct {
	Updates int           `json:"updates"`
	List    []*RemindItem `json:"list"`
}

type RemindItem struct {
	Cover       string `json:"cover"`
	SquareCover string `json:"square_cover"`
	UpdateDesc  string `json:"update_desc"`
	UpdateTitle string `json:"update_title"`
	Uri         string `json:"uri"`
	SeasonId    int64  `json:"season_id"`
	Epid        int64  `json:"epid"`
}

type EpPlayer struct {
	AID        int64             `json:"aid"`
	CID        int64             `json:"cid"`
	EpID       int64             `json:"episode_id"`
	Uri        string            `json:"url"`
	Cover      string            `json:"cover"`
	PlayerInfo *api.BvcVideoItem `json:"player_info"`
	NewDesc    string            `json:"new_desc"`
	IsPreview  int32             `json:"is_preview"`
	Duration   int64             `json:"duration"`
	RegionURI  string            `json:"region_uri"`
	Stat       *struct {
		Play    int64 `json:"play"`
		Reply   int64 `json:"reply"`
		Danmaku int64 `json:"danmaku"`
	}
	Season *struct {
		SeasonID int64  `json:"season_id"`
		Type     int32  `json:"type"`
		Cover    string `json:"cover"`
		Title    string `json:"title"`
		TypeName string `json:"type_name"`
	} `json:"season"`
}
