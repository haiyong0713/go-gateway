package invite

import (
	"git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/invite"

	acp "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	silverbulletapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

// Service ...
type Service struct {
	c                                       *conf.Config
	invite                                  invite.Dao
	accClient                               api.AccountClient
	passportClient                          passportinfoapi.PassportUserClient
	silverbulletClient                      silverbulletapi.SilverbulletProxyClient
	acpClient                               acp.AccountControlPlaneClient
	tokenSalt, tokenExpire, faceTokenExpire string
	fan                                     *fanout.Fanout
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:               c,
		invite:          invite.New(c),
		tokenExpire:     c.Invite.TokenExpire,
		faceTokenExpire: c.Invite.FaceTokenExpire,
		fan:             fanout.New("activity-fan", fanout.Worker(2), fanout.Buffer(2048)),
	}
	var err error
	s.tokenSalt = s.c.Rule.TokenSalt

	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.silverbulletClient, err = silverbulletapi.NewClient(c.SilverClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passportinfoapi.NewClient(c.PassportClient); err != nil {
		panic(err)
	}
	if s.acpClient, err = acp.NewClient(c.AcpClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
	s.invite.Close()
}
