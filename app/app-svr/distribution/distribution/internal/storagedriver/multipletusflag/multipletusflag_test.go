package multipletusflag

import (
	"testing"

	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"

	"github.com/jhump/protoreflect/desc"
	"github.com/stretchr/testify/assert"
)

func TestToDynamic(t *testing.T) {
	testJsonToDms := []struct {
		jsonString string
		result     bool
	}{
		{
			jsonString: `{"url":{"value":"xxxx"},"urlV2":{"value":"xxxx2"}}`,
			result:     true,
		},
		{
			jsonString: `{"url":{"value":"xxxx"}, "unknown_url":{"value":"xxxxx"}, "badge":{"value":"badge"}} `, //unknown field
			result:     true,
		},
		{
			jsonString: `{"url":"xxxx", "url2":"{"value":"xxxx2"}}`, // wrong json format
			result:     false,
		},
	}
	pm, ok := preferenceproto.TryGetPreference("bilibili.app.distribution.experimental.v1.MultipleTusConfig")
	assert.Equal(t, true, ok)
	var topleftFieldDm *desc.MessageDescriptor
	for _, field := range pm.ProtoDesc.GetFields() {
		if field.GetName() == "topLeft" {
			topleftFieldDm = field.GetMessageType()
		}
	}
	for _, v := range testJsonToDms {
		dm, err := toDynamicMessage([]byte(v.jsonString), topleftFieldDm)
		if v.result {
			assert.NoError(t, err)
			dmj, _ := dm.MarshalJSON()
			t.Log(string(dmj))
		}
	}
}
