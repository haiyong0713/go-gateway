package small_cover_v6

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	"github.com/pkg/errors"
)

type V6TunnelBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V6TunnelBuilder
	SetBase(*jsoncard.Base) V6TunnelBuilder
	SetTunnel(*tunnelgrpc.FeedCard) V6TunnelBuilder
	SetRcmd(*ai.Item) V6TunnelBuilder
	Build() (*jsoncard.SmallCoverV6, error)
}

type v6TunnelBuilder struct {
	jsonbuilder.BuilderContext
	base   *jsoncard.Base
	rcmd   *ai.Item
	tunnel *tunnelgrpc.FeedCard
}

func NewV6TunnelBuilder(ctx jsonbuilder.BuilderContext) V6TunnelBuilder {
	return v6TunnelBuilder{BuilderContext: ctx}
}

func (b v6TunnelBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V6TunnelBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v6TunnelBuilder) SetBase(base *jsoncard.Base) V6TunnelBuilder {
	b.base = base
	return b
}

func (b v6TunnelBuilder) SetTunnel(in *tunnelgrpc.FeedCard) V6TunnelBuilder {
	b.tunnel = in
	return b
}

func (b v6TunnelBuilder) SetRcmd(in *ai.Item) V6TunnelBuilder {
	b.rcmd = in
	return b
}

func (b v6TunnelBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoWeb, device.Plat(), int(device.Build()), b.tunnel.Link, nil)
}

func (b v6TunnelBuilder) Build() (*jsoncard.SmallCoverV6, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.tunnel == nil {
		return nil, errors.Errorf("empty `tunnel` field")
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.tunnel.Cover).
		UpdateTitle(b.tunnel.Title).
		UpdateURI(b.constructURI()).
		UpdateGoto(appcardmodel.GotoGame).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV6{
		Base:  b.base,
		Desc1: b.tunnel.Intro,
	}
	return out, nil
}
