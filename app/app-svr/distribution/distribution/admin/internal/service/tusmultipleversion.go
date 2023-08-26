package service

import (
	"context"
	"fmt"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	vcm "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"go-common/library/log"

	"github.com/pkg/errors"
)

func (s *Service) FetchConfigVersionManagerByField(ctx context.Context, fieldName string) (*vcm.ConfigVersionManager, error) {
	return s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
}

func (s *Service) BatchFetchConfigVersionManager(ctx context.Context, fieldNames []string) ([]*vcm.ConfigVersionManager, error) {
	return s.tusMultipleVersionDao.BatchFetchVersionManager(ctx, fieldNames)
}

func (s *Service) AddVersion(ctx context.Context, fieldName string, buildLimit []*vcm.BuildLimit) (*vcm.VersionInfo, error) {
	if !vcm.BuildLimits(buildLimit).Valid() {
		log.Error("invalid buildLimit %+v", buildLimit)
		return nil, errors.Errorf("invalid buildLimit %+v", buildLimit)
	}
	//先取出当前存储的版本信息
	configVersionManager, err := s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	//拿最新的人群包信息
	tusValues, _, err := parseTusValuesAndDmByFieldName(fieldName)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	latestVersion, err := configVersionManager.VersionIncrease(buildLimit, tusValues)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if err := s.tusMultipleVersionDao.EditVersions(ctx, configVersionManager); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	s.AsyncLog(ctx)
	return &vcm.VersionInfo{
		ConfigVersion: latestVersion,
		BuildLimit:    buildLimit,
		TusValues:     tusValues,
	}, nil
}

func (s *Service) UpdateBuildLimit(ctx context.Context, fieldName string, in *vcm.VersionInfo) error {
	if !vcm.BuildLimits(in.BuildLimit).Valid() {
		log.Error("invalid buildLimit %+v", in.BuildLimit)
		return errors.Errorf("invalid buildLimit %+v", in.BuildLimit)
	}
	configVersionManager, err := s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	for _, v := range configVersionManager.VersionInfos {
		if v.ConfigVersion != in.ConfigVersion {
			continue
		}
		v.BuildLimit = in.BuildLimit
	}
	if err := s.tusMultipleVersionDao.EditVersions(ctx, configVersionManager); err != nil {
		log.Error("%+v", err)
		return err
	}
	s.AsyncLog(ctx)
	return nil
}

func (s *Service) DeleteConfigVersion(ctx context.Context, fieldName string, configVersion string) error {
	cmv, err := s.tusMultipleVersionDao.FetchVersionManager(ctx, fieldName)
	if err != nil {
		return err
	}
	var (
		versionInfoToDel  *vcm.VersionInfo
		versionInfoToSave []*vcm.VersionInfo
	)
	for _, v := range cmv.VersionInfos {
		if v.ConfigVersion == configVersion {
			versionInfoToDel = v
			continue
		}
		versionInfoToSave = append(versionInfoToSave, v)
	}
	cmv.VersionInfos = versionInfoToSave
	if versionInfoToDel == nil {
		return errors.Errorf("Failed to find version in config version(%s) field(%s)", configVersion, fieldName)
	}
	if err := s.tusMultipleVersionDao.EditVersions(ctx, cmv); err != nil {
		return err
	}
	if err := s.tusMultipleVersionDao.DeleteVersionConfig(ctx, fieldName, versionInfoToDel); err != nil {
		return err
	}
	return nil
}

func (s *Service) InitConfigVersion() {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		panic(fmt.Sprintf("Failed to fetch proto meta from %s", multipleTusPreference))
	}
	var (
		fieldNames []string
	)
	for _, v := range meta.ProtoDesc.GetFields() {
		fieldNames = append(fieldNames, v.GetFullyQualifiedName())
	}
	configVersionManagers, err := s.tusMultipleVersionDao.BatchFetchVersionManager(context.Background(), fieldNames)
	if err != nil {
		panic(err)
	}
	configVersionManagerMap := make(map[string]*vcm.ConfigVersionManager, len(configVersionManagers))
	for _, v := range configVersionManagers {
		configVersionManagerMap[v.Field] = v
	}
	for _, v := range fieldNames {
		configVersionManager, ok := configVersionManagerMap[v]
		if ok && existOriginVersion(configVersionManager) {
			continue
		}
		tusValues, _, err := parseTusValuesAndDmByFieldName(v)
		if err != nil {
			panic(err)
		}
		originConfigVersion := &vcm.ConfigVersionManager{
			Field: v,
			VersionInfos: []*vcm.VersionInfo{
				{
					ConfigVersion: "v1.0",
					TusValues:     tusValues,
					CreateTime:    time.Now().Unix(),
				},
			},
		}
		if err := s.tusMultipleVersionDao.EditVersions(context.Background(), originConfigVersion); err != nil {
			panic(err)
		}
	}
}

func existOriginVersion(configVersionManager *vcm.ConfigVersionManager) bool {
	if configVersionManager == nil {
		return false
	}
	return len(configVersionManager.VersionInfos) > 0
}
