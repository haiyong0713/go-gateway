package conf

import (
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestTomlUnmarshal(t *testing.T) {
	data := []struct {
		Name     string
		Raw      string
		Expected *Config
	}{
		{
			Name: "MapConfig解析测试",
			Raw: `
			[AppAuth]
				[AppAuth.AuthInfo.campus-billboard-h5]
					AppName = "campus-billboard-h5"
					AppKey = "testKey"
`,
			Expected: &Config{
				AppAuth: &AppAuth{
					AuthInfo: map[string]*AppAuthInfo{
						"campus-billboard-h5": {
							AppName: "campus-billboard-h5",
							AppKey:  "testKey",
						},
					},
				},
			},
		},
	}
	for _, tc := range data {
		c := &Config{}
		err := toml.Unmarshal([]byte(tc.Raw), c)
		if err != nil {
			t.Fatalf("Test %s failed: %v", tc.Name, err)
		}
		if !reflect.DeepEqual(tc.Expected, c) {
			t.Errorf("%s: expecting (%+v)\ngot (%+v)", tc.Name, tc.Expected, c)
		}
	}
}

func TestAppAuthGW(t *testing.T) {
	ai := &AppAuth{
		AuthInfo: map[string]*AppAuthInfo{
			"campus-billboard-h5": {
				AppName: "campus-billboard-h5",
				AppKey:  "testKey",
			},
		},
	}
	md := metadata.MD{
		"x-bili-internal-gw-auth": []string{"campus-billboard-h5 testKey"},
	}
	if ok, ret := ai.authGW(md, &grpc.UnaryServerInfo{FullMethod: "/bilibili.app.dynamic.v2.Dynamic/CampusBillboardInternal"}); ok {
		t.Log("PASS")
	} else {
		t.Errorf("got %v", ret)
	}
}
