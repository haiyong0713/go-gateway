package service

import (
	"context"
	"sort"

	"go-common/library/log"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_kvID = 2326
)

// Kv get baidu kv
func (s *Service) Kv(ctx context.Context) ([]*model.Kv, error) {
	reply, err := s.resgrpc.ResourceNew(ctx, &resourcegrpc.ResourceRequest{ResID: _kvID})
	if err != nil {
		return nil, err
	}
	if len(reply.GetResource().GetAssignments()) == 0 {
		return []*model.Kv{}, nil
	}
	var res []*model.Kv
	for _, assi := range reply.GetResource().GetAssignments() {
		res = append(res, &model.Kv{ID: assi.Id, Name: assi.Name, Pic: assi.Pic, URL: assi.Url, ResID: assi.ResourceId, STime: assi.Stime, ETime: assi.Etime})
	}
	return res, nil
}

// CmtBox get live dm box
func (s *Service) CmtBox(ctx context.Context, id int64) (*resourcegrpc.CmtboxReply, error) {
	reply, err := s.resgrpc.CmtboxNew(ctx, &resourcegrpc.CmtboxRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// AbServer get ab server info.
func (s *Service) AbServer(c context.Context, mid int64, platform int, channel, buvid string) (data model.AbServer, err error) {
	return s.dao.AbServer(c, mid, platform, channel, buvid)
}

func (s *Service) loadInformationRegionCard() {
	var (
		tmpm, tmpm2 map[int32][]*resourcegrpc.InformationRegionCard
		err         error
	)
	if tmpm, err = s.dao.InformationRegionCard(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	// 根据插入位置正序排序
	tmpm2 = make(map[int32][]*resourcegrpc.InformationRegionCard)
	for rid, tmps := range tmpm {
		var (
			idxs []int
			im   = make(map[int]*resourcegrpc.InformationRegionCard)
		)
		for _, tmp := range tmps {
			if tmp != nil {
				idxs = append(idxs, int(tmp.PositionIdx))
				im[int(tmp.PositionIdx)] = tmp
			}
		}
		sort.Ints(idxs)
		for _, idx := range idxs {
			tmpm2[rid] = append(tmpm2[rid], im[idx])
		}
	}
	s.informationRegionCardCache = tmpm2
}

func (s *Service) loadParamConfig() {
	tmp, err := s.dao.ParamConfig(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.paramConfigCache = tmp
}

func (s *Service) ParamConfig(_ context.Context, key string) (*model.ParamConfig, error) {
	return s.paramConfigCache[key], nil
}
