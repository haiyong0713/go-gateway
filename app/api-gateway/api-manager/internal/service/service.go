package service

import (
	"context"
	"go-common/library/sync/pipeline/fanout"
	"sync"

	"go-common/library/conf/paladin"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"

	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/api-manager/internal/dao"
	"go-gateway/app/api-gateway/api-manager/internal/model"
	"go-gateway/app/api-gateway/delay"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.ManagerServer), new(*Service)))

// Service service.
type Service struct {
	ac           *paladin.Map
	dao          dao.Dao
	delay        delay.Dao
	gitCfg       *gitCfg
	allDis       []string
	allProtos    map[string]*model.ProtoInfo
	mutex        sync.Mutex
	protoRg      *railgun.Railgun
	codeGenerate *fanout.Fanout
	wfFanout     *fanout.Fanout
}

type gitCfg struct {
	UserName string
	Token    string
	Url      string
	DirPath  string
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:           &paladin.TOML{},
		dao:          d,
		codeGenerate: fanout.New("codeGenerate", fanout.Worker(4), fanout.Buffer(1024)),
		wfFanout:     fanout.New("workflow", fanout.Worker(4), fanout.Buffer(1024)),
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&s.gitCfg); err != nil {
		panic(err)
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}

	var cfg delay.Cfg
	if err = paladin.Get("delay.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if s.delay, err = delay.NewDao(cfg); err != nil {
		panic(err)
	}

	s.getAllDis()

	_ = s.codeGenerate.Do(context.Background(), func(ctx context.Context) {
		_ = s.gitClone()
		if err = s.analysisProto(); err != nil {
			log.Error("s.analysisProto error: %+v", err)
			panic(err)
		}
		s.getAllDis()
	})

	s.initRg()
	return
}

func (s *Service) getAllDis() {
	pros, err := s.dao.GetAllProtos(context.Background())
	if err != nil {
		panic(err)
	}
	allProtos := make(map[string]*model.ProtoInfo)
	var allDis []string
	for _, p := range pros {
		allProtos[p.DiscoveryID] = p
		allDis = append(allDis, p.DiscoveryID)
	}
	s.allProtos = allProtos
	s.allDis = allDis
}

func (s *Service) ApiDetailList(ctx context.Context, req *pb.ApiDetailListReq) (reply *pb.ApiDetailListResp, err error) {
	reply = &pb.ApiDetailListResp{Res: make(map[string]*pb.ApiInfo)}
	switch req.Tp {
	case model.ApiTypeGrpc:
		if reply.Res, err = s.dao.GetProtoByDis(ctx, req.Name); err != nil {
			return
		}
		var tmpRes map[string][]string
		if tmpRes, err = s.dao.GetServiceName(ctx, req.Name); err != nil {
			return
		}
		for k, v := range tmpRes {
			if _, ok := reply.Res[k]; ok {
				reply.Res[k].ServiceName = v
			}
		}
	case model.ApiTypeHttp:
		if reply.Res, err = s.dao.GetHttpApisByPath(ctx, req.Name); err != nil {
			return
		}
	default:
		return nil, ecode.RequestErr
	}
	return
}

// Close close the resource.
func (s *Service) Close() {
	_ = s.codeGenerate.Close()
	_ = s.wfFanout.Close()
	s.protoRg.Close()
	s.delay.Close()
}
