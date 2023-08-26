package service

import (
	"context"
	"sort"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
	tmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"
	"go-gateway/app/app-svr/distribution/distribution/admin/tool"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"github.com/jhump/protoreflect/desc"
	"github.com/pkg/errors"
)

const (
	multipleTusPreference = "bilibili.app.distribution.experimental.v1.MultipleTusConfig"
	_default              = "default"
)

func (s *Service) GetMultipleTusFields(ctx context.Context) ([]*tmm.FieldInfo, error) {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "Failed to fetch proto meta from %s", multipleTusPreference)
	}
	var (
		fieldInfos []*tmm.FieldInfo
		fieldNames []string
	)
	for _, v := range meta.ProtoDesc.GetFields() {
		fieldInfo := &tmm.FieldInfo{
			Name:       v.GetFullyQualifiedName(),
			Descriptor: tool.RemoveCRLF(v.GetSourceInfo().GetLeadingComments()),
		}
		fieldNames = append(fieldNames, fieldInfo.Name)
		fieldInfos = append(fieldInfos, fieldInfo)
	}
	s.rename(ctx, fieldNames, &multipleTusTitleRenamer{fieldInfos: fieldInfos})
	return fieldInfos, nil
}

func (s *Service) fetchTusValueInUsedByVersion(ctx context.Context, allTusValues []string, fieldName, configVersion string) ([]string, error) {
	configVersionManager, err := s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
	if err != nil {
		return nil, err
	}
	var tusValuesInManager = make(map[string]struct{})
	for _, v := range configVersionManager.VersionInfos {
		if v.ConfigVersion != configVersion {
			continue
		}
		for _, tusValue := range v.TusValues {
			tusValuesInManager[tusValue] = struct{}{}
		}
	}
	if len(tusValuesInManager) == 0 {
		return nil, errors.Errorf("Failed to match config version %s", configVersion)
	}
	var exceptedTusValues []string
	for _, v := range allTusValues {
		if _, ok := tusValuesInManager[v]; ok {
			exceptedTusValues = append(exceptedTusValues, v)
		}
	}
	return exceptedTusValues, nil
}

func (s *Service) FetchMultipleTusDetail(ctx context.Context, fieldName, configVersion string) (*tmm.DetailReply, error) {
	allTusValues, dm, err := parseTusValuesAndDmByFieldName(fieldName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	var (
		eg             = errgroup.WithContext(ctx)
		configsWithKey map[string][]byte
		tusInfos       map[string]string
		details        []*tmm.Detail
	)
	tusValues, err := s.fetchTusValueInUsedByVersion(ctx, allTusValues, fieldName, configVersion)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	eg.Go(func(ctx context.Context) error {
		tmpTusValues := append(tusValues, _default)
		taishanKeyInfo := &tmm.TaishanKeyInfo{
			Filed:         fieldName,
			ConfigVersion: configVersion,
			TusValues:     tmpTusValues,
		}
		reply, err := s.dao.FetchConfigsFromTaishan(ctx, taishanKeyInfo.BuildKeys())
		if err != nil {
			return err
		}
		configsWithKey = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		reply, err := s.dao.BatchFetchTusInfos(ctx, tusValues)
		if err != nil {
			return err
		}
		tmpTusInfos := make(map[string]string, len(reply))
		for _, v := range reply {
			tmpTusInfos[v.TusValue] = v.Name
		}
		tusInfos = tmpTusInfos
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	for key, v := range configsWithKey {
		jr, err := tool.MessageDescriptorToJson(dm, v)
		if err != nil {
			log.Error("error(%+v)", err)
			return nil, err
		}
		if err != nil {
			log.Error("error(%+v)", err)
			return nil, err
		}
		detail, err := tmm.TaishanKeyStringToDetail(key)
		if err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		detail.TusName = tusInfos[detail.TusValue]
		if detail.TusValue == model.DefaultTusValue {
			detail.TusName = model.DefaultTusValueName
		}
		detail.Config = jr
		details = append(details, detail)
	}
	sort.Slice(details, func(i, j int) bool {
		return details[i].TusValue < details[j].TusValue
	})
	s.rename(ctx, []string{fieldName}, &multipleTusTabsRenamer{fieldName: fieldName, details: details})
	return &tmm.DetailReply{
		Details:        details,
		FieldBasicInfo: tool.FieldBasicInfo(dm),
	}, nil
}

func parseTusValuesAndDmByFieldName(fieldName string) ([]string, *desc.MessageDescriptor, error) {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		return nil, nil, errors.Wrapf(ecode.NothingFound, "Failed to fetch proto meta from %s", multipleTusPreference)
	}
	var fieldDes *desc.FieldDescriptor
	for _, v := range meta.ProtoDesc.GetFields() {
		if v.GetFullyQualifiedName() == fieldName {
			fieldDes = v
			break
		}
	}
	if fieldDes == nil {
		return nil, nil, errors.Wrapf(ecode.NothingFound, "Failed to fetch field desc from %s", multipleTusPreference)
	}
	tusValues, err := preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsTusValues(fieldDes)
	if err != nil {
		return nil, nil, err
	}
	return tusValues, fieldDes.GetMessageType(), nil
}

func (s *Service) SaveMultipleTusConfig(ctx context.Context, in []*tmm.Detail, fieldName, configVersion string) error {
	_, dm, err := parseTusValuesAndDmByFieldName(fieldName)
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
	configVersionManager, err := s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	var existConfigVersionExpect bool
	for _, v := range configVersionManager.VersionInfos {
		if v.ConfigVersion == configVersion {
			existConfigVersionExpect = true
		}
	}
	if !existConfigVersionExpect {
		err := errors.Errorf("Failed to find config version %s", configVersion)
		log.Error("%+v", err)
		return err
	}
	if err := s.dao.SaveMultipleTusConfigs(ctx, in, fieldName, configVersion); err != nil {
		log.Error("%+v", err)
		return err
	}
	s.AsyncLog(ctx)
	return nil
}
