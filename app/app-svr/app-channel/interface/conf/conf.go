package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/log/infoc"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	xtime "go-common/library/time"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/BurntSushi/toml"
)

var (
	WardenSDKBuilder *wardensdk.InterceptorBuilder
	Conf             = &Config{}
)

type Config struct {
	// Env
	Env string
	// show  XLog
	Log *log.Config
	// tick time
	Tick xtime.Duration
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpClinetAsyn
	HTTPClientAsyn *bm.ClientConfig
	// HTTPShopping
	HTTPShopping *bm.ClientConfig
	// bm http
	BM *HTTPServers
	// host
	Host *Host
	// db
	MySQL *MySQL
	// rpc client
	ArchiveRPC *rpc.ClientConfig
	// relationRPC
	RelationRPC *rpc.ClientConfig
	// rpc client
	TagRPC *rpc.ClientConfig
	// rpc Article
	ArticleRPC *rpc.ClientConfig
	// Infoc2
	FeedInfoc2    *infoc.Config
	ChannelInfoc2 *infoc.Config
	SquareInfoc2  *infoc.Config
	// memcache
	Memcache *Memcache
	// BroadcastRPC grpc
	PGCRPC *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	// gprc Article
	ArticleGRPC *warden.ClientConfig
	// Square Count
	SquareCount int
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	ThumbupGRPC  *warden.ClientConfig
	ChannelGRPC  *warden.ClientConfig
	// beijixing
	NewChannelCardShowInfoc *infoc.Config
	// locationGRPC
	LocationClient *warden.ClientConfig
	// build limt
	BuildLimit *BuildLimit
	// fav grpc
	FavoriteGRPC     *warden.ClientConfig
	CoinGRPC         *warden.ClientConfig
	TagGRPC          *warden.ClientConfig
	ResourceGRPC     *warden.ClientConfig
	DynamicTopicGRPC *warden.ClientConfig
	NewTopicGRPC     *warden.ClientConfig
	DynGRPC          *warden.ClientConfig
	CfcGRPC          *warden.ClientConfig
	BaikeGRPC        *warden.ClientConfig
	// share
	Share *Share
	// switch
	Switch *Switch
	// square model
	Square *Square
	// cron
	Cron *Cron
	// nat grpc
	NatClient *warden.ClientConfig
	// PR limit
	PRLimit *PRLimit
	// feature平台
	Feature *Feature
	// content.flow.control.service 配置
	CfcSvrConfig *CfcSvrConfig
}

// 由服务方定义
type CfcSvrConfig struct {
	BusinessID int64
	Secret     string
	Source     string
}

type Host struct {
	Bangumi  string
	Data     string
	APICo    string
	Activity string
	LiveAPI  string
	Shopping string
	VcCo     string
}

type HTTPServers struct {
	Outer *bm.ServerConfig
}

type MySQL struct {
	Show    *sql.Config
	Manager *sql.Config
}

type Memcache struct {
	Channels *struct {
		*memcache.Config
		Expire xtime.Duration
	}
}

type BuildLimit struct {
	MiaokaiIOS           int
	MiaokaiAndroid       int
	TabSimilarIOS        int
	TabSimilarAndroid    int
	NoSquareFeedIOS      int
	NoSquareFeedAndroid  int
	MineNewSubIOS        int
	MineNewSubAndroid    int
	ArcWithPlayerAndroid int
	ArcWithPlayerIOS     int
	OGVChanIOSBuild      int64
	OGVChanAndroidBuild  int64
}

type Share struct {
	Items   *ShareItem
	JumpURI string
}

type ShareItem struct {
	Weibo         bool
	Wechat        bool
	WechatMonment bool
	QQ            bool
	QZone         bool
	Copy          bool
	More          bool
}

type Switch struct {
	DetailVerify bool
	ListOGVMore  bool
	ListOGVFold  bool
	SquareActive bool
	MineActive   bool
}

type Square struct {
	Models []string
}

type Cron struct {
	LoadMenusCacheV2      string
	LoadAuditCache        string
	LoadRegionlist        string
	LoadCardCache         string
	LoadConvergeCache     string
	LoadSpecialCache      string
	LoadLiveCardCache     string
	LoadGameDownloadCache string
	LoadCardSetCache      string
	LoadMenusCache        string
}

type PRLimit struct {
	ChannelList []int64
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch               bool
	ChannelListFirstItem string
	ChannelIndex         string
	TabSimila            string // del
	NewSub               string // del
	NoFeed               string // del
	Miaokai              string // del
	TcTranslateRequired  string // del
}

func InitWardenSDKBuilder(sdkBuilderConfig wardensdk.SDKBuilderConfig) {
	WardenSDKBuilder = wardensdk.NewBuilder(sdkBuilderConfig)
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
