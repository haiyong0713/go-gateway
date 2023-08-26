package multipletusflag

import (
	"context"
	"encoding/json"
	"sort"

	"go-gateway/app/app-svr/distribution/distribution/internal/extension/tusvalue"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"
	tmv "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"go-common/library/log"

	"github.com/pkg/errors"
)

func (m MultipleTusFlag) findTusVersion(ctx context.Context) (map[string]*tmv.VersionInfo, error) {
	var fieldNames []string
	for fieldName := range tusvalue.TusValues {
		fieldNames = append(fieldNames, fieldName)
	}
	configVersionManagers, err := m.batchFetchVersionFromKV(ctx, fieldNames)
	if err != nil {
		return nil, err
	}
	reply := make(map[string]*tmv.VersionInfo, len(configVersionManagers))
	for _, configVersionManger := range configVersionManagers {
		versionInfo := chooseVersionByDevice(ctx, configVersionManger.VersionInfos)
		if versionInfo == nil {
			log.Error("failed to choose a version for:%+v", configVersionManger.Field)
			continue
		}
		reply[configVersionManger.Field] = versionInfo
	}
	return reply, nil
}

func (m MultipleTusFlag) batchFetchVersionFromKV(ctx context.Context, fieldNames []string) ([]*tmv.ConfigVersionManager, error) {
	var keys []string
	for _, v := range fieldNames {
		keys = append(keys, tmv.NewTaishanKey(v))
	}
	req := m.kvStore.NewBatchGetReq(ctx, keys)
	resp, err := m.kvStore.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	if !resp.AllSucceed {
		return nil, errors.Errorf("Failed to Fetch all configs")
	}
	var configVersionManagers []*tmv.ConfigVersionManager
	for _, v := range resp.Records {
		cvm := &tmv.ConfigVersionManager{}
		if err := json.Unmarshal(v.Columns[0].Value, cvm); err != nil {
			return nil, err
		}
		configVersionManagers = append(configVersionManagers, cvm)
	}
	return configVersionManagers, nil
}

func chooseVersionByDevice(ctx context.Context, versionInfos []*tmv.VersionInfo) *tmv.VersionInfo {
	ssCtx, _ := sessioncontext.FromContext(ctx)
	plat := tmv.PlatConverter(ssCtx.Device().RawMobiApp, ssCtx.Device().Device)
	sort.Slice(versionInfos, func(i, j int) bool {
		return versionInfos[i].ConfigVersion > versionInfos[j].ConfigVersion
	})
	var versionInfo *tmv.VersionInfo
	for _, v := range versionInfos {
		if v.ConfigVersion == tmv.FirstVersion {
			versionInfo = v
			break
		}
		if tmv.BuildLimits(v.BuildLimit).AllowDeviceToUse(plat, ssCtx.Device().Build) {
			versionInfo = v
			break
		}
	}
	return versionInfo
}
