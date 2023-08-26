package dynamicV2

import (
	"encoding/json"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"reflect"
	"testing"
)

func TestService_titleSearchWordProc(t *testing.T) {
	type args struct {
		title  string
		dynCtx *mdlv2.DynamicContext
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "searchWord1",
			args: args{
				title: "哔哩哔哩动画uat测试",
				dynCtx: &mdlv2.DynamicContext{
					SearchWords: []string{"哔哩", "uat"},
				},
			},
			want: []string{
				"<font color=\"#fb7299\">哔哩</font>",
				"<font color=\"#fb7299\">哔哩</font>",
				"动画",
				"<font color=\"#fb7299\">uat</font>",
				"测试",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{}
			if got := s.titleSearchWordProc(tt.args.title, tt.args.dynCtx); !reflect.DeepEqual(got, tt.want) {
				getJsonData, _ := json.Marshal(got)
				wantJsonData, _ := json.Marshal(tt.want)
				t.Errorf("descSearchWordProc() = %s,\n want %s", string(getJsonData), string(wantJsonData))
			}
		})
	}
}
