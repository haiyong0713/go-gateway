package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/baas/api"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) AddImport(ctx context.Context, req *api.AddImportRequest) (*empty.Empty, error) {
	if err := s.dao.AddImport(ctx, req); err != nil {
		log.Error("Failed to add import: %+v,%+v", req, err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) UpdateImport(ctx context.Context, req *api.UpdateImportRequest) (*empty.Empty, error) {
	if err := s.dao.UpdateImport(ctx, req); err != nil {
		log.Error("Failed to update import: %+v,%+v", req, err)
		return nil, err
	}
	return &empty.Empty{}, nil
}
