package dao

import (
	"context"
	"github.com/google/wire"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/elastic"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
	"go-gateway/app/app-svr/app-gw/management/internal/model/tree"
)

var Provider = wire.NewSet(New, NewKV)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) error
	FetchRoleTree(ctx context.Context, username, cookie string) ([]*tree.Node, error)
	GetAuthUsers() *AuthUsers
	SetGateway(ctx context.Context, req *pb.SetGatewayReq) error
	ListGateway(ctx context.Context, node string) ([]*pb.Gateway, error)
	DeleteGateway(ctx context.Context, req *pb.DeleteGatewayReq) error
	EnableALLGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) error
	ListLog(ctx context.Context, group, name string, object, pn, ps int64) (*model.ListLogReply, error)
	TryLock(ctx context.Context, key string, oldVal, newVal []byte, ttl uint32) error
	TokenSecret(ctx context.Context) ([]byte, error)
	InitialTokenSecret(ctx context.Context) error
	ProxyPage(ctx context.Context, host, suffix string) (*pb.GatewayProxyReply, error)
	GatewayProfile(ctx context.Context, host string, isGRPC bool) (*model.GatewayProfile, error)
	AddGatewayConfigFile(ctx context.Context, req *model.AddConfigFileReq) error
	CreateGatewayConfigBuild(ctx context.Context, req *model.CreateConfigBuildReq) error
	FetchConfig(ctx context.Context, id int64, cookie string) ([]*model.ConfigBuildItem, error)
	ServerMetadata(ctx context.Context, host string) ([]string, error)
	CreateDeploymentMeta(ctx context.Context, meta *pb.DeploymentMeta) error
	SetDeploymentMeta(ctx context.Context, meta *pb.DeploymentMeta) error
	UpdateDeploymentState(ctx context.Context, src, dst *pb.DeploymentMeta) error
	SetDeploymentConfirm(ctx context.Context, req *pb.DeploymentReq, confirm *pb.DeploymentConfirm) error
	GetDeploymentMeta(ctx context.Context, req *pb.DeploymentReq) (*pb.DeploymentMeta, error)
	GetDeploymentConfirm(ctx context.Context, req *pb.DeploymentReq) (*pb.DeploymentConfirm, error)
	DeploymentIsConfirmed(ctx context.Context, req *pb.DeploymentReq) (bool, error)
	GetDeploymentActionLog(ctx context.Context, req *pb.DeploymentReq) ([]*pb.ActionLog, error)
	ReloadConfig(ctx context.Context, req *model.ReloadConfigReq) (*model.ReloadConfigReply, error)
	AddActionLog(ctx context.Context, req *pb.AddActionLogReq)
	ListDeployment(ctx context.Context, req *pb.ListDeploymentReq) ([]*pb.DeploymentMeta, error)
	BatchSetBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.SetBreakerAPIReq, dpReq []*pb.SetDynPathReq) error
	BatchDelBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.DeleteBreakerAPIReq, dpReq []*pb.DeleteDynPathReq) error
	GetPlugin(ctx context.Context, pluginName, field string) (*pb.Plugin, error)
	SetupPlugin(ctx context.Context, pluginName, field string, value *pb.Plugin) error
	QuotaResources(ctx context.Context, id, token string) ([]*pb.Limiter, error)
	AddQuotaResources(ctx context.Context, req *pb.Limiter, token string) error
	UpdateQuotaResources(ctx context.Context, req *pb.Limiter, token string) error
	DeleteQuotaResources(ctx context.Context, id, token string) error
	PluginList(ctx context.Context, req *pb.PluginListReq) ([]*pb.PluginListItem, error)
	GRPCServerMethods(ctx context.Context, appIDs []string) (map[string][]string, bool)
	GRPCServerPackages(ctx context.Context, appIDs []string) (map[string]map[string][]string, bool)
	EnableAllGRPCGatewayConfig(ctx context.Context, req *pb.UpdateALLGatewayConfigReq) error

	// snapshot dao
	CreateSnapshotDao() SnapshotDao
	CreateHTTPResourceDao() ResourceDao
	CreateGRPCResourceDao() ResourceDao
}

type ResourceDao interface {
	ListBreakerAPI(ctx context.Context, node string, last string) ([]*pb.BreakerAPI, error)
	SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq) error
	EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq) error
	DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq) error
	ListDynPath(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error)
	SetDynPath(ctx context.Context, req *pb.SetDynPathReq) error
	DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq) error
	EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq) error
	GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error)
	SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error
}

