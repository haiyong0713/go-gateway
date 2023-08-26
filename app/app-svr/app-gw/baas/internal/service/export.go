package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/baas/api"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) ExportList(ctx context.Context, req *api.ExportListRequest) (*api.ExportListReply, error) {
	export, err := s.exportList(ctx, req.TreeId, req.ExportApi)
	if err != nil {
		log.Error("Failed to export by condition: %s, %+v", req.ExportApi, err)
		return nil, err
	}
	ids := make([]int64, 0, len(export))
	for _, v := range export {
		ids = append(ids, v.Id)
	}
	import_, err := s.dao.ImportByExportIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]*api.ExportList, 0, len(export))
	for _, v := range export {
		item := &api.ExportList{
			Export:  api.ConstructExportItem(v),
			Imports: constructImportItems(import_[v.Id]),
		}
		list = append(list, item)
	}
	out := &api.ExportListReply{
		List: list,
	}
	return out, nil
}

func constructImportItems(in []*api.BaasImport) []*api.ImportItem {
	out := make([]*api.ImportItem, 0)
	for _, val := range in {
		out = append(out, api.ConstructImportItem(val))
	}
	return out
}

func (s *Service) exportList(ctx context.Context, treeID int64, exportAPI string) ([]*api.BaasExport, error) {
	result, err := s.dao.ExportList(ctx)
	if err != nil {
		return nil, err
	}
	listByTreeID := make([]*api.BaasExport, 0)
	for _, v := range result {
		if v.TreeId == treeID {
			listByTreeID = append(listByTreeID, v)
		}
	}
	if treeID == 0 {
		listByTreeID = result
	}
	if exportAPI != "" {
		out := make([]*api.BaasExport, 0)
		for _, v := range listByTreeID {
			if v.ExportApi == exportAPI {
				out = append(out, v)
			}
		}
		return out, nil
	}
	return listByTreeID, nil
}

func (s *Service) AddExport(ctx context.Context, req *api.AddExportRequest) (*empty.Empty, error) {
	if err := s.dao.AddExport(ctx, req); err != nil {
		log.Error("Failed to add export: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) UpdateExport(ctx context.Context, req *api.UpdateExportRequest) (*empty.Empty, error) {
	if err := s.dao.UpdateExport(ctx, req); err != nil {
		log.Error("Failed to update export: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}
