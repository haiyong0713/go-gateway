package small_cover_v7

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	"github.com/pkg/errors"
)

type V7TunnelBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V7TunnelBuilder
	SetBase(*jsoncard.Base) V7TunnelBuilder
	SetTunnel(*tunnelgrpc.FeedCard) V7TunnelBuilder
	SetRcmd(*ai.Item) V7TunnelBuilder
	Build() (*jsoncard.SmallCoverV7, error)
}

type v7TunnelBuilder struct {
	jsonbuilder.BuilderContext
	base   *jsoncard.Base
	rcmd   *ai.Item
	tunnel *tunnelgrpc.FeedCard
}

func NewV7TunnelBuilder(ctx jsonbuilder.BuilderContext) V7TunnelBuilder {
	return v7TunnelBuilder{BuilderContext: ctx}
}

func (b v7TunnelBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V7TunnelBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v7TunnelBuilder) SetBase(base *jsoncard.Base) V7TunnelBuilder {
	b.base = base
	return b
}

func (b v7TunnelBuilder) SetTunnel(in *tunnelgrpc.FeedCard) V7TunnelBuilder {
	b.tunnel = in
	return b
}

func (b v7TunnelBuilder) SetRcmd(in *ai.Item) V7TunnelBuilder {
	b.rcmd = in
	return b
}

func (b v7TunnelBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoWeb, device.Plat(), int(device.Build()), b.tunnel.Link, nil)
}

func (b v7TunnelBuilder) Build() (*jsoncard.SmallCoverV7, error) {
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
		Update(); err != nil {
		return nil, err
	}
	if b.tunnel.Goto != "" {
		if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
			UpdateGoto(appcardmodel.Gt(b.tunnel.Goto)).
			Update(); err != nil {
			return nil, err
		}
	}
	resourceType := cvtTunnelResourceType(b.tunnel.ResourceType)
	out := &jsoncard.SmallCoverV7{
		Base:         b.base,
		DestroyCard:  int8(b.tunnel.Destroy),
		Desc:         b.tunnel.Intro,
		ResourceType: resourceType,
	}
	switch resourceType {
	case "game_tunnel":
		if b.BuilderContext.VersionControl().Can("feed.compatibleWithGameID") {
			if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
				UpdateParam(strconv.FormatInt(b.tunnel.UniqueId, 10)).
				Update(); err != nil {
				return nil, err
			}
		}
		out.GameID = b.tunnel.UniqueId
	}
	return out, nil
}

func cvtTunnelResourceType(in string) string {
	if in == "game" {
		return "game_tunnel" // 曾经和客户端约定的游戏预约卡为 game_tunnel
	}
	return in
}
