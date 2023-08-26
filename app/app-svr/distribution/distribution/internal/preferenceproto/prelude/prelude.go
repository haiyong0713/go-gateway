package prelude

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
)

func init() {
	protoStore, err := preferenceproto.NewEmbed()
	if err != nil {
		panic(err)
	}
	if err := preferenceproto.InitPreferenceRegistry(protoStore); err != nil {
		panic(err)
	}
}
