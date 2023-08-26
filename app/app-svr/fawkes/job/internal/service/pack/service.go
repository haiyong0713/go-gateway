package pack

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-common/library/railgun"

	"go-common/library/conf/paladin.v2"

	"github.com/google/wire"

	"go-gateway/app/app-svr/fawkes/job/internal/dao"
	"go-gateway/app/app-svr/fawkes/job/internal/model"
	"go-gateway/app/app-svr/fawkes/job/internal/model/pack"
)

var Provider = wire.NewSet(New)

// Service struct.
type Service struct {
	ac          *paladin.Map
	out         *model.OutCfg
	jobCfg      *model.ClearNasJob
	dao         dao.Dao
	closeChan   chan struct{}
	packRailgun *railgun.Railgun
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:        &paladin.TOML{},
		out:       &model.OutCfg{},
		jobCfg:    &model.ClearNasJob{},
		closeChan: make(chan struct{}),
		dao:       d,
	}

	if err = paladin.Watch("job.toml", s.ac); err != nil {
		panic(err)
	}
	oc := &paladin.TOML{}
	if err = paladin.Watch("out.toml", oc); err != nil {
		panic(err)
	}
	if err = s.ac.Get("clearNas").UnmarshalTOML(&s.jobCfg); err != nil {
		return
	}
	if err = oc.Get("Out").UnmarshalTOML(&s.out); err != nil {
		return
	}
	s.initCleanRailgun()
	cf = s.Close
	return
}

// Close close the resource.
func (s *Service) Close() {
	close(s.closeChan)
	s.packRailgun.Close()
	log.Warn("fawkes-job pack close exit")
}

func (s *Service) CleanHttp(ctx *bm.Context) {
	var r = &pack.ClearPackReq{}
	if err := ctx.BindWith(r, binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	ctx.JSON(s.Clean(ctx, r.Tstart, r.Tend, r.PkgType, r.AppKey))
}
