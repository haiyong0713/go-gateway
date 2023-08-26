package splash

import (
	"context"
	"math"
	"sort"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/manager"
)

const (
	// order：顺序展示、probability：概率展示
	_ruleOrder       = "order"
	_ruleProbability = "probability"
	_normalSize      = float64(1600) / float64(2560)
	_fullSize        = float64(1280) / float64(2560)
	_padSize         = float64(2732) / float64(2048)
	_half            = 1
	_full            = 2
)

type Brand struct {
	PullInterval    int64           `json:"pull_interval"`
	Forcibly        bool            `json:"forcibly"`
	Rule            string          `json:"rule"`
	List            []*BrandList    `json:"list,omitempty"`
	CollectionList  []*BrandList    `json:"collection_list,omitempty"`
	Show            []*BrandShow    `json:"show,omitempty"`
	Preload         []*BrandPreload `json:"preload"`
	HasNewSplashSet bool            `json:"has_new_splash_set"`
	NewSplashHash   string          `json:"new_splash_hash"`
	ForceShowTimes  int64           `json:"force_show_times"`
	ShowHash        string          `json:"show_hash"`
	BadgeFrom       string          `json:"badge_from"`
}

type BrandSetOption struct {
	Prompt         string `json:"prompt"`
	MaxSelected    int64  `json:"max_selected"`
	MaxPrompt      string `json:"max_prompt"`
	SelectedPrompt string `json:"selected_prompt"`
	SelectedText   string `json:"selected_text"`
	OverflowToast  string `json:"overflow_toast"`
	EmptyToast     string `json:"empty_toast"`
	ExitDialog     struct {
		Empty struct {
			Text string `json:"text"`
			YES  string `json:"yes"`
			NO   string `json:"no"`
		} `json:"empty"`
		Unsaved struct {
			Text string `json:"text"`
			YES  string `json:"yes"`
			NO   string `json:"no"`
		} `json:"selected"`
	} `json:"exit_dialog"`
	HasBadge         bool   `json:"has_badge"`
	BadgeHash        string `json:"badge_hash"`
	BottomSaveButton struct {
		Text         string `json:"text"`
		SuccessToast string `json:"success_toast"`
	} `json:"bottom_save_button"`
}

type BrandPreload struct {
	Show           []*BrandShow `json:"show,omitempty"`
	PullInterval   int64        `json:"pull_interval"`
	Forcibly       bool         `json:"forcibly"`
	Rule           string       `json:"rule"`
	Start          xtime.Time   `json:"begin_time,omitempty"`
	End            xtime.Time   `json:"end_time,omitempty"`
	ForceShowTimes int64        `json:"force_show_times"`
	ShowHash       string       `json:"show_hash"`
}

type BrandList struct {
	ID      int64  `json:"id"`
	Image   string `json:"thumb"`
	LogoURL string `json:"logo_url"`
	Mode    string `json:"mode"`
}

type BrandShow struct {
	ID              int64                     `json:"id"`
	Start           xtime.Time                `json:"begin_time,omitempty"`
	End             xtime.Time                `json:"end_time,omitempty"`
	Probability     int                       `json:"probability"`
	Duration        int64                     `json:"duration"`
	Position        int                       `json:"-"`
	Mode            string                    `json:"mode"`
	ShowLogo        bool                      `json:"show_logo"`
	LogoPosition    string                    `json:"logo_position"`
	NewSplash       bool                      `json:"new_splash"`
	Categories      []*manager.SplashCategory `json:"categories"`
	InitialPushTime int64                     `json:"initial_push_time"`
}

type BrandSet struct {
	Show                []*BrandShow              `json:"show,omitempty"`
	CollectionShow      []*BrandShow              `json:"collection_show,omitempty"`
	Config              []*ConfigItem             `json:"config,omitempty"`
	Desc                string                    `json:"desc,omitempty"`
	AllCategories       []*manager.SplashCategory `json:"all_categories,omitempty"`
	BrandSetOption      *BrandSetOption           `json:"brand_set_option"`
	ShowTitle           string                    `json:"show_title,omitempty"`
	CollectionShowTitle string                    `json:"collection_show_title,omitempty"`
}

