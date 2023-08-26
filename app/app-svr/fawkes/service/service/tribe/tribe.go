package tribe

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) AddTribeInfo(ctx context.Context, req *tribe.AddTribeInfoReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	if _, err = s.fkDao.AddTribe(ctx, req.AppKey, req.Name, req.CName, req.Owners, req.Description, req.NoHost, req.Priority, req.IsBuildIn); err != nil {
		log.Errorc(ctx, "AddTribe error: %v", err)
	}
	return
}

func (s *Service) GetTribeInfo(ctx context.Context, req *tribe.GetTribeInfoReq) (resp *tribe.GetTribeInfoResp, err error) {
	var row *tribemdl.Tribe
	if row, err = s.fkDao.SelectTribeById(ctx, req.Id); err != nil {
		log.Errorc(ctx, "AddTribe id[%d] error: %v", req.Id, err)
		return
	}
	resp = &tribe.GetTribeInfoResp{
		TribeInfo: tribePO2DTO(row),
	}
	return
}

func (s *Service) ListTribeInfo(ctx context.Context, req *tribe.ListTribeInfoReq) (resp *tribe.ListTribeInfoResp, err error) {
	var (
		rows  []*tribemdl.Tribe
		infos []*tribe.TribeInfo
		total int64
	)
	if total, err = s.fkDao.CountTribeByArg(ctx, req.AppKey, req.Name, req.CName); err != nil {
		log.Errorc(ctx, "QueryTribeInfo error: %v", err)
		return
	}
	if rows, err = s.fkDao.SelectTribeByArg(ctx, req.AppKey, req.Name, req.CName, req.Ps, req.Pn); err != nil {
		log.Errorc(ctx, "QueryTribeInfo error: %v", err)
		return
	}
	for _, v := range rows {
		dto := tribePO2DTO(v)
		infos = append(infos, dto)
	}
	resp = &tribe.ListTribeInfoResp{
		PageInfo:  &tribe.PageInfo{Total: total, Pn: req.Pn, Ps: req.Ps},
		TribeInfo: infos,
	}
	return
}

func (s *Service) DeleteTribeInfo(ctx context.Context, req *tribe.DeleteTribeInfoReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	if _, err = s.fkDao.DeleteTribeById(ctx, req.Id); err != nil {
		log.Errorc(ctx, "DeleteTribeInfo id[%d] error: %v", req.Id, err)
	}
	return
}

func (s *Service) UpdateTribeInfo(ctx context.Context, req *tribe.UpdateTribeInfoReq) (resp *empty.Empty, err error) {
	resp = new(empty.Empty)
	if _, err = s.fkDao.UpdateTribe(ctx, req.Id, req.AppKey, req.Name, req.CName, req.Owners, req.Description, req.NoHost, req.Priority, req.IsBuildIn); err != nil {
		log.Errorc(ctx, "UpdateTribeInfo error: %v", err)
	}
	return
}

func tribePO2DTO(po *tribemdl.Tribe) (tribeInfo *tribe.TribeInfo) {
	if po == nil {
		return
	}
	tribeInfo = &tribe.TribeInfo{
		Id:          po.Id,
		AppKey:      po.AppKey,
		Name:        po.Name,
		CName:       po.CName,
		Owners:      po.Owners,
		Description: po.Description,
		NoHost:      po.NoHost,
		Priority:    po.Priority,
		IsBuildIn:   po.IsBuildIn,
	}
	return
}
