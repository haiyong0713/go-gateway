package dynamicV2

import (
	"encoding/json"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"reflect"
	"testing"
)

func TestService_descSearchWordProc(t *testing.T) {
	type args struct {
		desc   string
		dynCtx *mdlv2.DynamicContext
	}
	tests := []struct {
		name string
		args args
		want []*api.Description
	}{
		{
			name: "searchWord1",
			args: args{
				desc: "哔哩哔哩动画uat测试",
				dynCtx: &mdlv2.DynamicContext{
					SearchWords: []string{"uat", "哔哩"},
				},
			},
			want: []*api.Description{
				{Text: "哔哩", Type: api.DescType_desc_type_search_word},
				{Text: "哔哩", Type: api.DescType_desc_type_search_word},
				{Text: "动画", Type: api.DescType_desc_type_text},
				{Text: "uat", Type: api.DescType_desc_type_search_word},
				{Text: "测试", Type: api.DescType_desc_type_text},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := new(Service)
			if got := s.descSearchWordProc(tt.args.desc, tt.args.dynCtx); !reflect.DeepEqual(got, tt.want) {
				getJsonData, _ := json.Marshal(got)
				wantJsonData, _ := json.Marshal(tt.want)
				t.Errorf("descSearchWordProc() = %s,\n want %s", string(getJsonData), string(wantJsonData))
			}
		})
	}
}