type ConfigItem struct {
	Title     string `json:"title,omitempty"`
	Type      string `json:"type,omitempty"`
	MainTitle string `json:"main_title,omitempty"`
	Subtitle  string `json:"subtitle,omitempty"`
}

type SplashParam struct {
	MobiApp      string `form:"mobi_app"`
	Platform     string `form:"platform"`
	Device       string `form:"device"`
	Build        int    `form:"build"`
	Network      string `form:"network"`
	ScreenWidth  int64  `form:"screen_width"`
	ScreenHeight int64  `form:"screen_height"`
	LastReadAt   int64  `form:"last_read_at"`
	Mid          int64
}

type SplashSaveParam struct {
	ID                 []int64 `form:"id,split"`
	MobiApp            string  `form:"mobi_app"`
	Platform           string  `form:"platform"`
	Device             string  `form:"device"`
	Build              int     `form:"build"`
	Network            string  `form:"network"`
	CollectionSplashID []int64 `form:"collection_splash_id,split"`
}

func BrandListChange(c context.Context, config *conf.Feature, ms map[string]*manager.ImgInfo, logoURL string, param *SplashParam) []*BrandList {
	removeFull := model.SplashRemoveFull(c, config, param.MobiApp, param.Build)
	var list []*BrandList
	for _, v := range ms {
		if removeFull && v.Mode == _full { // 6.10之前版本过滤全屏
			continue
		}
		l := &BrandList{
			ID:      v.ID,
			Image:   BuildImgURL(v, param),
			LogoURL: BuildLogoURL(v, logoURL),
			Mode:    SwitchMode(v.Mode),
		}
		list = append(list, l)
	}
	return list
}

func BuildLogoURL(img *manager.ImgInfo, url string) string {
	if img.LogoConfig.ImgURL != "" {
		return img.LogoConfig.ImgURL
	}
	return url
}

func BuildImgURL(img *manager.ImgInfo, param *SplashParam) string {
	switch img.Mode {
	case _half: // 半屏
		return img.ImgURL
	case _full: // 全屏
		imgMap := map[string]string{
			"normal": img.FullScreenImgURL.Normal,
			"full":   img.FullScreenImgURL.Full,
			"pad":    img.FullScreenImgURL.Pad,
		}
		if param.ScreenWidth == 0 || param.ScreenHeight == 0 {
			return img.FullScreenImgURL.Normal
		}
		size := float64(param.ScreenWidth) / float64(param.ScreenHeight)
		rankScreen := []BrandScreen{
			{Delta: math.Abs(_normalSize - size), Screen: "normal"},
			{Delta: math.Abs(_fullSize - size), Screen: "full"},
			{Delta: math.Abs(_padSize - size), Screen: "pad"},
		}
		sort.Slice(rankScreen, func(i, j int) bool {
			return rankScreen[i].Delta < rankScreen[j].Delta
		})
		return imgMap[rankScreen[0].Screen]
	default:
		log.Error("Unrecognized mode: %d", img.Mode)
		return img.ImgURL
	}
}

type BrandScreen struct {
	Delta  float64 `json:"delta"`
	Screen string  `json:"screen"`
}

func BrandListShowChange(c context.Context, config *conf.Feature, cfg *manager.SplashConfig, imgMap map[int64]*manager.ImgInfo, brandConf *conf.BrandSplash, param *SplashParam, now time.Time) ([]*BrandShow, bool, string, int64) {
	var (
		list     []*BrandShow
		rule     string
		forcibly bool
	)
	if cfg == nil {
		return list, forcibly, rule, 0
	}
	// 1：强制-顺序
	// 2：强制-概率
	// 3：默认-顺序
	// 4：默认-概率
	// 5：用户自选
	//nolint:gomnd
	switch cfg.ShowMode {
	case 1:
		forcibly = true
		rule = _ruleOrder
	case 2:
		rule = _ruleProbability
		forcibly = true
	case 3:
		rule = _ruleOrder
	case 4:
		rule = _ruleProbability
	default:
		return list, forcibly, rule, 0
	}
	// list 这里暂时不需要输出分类
	list = showChange(c, config, cfg, imgMap, brandConf, forcibly, param, now, nil)
	return list, forcibly, rule, cfg.ForceShowTimes
}

