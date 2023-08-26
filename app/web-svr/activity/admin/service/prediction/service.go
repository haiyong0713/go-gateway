package prediction

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/conf"
	predao "go-gateway/app/web-svr/activity/admin/dao/prediction"
	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"
)

// Service struct
type Service struct {
	c   *conf.Config
	dao *predao.Dao
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: predao.New(c),
	}
	return
}

// BatchAdd .
func (s *Service) BatchAdd(c context.Context, add []*premdl.BatchAdd) (err error) {
	list := make([]*premdl.Prediction, 0, len(add))
	for _, v := range add {
		if v.Sid > 0 && v.Pid >= 0 && v.Name != "" {
			list = append(list, &premdl.Prediction{Sid: v.Sid, Min: v.Min, Max: v.Max, Type: v.Type, Name: v.Name, State: v.State, Pid: v.Pid})
		}
	}
	if len(list) > 0 {
		if err = s.dao.BatchAdd(c, list); err != nil {
			log.Error("s.dao.BatchAdd() error(%+v)", err)
		}
	}
	return
}

// PredSearch .
func (s *Service) PredSearch(c context.Context, search *premdl.PredSearch) (list *premdl.SearchRes, err error) {
	if list, err = s.dao.Search(c, search); err != nil {
		log.Error("s.dao.Search(%v) error(%v)", search, err)
	}
	return
}

// PresUp .
func (s *Service) PresUp(c context.Context, up *premdl.PresUp) (err error) {
	if err = s.dao.PresUp(c, up); err != nil {
		log.Error("s.dao.PresUp(%v) error(%v)", up, err)
	}
	return
}

// ItemAdd .
func (s *Service) ItemAdd(c context.Context, items []*premdl.ItemAdd) (err error) {
	itemList := make([]*premdl.PredItem, 0, len(items))
	for _, v := range items {
		itemList = append(itemList, &premdl.PredItem{Sid: v.Sid, Pid: v.Pid, State: v.State, Desc: v.Desc, Image: v.Image})
	}
	if len(itemList) > 0 {
		if err = s.dao.BatchItem(c, itemList); err != nil {
			log.Error("s.dao.BatchItem(%v) error(%v)", itemList, err)
			return
		}
	}
	return
}

// ItemUp .
func (s *Service) ItemUp(c context.Context, up *premdl.ItemUp) (err error) {
	if err = s.dao.ItemUp(c, up); err != nil {
		log.Error("s.dao.ItemUp(%v) error(%v)", up, err)
	}
	return
}

// ItemSearch .
func (s *Service) ItemSearch(c context.Context, arg *premdl.ItemSearch) (res *premdl.ItemSearchRes, err error) {
	if res, err = s.dao.ItemSearch(c, arg); err != nil {
		log.Error("s.dao.ItemSearch(%v) error(%v)", arg, err)
	}
	return
}
