package jsonsmallcoverv1

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	"github.com/pkg/errors"
)

type V1BangumiUpdateBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V1BangumiUpdateBuilder
	SetBase(*jsoncard.Base) V1BangumiUpdateBuilder
	SetRcmd(*ai.Item) V1BangumiUpdateBuilder
	SetBangumiUpdate(*bangumi.Update) V1BangumiUpdateBuilder
	Build() (*jsoncard.SmallCoverV1, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV1)) V1BangumiUpdateBuilder
}

type v1BangumiUpdateBuilder struct {
	jsonbuilder.BuilderContext
	base    *jsoncard.Base
	rcmd    *ai.Item
	update  *bangumi.Update
	afterFn []func(*jsoncard.SmallCoverV1)
}

func NewV1BangumiUpdateBuilder(ctx jsonbuilder.BuilderContext) V1BangumiUpdateBuilder {
	return v1BangumiUpdateBuilder{BuilderContext: ctx}
}

func (b v1BangumiUpdateBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V1BangumiUpdateBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v1BangumiUpdateBuilder) SetBase(base *jsoncard.Base) V1BangumiUpdateBuilder {
	b.base = base
	return b
}

func (b v1BangumiUpdateBuilder) SetRcmd(in *ai.Item) V1BangumiUpdateBuilder {
	b.rcmd = in
	return b
}

func (b v1BangumiUpdateBuilder) SetBangumiUpdate(in *bangumi.Update) V1BangumiUpdateBuilder {
	b.update = in
	return b
}

func (b v1BangumiUpdateBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV1)) V1BangumiUpdateBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v1BangumiUpdateBuilder) Build() (*jsoncard.SmallCoverV1, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.update == nil {
		return nil, errors.Errorf("empty `update` field")
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateParam("").
		UpdateCover(b.update.SquareCover).
		UpdateTitle("你的追番更新啦").
		Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.SmallCoverV1{
		Base:          b.base,
		TitleRightPic: appcardmodel.IconTV,
		Desc1:         b.update.Title,
	}
	updates := int64(b.update.Updates)
	//nolint:gomnd
	if b.update.Updates > 99 {
		updates = 99
		out.TitleRightPic = appcardmodel.IconBomb
	}
	out.TitleRightText = strconv.FormatInt(updates, 10)
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}
