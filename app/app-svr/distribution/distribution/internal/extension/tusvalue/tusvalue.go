package tusvalue

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
)

var (
	TusValues             = make(map[string][]string)
	multipleTusPreference = "bilibili.app.distribution.experimental.v1.MultipleTusConfig"
)

func Init() {
	meta, ok := preferenceproto.TryGetPreference(multipleTusPreference)
	if !ok {
		panic("failed to get preference with tus value")
	}
	for _, v := range meta.ProtoDesc.GetFields() {
		tusvalues, err := preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsTusValues(v)
		if err != nil {
			panic(err)
		}
		TusValues[v.GetFullyQualifiedName()] = append(TusValues[v.GetFullyQualifiedJSONName()], tusvalues...)
	}
}
