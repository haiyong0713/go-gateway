package notify_tunnel

import (
	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	"github.com/pkg/errors"
)

type NotifyTunnelV1Builder interface {
	ReplaceContext(jsonbuilder.BuilderContext) NotifyTunnelV1Builder
	SetBase(*jsoncard.Base) NotifyTunnelV1Builder
	SetTunnel(map[int64]*tunnelgrpc.FeedCard) NotifyTunnelV1Builder
	SetRcmd(*ai.Item) NotifyTunnelV1Builder

	Build() (*jsoncard.UniversalNotifyTunnelV1, error)
}

type notifyTunnelV1Builder struct {
	jsonbuilder.BuilderContext
	rcmd       *ai.Item
	base       *jsoncard.Base
	threePoint jsoncommon.ThreePoint
	tunnel     map[int64]*tunnelgrpc.FeedCard
}

func NewNotifyTunnelV1Builder(ctx jsonbuilder.BuilderContext) NotifyTunnelV1Builder {
	return notifyTunnelV1Builder{BuilderContext: ctx}
}

func (b notifyTunnelV1Builder) ReplaceContext(ctx jsonbuilder.BuilderContext) NotifyTunnelV1Builder {
	b.BuilderContext = ctx
	return b
}

func (b notifyTunnelV1Builder) SetBase(base *jsoncard.Base) NotifyTunnelV1Builder {
	b.base = base
	return b
}

func (b notifyTunnelV1Builder) SetTunnel(tunnel map[int64]*tunnelgrpc.FeedCard) NotifyTunnelV1Builder {
	b.tunnel = tunnel
	return b
}

func (b notifyTunnelV1Builder) SetRcmd(rcmd *ai.Item) NotifyTunnelV1Builder {
	b.rcmd = rcmd
	return b
}

func (b notifyTunnelV1Builder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	enableSwitchColumn := b.BuilderContext.FeatureGates().FeatureEnabled(cardschema.FeatureSwitchColumnThreePoint)
	if b.BuilderContext.VersionControl().Can("feed.usingNewThreePointV2") {
		return b.threePoint.ConstructDefaultThreePointV2(b.BuilderContext, enableSwitchColumn)
	}
	return b.threePoint.ConstructDefaultThreePointV2Legacy(b.BuilderContext, enableSwitchColumn)
}

func (b notifyTunnelV1Builder) Build() (*jsoncard.UniversalNotifyTunnelV1, error) {
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	msgIDs, err := card.ConstructMsgIDs(b.rcmd.MsgIDs)
	if err != nil {
		return nil, errors.Errorf("invalid ai msg id: %q: %+v", b.rcmd.MsgIDs, err)
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateArgs(jsoncard.Args{}).
		UpdateThreePoint(b.threePoint.ConstructDefaultThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.UniversalNotifyTunnelV1{
		Base:  b.base,
		Items: make([]*jsoncard.NotifyTunnelItemV1, 0, len(msgIDs)),
	}
	for _, oid := range msgIDs {
		tc, ok := b.tunnel[oid]
		if !ok {
			log.Warn("Failed to get tunnel card with id: %d", oid)
			continue
		}
		item := &jsoncard.NotifyTunnelItemV1{}
		item.FromTunnelCard(tc)
		out.Items = append(out.Items, item)
	}
	if len(out.Items) <= 0 {
		return nil, errors.Errorf("none tunnel card exist with msg_ids: %+v", msgIDs)
	}
	return out, nil
}
