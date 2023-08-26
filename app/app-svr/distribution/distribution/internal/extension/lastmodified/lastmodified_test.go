package lastmodified

import (
	"testing"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/internal/extension/util"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/stretchr/testify/assert"
)

func init() {
	Init()
}

func TestSetPreferenceValueLastModified(t *testing.T) {
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

		now := time.Now()
		t.Logf("before: %s", util.MessageJSONify(dm))
		SetPreferenceValueLastModified(dm, now)
		t.Logf("after: %s", util.MessageJSONify(dm))

		for _, field := range msgDesc.GetFields() {
			fieldV, err := util.GetFieldAsDynamicMessage(dm, field)
			assert.NoError(t, err)
			assert.Equal(t, now.Unix(), fieldV.GetFieldByName("last_modified"))
		}
	})
	t.Run("StructConfig-compare", func(t *testing.T) {
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

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
		"field_double": { "value": 11.0,"last_modified":1234 },
		"field_float": { "value": 12.0,"last_modified":1234 },
		"field_int64": { "value": 13,"last_modified":1234 },
		"field_uint64": { "value": 14,"last_modified":1234 },
		"field_int32": { "value": 15,"last_modified":1234 },
		"field_uint32": { "value": 16,"last_modified":1234 },
		"field_bool": { "value": false,"last_modified":1234 },
		"field_string": { "value": "17","last_modified":1234 },
		"field_bytes": { "value": "MTg=","last_modified":1234 }
}
`)); err != nil {
			assert.NoError(t, err)
			return
		}

		now := time.Now()
		t.Logf("before: %s %s", util.MessageJSONify(dm), util.MessageJSONify(origin))
		CompareAndSetPreferenceValueLastModified(dm, origin, now)
		t.Logf("after: %s %s", util.MessageJSONify(dm), util.MessageJSONify(origin))

		for _, field := range msgDesc.GetFields() {
			fieldV, err := util.GetFieldAsDynamicMessage(dm, field)
			assert.NoError(t, err)
			assert.Equal(t, now.Unix(), fieldV.GetFieldByName("last_modified"))
		}
	})
	t.Run("StructConfig-compareequal", func(t *testing.T) {
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

		origin := dynamic.NewMessage(msgDesc)
		if err := origin.UnmarshalJSON([]byte(`{
		"field_double": { "value": 1.0, "last_modified":1234 },
		"field_float": { "value": 2.0,"last_modified":1234 },
		"field_int64": { "value": 3,"last_modified":1234 },
		"field_uint64": { "value": 4,"last_modified":1234 },
		"field_int32": { "value": 5,"last_modified":1234 },
		"field_uint32": { "value": 6,"last_modified":1234 },
		"field_bool": { "value": true,"last_modified":1234 },
		"field_string": { "value": "7","last_modified":1234 },
		"field_bytes": { "value": "OA==","last_modified":1234 }
}
`)); err != nil {
			assert.NoError(t, err)
			return
		}

		now := time.Now()
		t.Logf("before: %s %s", util.MessageJSONify(dm), util.MessageJSONify(origin))
		CompareAndSetPreferenceValueLastModified(dm, origin, now)
		t.Logf("after: %s %s", util.MessageJSONify(dm), util.MessageJSONify(origin))

		for _, field := range msgDesc.GetFields() {
			fieldV, err := util.GetFieldAsDynamicMessage(dm, field)
			assert.NoError(t, err)
			assert.Equal(t, int64(1234), fieldV.GetFieldByName("last_modified"))
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

		originDM := dynamic.NewMessage(msgDesc)
		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}
		if err := originDM.UnmarshalJSON(originFixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}

		now := time.Now()
		CompareAndSetPreferenceValueLastModified(dm, originDM, now)

		t.Run("field_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)

			t.Logf("origin: %s\nafter: %s", util.MessageJSONify(origin), util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_struct").GetMessageType().GetFields() {
				fieldV := dm.GetFieldByName("field_struct").(*dynamic.Message).GetField(field).(*dynamic.Message)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), fieldV.GetFieldByName("last_modified"))
			}
		})

		t.Run("field_default_struct", func(t *testing.T) {
			dm, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_default_struct"))
			assert.NoError(t, err)

			for _, field := range dm.GetMessageDescriptor().GetFields() {
				fieldDM, err := util.GetFieldAsDynamicMessage(dm, field)
				assert.NoError(t, err)
				assert.Equalf(t, now.Unix(), fieldDM.GetFieldByName("last_modified"), "fieldDM: %+v", fieldDM)
			}
		})

		t.Run("field_repeated_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)

			t.Logf("origin: %s\nafter: %s", util.MessageJSONify(origin), util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_repeated_struct").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm.GetFieldByName("field_repeated_struct").(*dynamic.Message), field)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), dmSlice[0].GetFieldByName("last_modified"))
				assert.Equal(t, int64(0), dmSlice[1].GetFieldByName("last_modified"))
			}
		})

		t.Run("field_repeated_default_struct_slice", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)

			t.Logf("origin: %s\nafter: %s", util.MessageJSONify(origin), util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_repeated_default_struct_slice").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm.GetFieldByName("field_repeated_default_struct_slice").(*dynamic.Message), field)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), dmSlice[0].GetFieldByName("last_modified"))
				assert.Equal(t, int64(0), dmSlice[1].GetFieldByName("last_modified"))
			}
		})

		t.Run("field_repeated_repeated_default_struct_slice", func(t *testing.T) {
			afterDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, afterDMSlice, 2)

			for _, field := range afterDMSlice[0].GetMessageDescriptor().GetFields() {
				fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(afterDMSlice[0], field)
				assert.NoError(t, err)
				assert.Equalf(t, now.Unix(), fieldDMSlice[0].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
				assert.Equalf(t, int64(0), fieldDMSlice[1].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
			}

			beforeDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(beforeDMSlice[1]), util.MessageJSONify(afterDMSlice[1]))
		})
	})

	t.Run("EmbedStructConfig-comparequal", func(t *testing.T) {
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
		originFixtureBytes := []byte(`{
			"field_struct": {
				"field_double": { "value": 1.0,"last_modified":1234 },
				"field_float": { "value": 2.0,"last_modified":1234 },
				"field_int64": { "value": 3,"last_modified":1234 },
				"field_uint64": { "value": 4,"last_modified":1234 },
				"field_int32": { "value": 5,"last_modified":1234 },
				"field_uint32": { "value": 6,"last_modified":1234 },
				"field_bool": { "value": true,"last_modified":1234 },
				"field_string": { "value": "7","last_modified":1234 },
				"field_bytes": { "value": "OA==","last_modified":1234 }
			},
			"field_default_struct": {
				"field_double": { "value": 1.0,"last_modified":1234 },
				"field_float": { "value": 2.0,"last_modified":1234 },
				"field_int64": { "value": 3,"last_modified":1234 },
				"field_uint64": { "value": 4,"last_modified":1234 },
				"field_int32": { "value": 5,"last_modified":1234 },
				"field_uint32": { "value": 6,"last_modified":1234 },
				"field_bool": { "value": true,"last_modified":1234 },
				"field_string": { "value": "7","last_modified":1234 },
				"field_bytes": { "value": "OA==","last_modified":1234 }
			},
			"field_repeated_struct": {
				"field_double": [{ "value": 1.0,"last_modified":1234 }, {}],
				"field_float": [{ "value": 2.0,"last_modified":1234 }, {}],
				"field_int64": [{ "value": 3,"last_modified":1234 }, {}],
				"field_uint64": [{ "value": 4,"last_modified":1234 }, {}],
				"field_int32": [{ "value": 5,"last_modified":1234 }, {}],
				"field_uint32": [{ "value": 6,"last_modified":1234 }, {}],
				"field_bool": [{ "value": true,"last_modified":1234 }, {}],
				"field_string": [{ "value": "7","last_modified":1234 }, {}],
				"field_bytes": [{ "value": "OA==","last_modified":1234 }, {}]
			},
			"field_repeated_default_struct_slice": {
				"field_double": [{ "value": 1.0,"last_modified":1234 }, {}],
				"field_float": [{ "value": 2.0,"last_modified":1234 }, {}],
				"field_int64": [{ "value": 3,"last_modified":1234 }, {}],
				"field_uint64": [{ "value": 4,"last_modified":1234 }, {}],
				"field_int32": [{ "value": 5,"last_modified":1234 }, {}],
				"field_uint32": [{ "value": 6,"last_modified":1234 }, {}],
				"field_bool": [{ "value": true,"last_modified":1234 }, {}],
				"field_string": [{ "value": "7","last_modified":1234 }, {}],
				"field_bytes": [{ "value": "OA==","last_modified":1234 }, {}]
			},
			"field_repeated_repeated_default_struct_slice": [{
				"field_double": [{ "value": 1.0,"last_modified":1234 }, {}],
				"field_float": [{ "value": 2.0,"last_modified":1234 }, {}],
				"field_int64": [{ "value": 3,"last_modified":1234 }, {}],
				"field_uint64": [{ "value": 4,"last_modified":1234 }, {}],
				"field_int32": [{ "value": 5,"last_modified":1234 }, {}],
				"field_uint32": [{ "value": 6,"last_modified":1234 }, {}],
				"field_bool": [{ "value": true,"last_modified":1234 }, {}],
				"field_string": [{ "value": "7","last_modified":1234 }, {}],
				"field_bytes": [{ "value": "OA==","last_modified":1234 }, {}]
			}, {}]
		}
`)

		originDM := dynamic.NewMessage(msgDesc)
		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}
		if err := originDM.UnmarshalJSON(originFixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}

		now := time.Now()
		CompareAndSetPreferenceValueLastModified(dm, originDM, now)

		t.Run("field_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(origin), util.MessageJSONify(after))
			for _, field := range after.GetMessageDescriptor().GetFields() {
				fieldDM, err := util.GetFieldAsDynamicMessage(after, field)
				assert.NoError(t, err)
				assert.Equalf(t, int64(1234), fieldDM.GetFieldByName("last_modified"), "fieldDM: %+v", fieldDM)
			}
		})

		t.Run("field_default_struct", func(t *testing.T) {
			dm, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_default_struct"))
			assert.NoError(t, err)

			for _, field := range dm.GetMessageDescriptor().GetFields() {
				fieldDM, err := util.GetFieldAsDynamicMessage(dm, field)
				assert.NoError(t, err)
				assert.Equalf(t, int64(1234), fieldDM.GetFieldByName("last_modified"), "fieldDM: %+v", fieldDM)
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
				assert.Equalf(t, int64(1234), fieldDMSlice[0].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
				assert.Equalf(t, int64(0), fieldDMSlice[1].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
			}
		})

		t.Run("field_repeated_repeated_default_struct_slice", func(t *testing.T) {
			afterDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, afterDMSlice, 2)

			for _, field := range afterDMSlice[0].GetMessageDescriptor().GetFields() {
				fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(afterDMSlice[0], field)
				assert.NoError(t, err)
				assert.Equalf(t, int64(1234), fieldDMSlice[0].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
				assert.Equalf(t, int64(0), fieldDMSlice[1].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
			}

			beforeDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)

			assert.Equal(t, util.MessageJSONify(beforeDMSlice[1]), util.MessageJSONify(afterDMSlice[1]))
		})
	})
}

func TestSetPreferenceValueLastModifiedWithArbitraryRepeated(t *testing.T) {
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
			"field_repeated_struct": {
				"field_double": [{ "value": 1.0 }, { "value": 1.0 }],
				"field_float": [{ "value": 2.0 }, { "value": 2.0 }],
				"field_int64": [{ "value": 3 }, { "value": 3 }],
				"field_uint64": [{ "value": 4 }, { "value": 4 }],
				"field_int32": [{ "value": 5 }, { "value": 5 }],
				"field_uint32": [{ "value": 6 }, { "value": 6 }],
				"field_bool": [{ "value": true }, { "value": true }],
				"field_string": [{ "value": "7" }, { "value": "7" }],
				"field_bytes": [{ "value": "OA==" }, { "value": "OA==" }]
			},
			"field_repeated_default_struct_slice": {
				"field_double": [{ "value": 1.0 }, { "value": 1.0 }],
				"field_float": [{ "value": 2.0 }, { "value": 2.0 }],
				"field_int64": [{ "value": 3 }, { "value": 3 }],
				"field_uint64": [{ "value": 4 }, { "value": 4 }],
				"field_int32": [{ "value": 5 }, { "value": 5 }],
				"field_uint32": [{ "value": 6 }, { "value": 6 }],
				"field_bool": [{ "value": true }, { "value": true }],
				"field_string": [{ "value": "7" }, { "value": "7" }],
				"field_bytes": [{ "value": "OA==" }, { "value": "OA==" }]
			},
			"field_repeated_repeated_default_struct_slice": [{
				"field_double": [{ "value": 1.0 }, { "value": 1.0 }],
				"field_float": [{ "value": 2.0 }, { "value": 2.0 }],
				"field_int64": [{ "value": 3 }, { "value": 3 }],
				"field_uint64": [{ "value": 4 }, { "value": 4 }],
				"field_int32": [{ "value": 5 }, { "value": 5 }],
				"field_uint32": [{ "value": 6 }, { "value": 6 }],
				"field_bool": [{ "value": true }, { "value": true }],
				"field_string": [{ "value": "7" }, { "value": "7" }],
				"field_bytes": [{ "value": "OA==" }, { "value": "OA==" }]
			}, {}]
		}
`)
		originFixtureBytes := []byte(`{
	"field_repeated_struct": {
		"field_double": [{ "value": 11.0 }],
		"field_float": [{ "value": 12.0 }],
		"field_int64": [{ "value": 13 }],
		"field_uint64": [{ "value": 14 }],
		"field_int32": [{ "value": 15 }],
		"field_uint32": [{ "value": 16 }],
		"field_bool": [{ "value": false }],
		"field_string": [{ "value": "17" }],
		"field_bytes": [{ "value": "MTg=" }]
	},
	"field_repeated_default_struct_slice": {
		"field_double": [{ "value": 11.0 }],
		"field_float": [{ "value": 12.0 }],
		"field_int64": [{ "value": 13 }],
		"field_uint64": [{ "value": 14 }],
		"field_int32": [{ "value": 15 }],
		"field_uint32": [{ "value": 16 }],
		"field_bool": [{ "value": false }],
		"field_string": [{ "value": "17" }],
		"field_bytes": [{ "value": "MTg=" }]
	},
	"field_repeated_repeated_default_struct_slice": [{
		"field_double": [{ "value": 11.0 }],
		"field_float": [{ "value": 12.0 }],
		"field_int64": [{ "value": 13 }],
		"field_uint64": [{ "value": 14 }],
		"field_int32": [{ "value": 15 }],
		"field_uint32": [{ "value": 16 }],
		"field_bool": [{ "value": false }],
		"field_string": [{ "value": "17" }],
		"field_bytes": [{ "value": "MTg=" }]
	}]
}
`)

		originDM := dynamic.NewMessage(msgDesc)
		dm := dynamic.NewMessage(msgDesc)
		if err := dm.UnmarshalJSON(fixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}
		if err := originDM.UnmarshalJSON(originFixtureBytes); err != nil {
			assert.NoError(t, err)
			return
		}

		now := time.Now()
		CompareAndSetPreferenceValueLastModified(dm, originDM, now)

		t.Run("field_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_struct"))
			assert.NoError(t, err)

			t.Logf("origin: %s\nafter: %s", util.MessageJSONify(origin), util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_struct").GetMessageType().GetFields() {
				fieldV := dm.GetFieldByName("field_struct").(*dynamic.Message).GetField(field).(*dynamic.Message)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), fieldV.GetFieldByName("last_modified"))
			}
		})

		t.Run("field_repeated_struct", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_struct"))
			assert.NoError(t, err)

			t.Logf("origin: %s", util.MessageJSONify(origin))
			t.Logf("after: %s", util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_repeated_struct").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm.GetFieldByName("field_repeated_struct").(*dynamic.Message), field)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), dmSlice[0].GetFieldByName("last_modified"))
				assert.Equal(t, now.Unix(), dmSlice[1].GetFieldByName("last_modified"))
			}
		})

		t.Run("field_repeated_default_struct_slice", func(t *testing.T) {
			origin, err := util.GetFieldAsDynamicMessage(originDM, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)
			after, err := util.GetFieldAsDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_default_struct_slice"))
			assert.NoError(t, err)

			t.Logf("origin: %s", util.MessageJSONify(origin))
			t.Logf("after: %s", util.MessageJSONify(after))

			for _, field := range msgDesc.FindFieldByName("field_repeated_default_struct_slice").GetMessageType().GetFields() {
				dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm.GetFieldByName("field_repeated_default_struct_slice").(*dynamic.Message), field)
				assert.NoError(t, err)
				assert.Equal(t, now.Unix(), dmSlice[0].GetFieldByName("last_modified"))
				assert.Equal(t, now.Unix(), dmSlice[1].GetFieldByName("last_modified"))
			}
		})

		t.Run("field_repeated_repeated_default_struct_slice", func(t *testing.T) {
			afterDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(dm, msgDesc.FindFieldByName("field_repeated_repeated_default_struct_slice"))
			assert.NoError(t, err)
			assert.Len(t, afterDMSlice, 2)

			for _, field := range afterDMSlice[0].GetMessageDescriptor().GetFields() {
				fieldDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(afterDMSlice[0], field)
				assert.NoError(t, err)
				assert.Equalf(t, now.Unix(), fieldDMSlice[0].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
				assert.Equalf(t, now.Unix(), fieldDMSlice[1].GetFieldByName("last_modified"), "fieldDMSlice: %+v", fieldDMSlice)
			}
		})
	})
}

