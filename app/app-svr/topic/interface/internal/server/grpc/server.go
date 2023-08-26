package grpc

import (
	mauth "go-common/component/auth/middleware/grpc"
	mrestrict "go-common/component/restriction/middleware/grpc"
	abtest "go-common/component/tinker/middleware/grpc"
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"

	"go-gateway/app/app-svr/app-card/middleware/anticrawler"
	pb "go-gateway/app/app-svr/topic/interface/api"

	idtv1 "git.bilibili.co/bapis/bapis-go/passport/service/identify"

	"github.com/pkg/errors"
)

// New new a grpc server.
func New(svc pb.TopicServer) (ws *warden.Server, err error) {
	var (
		cfg         warden.ServerConfig
		ct          paladin.TOML
		quotaConfig quota.Config
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	if err = ct.Get("Quota").UnmarshalTOML(&quotaConfig); err != nil {
		return
	}
	ws = warden.NewServer(&cfg)
	pb.RegisterTopicServer(ws.Server(), svc)
	ws.Use(quota.New(&quotaConfig).Limit())
	if _, err = idtv1.NewClient(nil); err != nil {
		panic(errors.Errorf("rpcClient NewClient error: %+v", err))
	}
	// 用户鉴权
	auther := mauth.New(nil)
	ws.Add("/bilibili.app.topic.v1.Topic/TopicDetailsAll", auther.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	ws.Add("/bilibili.app.topic.v1.Topic/TopicDetailsFold", auther.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), mrestrict.UnaryServerInterceptor())
	ws.Add("/bilibili.app.topic.v1.Topic/TopicSetDetails", auther.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	ws.Add("/bilibili.app.topic.v1.Topic/TopicMergedResource", auther.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	ws.Add("/bilibili.app.topic.v1.Topic/TopicReserveButtonClick", auther.UnaryServerInterceptor(true), anticrawler.ReportInterceptor(), abtest.UnaryServerInterceptor(), mrestrict.UnaryServerInterceptor())
	ws, err = ws.Start()
	return
}
