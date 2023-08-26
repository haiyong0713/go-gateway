package small_cover_v6

import (
	"strings"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	vip "git.bilibili.co/bapis/bapis-go/vip/service"

	"github.com/pkg/errors"
)

type V6VipBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V6VipBuilder
	SetBase(*jsoncard.Base) V6VipBuilder
	SetVip(*vip.TipsRenewReply) V6VipBuilder
	SetRcmd(*ai.Item) V6VipBuilder
	Build() (*jsoncard.SmallCoverV6, error)
}

type v6VipBuilder struct {
	jsonbuilder.BuilderContext
	base *jsoncard.Base
	rcmd *ai.Item
	vip  *vip.TipsRenewReply
}

func NewV6VipBuilder(ctx jsonbuilder.BuilderContext) V6VipBuilder {
	return v6VipBuilder{BuilderContext: ctx}
}

func (b v6VipBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V6VipBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v6VipBuilder) SetBase(base *jsoncard.Base) V6VipBuilder {
	b.base = base
	return b
}

func (b v6VipBuilder) SetVip(in *vip.TipsRenewReply) V6VipBuilder {
	b.vip = in
	return b
}

func (b v6VipBuilder) SetRcmd(in *ai.Item) V6VipBuilder {
	b.rcmd = in
	return b
}

func (b v6VipBuilder) constructURI() string {
	device := b.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoWeb, device.Plat(), int(device.Build()), b.vip.Link, nil)
}

func (b v6VipBuilder) constructCover() string {
	if b.vip.ImgUri != "" {
		return b.vip.ImgUri
	}
	return "https://i0.hdslb.com/bfs/archive/1b8deb69e4a9effc8f3be24107f925480afe3ade.png"
}

func (b v6VipBuilder) constructTitle() string {
	title := strings.Replace(b.vip.Title, "[", " \u003cem class=\"keyword\"\u003e", 1)
	return strings.Replace(title, "]", "\u003c/em\u003e ", 1)
}

func (b v6VipBuilder) Build() (*jsoncard.SmallCoverV6, error) {
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
	out := &jsoncard.SmallCoverV6{
		Base:  b.base,
		Desc1: b.vip.Tip,
	}
	return out, nil
}
