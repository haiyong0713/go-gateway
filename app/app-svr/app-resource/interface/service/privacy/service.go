package privacy

import (
	"context"
	"go-common/library/log"

	api "go-gateway/app/app-svr/app-resource/interface/api/privacy"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	dynamicdao "go-gateway/app/app-svr/app-resource/interface/dao/dynamic"
)

const (
	CityOpen  = int64(1)
	CityClose = int64(2)
)

type Service struct {
	c          *conf.Config
	dynamicDao *dynamicdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		dynamicDao: dynamicdao.New(c),
	}
	return
}

func (s *Service) PrivacyConfig(c context.Context, mid int64) (res *api.PrivacyConfigReply, err error) {
	res = new(api.PrivacyConfigReply)
	// 动态同城隐私开关
	res.PrivacyConfigItem = append(res.PrivacyConfigItem, s.cityPrivacyConfig(c, mid))
	return
}

func (s *Service) SetPrivacyConfig(c context.Context, mid int64, arg *api.SetPrivacyConfigRequest) (res *api.NoReply, err error) {
	res = new(api.NoReply)
	//nolint:exhaustive
	switch arg.PrivacyConfigType {
	case api.PrivacyConfigType_dynamic_city:
		if mid != 0 {
			if err = s.updateCityPrivacyConfig(c, mid, arg); err != nil {
				log.Error("%+v", err)
			}
		}
	}
	return
}

func (s *Service) cityPrivacyConfig(c context.Context, mid int64) *api.PrivacyConfigItem {
	pi := &api.PrivacyConfigItem{
		PrivacyConfigType: api.PrivacyConfigType_dynamic_city,
		Title:             s.c.Privacy.City.Title,
		State:             api.PrivacyConfigState_close,
		SubTitle:          s.c.Privacy.City.SubTitle,
		SubTitleUri:       s.c.Privacy.City.SubTitleURL,
	}
	if mid != 0 {
		state, err := s.dynamicDao.FetchUserPrivacy(c, mid)
		if err != nil {
			log.Error("%+v", err)
			return pi
		}
		if state == CityOpen {
			pi.State = api.PrivacyConfigState_open
		}
	}
	return pi
}

func (s *Service) updateCityPrivacyConfig(c context.Context, mid int64, arg *api.SetPrivacyConfigRequest) error {
	var state = CityClose
	//nolint:exhaustive
	switch arg.GetState() {
	case api.PrivacyConfigState_open:
		state = CityOpen
	}
	if err := s.dynamicDao.UpdateUserPrivacy(c, mid, state); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}
