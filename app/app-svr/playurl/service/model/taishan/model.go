package taishan

import v2 "go-gateway/app/app-svr/playurl/service/api/v2"

func FormatConf(in *PlayConf) (out *v2.CloudConf) {
	out = &v2.CloudConf{Show: in.Show}
	// 处理
	var tmpField *v2.FieldValue
	if in.FieldValue != nil {
		switch tval := in.FieldValue.Value.(type) {
		case *FieldValue_Switch:
			tmpField = &v2.FieldValue{Value: &v2.FieldValue_Switch{Switch: tval.Switch}}
		}
	}
	out.FieldValue = tmpField
	return
}

func ConvertConf(in *v2.PlayConfState) (out *PlayConf) {
	// 处理
	var tmpField *FieldValue
	if in.FieldValue != nil {
		switch tval := in.FieldValue.Value.(type) {
		case *v2.FieldValue_Switch:
			tmpField = &FieldValue{Value: &FieldValue_Switch{Switch: tval.Switch}}
		}
	}
	out = &PlayConf{Show: in.Show, FieldValue: tmpField}
	return
}

func DefaultFormatConf() (out *v2.CloudConf) {
	out = &v2.CloudConf{Show: true}
	out.FieldValue = &v2.FieldValue{Value: &v2.FieldValue_Switch{Switch: true}}
	return
}

func GrayDefault(isShow bool) []*v2.PlayConfState {
	grayConf := []v2.ConfType{
		v2.ConfType_DISLIKE,    //踩
		v2.ConfType_COIN,       //投币
		v2.ConfType_ELEC,       //充电
		v2.ConfType_SCREENSHOT, //截图/gif
	}
	var rly []*v2.PlayConfState
	for _, v := range grayConf {
		rly = append(rly, &v2.PlayConfState{ConfType: v, Show: isShow, FieldValue: &v2.FieldValue{Value: &v2.FieldValue_Switch{Switch: isShow}}})
	}
	return rly
}
