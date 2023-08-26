package jsonbanner

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	"github.com/pkg/errors"
)

type V5BannerBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V5BannerBuilder
	SetRcmd(*ai.Item) V5BannerBuilder
	SetBase(*jsoncard.Base) V5BannerBuilder
	SetBanners([]*banner.Banner) V5BannerBuilder
	SetVersion(string) V5BannerBuilder

	Build() (*jsoncard.Banner, error)
}

type v5BannerBuilder struct {
	jsonbuilder.BuilderContext
	base    *jsoncard.Base
	banners []*banner.Banner
	version string
	rcmd    *ai.Item
}

func NewBannerV5Builder(ctx jsonbuilder.BuilderContext) V5BannerBuilder {
	return v5BannerBuilder{BuilderContext: ctx}
}

func (b v5BannerBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V5BannerBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v5BannerBuilder) SetBase(base *jsoncard.Base) V5BannerBuilder {
	b.base = base
	return b
}

func (b v5BannerBuilder) SetBanners(banners []*banner.Banner) V5BannerBuilder {
	b.banners = banners
	return b
}

func (b v5BannerBuilder) SetVersion(in string) V5BannerBuilder {
	b.version = in
	return b
}

func (b v5BannerBuilder) SetRcmd(item *ai.Item) V5BannerBuilder {
	b.rcmd = item
	return b
}

func (b v5BannerBuilder) Build() (*jsoncard.Banner, error) {
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if len(b.banners) == 0 {
		return nil, errors.Errorf("empty `Banners` field")
	}
	return &jsoncard.Banner{
		Base:       b.base,
		BannerItem: b.banners,
		Hash:       b.version,
	}, nil
}
