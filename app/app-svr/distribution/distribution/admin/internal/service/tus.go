package service

import (
	"context"
	"sort"

	"go-common/library/log"
	tusm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tus"
	"go-gateway/app/app-svr/distribution/distribution/admin/tool"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
)

const (
	tusPreference = "bilibili.app.distribution.experimental.v1.TusConfig"
)

func (s *Service) BatchFetchTusInfos(ctx context.Context) ([]*tusm.Info, error) {
	tusValues, err := tool.GetFiledOptionValuesFromPreferenceproto(tusPreference, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsTusValue)
	if err != nil {
		return nil, err
	}
	tusInfos, err := s.dao.BatchFetchTusInfos(ctx, tusValues)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return tusInfos, nil
}

func (s *Service) FetchTusConfigDetail(ctx context.Context, tusValue string) (*tusm.DetailReply, error) {
	km := tusm.TaishanKeyInfos{
		TusValue: tusValue,
	}
	keys := km.BuildKeys()
	configsWithKey, err := s.dao.FetchConfigsFromTaishan(ctx, keys)
	if err != nil {
		log.Error("error(%+v) km(%v)", err, km)
		return nil, err
	}
	dm, err := tool.FindMessageDescriptor("TusConfig", km.TusValue, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsTusValue)
	if err != nil {
		log.Error("error(%+v) km(%v)", err, km)
		return nil, err
	}
	var details []*tusm.Detail
	for key, v := range configsWithKey {
		jr, err := tool.MessageDescriptorToJson(dm, v)
		if err != nil {
			log.Error("error(%+v) km(%v)", err, km)
			return nil, err
		}
		if err != nil {
			log.Error("error(%+v) km(%v)", err, km)
			return nil, err
		}
		detail, err := tusm.TaishanKeyStringToDetail(key)
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		detail.Config = jr
		details = append(details, detail)
	}
	sort.Slice(details, func(i, j int) bool {
		return details[i].Result < details[j].Result
	})
	return &tusm.DetailReply{
		Details:        details,
		FieldBasicInfo: tool.FieldBasicInfo(dm),
	}, nil
}

func (s *Service) SaveTusConfigs(ctx context.Context, in []*tusm.Detail) error {
	//for validate
	dm, err := tool.FindMessageDescriptor("TusConfig", in[0].TusValue, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsTusValue)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	for _, v := range in {
		if _, err := tool.MessageDescriptorToJson(dm, v.Config); err != nil {
			log.Error("%+v", err)
			return err
		}
	}
	if err := s.dao.SaveTusConfigs(ctx, in); err != nil {
		log.Error("%+v", err)
		return err
	}
	s.AsyncLog(ctx)
	return nil
}
