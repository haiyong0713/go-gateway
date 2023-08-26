package multipletusflag

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"
	tmv "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	"go-common/component/metadata/device"

	"github.com/stretchr/testify/assert"
)

type sessionContextImpl struct {
	mid          int64
	device       device.Device
	extraContext map[string]string
}

func (s *sessionContextImpl) Device() device.Device {
	return s.device
}

func (s *sessionContextImpl) Mid() int64 {
	return s.mid
}

func (s *sessionContextImpl) ExtraContext() map[string]string {
	return s.extraContext
}

func (s *sessionContextImpl) ExtraContextValue(key string) (string, bool) {
	v, ok := s.extraContext[key]
	return v, ok
}

func TestChooseVersionByDevice(t *testing.T) {
	tests := []struct {
		sessionContextImpl *sessionContextImpl
		versionInfos       []*tmv.VersionInfo
		expectedVersion    string
	}{
		{
			sessionContextImpl: &sessionContextImpl{
				device: device.Device{
					Build:      100,
					RawMobiApp: "android",
					Device:     "phone",
				},
			},
			versionInfos: []*tmv.VersionInfo{
				{
					ConfigVersion: "v1.0",
				},
				{
					ConfigVersion: "v2.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    99,
						},
					},
				},
				{
					ConfigVersion: "v3.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    101,
						},
					},
				},
			},
			expectedVersion: "v2.0",
		},
		{
			sessionContextImpl: &sessionContextImpl{
				device: device.Device{
					Build:      100,
					RawMobiApp: "android",
					Device:     "phone",
				},
			},
			versionInfos: []*tmv.VersionInfo{
				{
					ConfigVersion: "v1.0",
				},
				{
					ConfigVersion: "v2.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    101,
						},
					},
				},
				{
					ConfigVersion: "v3.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    200,
						},
					},
				},
			},
			expectedVersion: "v1.0",
		},
		{
			sessionContextImpl: &sessionContextImpl{
				device: device.Device{
					Build:      1000,
					RawMobiApp: "android",
					Device:     "phone",
				},
			},
			versionInfos: []*tmv.VersionInfo{
				{
					ConfigVersion: "v1.0",
				},
				{
					ConfigVersion: "v2.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    99,
						},
					},
				},
				{
					ConfigVersion: "v3.0",
					BuildLimit: []*tmv.BuildLimit{
						{
							Plat:     0,
							Operator: tmv.GT,
							Build:    101,
						},
					},
				},
			},
			expectedVersion: "v3.0",
		},
	}
	for _, v := range tests {
		sctx := sessioncontext.NewContext(context.Background(), v.sessionContextImpl)
		versionInfo := chooseVersionByDevice(sctx, v.versionInfos)
		assert.Equal(t, v.expectedVersion, versionInfo.ConfigVersion)
	}
}
