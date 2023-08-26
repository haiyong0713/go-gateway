package defaultvalue

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

func TestInitializeWithDefaultValue(t *testing.T) {
	t.Run("StructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.StructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 },
			"field_int64": { "value": 3 },
			"field_uint64": { "value": 4 },
			"field_int32": { "value": 5 },
			"field_uint32": { "value": 6 },
			"field_bool": { "value": true },
			"field_string": { "value": "7" },
			"field_bytes": { "value": "OA==" }
	}	
	`)); err != nil {
			assert.NoError(t, err)
			return
		}
		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)
		assert.Equal(t, originString, afterString)
	})

	t.Run("WithDefaultValueStructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.WithDefaultValueStructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
			"field_double": { "value": 1.0 },
			"field_float": { "value": 2.0 },
			"field_int64": { "value": 3 },
			"field_uint64": { "value": 4 },
			"field_int32": { "value": 5 },
			"field_uint32": { "value": 6 },
			"field_bool": { "value": true },
			"field_string": { "value": "7" },
			"field_bytes": { "value": "OA==" }
	}
	`)); err != nil {
			assert.NoError(t, err)
			return
		}

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		for _, field := range msgDesc.GetFields() {
			fieldDM, err := util.GetFieldAsDynamicMessage(dm, field)
			assert.NoError(t, err)
			assert.Equalf(t, fieldDM.GetFieldByName("value"), fieldDM.GetFieldByName("default_value"), "fieldDM: %+v", fieldDM)
		}
	})

	t.Run("RepeatedStructConfig-empty", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.RepeatedStructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
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

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		assert.Equal(t, originString, afterString)
	})

	t.Run("RepeatedStructConfig-emptyitem", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.RepeatedStructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
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

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		assert.Equal(t, originString, afterString)
	})

	t.Run("RepeatedWithDefaultValueStructConfig-empty", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.RepeatedWithDefaultValueStructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
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

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		assert.Equal(t, originString, afterString)
	})

	t.Run("RepeatedWithDefaultValueStructConfig-emptyitem", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.RepeatedWithDefaultValueStructConfig")
		assert.True(t, ok)

		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON([]byte(`{
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

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		for _, field := range msgDesc.GetFields() {
			fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, field)
			assert.NoError(t, err)
			assert.Equalf(t, fieldDMSlice[0].GetFieldByName("value"), fieldDMSlice[0].GetFieldByName("default_value"), "fieldDMSlice[0]: %+v", fieldDMSlice[0])
			assert.Equalf(t, fieldDMSlice[0].GetFieldByName("default_value"), fieldDMSlice[1].GetFieldByName("default_value"), "fieldDMSlice: %+v", fieldDMSlice)
		}
	})

	t.Run("EmbedStructConfig", func(t *testing.T) {
		msgDesc, ok := preferenceproto.GlobalRegistry.TryGetMessage("bilibili.app.distribution.fixture.v1.EmbedStructConfig")
		assert.True(t, ok)

		fixtureBytes := []byte(`{
			"field_struct": {
				"field_double": { "value": 1.0 },
				"field_float": { "value": 2.0 },
				"field_int64": { "value": 3 },
				"field_uint64": { "value": 4 },
				"field_int32": { "value": 5 },
				"field_uint32": { "value": 6 },
				"field_bool": { "value": true },
				"field_string": { "value": "7" },
				"field_bytes": { "value": "OA==" }
			},
			"field_default_struct": {
				"field_double": { "value": 1.0 },
				"field_float": { "value": 2.0 },
				"field_int64": { "value": 3 },
				"field_uint64": { "value": 4 },
				"field_int32": { "value": 5 },
				"field_uint32": { "value": 6 },
				"field_bool": { "value": true },
				"field_string": { "value": "7" },
				"field_bytes": { "value": "OA==" }
			},
			"field_repeated_struct": {
				"field_double": [{ "value": 1.0 }, {}],
				"field_float": [{ "value": 2.0 }, {}],
				"field_int64": [{ "value": 3 }, {}],
				"field_uint64": [{ "value": 4 }, {}],
				"field_int32": [{ "value": 5 }, {}],
				"field_uint32": [{ "value": 6 }, {}],
				"field_bool": [{ "value": true }, {}],
				"field_string": [{ "value": "7" }, {}],
				"field_bytes": [{ "value": "OA==" }, {}]
			},
			"field_repeated_default_struct_slice": {
				"field_double": [{ "value": 1.0 }, {}],
				"field_float": [{ "value": 2.0 }, {}],
				"field_int64": [{ "value": 3 }, {}],
				"field_uint64": [{ "value": 4 }, {}],
				"field_int32": [{ "value": 5 }, {}],
				"field_uint32": [{ "value": 6 }, {}],
				"field_bool": [{ "value": true }, {}],
				"field_string": [{ "value": "7" }, {}],
				"field_bytes": [{ "value": "OA==" }, {}]
			},
			"field_repeated_repeated_default_struct_slice": [{
				"field_double": [{ "value": 1.0 }, {}],
				"field_float": [{ "value": 2.0 }, {}],
				"field_int64": [{ "value": 3 }, {}],
				"field_uint64": [{ "value": 4 }, {}],
				"field_int32": [{ "value": 5 }, {}],
				"field_uint32": [{ "value": 6 }, {}],
				"field_bool": [{ "value": true }, {}],
				"field_string": [{ "value": "7" }, {}],
				"field_bytes": [{ "value": "OA==" }, {}]
			}, {}]
		}
`)

		originDM := dynamic.NewMessage(msgDesc)
		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}
		if err := originDM.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}

		originString := util.MessageJSONify(dm)
		t.Logf("origin: %s", originString)

		InitializeWithDefaultValue(dm)

		afterString := util.MessageJSONify(dm)
		t.Logf("after: %s", afterString)

		t.Run("field_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(after))
		})

		t.Run("field_default_struct", func(t *testing.T) {
			dm, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_default_struct"))
			assert.NoError(t, err)

			for _, field := range dm.GetMessageDescriptor().GetFields() {
				fieldDM, err := util.GetFieldAsDynamicMessage(dm, field)
				assert.NoError(t, err)
				assert.Equalf(t, fieldDM.GetFieldByName("value"), fieldDM.GetFieldByName("default_value"), "fieldDM: %+v", fieldDM)
			}
		})

		t.Run("field_repeated_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(after))
		})

		t.Run("field_repeated_default_struct_slice", func(t *testing.T) {
			dm, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)

			for _, field := range dm.GetMessageDescriptor().GetFields() {
				fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, field)
				assert.NoError(t, err)
				assert.Equalf(t, fieldDMSlice[0].GetFieldByName("value"), fieldDMSlice[0].GetFieldByName("default_value"), "fieldDMSlice[0]: %+v", fieldDMSlice[0])
				assert.Equalf(t, fieldDMSlice[0].GetFieldByName("default_value"), fieldDMSlice[1].GetFieldByName("default_value"), "fieldDMSlice: %+v", fieldDMSlice)
			}
		})

		t.Run("field_repeated_repeated_default_struct_slice", func(t *testing.T) {
			afterDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, afterDMSlice, 2)

			for _, field := range afterDMSlice[0].GetMessageDescriptor().GetFields() {
				fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(afterDMSlice[0], field)
				assert.NoError(t, err)
				assert.Equalf(t, fieldDMSlice[0].GetFieldByName("value"), fieldDMSlice[0].GetFieldByName("default_value"), "fieldDMSlice[0]: %+v", fieldDMSlice[0])
				assert.Equalf(t, fieldDMSlice[0].GetFieldByName("default_value"), fieldDMSlice[1].GetFieldByName("default_value"), "fieldDMSlice: %+v", fieldDMSlice)
			}

			beforeDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(beforeDMSlice[1]), util.MessageJSONify(afterDMSlice[1]))
		})

	})
}
