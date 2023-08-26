package dao

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	databusv2 "go-common/library/queue/databus.v2"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	activityapi "go-gateway/app/web-svr/activity/interface/api"
	espclient "go-gateway/app/web-svr/esports/interface/api/v1"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/component"
	"go-gateway/app/web-svr/esports/service/conf"
	"go-gateway/app/web-svr/esports/service/internal/model"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	favClient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"

	"github.com/jinzhu/gorm"
)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	// GRPC
	// 直播
	LiveRoomInfo(ctx context.Context, roomIds []int64) (roomInfos map[int64]*liveRoom.Infos, err error)
	// 初始化人群包
	InitBGroup(ctx context.Context, contest *model.ContestModel) (err error)
	// 天马卡
	// 初始化事件
	InitTunnelEvent(ctx context.Context, contest *model.ContestModel) (err error)
	// 创建卡片
	UpsertTunnelCard(ctx context.Context, contest *model.ContestModel) (err error)
	// 竞猜
	GetGuessDetail(ctx context.Context, contestIds []int64, mid int64) (guessMap map[int64]bool, err error)
	// 订阅
	// 添加订阅
	AddFav(ctx context.Context, contestId int64, mid int64) (err error)
	// 取消订阅
	DelFav(ctx context.Context, contestId int64, mid int64) (err error)
	// 推送订阅相关的消息
	BGroupDataBusPub(ctx context.Context, mid, contestID, state int64) (err error)
	// 游标翻页获取赛程的订阅人群
	GetSubscriberByContestId(ctx context.Context, contestId int64, cursor int64, cursorSize int32) (res *v1.ContestSubscribers, err error)
	// 获取用户对赛程的订阅关系
	GetSubscribeRelationByContests(ctx context.Context, contestIds []int64, mid int64) (relations map[int64]bool, err error)

	// HTTP
	// ES
	GetContestsCacheOrEs(ctx context.Context, contestsQueryParams *model.ContestsQueryParamsModel) (contestIds []int64, total int, cache bool, err error)

	// Redis
	RedisLock(ctx context.Context, key string, value string, ttl int64, retry int, internalMillSeconds int64) (err error)
	RedisUnLock(ctx context.Context, key string, value string) (err error)
	RedisUniqueValue() string
	GetSeasonInfoCache(ctx context.Context, seasonId int64) (seasonInfo *model.SeasonModel, err error)
	GetSeasonsInfoCache(ctx context.Context, seasonId []int64) (seasonsInfo map[int64]*model.SeasonModel, missIds []int64, err error)
	SetSeasonsInfoCache(ctx context.Context, seasonModels []*model.SeasonModel) (err error)
	SetSeasonInfoCache(ctx context.Context, seasonInfo *model.SeasonModel) (err error)
	GetGamesCache(ctx context.Context, gameIds []int64) (gamesInfoMap map[int64]*model.GameModel, missIds []int64, err error)
	SetGamesCache(ctx context.Context, gamesInfoMap map[int64]*model.GameModel) (err error)
	GetMatchCache(ctx context.Context, matchId int64) (matchModel *model.MatchModel, err error)
	SetMatchCache(ctx context.Context, matchModel model.MatchModel) (err error)
	GetTeamsCache(ctx context.Context, teamIds []int64) (teamsInfoMap map[int64]*model.TeamModel, missIds []int64, err error)
	SetTeamsCache(ctx context.Context, teamInfoMap map[int64]*model.TeamModel) (err error)
	GetContestCache(ctx context.Context, contestId int64) (contestModel *model.ContestModel, err error)
	GetContestsCache(ctx context.Context, contestIds []int64) (contestModelMap map[int64]*model.ContestModel, missIds []int64, err error)
	SetContestCache(ctx context.Context, contestModels []*model.ContestModel) (err error)
	DeleteContestCache(ctx context.Context, contestIds int64) (err error)
	DeleteSeasonCache(ctx context.Context, seasonId int64) (err error)
	DeleteMatchCache(ctx context.Context, matchId int64) (err error)
	DeleteTeamCache(ctx context.Context, teamId int64) (err error)
	DeleteGameCache(ctx context.Context, gameId int64) (err error)
	GetAllGamesCache(ctx context.Context) (gamesInfoMap map[int64]*model.GameModel, err error)
	GetSeriesCacheById(ctx context.Context, seriesId int64) (seriesModel *model.ContestSeriesModel, err error)
	SetSeriesCacheById(ctx context.Context, seriesModel *model.ContestSeriesModel) (err error)
	GetSeriesCacheByIds(ctx context.Context, seriesId []int64) (seriesModel map[int64]*model.ContestSeriesModel, missIds []int64, err error)
	SetSeriesCacheByIds(ctx context.Context, seriesModel []*model.ContestSeriesModel) (err error)
	GetSeasonSeriesListCache(ctx context.Context, seasonId int64) (seriesModels []*model.ContestSeriesModel, err error)
	SetSeasonSeriesListCache(ctx context.Context, seriesModels []*model.ContestSeriesModel) (err error)

	// MC
	GetActiveSeasonsCache(ctx context.Context) (seasonIds []int64, err error)
	StoreActiveSeasonsCache(ctx context.Context, seasonsMap map[int64]*model.SeasonModel) (err error)
	GetSeasonContestIdsCache(ctx context.Context, seasonId int64) (contestIds []int64, err error)
	StoreSeasonContestIdsCache(ctx context.Context, seasonId int64, contestIds []int64) (err error)
	GetSeasonTeamsCache(ctx context.Context, seasonId int64) (seasonTeams []*model.SeasonTeamModel, err error)
	StoreSeasonTeamsCache(ctx context.Context, seasonTeams []*model.SeasonTeamModel, seasonId int64) (err error)

	// DB
	ContestAddTransaction(ctx context.Context, contest *model.ContestModel, gameIds []int64, teamIds []int64, contestData []*model.ContestDataModel, adId int64) (err error)
	ContestUpdateTransaction(ctx context.Context, contest *model.ContestModel, gameIds []int64, teamIds []int64, contestData []*model.ContestDataModel) (err error)
	ContestContestStatusUpdate(ctx context.Context, contestId int64, contestStatus int64) (err error)
	GetSeasonsBySETime(ctx context.Context, startTime int64, endTime int64) (seasons []*model.SeasonModel, err error)
	GetSeasonContestIds(ctx context.Context, seasonId int64) (contestIds []int64, err error)
	GetContestById(ctx context.Context, contestId int64, valid bool) (contestModel *model.ContestModel, err error)
	GetContestGameById(ctx context.Context, contestId int64) (gameModel *model.GameModel, err error)
	GetContestsByIds(ctx context.Context, contestIds []int64, valid bool) (contestModels []*model.ContestModel, err error)
	GetSeasonByID(ctx context.Context, id int64) (season *model.SeasonModel, err error)
	GetSeasonsByIDs(ctx context.Context, ids []int64) (season []*model.SeasonModel, err error)
	GetAllGames(ctx context.Context) (gameModels []*model.GameModel, err error)
	GetGamesByIds(ctx context.Context, gameIds []int64) (gameModels []*model.GameModel, err error)
	GetMatchModel(ctx context.Context, matchId int64) (matchModel *model.MatchModel, err error)
	GetTeamsByIds(ctx context.Context, teamIds []int64) (teamModels []*model.TeamModel, err error)
	GetSeriesById(ctx context.Context, seriesId int64) (contestSeriesModel *model.ContestSeriesModel, err error)
	GetSeriesByIds(ctx context.Context, seriesIds []int64) (contestSeriesModel map[int64]*model.ContestSeriesModel, err error)
	GetSeriesBySeasonId(ctx context.Context, seasonId int64) (contestSeriesModels []*model.ContestSeriesModel, err error)
	GetSeasonTeamsModel(ctx context.Context, seasonId int64) (seasonTeams []*model.SeasonTeamModel, err error)
	RawReplyWall() (list []*model.ReplyWallModel, err error)
	ReplyWallList(ctx context.Context) (res []*model.ReplyWallModel, err error)
	DelCacheReplyWallList(ctx context.Context) (err error)
	ReplyWallUpdateTransaction(ctx context.Context, req *v1.SaveReplyWallModel) (err error)

	GetAccountInfos(ctx context.Context, mids []int64) (userInfoMap *accapi.InfosReply, err error)
	GetDistinctSeasonByTime(ctx context.Context, beginTime int64, endTime int64) (seasonIds []int64, err error)
}

