package small_cover_v7

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	"strings"

	vip "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/pkg/errors"
)

type V7VipBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V7VipBuilder
	SetBase(*jsoncard.Base) V7VipBuilder
	SetVip(*vip.TipsRenewReply) V7VipBuilder
	SetRcmd(*ai.Item) V7VipBuilder
	Build() (*jsoncard.SmallCoverV7, error)
}

type v7VipBuilder struct {
	jsonbuilder.BuilderContext
	base *jsoncard.Base
	rcmd *ai.Item
	vip  *vip.TipsRenewReply
}

func NewV7VipBuilder(ctx jsonbuilder.BuilderContext) V7VipBuilder {
	return v7VipBuilder{BuilderContext: ctx}
}

func (b v7VipBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V7VipBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v7VipBuilder) SetBase(base *jsoncard.Base) V7VipBuilder {
	b.base = base
	return b
}

func (b v7VipBuilder) SetVip(in *vip.TipsRenewReply) V7VipBuilder {
	b.vip = in
	return b
}

func (b v7VipBuilder) SetRcmd(in *ai.Item) V7VipBuilder {
	b.rcmd = in
	return b
}

func (b v7VipBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoWeb, device.Plat(), int(device.Build()), b.vip.Link, nil)
}

func (b v7VipBuilder) constructCover() string {
	if b.vip.ImgUri != "" {
		return b.vip.ImgUri
	}
	return "https://i0.hdslb.com/bfs/archive/1b8deb69e4a9effc8f3be24107f925480afe3ade.png"
}

func (b v7VipBuilder) constructTitle() string {
	title := strings.Replace(b.vip.Title, "[", " \u003cem class=\"keyword\"\u003e", 1)
	return strings.Replace(title, "]", "\u003c/em\u003e ", 1)
}

func (b v7VipBuilder) Build() (*jsoncard.SmallCoverV7, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.vip == nil {
		return nil, errors.Errorf("empty `vip` field")
	}
	if b.vip.Title == "" {
		return nil, errors.Errorf("empty `Title`")
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.constructCover()).
		UpdateTitle(b.constructTitle()).
		UpdateURI(b.constructURI()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV7{
		Base:         b.base,
		DestroyCard:  1,
		Desc:         b.vip.Tip,
		ResourceType: "vip_renew",
	}
	return out, nil
}
