package pwd_appeal

import (
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/boss"
	"go-gateway/app/app-svr/app-feed/admin/dao/pwd_appeal"
	"go-gateway/app/app-svr/app-feed/admin/dao/sms"
	"go-gateway/app/app-svr/app-feed/admin/service/spmode"
)

type Service struct {
	cfg          *conf.Config
	dao          *pwd_appeal.Dao
	boss         *boss.Dao
	sms          *sms.Dao
	worker       *fanout.Fanout
	EncryptedPwd map[string]string
}

func NewService(cfg *conf.Config) *Service {
	return &Service{
		cfg:          cfg,
		dao:          pwd_appeal.NewDao(cfg),
		boss:         boss.NewDao(cfg.PwdAppeal.Boss),
		sms:          sms.NewDao(cfg),
		worker:       fanout.New("pwd-appeal"),
		EncryptedPwd: spmode.EncryptedPwd(),
	}
}
