package tribe

import (
	"context"

	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) GetPackRelations(ctx context.Context, req *tribe.GetPackRelationsReq) (resp *tribe.GetPackRelationsResp, err error) {
	var rows []*tribemdl.HostRelations
	var relations []*tribe.Relation
	if rows, err = s.fkDao.SelectTribeHostRelation(ctx, req.AppKey, req.Feature); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	for _, v := range rows {
		relations = append(relations, po2to(v))
	}
	resp = &tribe.GetPackRelationsResp{
		Relations: relations,
	}
	return
}

func po2to(po *tribemdl.HostRelations) (relation *tribe.Relation) {
	if po == nil {
		return
	}
	return &tribe.Relation{
		Id:             po.Id,
		CurrentBuildId: po.CurrentBuildId,
		ParentBuildId:  po.ParentBuildId,
		Feature:        po.Feature,
	}
}
