package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/app-player/job/conf"
	"go-gateway/app/app-svr/app-player/job/dao"
	"go-gateway/app/app-svr/app-player/job/model"

	bcgrpc "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"

	"google.golang.org/grpc"
)

// Service is service.
type Service struct {
	c                     *conf.Config
	dao                   *dao.Dao
	trafficControlRailGun *railgun.Railgun
	bcClient              bcgrpc.BroadcastVideoAPIClient
}

// New new a service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}
	var err error
	if s.bcClient, err = func(cfg *warden.ClientConfig, opts ...grpc.DialOption) (bcgrpc.BroadcastVideoAPIClient, error) {
		client := warden.NewClient(cfg, opts...)
		conn, err := client.Dial(context.Background(), "discovery://default/"+"push.service.broadcast")
		if err != nil {
			return nil, err
		}
		return bcgrpc.NewBroadcastVideoAPIClient(conn), nil
	}(c.Broadcast); err != nil {
		panic(fmt.Sprintf("env:%s no BroadcastVideoAPIClient grpc newClient error(%v)", env.DeployEnv, err))
	}
	s.initTrafficControlRailGun(c.ArcControlRailGun.Databus, c.ArcControlRailGun.SingleConfig, c.ArcControlRailGun.Cfg)
	return
}

func (s *Service) initTrafficControlRailGun(databus *railgun.DatabusV1Config, singleConfig *railgun.SingleConfig, cfg *railgun.Config) {
	inputer := railgun.NewDatabusV1Inputer(databus)
	processor := railgun.NewSingleProcessor(singleConfig, s.trafficControlRailGunUnpack, s.trafficControlRailGunDo)
	g := railgun.NewRailGun("稿件流量管控消息推送", cfg, inputer, processor)
	s.trafficControlRailGun = g
	g.Start()
}

func (s *Service) trafficControlRailGunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	tcInfo := new(model.TrafficControl)
	if err := json.Unmarshal(msg.Payload(), &tcInfo); err != nil {
		log.Error("trafficControlRailGunUnpack json.Unmarshal error(%+v)", err)
		return nil, err
	}
	if !tcInfo.LegalTCInfo() {
		return nil, nil
	}
	return &railgun.SingleUnpackMsg{
		Group: tcInfo.Data.Oid,
		Item:  tcInfo,
	}, nil
}

func (s *Service) trafficControlRailGunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	tcInfo, ok := item.(*model.TrafficControl)
	if !ok {
		return railgun.MsgPolicyIgnore
	}
	log.Warn("trafficControlRailGunUnpack aid(%d), flowState(%+v)", tcInfo.Data.Oid, tcInfo.Data.FlowState)
	//手动开启在线人数平滑策略
	if tcInfo.ManualControl() {
		s.SaveOnlineInfo(ctx, true, tcInfo.Data.Oid)
		return railgun.MsgPolicyNormal
	}
	//自动触发在线人数平滑策略
	if tcInfo.AutoControl() {
		s.SaveOnlineInfo(ctx, false, tcInfo.Data.Oid)
		return railgun.MsgPolicyNormal
	}
	return railgun.MsgPolicyIgnore
}

// Close Databus consumer close.
func (s *Service) Close() {
	s.trafficControlRailGun.Close()
}

// Ping is
func (s *Service) Ping(c context.Context) (err error) {
	return
}
