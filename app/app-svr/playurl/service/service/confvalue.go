package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"

	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/model"

	"git.bilibili.co/bapis/bapis-go/bilibili/app/distribution"
	disbase "git.bilibili.co/bapis/bapis-go/bilibili/app/distribution"
	dp "git.bilibili.co/bapis/bapis-go/bilibili/app/distribution/setting/play"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

const (
	//typeurl
	_distributionPlayConf      = "type.googleapis.com/bilibili.app.distribution.play.v1.PlayConfig"
	_distributionCloudPlayConf = "type.googleapis.com/bilibili.app.distribution.play.v1.CloudPlayConfig"
)

func convertConfValueToAnys(confValueToSave []*model.ConfValueEdit) ([]*types.Any, error) {
	//播放配置的值有修改
	var (
		playConf      = &dp.PlayConfig{}
		cloudPlayConf = &dp.CloudPlayConfig{}
	)
	for _, v := range confValueToSave {
		switch v.ConfType {
		case v2.ConfType_COLORFILTER:
			colorFilter, ok := v.ConfValue.Value.(*v2.ConfValue_SelectedVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			playConf.ColorFilter = &disbase.Int64Value{Value: colorFilter.SelectedVal}
		case v2.ConfType_SUBTITLE:
			subTitle, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			playConf.EnableSubtitle = &disbase.BoolValue{Value: subTitle.SwitchVal}
		case v2.ConfType_LOSSLESS:
			lossless, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			cloudPlayConf.EnableLossLess = &disbase.BoolValue{Value: lossless.SwitchVal}
		case v2.ConfType_DOLBY:
			dolby, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			cloudPlayConf.EnableDolby = &disbase.BoolValue{Value: dolby.SwitchVal}
		case v2.ConfType_BACKGROUNDPLAY:
			background, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			cloudPlayConf.EnableBackground = &disbase.BoolValue{Value: background.SwitchVal}
		case v2.ConfType_PANORAMA:
			panorama, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			cloudPlayConf.EnablePanorama = &disbase.BoolValue{Value: panorama.SwitchVal}
		case v2.ConfType_SHAKE:
			shake, ok := v.ConfValue.Value.(*v2.ConfValue_SwitchVal)
			if !ok {
				return nil, errors.WithMessagef(ecode.RequestErr, "Failed to convert confvalue,conftype is(%d)", v.ConfType)
			}
			cloudPlayConf.EnableShake = &disbase.BoolValue{Value: shake.SwitchVal}
		default:
			log.Warn("Failed to match confType(%d)", v.ConfType)
			continue
		}
	}
	return convertProtoMessageToAnys(playConf, cloudPlayConf)
}

func convertProtoMessageToAnys(playConf *dp.PlayConfig, cloudPlayConf *dp.CloudPlayConfig) ([]*types.Any, error) {
	playConfAnyValue, err := playConf.Marshal()
	if err != nil {
		return nil, err
	}
	cloudConfAnyValue, err := cloudPlayConf.Marshal()
	if err != nil {
		return nil, err
	}
	anys := []*types.Any{
		{
			TypeUrl: _distributionPlayConf,
			Value:   playConfAnyValue,
		},
		{
			TypeUrl: _distributionCloudPlayConf,
			Value:   cloudConfAnyValue,
		},
	}
	return anys, nil
}

// 分离需要更新的 FieldValue 和 ConfValue
// FieldValue: 开关显影
// ConfValue: 配置的值/开关的状态
func separateEditedValue(confStates []*v2.PlayConfState) ([]*model.ConfValueEdit, []*v2.FieldValue) {
	var (
		confValueToSave  []*model.ConfValueEdit
		fieldValueToSave []*v2.FieldValue
	)
	for _, v := range confStates {
		if v == nil {
			continue
		}
		if v.FieldValue != nil {
			fieldValueToSave = append(fieldValueToSave, v.FieldValue)
		}
		if v.ConfValue != nil {
			ce := &model.ConfValueEdit{
				ConfType:  v.ConfType,
				ConfValue: v.ConfValue,
			}
			confValueToSave = append(confValueToSave, ce)
		}
	}
	return confValueToSave, fieldValueToSave
}

func abilityConfBoolValSetter(ctx context.Context, experiment Experiment) *v2.ConfValue {
	experiment.Exp(ctx)
	out := &v2.ConfValue{
		Value: &v2.ConfValue_SwitchVal{
			SwitchVal: experiment.GetResultAfterExp().(bool),
		},
	}
	return out
}

func abilityConfIntValSetter(ctx context.Context, experiment Experiment) *v2.ConfValue {
	experiment.Exp(ctx)
	out := &v2.ConfValue{
		Value: &v2.ConfValue_SelectedVal{
			SelectedVal: experiment.GetResultAfterExp().(int64),
		},
	}
	return out
}

func translateDistributionReply(in *distribution.GetUserPreferenceReply) (*dp.PlayConfig, *dp.CloudPlayConfig, error) {
	playConf := &dp.PlayConfig{}
	cloudPlayConf := &dp.CloudPlayConfig{}
	for _, v := range in.Value {
		switch v.TypeUrl {
		case _distributionPlayConf:
			if err := playConf.Unmarshal(v.Value); err != nil {
				return nil, nil, errors.WithMessagef(err, "Failed to Unmarshal play any(%s)", v.TypeUrl)
			}
		case _distributionCloudPlayConf:
			if err := cloudPlayConf.Unmarshal(v.Value); err != nil {
				return nil, nil, errors.WithMessagef(err, "Failed to Unmarshal cloud play any(%s)", v.TypeUrl)
			}
		default:
			return nil, nil, errors.WithMessagef(ecode.NothingFound, "Failed to match typeUrl(%s)", v.TypeUrl)
		}
	}
	return playConf, cloudPlayConf, nil
}
