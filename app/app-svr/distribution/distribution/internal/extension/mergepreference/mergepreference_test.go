package mergepreference

import (
	"testing"

	"go-gateway/app/app-svr/distribution/distribution/internal/extension/util"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/stretchr/testify/assert"
)

func init() {
	Init()
}

func TestMergePreference(t *testing.T) {

	t.Run("StructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.StructConfig")
		assert.True(t, ok)

		dst := dynamic.NewMessage(msgDesc)
		if err := dst.UnmarshalJSON([]byte(`{
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 },
			"field_int64": { "value": 3 },
			"field_uint64": { "value": 4 },
			"field_int32": { "value": 5 },
			"field_uint32": { "value": 6 },
			"field_bool": { "value": true },
			"field_string": { "value": "7" },
			"field_bytes": { "value": "OA==" }
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		MergePreference(dst, origin)

		t.Logf("dst: %s", util.MessageJSONify(dst))
		t.Logf("origin: %s", util.MessageJSONify(origin))

		assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))
	})

	t.Run("StructConfig-chunk", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.StructConfig")
		assert.True(t, ok)

		dst := dynamic.NewMessage(msgDesc)
		if err := dst.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 }
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 },
			"field_int64": { "value": 3 },
			"field_uint64": { "value": 4 },
			"field_int32": { "value": 5 },
			"field_uint32": { "value": 6 },
			"field_bool": { "value": true },
			"field_string": { "value": "7" },
			"field_bytes": { "value": "OA==" }
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		MergePreference(dst, origin)

		t.Logf("dst: %s", util.MessageJSONify(dst))
		t.Logf("origin: %s", util.MessageJSONify(origin))

		assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))
	})

	t.Run("StructConfig-diffchunk", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.StructConfig")
		assert.True(t, ok)

		dst := dynamic.NewMessage(msgDesc)
		if err := dst.UnmarshalJSON([]byte(`{
			"field_double": { "value": 11.0 },
			"field_float": { "value": 12.0 }
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 },
			"field_int64": { "value": 3 },
			"field_uint64": { "value": 4 },
			"field_int32": { "value": 5 },
			"field_uint32": { "value": 6 },
			"field_bool": { "value": true },
			"field_string": { "value": "7" },
			"field_bytes": { "value": "OA==" }
	}`)); err != nil {
			assert.NoError(t, err)
			return
		}

		MergePreference(dst, origin)

		t.Logf("dst: %s", util.MessageJSONify(dst))
		t.Logf("origin: %s", util.MessageJSONify(origin))

		for _, field := range msgDesc.GetFields() {
			if field.GetName() == "field_double" {
				assert.Equal(t, float64(11.0), dst.GetField(field).(*dynamic.Message).GetFieldByName("value"))
				continue
			}
			if field.GetName() == "field_float" {
				assert.Equal(t, float32(12.0), dst.GetField(field).(*dynamic.Message).GetFieldByName("value"))
				continue
			}
			assert.Equal(t,
				origin.GetField(field).(*dynamic.Message).GetFieldByName("value"),
				origin.GetField(field).(*dynamic.Message).GetFieldByName("value"))
		}
	})

	t.Run("RepeatedStructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.RepeatedStructConfig")
		assert.True(t, ok)

		dst := dynamic.NewMessage(msgDesc)
		if err := dst.UnmarshalJSON([]byte(`{
			"field_double": [],
			"field_float": [],
			"field_int64": [],
			"field_uint64": [],
			"field_int32": [],
			"field_uint32": [],
			"field_bool": [],
			"field_string": [],
			"field_bytes": []
		}
	`)); err != nil {
			assert.NoError(t, err)
			return
		}

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
			"field_double": [{"value":1.0}, {}],
			"field_float": [{"value":2.0}, {}],
			"field_int64": [{"value":3}, {}],
			"field_uint64": [{"value":4}, {}],
			"field_int32": [{"value":5}, {}],
			"field_uint32": [{"value":6}, {}],
			"field_bool": [{"value":true}, {}],
			"field_string": [{"value":"7"}, {}],
			"field_bytes": [{"value": "OA=="}, {}]
		}				
	`)); err != nil {
			assert.NoError(t, err)
			return
		}

		MergePreference(dst, origin)

		t.Logf("dst: %s", util.MessageJSONify(dst))
		t.Logf("origin: %s", util.MessageJSONify(origin))

		for _, field := range msgDesc.GetFields() {
			assert.Len(t, dst.GetField(field), 0)
		}
	})

	t.Run("EmbedStructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.EmbedStructConfig")
		assert.True(t, ok)

		fixtureBytes := []byte(`{
		}
`)
		originFixtureBytes := []byte(`{
	"field_struct": {
		"field_double": { "value": 11.0 },
		"field_float": { "value": 12.0 },
		"field_int64": { "value": 13 },
		"field_uint64": { "value": 14 },
		"field_int32": { "value": 15 },
		"field_uint32": { "value": 16 },
		"field_bool": { "value": false },
		"field_string": { "value": "17" },
		"field_bytes": { "value": "MTg=" }
	},
	"field_default_struct": {
		"field_double": { "value": 11.0 },
		"field_float": { "value": 12.0 },
		"field_int64": { "value": 13 },
		"field_uint64": { "value": 14 },
		"field_int32": { "value": 15 },
		"field_uint32": { "value": 16 },
		"field_bool": { "value": false },
		"field_string": { "value": "17" },
		"field_bytes": { "value": "MTg=" }
	},
	"field_repeated_struct": {
		"field_double": [{ "value": 11.0 }, {}],
		"field_float": [{ "value": 12.0 }, {}],
		"field_int64": [{ "value": 13 }, {}],
		"field_uint64": [{ "value": 14 }, {}],
		"field_int32": [{ "value": 15 }, {}],
		"field_uint32": [{ "value": 16 }, {}],
		"field_bool": [{ "value": false }, {}],
		"field_string": [{ "value": "17" }, {}],
		"field_bytes": [{ "value": "MTg=" }, {}]
	},
	"field_repeated_default_struct_slice": {
		"field_double": [{ "value": 11.0 }, {}],
		"field_float": [{ "value": 12.0 }, {}],
		"field_int64": [{ "value": 13 }, {}],
		"field_uint64": [{ "value": 14 }, {}],
		"field_int32": [{ "value": 15 }, {}],
		"field_uint32": [{ "value": 16 }, {}],
		"field_bool": [{ "value": false }, {}],
		"field_string": [{ "value": "17" }, {}],
		"field_bytes": [{ "value": "MTg=" }, {}]
	},
	"field_repeated_repeated_default_struct_slice": [{
		"field_double": [{ "value": 11.0 }, {}],
		"field_float": [{ "value": 12.0 }, {}],
		"field_int64": [{ "value": 13 }, {}],
		"field_uint64": [{ "value": 14 }, {}],
		"field_int32": [{ "value": 15 }, {}],
		"field_uint32": [{ "value": 16 }, {}],
		"field_bool": [{ "value": false }, {}],
		"field_string": [{ "value": "17" }, {}],
		"field_bytes": [{ "value": "MTg=" }, {}]
	}, {}]
}
`)

		dst := dynamic.NewMessage(msgDesc)
		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON(originFixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}
		if err := dst.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}

		MergePreference(dst, origin)

		t.Run("field_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(origin, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)
			dst, err := util.GetFieldAsDynamicMessage(dst, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)

			t.Logf("dst: %s", util.MessageJSONify(dst))
			t.Logf("origin: %s", util.MessageJSONify(origin))

			assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))

			for _, field := range msgDesc.FindFieldByName("field_struct").GetMessageType().GetFields() {
				assert.Equal(t, origin.GetField(field), dst.GetField(field))
			}
		})

		t.Run("field_default_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(origin, msgDesc.FindFieldByName("field_default_struct"))
			assert.NoError(t, err)
			dst, err := util.GetFieldAsDynamicMessage(dst, msgDesc.FindFieldByName("field_default_struct"))
			assert.NoError(t, err)

			t.Logf("dst: %s", util.MessageJSONify(dst))
			t.Logf("origin: %s", util.MessageJSONify(origin))

			assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))

			for _, field := range msgDesc.FindFieldByName("field_default_struct").GetMessageType().GetFields() {
				assert.Equal(t, origin.GetField(field), dst.GetField(field))
			}
		})

		t.Run("field_repeated_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(origin, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)
			dst, err := util.GetFieldAsDynamicMessage(dst, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)

			t.Logf("dst: %s", util.MessageJSONify(dst))
			t.Logf("origin: %s", util.MessageJSONify(origin))

			for _, field := range msgDesc.FindFieldByName("field_repeated_struct").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, field)
				assert.NoError(t, err)
				assert.Len(t, dmSlice, 2)
				assert.Equal(t, origin.GetField(field), dst.GetField(field))
				assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))
			}
		})

		t.Run("field_repeated_default_struct_slice", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(origin, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)
			dst, err := util.GetFieldAsDynamicMessage(dst, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)

			t.Logf("dst: %s", util.MessageJSONify(dst))
			t.Logf("origin: %s", util.MessageJSONify(origin))

			for _, field := range msgDesc.FindFieldByName("field_repeated_default_struct_slice").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, field)
				assert.NoError(t, err)
				assert.Len(t, dmSlice, 2)
				assert.Equal(t, origin.GetField(field), dst.GetField(field))
				assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(dst))
			}
		})

		t.Run("field_repeated_repeated_default_struct_slice", func(t *testing.T) {
			originDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(origin, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, originDMSlice, 2)
			dstDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, dstDMSlice, 0)
		})
	})

}
