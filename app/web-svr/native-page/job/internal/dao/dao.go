package dao

import (
	"context"
	railgunv2 "go-common/library/railgun.v2"
	"go-common/library/railgun.v2/processor/single"
	"go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/job/internal/dao/binlog"
	"go-gateway/app/web-svr/native-page/job/internal/dao/dbcommon"
	"go-gateway/app/web-svr/native-page/job/internal/model"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/google/wire"
	"go-common/library/cache/credis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/sync/pipeline/fanout"
)

var Provider = wire.NewSet(New, NewDB, NewRedis)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	GetCfg() Config
	// progressDao
	GetProgressParams(c context.Context) ([]*model.ProgressParam, error)
	GetProgressParamsFromClick(c context.Context) ([]*model.ProgressParam, error)
	PushProgress(c context.Context, param *model.ProgressParam, progress int64, mids []int64, dimension model.ProgressDimension) (int64, error)
	LoadProgressParamsExtra(c context.Context)
	LoadPageRelations(c context.Context)
	GetParentPageID(pageID int64) (int64, bool)
	BatchActivityProgress(c context.Context, params []*model.ProgressParam, strictMod bool) (map[int64]*actGRPC.ActivityProgressReply, error)
	// activityDao
	GetReserveProgress(c context.Context, sid, mid, ruleID, typ, dataType int64, dimension actGRPC.GetReserveProgressDimension) (int64, error)
	ActivityProgress(c context.Context, sid, typ, mid int64, gids []int64) (*actGRPC.ActivityProgressReply, error)
	NewTopicPage(c context.Context) ([]*api.NativePage, map[string]struct{}, error)
	OfflinePage(c context.Context) ([]*api.NativePage, error)
	OnlinePage(c context.Context)
	ResetUserSpace(c context.Context, mid, pageID int64, newState string) error
	SpaceOffline(c context.Context, mid, pageID int64, tabType string) error
	UpActivityTab(c context.Context, mid int64, state int32, title string, pageID int64) (bool, error)
	// dbCommonDao
	PagingAutoAuditTsPages(c context.Context, pn, ps int64) ([]*api.NativeTsPage, error)
	AttemptPagingNatPages(c context.Context, lastID, limit int64) ([]*api.NativePage, error)
	// managerDao
	TsOnline(c context.Context, tsID, pid, auditTime int64) error
	// lockDao
	Lock(c context.Context, key string, expire int64) (string, bool, error)
	Unlock(c context.Context, key, id string) error
	// redis
	AddCacheSponsoredUp(c context.Context, mid int64) error
}

// dao dao.
type dao struct {
	cfg   Config
	db    *sql.DB
	redis credis.Redis
	cache *fanout.Fanout
	// dao
	progressDao *progressDao
	activityDao *activityDao
	binlogDao   *binlog.Dao
	dbCommonDao *dbcommon.Dao
	managerDao  *managerDao
	lockDao     *lockDao
	spaceDao    *spaceDao
	// databus
	pointRailgun  railgunv2.Consumer
	binlogRailgun railgunv2.Consumer
}

type Config struct {
	Progress        *progressCfg
	Activity        *activityCfg
	Space           *spaceCfg
	ProgressDatabus *progressDatabusCfg
	CronInterval    *cronInterval
	Binlog          *binlog.Config
	DBCommon        *dbcommon.Config
	Manager         *managerCfg
	Expire          *expire
	Mail            *HandWriteEmail
}

// HandWriteEmail 手书活动邮箱配置
type HandWriteEmail struct {
	Host    string
	Port    int
	Address string
	Pwd     string
	Name    string
	Switch  bool
	MaxNum  int
}

type progressDatabusCfg struct {
	Point        *databus.Config
	PointRailgun *railgun.SingleConfig
}

type cronInterval struct {
	PageRelationCron        string
	ProgressParamsExtraCron string
	UpAutoAuditCron         string
	UpDownCron              string
	NewPageCron             string
	BroadProgressCron       string
	BroadClickProgressCron  string
}

type expire struct {
	AutoAuditLockExpire int64
	SendMailLockExpire  int64
}

// New new a dao and return.
func New(r credis.Redis, db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(r, db)
}

func newDao(r credis.Redis, db *sql.DB) (d *dao, cf func(), err error) {
	cfg := Config{}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		cfg:         cfg,
		db:          db,
		redis:       r,
		cache:       fanout.New("cache"),
		activityDao: newActivityDao(cfg.Activity),
		spaceDao:    newSpaceDao(cfg.Space),
		managerDao:  newManagerDao(cfg.Manager),
	}
	d.dbCommonDao = dbcommon.NewDao(cfg.DBCommon, db, d.cache, d.redis)
	d.progressDao = newProgressDao(cfg.Progress, db, d.cache, d.redis, d.activityDao)
	d.binlogDao = binlog.NewDao(cfg.Binlog, d.redis)
	d.lockDao = newLockDao(d.redis)
	if err = d.startDatabus(); err != nil { //错误抛出，panic
		return
	}
	cf = d.Close
	return
}

