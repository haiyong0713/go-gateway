package cm

import (
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
)

// CmV1BuilderFactory is
type CmV1BuilderFactory interface {
	ReplaceContext(jsonbuilder.BuilderContext) CmV1BuilderFactory

	DeriveAdAvBuilder() V1AdAvBuilder
	DeriveAdWebBuilder() V1AdWebBuilder
	DeriveAdPlayerBuilder() V1AdPlayerBuilder
}

type cmV1BuilderFactory struct {
	jsonbuilder.BuilderContext
}

// NewCmV1BuilderFactory is
func NewCmV1BuilderFactory(ctx jsonbuilder.BuilderContext) CmV1BuilderFactory {
	return cmV1BuilderFactory{BuilderContext: ctx}
}

func (b cmV1BuilderFactory) ReplaceContext(ctx jsonbuilder.BuilderContext) CmV1BuilderFactory {
	b.BuilderContext = ctx
	return b
}

func (b cmV1BuilderFactory) DeriveAdAvBuilder() V1AdAvBuilder {
	return v1AdAvBuilder{parent: &b}
}

func (b cmV1BuilderFactory) DeriveAdWebBuilder() V1AdWebBuilder {
	return v1AdWebBuilder{parent: &b}
}

func (b cmV1BuilderFactory) DeriveAdPlayerBuilder() V1AdPlayerBuilder {
	return v1AdPlayerBuilder{parent: &b}
}
