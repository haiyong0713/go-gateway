package multipletusflag

import (
	"context"
	"crypto/md5"
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/distribution/distribution/internal/dao/kv"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"
	tmv "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"go-common/library/log"

	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

const (
	_defaultTus = "default"
)

type MultipleTusFlag struct {
	//tus client
	kvStore *kv.Taishan
	tus     tus.TitanUserServerClient
}

func New(tus tus.TitanUserServerClient, kv *kv.Taishan) *MultipleTusFlag {
	return &MultipleTusFlag{
		tus:     tus,
		kvStore: kv,
	}
}

func (m MultipleTusFlag) Name() string {
	return "multiple-tus-flag"
}

func (m MultipleTusFlag) GetUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta) ([]*preferenceproto.Preference, error) {
	versionInfos, err := m.findTusVersion(ctx)
	if err != nil {
		return nil, err
	}
	var tusValues []string
	for _, v := range versionInfos {
		tusValues = append(tusValues, v.TusValues...)
	}
	// 调用tus接口拿到这个人所有命中的人群包
	tusResult, err := m.fetchTusResult(ctx, tusValues)
	if err != nil {
		return nil, err
	}

	var (
		dmValueWithName = make(map[string]*dynamic.Message)
	)

	for _, meta := range metas {
		tmpDmValueWithName := m.fetchConfigFromKVAndParsed(ctx, meta, tusResult, versionInfos)
		//merged
		for name, v := range tmpDmValueWithName {
			dmValueWithName[name] = v
		}
	}
	out := make([]*preferenceproto.Preference, 0, len(metas))
	for _, meta := range metas {
		dm := dynamic.NewMessage(meta.ProtoDesc)
		for name, value := range dmValueWithName {
			if err := dm.TrySetFieldByName(name, value); err != nil {
				log.Error("Failed to set tus field: %+v: %+v", name, err)
				continue
			}
		}
		out = append(out, &preferenceproto.Preference{
			Meta:    *meta,
			Message: dm,
		})
	}
	return out, nil
}

func getHittedTusValue(tusResults map[string]struct{}, tusValues []string) string {
	for _, v := range tusValues {
		if _, ok := tusResults[v]; ok {
			return v
		}
	}
	//有没有命中该功能下的某个人群包,没有则默认用default配置
	return _defaultTus
}

func (m MultipleTusFlag) fetchConfigFromKVAndParsed(ctx context.Context, meta *preferenceproto.PreferenceMeta, tusResults map[string]struct{}, versionInfos map[string]*tmv.VersionInfo) map[string]*dynamic.Message {
	out := make(map[string]*dynamic.Message)
	for _, filed := range meta.ProtoDesc.GetFields() {
		//取出某个功能所有配置的人群包
		//有error continue+日志记录，防止因为一个功能配置有影响导致所有功能配置都没有解析
		versionInfo, ok := versionInfos[filed.GetFullyQualifiedName()]
		if !ok {
			log.Error("filed to get version Info , filed(%+v)", filed)
			continue
		}
		key := keyFormat(filed.GetFullyQualifiedName(), versionInfo.ConfigVersion, getHittedTusValue(tusResults, versionInfo.TusValues))
		kvReq := m.kvStore.NewGetReq([]byte(key))
		kvReply, err := m.kvStore.Get(ctx, kvReq)
		if err != nil {
			log.Error("filed to get kv reply error(%+v), filed(%+v)", err, filed)
			continue
		}
		dm, err := toDynamicMessage(kvReply.Columns[0].Value, filed.GetMessageType())
		if err != nil {
			log.Error("parseKVValueToDynamicMessage error(%+v), filed(%+v)", err, filed)
			continue
		}
		out[filed.GetName()] = dm
	}
	return out
}

func toDynamicMessage(in []byte, md *desc.MessageDescriptor) (*dynamic.Message, error) {
	dm := dynamic.NewMessage(md)
	if err := dm.UnmarshalJSONPB(&jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}, in); err != nil {
		return nil, err
	}
	return dm, nil
}

func (m MultipleTusFlag) fetchTusResult(ctx context.Context, tusValues []string) (map[string]struct{}, error) {
	ssCtx, _ := sessioncontext.FromContext(ctx)
	tusReply, err := m.tus.CheckTagBatch(ctx, &tus.TusBatchRequest{
		Uid:       strconv.FormatInt(ssCtx.Mid(), 10),
		Condition: buildConditionForTus(tusValues),
		BizType:   "gateway",
		UidType:   "buvid || mid",
		Sign:      fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%s", "2dd8dac606d1", strconv.FormatInt(ssCtx.Mid(), 10))))),
	})
	if err != nil {
		return nil, err
	}
	if tusReply.Code != 0 {
		return nil, errors.Errorf("tus CheckTagBatch error code(%d)", tusReply.Code)
	}
	if len(tusReply.Hits) != len(tusValues) {
		return nil, errors.Errorf("tus reply(%+v) length not match with tus values(%+v)", tusReply.Hits, tusValues)
	}

	allHitTus := make(map[string]struct{})
	for index, hit := range tusReply.Hits {
		if !hit {
			continue
		}
		allHitTus[tusValues[index]] = struct{}{}
	}

	return allHitTus, nil
}

func buildConditionForTus(tusValues []string) []string {
	var conditions []string
	for _, v := range tusValues {
		conditions = append(conditions, fmt.Sprintf("tag_%s==1", v))
	}
	return conditions
}

func keyFormat(field, configVersion, tusValue string) string {
	if configVersion == "v1.0" { //1.0版本还是使用老key，不用进行配置迁移
		return fmt.Sprintf("tusmultiple_%s_%s", field, tusValue)
	}
	return fmt.Sprintf("tusmultiple_%s_%s_%s", field, tusValue, configVersion)
}

func (m MultipleTusFlag) SetUserPreference(ctx context.Context, preferences []*preferenceproto.Preference) error {
	return nil
}