type SnapshotDao interface {
	ListBreakerAPI(ctx context.Context, node string, last string, uuid string) ([]*pb.BreakerAPI, error)
	SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) error
	EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) error
	DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) error
	ListDynPath(ctx context.Context, node string, gateway string, uuid string) ([]*pb.DynPath, error)
	SetDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) error
	DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) error
	EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) error
	AddSnapshot(ctx context.Context, req *pb.AddSnapshotReq) (*pb.AddSnapshotReply, error)
	GetSnapshotMeta(ctx context.Context, node string, gateway string, uuid string) (*pb.SnapshotMeta, error)
	BuildPlan(ctx context.Context, node, gateway, uuid string) (*pb.SnapshotRunPlan, error)
	RunPlan(ctx context.Context, req *pb.SnapshotRunPlan) error
	CreateSnapshotGRPCDao() SnapshotGRPCDao

	GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error)
	SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error
}

type SnapshotGRPCDao interface {
	ListBreakerAPI(ctx context.Context, node string, last string, uuid string) ([]*pb.BreakerAPI, error)
	SetBreakerAPI(ctx context.Context, req *pb.SetBreakerAPIReq, uuid string) error
	EnableBreakerAPI(ctx context.Context, req *pb.EnableBreakerAPIReq, uuid string) error
	DeleteBreakerAPI(ctx context.Context, req *pb.DeleteBreakerAPIReq, uuid string) error
	ListDynPath(ctx context.Context, node string, gateway string, uuid string) ([]*pb.DynPath, error)
	SetDynPath(ctx context.Context, req *pb.SetDynPathReq, uuid string) error
	DeleteDynPath(ctx context.Context, req *pb.DeleteDynPathReq, uuid string) error
	EnableDynPath(ctx context.Context, req *pb.EnableDynPathReq, uuid string) error
	BuildPlan(ctx context.Context, node, gateway, uuid string) (*pb.SnapshotRunPlan, error)
	RunPlan(ctx context.Context, req *pb.SnapshotRunPlan) error
	BatchSetBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.SetBreakerAPIReq, dpReq []*pb.SetDynPathReq) error
	BatchDelBreakerAPIAndDynPath(ctx context.Context, bapiReq []*pb.DeleteBreakerAPIReq, dpReq []*pb.DeleteDynPathReq) error

	GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error)
	SetQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	DeleteQuotaMethod(ctx context.Context, req *pb.QuotaMethod) error
	EnableQuotaMethod(ctx context.Context, req *pb.EnableLimiterReq) error
}

type httpResourceDao struct {
	dao *resourceDao
}

type grpcResourceDao struct {
	dao *resourceDao
}

type resourceDao struct {
	taishan *Taishan
}

type snapshotDao struct {
	dao     *dao // delegate to origin
	grpcDao *snapshotGRPCDao
}

type snapshotGRPCDao struct {
	dao *dao
}

// dao dao.
type dao struct {
	taishan *Taishan
	http    *bm.Client
	es      *elastic.Elastic
	// http host
	Hosts struct {
		Config string
		Easyst string
		ApiCo  string
	}
	sideBarUsers *SideBarUsers
	resource     *resourceDao
}

// New new a dao and return.
func New(taishan *Taishan) (d Dao, cf func(), err error) {
	return newDao(taishan)
}

func newDao(taishan *Taishan) (d *dao, cf func(), err error) {
	resolver := resolver.New(nil, discovery.Builder())
	var cfg struct {
		HTTPClient *bm.ClientConfig
		Elastic    *elastic.Config
	}
	if err = paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		taishan:      taishan,
		http:         bm.NewClient(cfg.HTTPClient, bm.SetResolver(resolver)),
		es:           elastic.NewElastic(cfg.Elastic),
		sideBarUsers: newSideBarUsers(),
		resource:     newResourceDao(taishan),
	}
	ac := &paladin.TOML{}
	if err = paladin.Watch("application.toml", ac); err != nil {
		return
	}
	if err := ac.Get("hosts").UnmarshalTOML(&d.Hosts); err != nil {
		panic(err)
	}
	if err := paladin.Watch("auth.toml", d.sideBarUsers); err != nil {
		panic(err)
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) CreateSnapshotDao() SnapshotDao {
	grpcDao := &snapshotGRPCDao{dao: d}
	snapshotDao := &snapshotDao{
		dao:     d,
		grpcDao: grpcDao,
	}
	return snapshotDao
}

func (d *dao) CreateHTTPResourceDao() ResourceDao {
	return &httpResourceDao{dao: d.resource}
}

func (d *dao) CreateGRPCResourceDao() ResourceDao {
	return &grpcResourceDao{dao: d.resource}
}

func newResourceDao(taishan *Taishan) *resourceDao {
	return &resourceDao{taishan: taishan}
}
