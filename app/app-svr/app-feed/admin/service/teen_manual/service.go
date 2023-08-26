package teen_manual

import (
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/account"
	"go-gateway/app/app-svr/app-feed/admin/dao/member"
	"go-gateway/app/app-svr/app-feed/admin/dao/spmode"
	"go-gateway/app/app-svr/app-feed/admin/dao/teen_manual"
)

type Service struct {
	cfg        *conf.Config
	dao        *teen_manual.Dao
	spmodeDao  *spmode.Dao
	accountDao *account.Dao
	memberDao  *member.Dao
	worker     *fanout.Fanout
}

func NewService(cfg *conf.Config) *Service {
	return &Service{
		cfg:        cfg,
		dao:        teen_manual.NewDao(cfg),
		spmodeDao:  spmode.NewDao(cfg),
		accountDao: account.New(cfg),
		memberDao:  member.NewDao(cfg),
		worker:     fanout.New("teen-manual-force"),
	}
}
