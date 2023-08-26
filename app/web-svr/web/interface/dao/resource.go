package dao

import (
	"context"

	"go-common/library/log"

	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	resourmdl "go-gateway/app/app-svr/resource/service/model"
	"go-gateway/app/web-svr/web/interface/model"
)

func (d *Dao) InformationRegionCard(c context.Context) (res map[int32][]*resourcegrpc.InformationRegionCard, err error) {
	var (
		args   = &resourcegrpc.NoArgRequest{}
		resTmp *resourcegrpc.InformationRegionCardReply
	)
	if resTmp, err = d.resourctClient.InformationRegionCard(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[int32][]*resourcegrpc.InformationRegionCard)
	for _, regionCard := range resTmp.GetInformationRegionCards() {
		if regionCard == nil {
			continue
		}
		if regionID := resourmdl.InformationRegionChange(regionCard.GetCardPosition()); regionID != 0 {
			res[regionID] = append(res[regionID], regionCard)
		}
	}
	return
}

func (d *Dao) ParamConfig(c context.Context) (map[string]*model.ParamConfig, error) {
	resTmp, err := d.resourctClient.ParamList(c, &resourcegrpc.ParamReq{Plats: []int64{30}})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	res := make(map[string]*model.ParamConfig)
	for _, list := range resTmp.List {
		if list == nil {
			continue
		}
		res[list.GetName()] = &model.ParamConfig{
			ID:         list.GetID(),
			Name:       list.GetName(),
			Value:      list.GetValue(),
			Remark:     list.GetRemark(),
			Department: list.GetDepartment(),
		}
	}
	return res, nil
}
