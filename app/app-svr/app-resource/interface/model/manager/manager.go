package manager

import (
	"encoding/json"
	"time"

	xtime "go-common/library/time"
)

type SplashList struct {
	ImgMap                map[string]*ImgInfo `json:"img_map"`
	ImgMapV2              map[string]*ImgInfo `json:"-"`
	DefaultConfig         *SplashConfig       `json:"default_config"`
	SelectConfig          *SplashConfig       `json:"select_config"`
	PrepareDefaultConfigs []*SplashConfig     `json:"prepare_default_configs"`
	PrepareSelectConfigs  []*SplashConfig     `json:"prepare_select_configs"`
	BaseDefaultConfig     *SplashConfig       `json:"base_default_config"`
	Categories            []*SplashCategory   `json:"categories"`
}

type CollectionSplashList struct {
	ImgMap map[int64]*ImgInfo `json:"img_map"`
}

type SplashCategory struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Sort  int64  `json:"sort"`
	Count int64  `json:"count,omitempty"`
}

type FullScreenImgURL struct {
	Normal string `json:"normal"`
	Full   string `json:"full"`
	Pad    string `json:"pad"`
}

type ImgInfo struct {
	ID               int64            `json:"id"`
	ImgName          string           `json:"img_name"`
	ImgURL           string           `json:"img_url"`
	Mode             int64            `json:"mode"`
	LogoConfig       LogoConfig       `json:"logo_config"`
	FullScreenImgURL FullScreenImgURL `json:"full_screen_img_url"`
	InitialPushTime  int64            `json:"initial_push_time"`
	KeepNewDays      int64            `json:"keep_new_days"`
	CategoryIDs      []int64          `json:"category_ids"`
}

func (i *ImgInfo) IsNew(at time.Time) bool {
	if i.KeepNewDays <= 0 {
		return false
	}
	if i.InitialPushTime <= 0 {
		return false
	}
	//nolint:gomnd
	asNewDur := (time.Hour * 24) * time.Duration(i.KeepNewDays)
	return at.Sub(time.Unix(i.InitialPushTime, 0)) <= asNewDur
}

type LogoConfig struct {
	Show   bool   `json:"show"`
	Mode   int64  `json:"mode"`
	ImgURL string `json:"img_url"`
}

type SplashConfig struct {
	Stime          xtime.Time    `json:"stime"`
	Etime          xtime.Time    `json:"etime"`
	ShowMode       int           `json:"show_mode"`
	ForceShowTimes int64         `json:"force_show_times"`
	ConfigStr      string        `json:"config_json"`
	Config         []*ConfigJSON `json:"-"`
}

type ConfigJSON struct {
	ImgID    int64 `json:"img_id"`
	Position int   `json:"position"`
	Rate     int   `json:"rate"`
	Sort     int64 `json:"sort"`
}

func (s *SplashConfig) SplashConfigChange() error {
	var tmp []*ConfigJSON
	if err := json.Unmarshal([]byte(s.ConfigStr), &tmp); err != nil {
		return err
	}
	s.Config = tmp
	return nil
}
