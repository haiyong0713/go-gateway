package dao

import (
	"context"
	bm "go-common/library/net/http/blademaster"

	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/api-manager/internal/model"

	"go-common/library/database/sql"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB)

// Dao dao interface
//
//go:generate mockgen -source dao.go -destination ./mock/dao.mock.go
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	AddApi(c context.Context, apis *model.ApiRawInfo) (err error)
	GetHttpApis(c context.Context) (res []*model.ApiRawInfo, err error)
	GetGrpcApis(c context.Context, discoveryID string) (res []*model.ApiRawInfo, err error)
	UpApi(c context.Context, id int64) (err error)
	GetHttpApisByPath(c context.Context, paths []string) (res map[string]*pb.ApiInfo, err error)
	GetServiceName(c context.Context, discoveryIDs []string) (res map[string][]string, err error)
	AddProto(c context.Context, pros *model.ProtoInfo) (err error)
	GetAllProtos(c context.Context) (res []*model.ProtoInfo, err error)
	GetProto(c context.Context, discoveryID string) (res []*model.ProtoInfo, err error)
	GetProtoByDis(c context.Context, discoveryIDs []string) (res map[string]*pb.ApiInfo, err error)

	GWAddPath(c context.Context, pathInfo *model.DynpathParam, discovery string) (err error)

	GroupCount(c context.Context) (int64, error)
	GroupByIDs(c context.Context, ids []int64) (res []*model.ContralGroup, err error)
	GroupByName(c context.Context, groupName string) (res []*model.ContralGroup, err error)
	GroupList(c context.Context, groupName string, pageNum, pageSize int64) (res []*model.ContralGroup, err error)
	GroupInsert(c context.Context, req *model.ContralGroup) (int64, error)
	GroupUpdate(c context.Context, req *model.ContralGroup) (int64, error)
	GroupFollowAdd(c context.Context, req *model.ContralGroupFollowActionPeq, username string) (int64, error)
	GroupFollowDel(c context.Context, req *model.ContralGroupFollowActionPeq, username string) (int64, error)
	GroupFollowList(c context.Context, uname string) (res []int64, err error)
	ApiCount(c context.Context, gid int64) (int64, error)
	ApiByIDs(c context.Context, ids []int64) ([]*model.ContralApi, error)
	ApiByName(c context.Context, apiName string) ([]*model.ContralApi, error)
	ApiList(c context.Context, gid int64, apiName string, pageNum, pageSize int64) ([]*model.ContralApi, error)
	ApiInsert(c context.Context, req *model.ContralApi) (int64, error)
	ApiUpdate(c context.Context, req *model.ContralApi) (int64, error)
	ApiConfigCount(c context.Context, id int64) (int64, error)
	ApiConfigByID(c context.Context, id int64) ([]*model.ContralApiConfig, error)
	ApiConfigByVersion(c context.Context, apiID int64, version string) ([]*model.ContralApiConfig, error)
	ApiConfigList(c context.Context, apiID, pageNum, pageSize int64) ([]*model.ContralApiConfig, error)
	ApiConfigInsert(c context.Context, req *model.ContralApiConfig) (int64, error)
	ApiPublishSave(c context.Context, req *model.ContralApiPublish) (int64, error)
	ApiPublishCount(c context.Context, id int64) (int64, error)
	ApiPublishList(c context.Context, apiID, pageNum, pageSize int64) ([]*model.ContralApiPublish, error)
}

// dao dao.
type dao struct {
	db      *sql.DB
	httpCli *bm.Client
}

// New new a dao and return.
func New(db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(db)
}

//nolint:unparam
func newDao(db *sql.DB) (d *dao, cf func(), err error) {
	d = &dao{
		db: db,
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(_ context.Context) (err error) {
	return nil
}
