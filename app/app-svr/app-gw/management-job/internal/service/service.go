package service

import (
	"context"
	"time"

	"go-common/library/conf/env"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	pb "go-gateway/app/app-svr/app-gw/management-job/api"
	"go-gateway/app/app-svr/app-gw/management-job/internal/dao"
	"go-gateway/app/app-svr/app-gw/management-job/internal/election"
	logutil "go-gateway/app/app-svr/app-gw/management/audit"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.ManagementJobServer), new(*Service)))

var LockKey = "{management-job}/election"

// Service service.
type Service struct {
	ac        *paladin.Map
	dao       dao.Dao
	cron      *cron.Cron
	worker    *fanout.Fanout
	candidate *election.Candidate
	follower  *election.Follower
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:     &paladin.TOML{},
		dao:    d,
		cron:   cron.New(),
		worker: fanout.New("app-gw-management-job", fanout.Worker(10), fanout.Buffer(1024)),
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	s.candidate, s.follower = election.NewCandidateAndFollower(s.dao.GetTaishan(), LockKey, dao.InstanceValue(), 10)
	if env.DeployEnv == "pre" {
		log.Warn("replication controller and cronjobs will not be dispatched in pre env")
		return
	}
	election.SetupReplication(context.Background(), s.candidate, s.follower)
	s.createCron()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.candidate.Stop()
	s.cron.Stop()
}

func (s *Service) cronPushAll(ctx context.Context) error {
	req := &pb.TaskDoReq{
		Sponsor: logutil.LogMngSponsor,
		Params: &pb.Params{
			Ctime: time.Now().Unix(),
			Mtime: time.Now().Unix(),
		},
	}
	return s.PushAll(ctx, req)
}

func (s *Service) cronGRPCPushAll(ctx context.Context) error {
	req := &pb.TaskDoReq{
		Sponsor: logutil.LogMngSponsor,
		Params: &pb.Params{
			Ctime: time.Now().Unix(),
			Mtime: time.Now().Unix(),
		},
	}
	return s.GRPCPushAll(ctx, req)
}

func loggingError(in func() error) func() {
	return func() {
		if err := in(); err != nil {
			log.Error("Failed to execute cron job: %+v", err)
		}
	}
}

func cronWithContext(ctx context.Context, in func(context.Context) error) func() error {
	return func() error {
		return in(ctx)
	}
}

func (s *Service) runOnLeader(in func() error) func() error {
	return func() error {
		if !s.candidate.IsLeader() {
			log.Info("Skip running cron job on follower node")
			return nil
		}
		return in()
	}
}

type CronTask struct {
	Spec string
	Fn   func()
}

func (s *Service) createCron() {
	crons := new(pb.CronsConfig)
	if err := s.ac.Get("CronsConfig").UnmarshalTOML(crons); err != nil {
		panic(err)
	}
	cronTasks := []CronTask{
		{crons.LoadProxyPassConfigsCron, loggingError(s.runOnLeader(cronWithContext(context.Background(), s.cronPushAll)))},
		{crons.LoadProxyPassConfigsCron, loggingError(s.runOnLeader(cronWithContext(context.Background(), s.cronGRPCPushAll)))},
	}
	for _, task := range cronTasks {
		if err := s.cron.AddFunc(task.Spec, task.Fn); err != nil {
			panic(err)
		}
	}
	s.cron.Start()
}
