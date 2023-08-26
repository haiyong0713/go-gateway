package service

import (
	"context"
	"math/rand"
	"time"

	"go-common/library/log/infoc"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/steins-gate/service/conf"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/account"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/archive"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/audit"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/hidden_vars"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/mark"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/ogv"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/record"
	"go-gateway/app/app-svr/steins-gate/service/internal/dao/steins"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// Service service.
type Service struct {
	c            *conf.Config
	dao          *steins.Dao
	ogvDao       *ogv.Dao
	markDao      *mark.Dao
	arcDao       *archive.Dao
	accDao       *account.Dao
	recordDao    *record.Dao
	auditDao     *audit.Dao
	hvaDao       *hidden_vars.Dao
	inCh         chan interface{}
	infocNode    *infoc.Infoc
	infocMark    *infoc.Infoc
	rand         *rand.Rand
	preGraphMids map[int64]struct{}
	promBusiness *prom.Prom
	skinList     []*model.Skin
}

// New new a service and return.n
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		dao:          steins.New(c),
		ogvDao:       ogv.New(c),
		markDao:      mark.New(c),
		arcDao:       archive.New(c),
		accDao:       account.New(c),
		recordDao:    record.New(c),
		hvaDao:       hidden_vars.New(c),
		auditDao:     audit.New(c),
		inCh:         make(chan interface{}, 1024),
		infocNode:    infoc.New(c.Node),
		infocMark:    infoc.New(c.Mark),
		rand:         rand.New(rand.NewSource(time.Now().UnixNano())),
		promBusiness: prom.BusinessInfoCount,
	}
	tmp := make(map[int64]struct{}, len(c.Rule.GraphMids))
	for _, v := range c.Rule.GraphMids {
		tmp[v] = struct{}{}
	}
	s.preGraphMids = tmp
	s.loadSkinList()
	//nolint:biligowordcheck
	go s.loadSkinListproc()
	return s
}

// IncrPromBusiness .
func (s *Service) IncrPromBusiness(key string) {
	s.promBusiness.Incr(key)
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return nil
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()

}
