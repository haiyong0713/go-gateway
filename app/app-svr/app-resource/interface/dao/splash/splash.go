package splash

import (
	"context"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	csgrcp "git.bilibili.co/bapis/bapis-go/collection-splash/service"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

// Dao is splash dao.
type Dao struct {
	account accountgrpc.AccountClient
	garb    garbgrpc.GarbClient
	cs      csgrcp.CollectionSplashClient
}

// New new splash dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{}
	account, err := accountgrpc.NewClient(c.AccountClient)
	if err != nil {
		panic(err)
	}
	d.account = account
	garb, err := garbgrpc.NewClient(c.GarbClient)
	if err != nil {
		panic(err)
	}
	d.garb = garb
	cs, err := csgrcp.NewClient(c.CollectionSplashClient)
	if err != nil {
		panic(err)
	}
	d.cs = cs
	return d
}

// Close close memcache resource.
func (dao *Dao) Close() {
}

func (d *Dao) AccountProfile(ctx context.Context, mid int64) (*accountgrpc.Profile, error) {
	reply, err := d.account.Profile3(ctx, &accountgrpc.MidReq{
		Mid:    mid,
		RealIp: metadata.String(ctx, metadata.RemoteIP),
	})
	if err != nil {
		return nil, err
	}
	if reply.Profile == nil {
		return nil, errors.Errorf("Invalid reply: %+v", reply)
	}
	return reply.Profile, nil
}

func (d *Dao) UserCollectionSplashList(ctx context.Context, mid int64) ([]int64, error) {
	reply, err := d.garb.UserSplashIds(ctx, &garbgrpc.UserSplashIdsReq{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	return reply.SplashIds, nil
}

func (d *Dao) CollectionSplash(ctx context.Context) ([]*csgrcp.Splash, error) {
	reply, err := d.cs.SplashList(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}
	return reply.GetList(), nil
}

func (d *Dao) EventSplash(ctx context.Context, mid int64, ip, mobiApp string) (*garbgrpc.EventSplashListReply, error) {
	reply, err := d.garb.EventSplashList(ctx, &garbgrpc.EventSplashListReq{
		Mid:      mid,
		Uip:      ip,
		Platform: mobiApp,
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}