func BrandSetShowChange(c context.Context, config *conf.Feature, cfg *manager.SplashConfig, imgMap map[int64]*manager.ImgInfo, brandConf *conf.BrandSplash, param *SplashParam, now time.Time, categories []*manager.SplashCategory) []*BrandShow {
	// 1：强制-顺序
	// 2：强制-概率
	// 3：默认-顺序
	// 4：默认-概率
	// 5：用户自选
	if cfg == nil || cfg.ShowMode != 5 {
		return nil
	}
	//nolint:gosimple
	var list []*BrandShow
	list = showChange(c, config, cfg, imgMap, brandConf, false, param, now, categories)
	return list
}

func asCategoryMap(in []*manager.SplashCategory) map[int64]*manager.SplashCategory {
	out := map[int64]*manager.SplashCategory{}
	for _, v := range in {
		out[v.ID] = v
	}
	return out
}

func showChange(c context.Context, config *conf.Feature, cfg *manager.SplashConfig, imgMap map[int64]*manager.ImgInfo, brandConf *conf.BrandSplash, forcibly bool, param *SplashParam, now time.Time, categories []*manager.SplashCategory) []*BrandShow {
	if cfg == nil {
		return nil
	}
	var (
		list []*BrandShow
	)
	categoryMap := asCategoryMap(categories)
	removeFull := model.SplashRemoveFull(c, config, param.MobiApp, param.Build)
	for _, v := range cfg.Config {
		img, ok := imgMap[v.ImgID]
		if !ok {
			continue
		}
		if removeFull && img.Mode == _full { // 6.10之前版本过滤全屏
			continue
		}
		s := &BrandShow{
			ID:              v.ImgID,
			Start:           cfg.Stime,
			End:             cfg.Etime,
			Probability:     v.Rate,
			Duration:        brandConf.Duration,
			Position:        v.Position,
			Mode:            SwitchMode(img.Mode),
			ShowLogo:        img.LogoConfig.Show,
			LogoPosition:    "center",
			NewSplash:       img.IsNew(now),
			InitialPushTime: img.InitialPushTime,
		}
		switch s.Mode {
		case "half":
			if brandConf.SplitDuration.HalfDuration > 0 {
				s.Duration = brandConf.SplitDuration.HalfDuration
			}
		case "full":
			if brandConf.SplitDuration.FullDuration > 0 {
				s.Duration = brandConf.SplitDuration.FullDuration
			}
		}
		for _, categoryID := range img.CategoryIDs {
			cg, ok := categoryMap[categoryID]
			if !ok {
				continue
			}
			s.Categories = append(s.Categories, &manager.SplashCategory{
				ID:   cg.ID,
				Name: cg.Name,
				Sort: cg.Sort,
			})
		}
		list = append(list, s)
	}
	// 顺序的时候排序
	if !forcibly {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Position < list[j].Position
		})
	}
	// 把用户没见过的新图放前面
	// 初始状态提交 -1，老版本提交为 0
	if param.LastReadAt != 0 {
		head := []*BrandShow{}
		tail := []*BrandShow{}
		for _, b := range list {
			if !b.NewSplash {
				tail = append(tail, b)
				continue
			}
			if b.InitialPushTime > param.LastReadAt {
				head = append(head, b)
				continue
			}
			b.NewSplash = false // 同样认为不是新闪屏
			tail = append(tail, b)
		}
		list = append(head, tail...)
	}
	return list
}

func SwitchMode(mode int64) string {
	switch mode {
	case _half:
		return "half"
	case _full:
		return "full"
	default:
		log.Error("Unrecognized mode: %d", mode)
		return "half"
	}
}
