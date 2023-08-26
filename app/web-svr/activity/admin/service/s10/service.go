package s10

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/s10"
	"go-gateway/app/web-svr/activity/admin/model/component"
)

type Service struct {
	s10FilePath  string
	dao          *s10.Dao
	s10MaiInfo   *component.EmailInfo
	subTabSwitch bool
	robins       []int64
}

func New(c *conf.Config) *Service {
	s := &Service{
		s10FilePath:  c.S10Mail.FilePath,
		dao:          s10.New(c),
		subTabSwitch: c.S10General.SubTabSwitch,
		s10MaiInfo:   c.S10Mail.MailInfo,
		robins:       c.S10General.Robins,
	}
	return s
}
