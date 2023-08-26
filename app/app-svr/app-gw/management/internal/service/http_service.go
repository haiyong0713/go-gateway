package service

import (
	"go-gateway/app/app-svr/app-gw/management/internal/dao"
)

// HttpService service.
type HttpService struct {
	*SnapshotService
	common      *CommonService
	dao         dao.Dao
	resourceDao dao.ResourceDao
}

func newHTTP(d dao.Dao) *HttpService {
	s := &HttpService{
		SnapshotService: newSnapshot(d.CreateSnapshotDao()),
		common:          newCommon(d),
		dao:             d,
		resourceDao:     d.CreateHTTPResourceDao(),
	}
	return s
}
