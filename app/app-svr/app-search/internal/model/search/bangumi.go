package search

import (
	"encoding/json"

	"go-gateway/app/app-svr/app-search/internal/model"
)

// Season for bangumi.
type Season struct {
	AllowDownload string          `json:"allow_download,omitempty"`
	SeasonID      string          `json:"season_id"`
	IsJump        int             `json:"is_jump"`
	EpisodeStatus int             `json:"episode_status"`
	Title         string          `json:"title"`
	Cover         string          `json:"cover"`
	IsFinish      string          `json:"is_finish"`
	IsStarted     int             `json:"is_started"`
	NewestEpID    string          `json:"newest_ep_id"`
	NewestEpIndex string          `json:"newest_ep_index"`
	TotalCount    string          `json:"total_count"`
	Weekday       string          `json:"weekday"`
	Evaluate      string          `json:"evaluate"`
	Bp            json.RawMessage `json:"rank,omitempty"`
	UserSeason    *struct {
		Attention string `json:"attention"`
	} `json:"user_season,omitempty"`
}

// Recommend for bangumi.
type Recommend struct {
	Aid    int64  `json:"aid"`
	Cover  string `json:"cover"`
	Status int    `json:"status"`
	Title  string `json:"title"`
}

// Card for bangumi.
type Card struct {
	SeasonID       int64                `json:"season_id"`
	SeasonType     int                  `json:"season_type"`
	IsFollow       int                  `json:"is_follow"`
	IsSelection    int                  `json:"is_selection"`
	Episodes       []*Episode           `json:"episodes"`
	SeasonTypeName string               `json:"season_type_name"`
	Badges         []*model.ReasonStyle `json:"badges"`
	URL            string               `json:"url"`
}

// Episode for bangumi card.
type Episode struct {
	ID         int64                `json:"id"`
	Status     int                  `json:"status"`
	Cover      string               `json:"cover"`
	Index      string               `json:"index"`
	IndexTitle string               `json:"index_title"`
	Badges     []*model.ReasonStyle `json:"badges"`
	URL        string               `json:"url"`
}
