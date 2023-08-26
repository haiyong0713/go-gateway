package spmode

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/account"
	"go-gateway/app/app-svr/app-feed/admin/dao/spmode"
)

type Service struct {
	cfg          *conf.Config
	dao          *spmode.Dao
	EncryptedPwd map[string]string
	worker       *fanout.Fanout
	accountDao   *account.Dao
}

func NewService(cfg *conf.Config) *Service {
	return &Service{
		cfg:          cfg,
		dao:          spmode.NewDao(cfg),
		EncryptedPwd: EncryptedPwd(),
		worker:       fanout.New("special-mode"),
		accountDao:   account.New(cfg),
	}
}

func EncryptedPwd() map[string]string {
	data := make(map[string]string, 10000)
	for i := 0; i < 10000; i++ {
		pwd := fmt.Sprintf("%04d", i)
		h := md5.New()
		h.Write([]byte(pwd))
		encrypted := hex.EncodeToString(h.Sum(nil))
		data[encrypted] = pwd
	}
	return data
}
