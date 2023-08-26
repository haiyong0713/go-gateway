package dao

import (
	"fmt"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/space/interface/conf"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	memberclient "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	dynamicFeed "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynamicSearchgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/search"
	liveusergrpc "git.bilibili.co/bapis/bapis-go/live/xuserex/v1"
	napagerpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgcappcard "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	spacegrpc "git.bilibili.co/bapis/bapis-go/space/service/v1"
	seasonGRPC "git.bilibili.co/bapis/bapis-go/ugc-season/service"

	"go-common/library/database/hbase.v2"
)

// Dao dao struct.
type Dao struct {
	// config
	c *conf.Config
	// db
	db *sql.DB
	// hbase
	hbase *hbase.Client
	// stmt
	channelStmt       []*sql.Stmt
	channelListStmt   []*sql.Stmt
	channelCntStmt    []*sql.Stmt
	channelArcCntStmt []*sql.Stmt
	// redis
	redis *redis.Pool
	// http client
	httpR       *bm.Client
	httpW       *bm.Client
	httpGame    *bm.Client
	httpDynamic *bm.Client
	// bfs client
	bfsClient *bfs.BFS
	// api URL
	bangumiConcernURL   string
	bangumiUnConcernURL string
	favArcURL           string
	favAlbumURL         string
	shopURL             string
	shopLinkURL         string
	albumCountURL       string
	albumListURL        string
	tagSubURL           string
	tagCancelSubURL     string
	tagSubListURL       string
	accTagsURL          string
	accTagsSetURL       string
	isAnsweredURL       string
	lastPlayGameURL     string
	appPlayedGameURL    string
	webTopPhotoURL      string
	topPhotoURL         string
	setTopPhotoURL      string
	liveURL             string
	groupsCountURL      string
	audioCardURL        string
	audioUpperCertURL   string
	audioCntURL         string
	dynamicListURL      string
	dynamicURL          string
	dynamicCntURL       string
	dynamicInfoURL      string
	creativeViewDataURL string
	// expire
	clExpire               int32
	settingExpire          int32
	noticeExpire           int32
	topArcExpire           int32
	mpExpire               int32
	themeExpire            int32
	topDyExpire            int32
	redisOfficialExpire    int32
	redisTopPhotoArcExpire int32
	// UserTab
	redisUserTabExpire   int32
	redisWhitelistExpire int32
	redisMinExpire       int
	redisMaxExpire       int
	// cache
	cache *fanout.Fanout
	rand  *rand.Rand
	// mid info
	midClient midrpc.AccountClient
	// native page
	naPageClient        napagerpc.NaPageClient
	seasonClient        seasonGRPC.UGCSeasonClient
	accGRPC             accgrpc.AccountClient
	relGRPC             relationgrpc.RelationClient
	liveUserGRPC        liveusergrpc.LabsClient
	dynamicSearchClient dynamicSearchgrpc.DynamicSearchServiceClient
	memberGRPC          memberclient.MemberClient
	pgcAppCardClient    pgcappcard.AppCardClient
	spaceClient         spacegrpc.SpaceClient
	galleryClient       gallerygrpc.GalleryServiceClient
	dynamicFeedClient   dynamicFeed.FeedClient
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:                   c,
		db:                  sql.NewMySQL(c.Mysql),
		hbase:               hbase.NewClient(c.HBase.Config),
		redis:               redis.NewPool(c.Redis.Config),
		httpR:               bm.NewClient(c.HTTPClient.Read),
		httpW:               bm.NewClient(c.HTTPClient.Write),
		httpGame:            bm.NewClient(c.HTTPClient.Game),
		httpDynamic:         bm.NewClient(c.HTTPClient.Dynamic),
		bfsClient:           bfs.New(nil),
		bangumiConcernURL:   c.Host.Bangumi + _bangumiConcernURI,
		bangumiUnConcernURL: c.Host.Bangumi + _bangumiUnConcernURI,
		favArcURL:           c.Host.API + _favArchiveURI,
		favAlbumURL:         c.Host.APILive + _favAlbumURI,
		shopURL:             c.Host.Mall + _shopURI,
		shopLinkURL:         c.Host.Mall + _shopLinkURI,
		albumCountURL:       c.Host.LinkDraw + _albumCountURI,
		albumListURL:        c.Host.LinkDraw + _albumListURI,
		tagSubURL:           c.Host.API + _tagSubURI,
		tagCancelSubURL:     c.Host.API + _tagCancelSubURI,
		tagSubListURL:       c.Host.API + _subTagListURI,
		accTagsURL:          c.Host.Acc + _accTagsURI,
		accTagsSetURL:       c.Host.Acc + _accTagsSetURI,
		isAnsweredURL:       c.Host.API + _isAnsweredURI,
		lastPlayGameURL:     c.Host.Game + _lastPlayGameURI,
		appPlayedGameURL:    c.Host.AppGame + _appPlayedGameURI,
		webTopPhotoURL:      c.Host.Space + _webTopPhotoURI,
		topPhotoURL:         c.Host.Space + _topPhotoURI,
		setTopPhotoURL:      c.Host.Space + _setTopPhotoURI,
		liveURL:             c.Host.APILive + _liveURI,
		groupsCountURL:      c.Host.APIVc + _groupsCountURI,
		audioCardURL:        c.Host.API + _audioCardURI,
		audioUpperCertURL:   c.Host.API + _audioUpperCertURI,
		audioCntURL:         c.Host.API + _audioCntURI,
		dynamicListURL:      c.Host.APIVc + _dynamicListURI,
		dynamicURL:          c.Host.APIVc + _dynamicURI,
		dynamicCntURL:       c.Host.APIVc + _dynamicCntURI,
		dynamicInfoURL:      c.Host.Dynamic + _dynamicInfoURI,
		creativeViewDataURL: c.Host.API + _creativeViewDataURI,
		// expire
		clExpire:               int32(time.Duration(c.Redis.ClExpire) / time.Second),
		settingExpire:          int32(time.Duration(c.Redis.SettingExpire) / time.Second),
		noticeExpire:           int32(time.Duration(c.Redis.NoticeExpire) / time.Second),
		topArcExpire:           int32(time.Duration(c.Redis.TopArcExpire) / time.Second),
		mpExpire:               int32(time.Duration(c.Redis.MpExpire) / time.Second),
		themeExpire:            int32(time.Duration(c.Redis.ThemeExpire) / time.Second),
		topDyExpire:            int32(time.Duration(c.Redis.TopDyExpire) / time.Second),
		redisOfficialExpire:    int32(time.Duration(c.Redis.OfficialExpire) / time.Second),
		redisTopPhotoArcExpire: int32(time.Duration(c.Redis.TopPhotoArcExpire) / time.Second),
		redisUserTabExpire:     int32(time.Duration(c.Redis.UserTabExpire) / time.Second),
		redisWhitelistExpire:   int32(time.Duration(c.Redis.WhitelistExpire) / time.Second),
		redisMinExpire:         int(time.Duration(c.Redis.MinExpire) / time.Second),
		redisMaxExpire:         int(time.Duration(c.Redis.MaxExpire) / time.Second),
		// cache
		cache: fanout.New("cache"),
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.channelStmt = make([]*sql.Stmt, _chSub)
	d.channelListStmt = make([]*sql.Stmt, _chSub)
	d.channelCntStmt = make([]*sql.Stmt, _chSub)
	d.channelArcCntStmt = make([]*sql.Stmt, _chSub)
	for i := 0; i < _chSub; i++ {
		d.channelStmt[i] = d.db.Prepared(fmt.Sprintf(_chSQL, i))
		d.channelListStmt[i] = d.db.Prepared(fmt.Sprintf(_chListSQL, i))
		d.channelCntStmt[i] = d.db.Prepared(fmt.Sprintf(_chCntSQL, i))
		d.channelArcCntStmt[i] = d.db.Prepared(fmt.Sprintf(_chArcCntSQL, i))
	}
	var err error
	if d.midClient, err = midrpc.NewClient(c.MidGRPC); err != nil {
		panic(fmt.Sprintf("mid NewClient err(%v)", err))
	}
	if d.naPageClient, err = napagerpc.NewClient(c.NaPageRPC); err != nil {
		panic(fmt.Sprintf("naPage NewClient err(%v)", err))
	}
	if d.seasonClient, err = seasonGRPC.NewClient(c.UGCSeasonGRPC); err != nil {
		panic(fmt.Sprintf("ugc-season NewClient err(%v)", err))
	}
	if d.accGRPC, err = accgrpc.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClient error (%+v)", err))
	}
	if d.liveUserGRPC, err = liveusergrpc.NewClient(c.LiveUserGRPC); err != nil {
		panic(fmt.Sprintf("liveusergrpc NewClient error (%+v)", err))
	}
	if d.dynamicSearchClient, err = dynamicSearchgrpc.NewClient(c.DynamicSearchGRPC); err != nil {
		panic(fmt.Sprintf("dynamicsearchgrpc NewClient error (%+v)", err))
	}
	if d.memberGRPC, err = memberclient.NewClient(c.MemberClient); err != nil {
		panic(fmt.Sprintf("memberclient NewClient error (%+v)", err))
	}
	if d.pgcAppCardClient, err = pgcappcard.NewClient(c.PgcCardClient); err != nil {
		panic(fmt.Sprintf("pgcappcard NewClient error (%+v)", err))
	}
	if d.spaceClient, err = spacegrpc.NewClient(c.SpaceGRPC); err != nil {
		panic(fmt.Sprintf("spacegrpc NewClientSpace error (%+v)", err))
	}
	if d.galleryClient, err = gallerygrpc.NewClient(c.GalleryGRPC); err != nil {
		panic(fmt.Sprintf("gallerygrpc NewClient error (%+v)", err))
	}
	if d.dynamicFeedClient, err = dynamicFeed.NewClient(c.DynamicFeedGRPC); err != nil {
		panic(fmt.Sprintf("dynamicFeedClient NewClient error (%+v)", err))
	}
	return
}