func TestCompareValueField(t *testing.T) {
	t.Run(preferenceproto.StringValue, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.StringValue)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.StringValue)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"12345"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.DoubleValue, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.DoubleValue)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.DoubleValue)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1.234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1.2345"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1.234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.FloatValue, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.FloatValue)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.FloatValue)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1.234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1.2345"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1.234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.Int64Value, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.Int64Value)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.Int64Value)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234000"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.UInt64Value, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.UInt64Value)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.UInt64Value)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234000"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.Int32Value, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.Int32Value)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.Int32Value)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234000"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.UInt32Value, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.UInt32Value)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.UInt32Value)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"1234","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234000"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"1234"}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.BoolValue, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.BoolValue)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.BoolValue)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":true,"last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":false}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":true}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
	t.Run(preferenceproto.BytesValue, func(t *testing.T) {
		dst := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.BytesValue)
		origin := dynamic.NewMessage(preferenceproto.DistributionPrimitiveType.BytesValue)

		assert.NoError(t, origin.UnmarshalJSON([]byte(`{"value":"dGVzdGRhdGE=","last_modified":1231}`)))
		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"dGVzdGRhdGEx"}`)))

		equal, err := compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.False(t, equal)

		assert.NoError(t, dst.UnmarshalJSON([]byte(`{"value":"dGVzdGRhdGE="}`)))
		equal, err = compareUpdateValueField(dst, origin)
		assert.NoError(t, err)
		assert.True(t, equal)
	})
}
