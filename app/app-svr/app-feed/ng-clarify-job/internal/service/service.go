package service

import (
	"context"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-feed/ng-clarify-job/api"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/dao"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/model"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.AppFeedNGClarifyJobServer), new(*Service)))

// Service service.
type Service struct {
	ac  *paladin.Map
	dao dao.Dao
}

// New new a service and return.
func New(d dao.Dao) (*Service, func(), error) {
	s := &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	cf := s.Close
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return s, cf, nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}

func (s *Service) SaveSession(ctx context.Context, session *model.IndexSession) error {
	return s.dao.SaveSession(ctx, session)
}

func (s *Service) DownloadURL(ctx context.Context, archivePath string, durSeconds int64) (*model.PresignedURLReply, error) {
	url, err := s.dao.DownloadURL(archivePath, time.Duration(durSeconds)*time.Second)
	if err != nil {
		return nil, err
	}
	reply := &model.PresignedURLReply{
		URL: url,
	}
	return reply, nil
}

func (s *Service) ScanArchvieIndex(ctx context.Context, startTS, endTS int64, lastKey string) (*model.ScanArchiveIndexReply, error) {
	reply, err := s.dao.ScanArchvieIndex(ctx, startTS, endTS, lastKey)
	if err != nil {
		return nil, err
	}
	for _, i := range reply.Index {
		urlStr, err := s.dao.DownloadURL(i.Path, time.Duration(600)*time.Second)
		if err != nil {
			log.Error("Failed to sign download url: %+v", err)
			continue
		}
		i.URL = urlStr
		i.CreatedTime = time.Unix(0, i.CreatedAt).String()
	}
	return reply, nil
}
