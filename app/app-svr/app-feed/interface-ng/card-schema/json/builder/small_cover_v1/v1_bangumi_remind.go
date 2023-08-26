package jsonsmallcoverv1

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"

	"github.com/pkg/errors"
)

type V1BangumiRemindBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V1BangumiRemindBuilder
	SetBase(*jsoncard.Base) V1BangumiRemindBuilder
	SetRcmd(*ai.Item) V1BangumiRemindBuilder
	SetBangumiRemind(*bangumi.Remind) V1BangumiRemindBuilder
	Build() (*jsoncard.SmallCoverV1, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV1)) V1BangumiRemindBuilder
}

type v1BangumiRemindBuilder struct {
	jsonbuilder.BuilderContext
	jsoncommon.BangumiNotify
	base    *jsoncard.Base
	rcmd    *ai.Item
	remind  *bangumi.Remind
	afterFn []func(*jsoncard.SmallCoverV1)
}

func NewV1BangumiRemindBuilder(ctx jsonbuilder.BuilderContext) V1BangumiRemindBuilder {
	return v1BangumiRemindBuilder{BuilderContext: ctx}
}

func (b v1BangumiRemindBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V1BangumiRemindBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v1BangumiRemindBuilder) SetBase(base *jsoncard.Base) V1BangumiRemindBuilder {
	b.base = base
	return b
}

func (b v1BangumiRemindBuilder) SetRcmd(in *ai.Item) V1BangumiRemindBuilder {
	b.rcmd = in
	return b
}

func (b v1BangumiRemindBuilder) SetBangumiRemind(in *bangumi.Remind) V1BangumiRemindBuilder {
	b.remind = in
	return b
}

func (b v1BangumiRemindBuilder) constructCover() string {
	return b.ConstructRemindCover(b.remind)
}

func (b v1BangumiRemindBuilder) constructURI() string {
	return b.ConstructRemindURI(b.remind)
}

func (b v1BangumiRemindBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV1)) V1BangumiRemindBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}

func (b v1BangumiRemindBuilder) Build() (*jsoncard.SmallCoverV1, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.remind == nil {
		return nil, errors.Errorf("empty `remind` field")
	}
	if len(b.remind.List) == 0 {
		return nil, errors.Errorf("empty `remind.List` field")
	}

	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateParam("").
		UpdateCover(b.constructCover()).
		UpdateURI(b.constructURI()).
		UpdateTitle(b.remind.List[0].UpdateTitle).
		Update(); err != nil {
		return nil, err
	}

	out := &jsoncard.SmallCoverV1{
		Base:          b.base,
		TitleRightPic: appcardmodel.IconTV,
		Desc1:         b.remind.List[0].UpdateDesc,
	}
	if len(b.remind.List) > 1 {
		out.Desc2 = b.remind.List[1].UpdateDesc
	}
	updates := int64(b.remind.Updates)
	//nolint:gomnd
	if b.remind.Updates > 99 {
		updates = 99
		out.TitleRightPic = appcardmodel.IconBomb
	}
	for _, v := range b.remind.List {
		out.SeasonId = append(out.SeasonId, v.SeasonId)
		out.Epid = append(out.Epid, v.Epid)
	}
	out.TitleRightText = strconv.FormatInt(updates, 10)
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}
