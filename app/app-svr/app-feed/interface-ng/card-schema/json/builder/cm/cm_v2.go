package cm

import (
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
)

// CmV2BuilderFactory is
type CmV2BuilderFactory interface {
	ReplaceContext(jsonbuilder.BuilderContext) CmV2BuilderFactory

	DeriveAdWebsBuilder() V2AdWebSBuilder
	DeriveAdAvBuilder() V2AdAvBuilder
	DeriveAdWebBuilder() V2AdWebBuilder
	DeriveAdPlayerBuilder() V2AdPlayerBuilder
	DeriveAdInlineBuilder() V2AdInlineLiveBuilder
	DeriveAdReservation() V2AdReservationBuilder
}

type cmV2BuilderFactory struct {
	jsonbuilder.BuilderContext
}

// NewCmV2BuilderFactory is
func NewCmV2BuilderFactory(ctx jsonbuilder.BuilderContext) CmV2BuilderFactory {
	return cmV2BuilderFactory{BuilderContext: ctx}
}

func (b cmV2BuilderFactory) ReplaceContext(ctx jsonbuilder.BuilderContext) CmV2BuilderFactory {
	b.BuilderContext = ctx
	return b
}

func (b cmV2BuilderFactory) DeriveAdWebsBuilder() V2AdWebSBuilder {
	return v2AdWebSBuilder{parent: &b}
}

func (b cmV2BuilderFactory) DeriveAdAvBuilder() V2AdAvBuilder {
	return v2AdAvBuilder{parent: &b}
}

func (b cmV2BuilderFactory) DeriveAdWebBuilder() V2AdWebBuilder {
	return v2AdWebBuilder{parent: &b}
}

func (b cmV2BuilderFactory) DeriveAdPlayerBuilder() V2AdPlayerBuilder {
	return v2AdPlayerBuilder{parent: &b}
}

func (b cmV2BuilderFactory) DeriveAdInlineBuilder() V2AdInlineLiveBuilder {
	return v2AdInlineLiveBuilder{parent: &b}
}

func (b cmV2BuilderFactory) DeriveAdReservation() V2AdReservationBuilder {
	return v2AdReservationBuilder{parent: &b}
}
