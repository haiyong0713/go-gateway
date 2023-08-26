package service

import (
	"context"
	abm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/abtest"
	"go-gateway/app/app-svr/distribution/distribution/admin/tool"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"sort"

	"github.com/pkg/errors"
)

const (
	abtestPreference = "bilibili.app.distribution.experimental.v1.ABTestConfig"
	expIDLen         = 1
)

func (s *Service) BatchFetchABTestInfo(ctx context.Context) ([]*abm.Infos, error) {
	expFlagValues, err := tool.GetFiledOptionValuesFromPreferenceproto(abtestPreference, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsABTestFlagValue)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	valueIDsMap, err := s.dao.BatchFetchAbtestExpID(ctx, expFlagValues)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for _, v := range valueIDsMap {
		if len(v) != expIDLen {
			log.Error("")
			return nil, errors.Errorf("ExpValues match too many ids")
		}
	}
	valueIDMap := make(map[string]int64, len(valueIDsMap))
	for expValue, expId := range valueIDsMap {
		valueIDMap[expValue] = expId[0]
	}
	abInfos, err := s.dao.BatchFetchAbtestExpInfo(ctx, valueIDMap)
	if err != nil {
		return nil, err
	}
	sort.Slice(abInfos, func(i, j int) bool {
		return abInfos[i].ID < abInfos[j].ID
	})
	return abInfos, nil
}

func (s *Service) FetchABTestConfigDetail(ctx context.Context, req *abm.DetailReq) (*abm.DetailReply, error) {
	groupIDWithName, err := s.dao.FetchAbtestGroupIDWithName(ctx, req.ExpID)
	if err != nil {
		log.Error("FetchABTestConfigDetail s.dao.FetchAbtestGroupInfos error(%+v)", err)
		return nil, err
	}
	var groupIDs []string
	for id := range groupIDWithName {
		groupIDs = append(groupIDs, id)
	}
	km := abm.TaishanKeyInfos{
		ID:       req.ExpID,
		GroupIDs: groupIDs,
	}
	keys := km.BuildKeys()
	configsWithKey, err := s.dao.FetchConfigsFromTaishan(ctx, keys)
	if err != nil {
		log.Error("FetchABTestConfigDetail s.dao.FetchABTestConfigs error(%+v) km(%v)", err, km)
		return nil, err
	}
	dm, err := tool.FindMessageDescriptor("ABTestConfig", req.FlagValue, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsABTestFlagValue)
	if err != nil {
		log.Error("FetchABTestConfigDetail FindMessageDescriptor error(%+v) km(%v)", err, km)
		return nil, err
	}
	var details []*abm.Detail
	for key, v := range configsWithKey {
		jr, err := tool.MessageDescriptorToJson(dm, v)
		if err != nil {
			log.Error("FetchABTestConfigDetail MessageDescriptorToJson error(%+v) km(%v)", err, km)
			return nil, err
		}
		detail, err := abm.TaishanKeyStringToDetail(key)
		if err != nil {
			log.Error("FetchABTestConfigDetail TaishanKeyStringToDetail error(%+v) km(%v)", err, km)
			return nil, err
		}
		detail.FlagValue = req.FlagValue
		detail.GroupName = groupIDWithName[detail.GroupID]
		detail.Config = jr
		details = append(details, detail)
	}
	sort.Slice(details, func(i, j int) bool {
		return details[i].GroupID < details[j].GroupID
	})
	return &abm.DetailReply{
		Details:        details,
		FieldBasicInfo: tool.FieldBasicInfo(dm),
	}, nil
}

func (s *Service) SaveABTestConfigs(ctx context.Context, in []*abm.Detail) error {
	//for validate
	dm, err := tool.FindMessageDescriptor("ABTestConfig", in[0].FlagValue, preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsABTestFlagValue)
	if err != nil {
		log.Error("SaveABTestConfigs FindMessageDescriptor error(%+v)", err)
		return err
	}
	for _, v := range in {
		if _, err := tool.MessageDescriptorToJson(dm, v.Config); err != nil {
			log.Error("SaveABTestConfigs MessageDescriptorToJson error(%+v)", err)
			return err
		}
	}
	if err := s.dao.SaveABTestConfigs(ctx, in); err != nil {
		log.Error("SaveABTestConfigs SaveABTestConfigs error(%+v)", err)
		return err
	}
	s.AsyncLog(ctx)
	return nil
}
