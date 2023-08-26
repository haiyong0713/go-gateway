package jsonsmallcover

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	"github.com/pkg/errors"
)

type V4TunnelBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V4TunnelBuilder
	SetBase(*jsoncard.Base) V4TunnelBuilder
	SetTunnel(*tunnelgrpc.FeedCard) V4TunnelBuilder
	SetRcmd(*ai.Item) V4TunnelBuilder
	Build() (*jsoncard.SmallCoverV4, error)
}

type v4TunnelBuilder struct {
	jsonbuilder.BuilderContext
	base   *jsoncard.Base
	tunnel *tunnelgrpc.FeedCard
	rcmd   *ai.Item
}

func NewV4TunnelBuilder(ctx jsonbuilder.BuilderContext) V4TunnelBuilder {
	return v4TunnelBuilder{BuilderContext: ctx}
}

func (b v4TunnelBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V4TunnelBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v4TunnelBuilder) SetBase(base *jsoncard.Base) V4TunnelBuilder {
	b.base = base
	return b
}

func (b v4TunnelBuilder) SetRcmd(rcmd *ai.Item) V4TunnelBuilder {
	b.rcmd = rcmd
	return b
}

func (b v4TunnelBuilder) SetTunnel(tunnel *tunnelgrpc.FeedCard) V4TunnelBuilder {
	b.tunnel = tunnel
	return b
}

func (b v4TunnelBuilder) Build() (*jsoncard.SmallCoverV4, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.tunnel == nil {
		return nil, errors.Errorf("empty `tunnel` field")
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.tunnel.Cover).
		UpdateTitle(b.tunnel.Title).
		UpdateParam(strconv.FormatInt(b.rcmd.ID, 10)).
		UpdateURI(b.constructTunnelURI()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV4{
		Base: b.base,
		Desc: b.tunnel.Intro,
	}
	return out, nil
}

func (b v4TunnelBuilder) constructTunnelURI() string {
	return appcardmodel.FillURI(appcardmodel.GotoWeb, 0, 0, b.tunnel.Link, nil)
}
