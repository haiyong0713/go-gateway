package service

import (
	"fmt"
	"testing"
	"time"

	"go-common/component/metadata/device"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/defaultvalue"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/lastmodified"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/stretchr/testify/assert"
)

func TestSetPreferenceValueLastModified(t *testing.T) {
	meta, ok := preferenceproto.TryGetPreference("bilibili.app.distribution.pegasus.v1.PegasusDeviceConfig")
	assert.True(t, ok)

	dm := dynamic.NewMessage(meta.ProtoDesc)
	lastmodified.SetPreferenceValueLastModified(dm, time.Now())

	pcvDesc, ok := preferenceproto.TryGetMessage("bilibili.app.distribution.pegasus.v1.PegasusColumnValue")
	assert.True(t, ok)
	pcv := dynamic.NewMessage(pcvDesc)
	_ = pcv.UnmarshalJSON([]byte(`{
    "value": {
        "value": 1
    }
}`))
	dm.SetFieldByName("column", pcv)

	lastmodified.SetPreferenceValueLastModified(dm, time.Now())

	pcvJSON, _ := pcv.MarshalJSON()
	t.Logf("field with last modified: %s", string(pcvJSON))

	dmJSON, _ := dm.MarshalJSON()
	t.Logf("message with last modified: %s", string(dmJSON))
}

func TestInitializeWithDefaultValue(t *testing.T) {
	defaultvalue.Init()

	meta, ok := preferenceproto.TryGetPreference("bilibili.app.distribution.pegasus.v1.PegasusDeviceConfig")
	assert.True(t, ok)

	dm := dynamic.NewMessage(meta.ProtoDesc)
	lastmodified.SetPreferenceValueLastModified(dm, time.Now())

	pcvDesc, ok := preferenceproto.TryGetMessage("bilibili.app.distribution.pegasus.v1.PegasusColumnValue")
	assert.True(t, ok)
	pcv := dynamic.NewMessage(pcvDesc)
	_ = pcv.UnmarshalJSON([]byte(`{
    "value": {
        "value": 1
    }
}`))
	dm.SetFieldByName("column", pcv)

	lastmodified.SetPreferenceValueLastModified(dm, time.Now())
	defaultvalue.InitializeWithDefaultValue(dm)

	pcvJSON, _ := pcv.MarshalJSON()
	t.Logf("field with last modified: %s", string(pcvJSON))

	dmJSON, _ := dm.MarshalJSON()
	t.Logf("message with last modified: %s", string(dmJSON))
}

type testSessionContext struct {
	mid          int64
	device       device.Device
	extraContext map[string]string
}

func (s *testSessionContext) Device() device.Device {
	return s.device
}

func (s *testSessionContext) Mid() int64 {
	return s.mid
}

func (s *testSessionContext) ExtraContext() map[string]string {
	return s.extraContext
}

func (s *testSessionContext) ExtraContextValue(key string) (string, bool) {
	v, ok := s.extraContext[key]
	return v, ok
}

func TestPreferenceMeta_KeyBuilder(t *testing.T) {
	meta, ok := preferenceproto.TryGetPreference("bilibili.app.distribution.play.v1.SpecificPlayConfig")
	assert.True(t, ok)
	tss := []*testSessionContext{
		{
			mid: 1234,
			device: device.Device{
				Buvid:   "ABCD1243",
				FpLocal: "11qhduiaf",
			},
			extraContext: map[string]string{
				"aid": "1234",
				"cid": "1234",
			},
		},
		{
			mid: 3456,
			device: device.Device{
				Buvid:   "jaksdjien",
				FpLocal: "11qhduiaf",
			},
			extraContext: map[string]string{
				"aid": "2231255",
				"cid": "334566",
			},
		},
		{
			mid: 3456,
			device: device.Device{
				Buvid:   "jaksdjien",
				FpLocal: "11qhduiaf",
			},
		},
	}
	for i, ts := range tss {
		key := meta.KeyBuilder()(ts)
		if i < len(tss)-1 {
			aidValue, ok := ts.ExtraContextValue("aid")
			assert.True(t, ok)
			cidValue, ok := ts.ExtraContextValue("cid")
			assert.True(t, ok)
			assert.Equal(t, key, fmt.Sprintf("{buvid:%s}/fp_local:%s/mid:%d/aid:%s/cid:%s/bilibili.app.distribution.play.v1.SpecificPlayConfig", ts.Device().Buvid, ts.Device().FpLocal, ts.Mid(), aidValue, cidValue))
			continue
		}
		assert.Equal(t, key, fmt.Sprintf("{buvid:%s}/fp_local:%s/mid:%d/bilibili.app.distribution.play.v1.SpecificPlayConfig", ts.Device().Buvid, ts.Device().FpLocal, ts.Mid()))
	}
}