const (
	_replyReg = "/x/internal/v2/reply/subject/regist"
)

// dao dao.
type dao struct {
	conf  *conf.Config
	db    *sql.DB
	orm   *gorm.DB
	redis *redis.Redis
	mc    *memcache.Memcache
	cache *fanout.Fanout
	// grpc client .
	liveRoomClient liveRoom.RoomClient
	favoriteClient favClient.FavoriteClient
	espClient      espclient.EsportsClient
	activityClient activityapi.ActivityClient
	accountClient  accapi.AccountClient
	replyClient    *bm.Client
	replyURL       string
	// http client .
	http                 *bm.Client
	sysInformsHTTPClient *bm.Client
	bGroupClient         bGroup.BGroupServiceClient
	tunnelV2Client       tunnelV2.TunnelClient
	demoExpire           int32
	elastic              *elastic.Elastic
	DataBusV2Client      databusv2.Client
	BGroupMessagePub     databusv2.Producer
}

// New new a dao and return.
func New(conf *conf.Config) (d Dao) {
	return newDao(conf)
}

func newDao(conf *conf.Config) (d *dao) {
	var cfg struct {
		DemoExpire xtime.Duration
	}

	d = &dao{
		conf:                 conf,
		db:                   component.GlobalDB,
		orm:                  component.GlobalOrm,
		redis:                component.GlobalRedis,
		mc:                   component.GlobalMemcached,
		cache:                fanout.New("cache"),
		demoExpire:           int32(time.Duration(cfg.DemoExpire) / time.Second),
		http:                 bm.NewClient(conf.HTTPClient),
		sysInformsHTTPClient: bm.NewClient(conf.SysInformsHTTPClient),
		replyClient:          bm.NewClient(conf.HTTPReply),
		replyURL:             conf.Host.APICo + _replyReg,
		elastic:              elastic.NewElastic(conf.Elastic),
	}
	initGrpcClients(d)
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
