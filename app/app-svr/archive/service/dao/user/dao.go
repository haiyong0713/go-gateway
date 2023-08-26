package user

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"

	user "git.bilibili.co/bapis/bapis-go/passport/service/user"

	"go-gateway/app/app-svr/archive/service/conf"
)

type Dao struct {
	c        *conf.Config
	puClient user.PassportUserClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.puClient, err = passportUserNewClient(c.PassportUserClient); err != nil {
		panic(fmt.Sprintf("passportUserNewClient NewClient error (%+v)", err))
	}
	return d
}

// NewClient new a grpc client
func passportUserNewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (user.PassportUserClient, error) {
	const (
		_appID = "passport.service.user"
	)
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+_appID)
	if err != nil {
		return nil, err
	}
	return user.NewPassportUserClient(conn), nil
}

func (d *Dao) UserFixedLocations(c context.Context) (res *user.UserFixedLocationsReply, err error) {
	req := &user.UserFixedLocationsReq{}
	if res, err = d.puClient.UserFixedLocations(c, req); err != nil {
		log.Error("d.puClient.UserFixedLocations error req(%+v) err(%+v)", req, err)
		return
	}
	return
}