func (d *dao) startDatabus() (err error) {
	//积分数据源进度推送
	processor := single.New(unpackPoint, d.progressDao.doPoint)
	d.pointRailgun, err = railgunv2.NewConsumer("ActPlatHistory-MainWebSvr-Prog-C", processor)
	if err != nil {
		return
	}
	//ative页binlog处理
	processorTwo := single.New(binlog.UnpackBinlog, d.binlogDao.DoBinlog)
	d.binlogRailgun, err = railgunv2.NewConsumer("Lottery-MainWebSvr-NatPage-C", processorTwo)
	if err != nil {
		// 需要关闭
		d.pointRailgun.Close()
		return
	}
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
	d.pointRailgun.Close()
	d.binlogRailgun.Close()
	d.db.Close()
	d.redis.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) GetProgressParams(c context.Context) ([]*model.ProgressParam, error) {
	return d.progressDao.GetProgressParams(c)
}

func (d *dao) GetProgressParamsFromClick(c context.Context) ([]*model.ProgressParam, error) {
	return d.progressDao.GetProgressParamsFromClick(c)
}

func (d *dao) PushProgress(c context.Context, param *model.ProgressParam, progress int64, mids []int64, dimension model.ProgressDimension) (int64, error) {
	return d.progressDao.PushProgress(c, param, progress, mids, dimension)
}

func (d *dao) GetReserveProgress(c context.Context, sid, mid, ruleID, typ, dataType int64, dimension actGRPC.GetReserveProgressDimension) (int64, error) {
	return d.activityDao.GetReserveProgress(c, sid, mid, ruleID, typ, dataType, dimension)
}

func (d *dao) ActivityProgress(c context.Context, sid, typ, mid int64, gids []int64) (*actGRPC.ActivityProgressReply, error) {
	return d.activityDao.ActivityProgress(c, sid, typ, mid, gids)
}

func (d *dao) LoadProgressParamsExtra(c context.Context) {
	d.progressDao.loadProgressParamsExtra(c)
}

func (d *dao) LoadPageRelations(c context.Context) {
	d.progressDao.loadPageRelations(c)
}

func (d *dao) GetParentPageID(pageID int64) (int64, bool) {
	return d.progressDao.GetParentPageID(pageID)
}

func (d *dao) GetCfg() Config {
	return d.cfg
}

func (d *dao) OfflinePage(c context.Context) ([]*api.NativePage, error) {
	return d.dbCommonDao.OfflinePage(c)
}

func (d *dao) NewTopicPage(c context.Context) ([]*api.NativePage, map[string]struct{}, error) {
	return d.dbCommonDao.NewTopicPage(c)
}

func (d *dao) OnlinePage(c context.Context) {
	d.dbCommonDao.OnlinePage(c)
}

func (d *dao) ResetUserSpace(c context.Context, mid, pageID int64, newState string) error {
	return d.dbCommonDao.ResetUserSpace(c, mid, pageID, newState)
}

func (d *dao) UpActivityTab(c context.Context, mid int64, state int32, title string, pageID int64) (bool, error) {
	return d.spaceDao.UpActivityTab(c, mid, state, title, pageID)
}

func (d *dao) PagingAutoAuditTsPages(c context.Context, pn, ps int64) ([]*api.NativeTsPage, error) {
	return d.dbCommonDao.PagingAutoAuditTsPages(c, pn, ps)
}

func (d *dao) TsOnline(c context.Context, tsID, pid, auditTime int64) error {
	return d.managerDao.TsOnline(c, tsID, pid, auditTime)
}

func (d *dao) SpaceOffline(c context.Context, mid, pageID int64, tabType string) error {
	return d.managerDao.SpaceOffline(c, mid, pageID, tabType)
}

func (d *dao) Lock(c context.Context, key string, expire int64) (string, bool, error) {
	return d.lockDao.Lock(c, key, expire)
}

func (d *dao) Unlock(c context.Context, key, id string) error {
	return d.lockDao.Unlock(c, key, id)
}

func (d *dao) BatchActivityProgress(c context.Context, params []*model.ProgressParam, strictMod bool) (map[int64]*actGRPC.ActivityProgressReply, error) {
	return d.progressDao.batchActivityProgress(c, params, strictMod)
}

func (d *dao) AttemptPagingNatPages(c context.Context, lastID, limit int64) ([]*api.NativePage, error) {
	return d.dbCommonDao.AttemptNewNatPages(c, lastID, limit)
}
