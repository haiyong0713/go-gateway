package extension

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/defaultvalue"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/lastmodified"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/mergepreference"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/tusvalue"
)

func Init() {
	defaultvalue.Init()
	lastmodified.Init()
	mergepreference.Init()
	tusvalue.Init()
}
