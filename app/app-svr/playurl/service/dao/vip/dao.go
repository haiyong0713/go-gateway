package vip

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/credis"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/playurl/service/conf"

	vipInforpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"google.golang.org/grpc"
)

// Dao is vip dao.
type Dao struct {
	// redis
	redis credis.Redis
	// rpc
	vipRPC vipInforpc.VipInfoClient
}

func vipBk(mid int64) string {
	return fmt.Sprintf("vi_d:%d", mid)
}

// New vip dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		redis: credis.NewRedis(c.Redis.Vip),
	}

	clientSDK := conf.WardenSDKBuilder.Build("vipinfo.service")
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(clientSDK.UnaryClientInterceptor()),
	}
	vipRPC, err := vipInforpc.NewClient(c.VipClient, opts...)
	if err != nil {
		panic(fmt.Sprintf("vip NewClient error(%v)", err))
	}
	d.vipRPC = vipRPC
	return
}

func (d *Dao) isVipBk(c context.Context, mid int64) bool {
	conn := d.redis.Conn(c)
	defer conn.Close()
	value, err := redis.Bytes(conn.Do("GET", vipBk(mid)))
	if err != nil {
		log.Error("IsVipBk conn.Do(GET key:%s) err(%+v)", vipBk(mid), err)
		return false
	}
	vipInfo := &vipInforpc.ModelInfo{}
	if err = json.Unmarshal(value, &vipInfo); err != nil {
		log.Error("IsVipBk json.Unmarshal key:%s err(%+v)", vipBk(mid), err)
		return false
	}
	log.Info("IsVipBk redis %s:%v", vipBk(mid), vipInfo)
	return vipInfo.IsValid()
}

// Info .
// nolint:govet
func (d *Dao) Info(c context.Context, mid int64, buvid string, withControl bool) (isVip bool, control *vipInforpc.ControlResult) {
	reply, err := d.vipRPC.Info(c, &vipInforpc.InfoReq{Mid: mid, Buvid: buvid, WithControl: withControl})
	if err != nil {
		log.Error("verifyArchive VipInfo mid(%d) error(%v)", mid, err)
		//如果vip err = -503 or -509 走缓存降级
		if ecode.EqualError(ecode.LimitExceed, err) || ecode.EqualError(ecode.ServiceUnavailable, err) {
			isVip = d.isVipBk(c, mid)
			log.Warn("Use VipBk mid(%d) vip(%d) err(%+v)", mid, isVip, err)
		}
		return
	}
	if reply != nil && reply.Res != nil {
		// 被管控时，不下发vip清晰度
		if reply.Res.IsValid() && (reply.Control == nil || !reply.Control.Control) {
			isVip = true
		}
		//vip 才下发管控信息
		if reply.Res.IsValid() {
			control = reply.Control
		}
	}
	return
}
