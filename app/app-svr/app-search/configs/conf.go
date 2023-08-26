package configs

import (
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/container/pool"
	"go-common/library/log"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	MemcacheDegradeConfig = &DegradeConfig{
		Expire: 54000,
		Memcache: &memcache.Config{
			Config: &pool.Config{
				Active: 5000,
				Idle:   1000,
			},
			Name:         "app-interface/search/degrade",
			Proto:        "tcp",
			Addr:         "127.0.0.1:11280",
			DialTimeout:  xtime.Duration(100 * time.Millisecond),
			ReadTimeout:  xtime.Duration(200 * time.Millisecond),
			WriteTimeout: xtime.Duration(200 * time.Millisecond),
		},
		NewSearchMemcacheYlf: &memcache.Config{
			Config: &pool.Config{
				Active:      5000,
				Idle:        1000,
				IdleTimeout: xtime.Duration(80 * time.Millisecond),
			},
			Name:         "shylf_main_app_interface_mc",
			Proto:        "tcp",
			Addr:         "127.0.0.1:27922",
			DialTimeout:  xtime.Duration(150 * time.Millisecond),
			ReadTimeout:  xtime.Duration(200 * time.Millisecond),
			WriteTimeout: xtime.Duration(200 * time.Millisecond),
		},
		NewSearchMemcacheJd: &memcache.Config{
			Config: &pool.Config{
				Active:      5000,
				Idle:        1000,
				IdleTimeout: xtime.Duration(80 * time.Millisecond),
			},
			Name:         "shjd_main_app_interface_mc",
			Proto:        "tcp",
			Addr:         "127.0.0.1:28019",
			DialTimeout:  xtime.Duration(150 * time.Millisecond),
			ReadTimeout:  xtime.Duration(200 * time.Millisecond),
			WriteTimeout: xtime.Duration(200 * time.Millisecond),
		},
	}
)

// Config struct
type Config struct {
	Search              *Search
	SearchBuildLimit    *SearchBuildLimit
	BuildLimit          *BuildLimit
	SearchDynamicSwitch *SearchDynamicSwitch
	Cfg                 *Cfg
	Switch              *Switch
	PlayerBuildLimit    map[string]int
	SearchPageTitle     *SearchPageTitle
	Custom              *Custom
	Cron                *Cron
	// 下发资源
	Resource *Resource
	// feature配置
	Feature *Feature
	// 天马搜索导航位配置
	SearchRcmdTagsConfig *SearchRcmdTagsConfig
}

type SearchRcmdTagsConfig struct {
	CloseRcmdTagsSwitch bool
	AiRcmdTimeout       string
}

// Custom is
type Custom struct {
	RecommendTimeout xtime.Duration
}

// Cfg def.
type Cfg struct {
	PgcSearchCard *PgcSearchCard
}

// PgcSearchCard def.
type PgcSearchCard struct {
	Epsize            int
	IpadEpSize        int
	IpadCheckMoreSize int
	OfflineWatch      string
	OnlineWatch       string
	CheckMoreContent  string
	CheckMoreSchema   string
	EpLabel           string
	// 宫格样式是否出角标
	GridBadge bool
}

// BuildLimit is
type BuildLimit struct {
	OGVChanIOSBuild     int64
	OGVChanAndroidBuild int64
}

// Search struct
type Search struct {
	SeasonNum             int
	MovieNum              int
	SeasonMore            int
	MovieMore             int
	UpUserNum             int
	UVLimit               int
	UserNum               int
	UserVideoLimit        int
	UserVideoLimitMix     int
	BiliUserNum           int
	BiliUserVideoLimit    int
	BiliUserVideoLimitMix int
	OperationNum          int
	IPadSearchBangumi     int
	IPadSearchFt          int
	TrendingLimit         int
	EggCloseCount         int
	BackgroundSwitch      bool
	LiveFaceSwitch        bool
	SearchRankingSwitch   bool
	SpaceEntrance         *SpaceEntrance
}

type SpaceEntrance struct {
	TextMore        string
	TextMoreWithNum string
	TextColor       string
	TextColorNight  string
}

// SearchBuildLimit struct
type SearchBuildLimit struct {
	PGCHighLightIOS          int
	PGCHighLightAndroid      int
	PGCALLIOS                int
	PGCALLAndroid            int
	SpecialerGuideIOS        int
	SpecialerGuideAndroid    int
	SearchArticleIOS         int
	SearchArticleAndroid     int
	ComicIOS                 int
	ComicAndroid             int
	ChannelIOS               int
	ChannelAndroid           int
	CooperationIOS           int
	CooperationAndroid       int
	CooperationIPadHD        int
	QueryCorIOS              int
	QueryCorAndroid          int
	SugDetailIOS             int
	SugDetailAndroid         int
	NewTwitterIOS            int
	NewTwitterAndroid        int
	NewOrderIOS              int
	NewOrderAndroid          int
	DefaultWordJumpIOS       int
	DefaultWordJumpAndroid   int
	DefaultWordJumpAndroidI  int
	NewChannelIOS            int
	NewChannelAndroid        int
	ESportsIOS               int
	ESportsAndroid           int
	VideoDurationIOS         int
	VideoDurationAndroid     int
	OGVURLAndroid            int
	OGVURLIOS                int
	SpecialCardIOS           int
	SpecialCardAndroid       int
	UpNewAndroid             int
	UpNewIOS                 int
	CardOptimizeAndroid      int
	CardOptimizeIPhone       int
	CardOptimizeIpadHD       int
	TipsCardIOS              int
	TipsCardAndroid          int
	ADCardIOS                int
	ADCardAndroid            int
	UserInlineLiveIOS        int
	UserInlineLiveAndroid    int
	FlowInlineCardIOS        int
	FlowInlineCardAndroid    int
	FlowOGVInlineCardIOS     int
	FlowOGVInlineCardAndroid int

	// type search
	TypeSearchWithPlayURLIOS     int
	TypeSearchWithPlayURLAndroid int
	TypeSearchChannelESIOS       int
	TypeSearchChannelESAndroid   int
}

// SearchDynamicSwitch .
type SearchDynamicSwitch struct {
	IsUP    bool
	IsCount bool
}

// Switch func switch.
type Switch struct {
	SearchRecommend bool
	SearchSuggest   bool
	SearchMainRcmd  bool
	// 商品店铺开关
	AdOpen bool
	// 大航海开关
	GuardOpen bool
	// 皮肤装扮开关
	SkinOpen bool
	// 空间投稿全部
	SpaceContributeAll bool
	// 搜索三点
	SearchThreePoint bool
	// 秒开新参数开启
	PlayerArgs bool
}

// DegradeConfig struct.
type DegradeConfig struct {
	Expire               int32
	Memcache             *memcache.Config
	NewSearchMemcacheYlf *memcache.Config
	NewSearchMemcacheJd  *memcache.Config
}

type SearchPageTitle struct {
	HistoryTitle string
	FindTitle    string
}

type Cron struct {
	LoadSidebar         string
	LoadBlacklist       string
	LoadHotCache        string
	LoadSearchTipsCache string
	LoadSpecialCache    string
	LoadUpRcmdBlockList string
	LoadSystemNotice    string
}

type Resource struct {
	SearchThreePoint *SearchThreePoint
}

type SearchThreePoint struct {
	WaitIcon   string
	WaitTitle  string
	ShareIcon  string
	ShareTitle string
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch         bool
	ShowLive       string
	SearchParamOGV string
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
