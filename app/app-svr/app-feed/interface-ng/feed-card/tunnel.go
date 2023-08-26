package feedcard

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/notify_tunnel"
	jsonsmallcover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover"
	v6 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v6"
	v7 "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/small_cover_v7"

	"github.com/pkg/errors"
)

func BuildSmallCoverV4FromTunnel(ctx cardschema.FeedContext, _ int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	tunnel, ok := fanoutResult.Tunnel[item.ID]
	if !ok {
		return nil, errors.Errorf("tunnel: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.SmallCoverV4).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoWeb).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}

	builder := jsonsmallcover.NewV4TunnelBuilder(ctx)
	card, err := builder.SetBase(base).SetRcmd(item).SetTunnel(tunnel).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV7FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	tunnel, ok := fanoutResult.Tunnel[item.ID]
	if !ok {
		return nil, errors.Errorf("tunnel: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV7).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoGame).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := v7.NewV7TunnelBuilder(ctx)
	card, err := builder.SetBase(base).SetTunnel(tunnel).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildSmallCoverV6FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	tunnel, ok := fanoutResult.Tunnel[item.ID]
	if !ok {
		return nil, errors.Errorf("tunnel: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetParam(strconv.FormatInt(item.ID, 10)).
		SetCardType(appcardmodel.SmallCoverV6).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetGoto(appcardmodel.GotoGame).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := v6.NewV6TunnelBuilder(ctx)
	card, err := builder.SetBase(base).SetTunnel(tunnel).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildNotifyTunnelV1FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Tunnel == nil {
		return nil, errors.Errorf("tunnel is not exist, %s", item.MsgIDs)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.NotifyTunnelV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := notify_tunnel.NewNotifyTunnelV1Builder(ctx)
	card, err := builder.SetBase(base).SetTunnel(fanoutResult.Tunnel).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildNotifyTunnelSingleV1FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	if fanoutResult.Tunnel == nil {
		return nil, errors.Errorf("tunnel is not exist, %s", item.MsgIDs)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.NotifyTunnelSingleV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := notify_tunnel.NewNotifyTunnelV1Builder(ctx)
	card, err := builder.SetBase(base).SetTunnel(fanoutResult.Tunnel).SetRcmd(item).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildNotifyTunnelLargeV1FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	tunnel, ok := fanoutResult.Tunnel[item.ID]
	if !ok {
		return nil, errors.Errorf("tunnel: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.NotifyTunnelLargeV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := notify_tunnel.NewNotifyTunnelLargeV1Builder(ctx)
	card, err := builder.SetBase(base).SetTunnel(tunnel).SetRcmd(item).SetArcPlayer(fanoutResult.Archive.Archive).
		SetEpisode(fanoutResult.Bangumi.InlinePGC).SetLiveRoom(fanoutResult.Live.InlineRoom).
		SetAuthorCard(fanoutResult.Account.Card).SetInline(fanoutResult.Inline).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}

func BuildNotifyTunnelLargeSingleV1FromTunnel(ctx cardschema.FeedContext, index int64, item *ai.Item, fanoutResult *FanoutResult) (cardschema.FeedCard, error) {
	tunnel, ok := fanoutResult.Tunnel[item.ID]
	if !ok {
		return nil, errors.Errorf("tunnel: %d not exist", item.ID)
	}
	base, err := jsonbuilder.NewBaseBuilder(ctx).
		SetCardType(appcardmodel.NotifyTunnelLargeSingleV1).
		SetCardGoto(appcardmodel.CardGt(item.Goto)).
		SetMetricRcmd(item).
		SetTrackID(item.TrackID).
		SetPosRecUniqueID(item.PosRecUniqueID).
		Build()
	if err != nil {
		return nil, err
	}
	builder := notify_tunnel.NewNotifyTunnelLargeV1Builder(ctx)
	card, err := builder.SetBase(base).SetTunnel(tunnel).SetRcmd(item).SetArcPlayer(fanoutResult.Archive.Archive).
		SetEpisode(fanoutResult.Bangumi.InlinePGC).SetLiveRoom(fanoutResult.Live.InlineRoom).
		SetAuthorCard(fanoutResult.Account.Card).SetInline(fanoutResult.Inline).Build()
	if err != nil {
		return nil, err
	}
	return card, nil
}
