package _select

import (
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	"github.com/pkg/errors"
)

type SelectBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) SelectBuilder
	SetBase(*jsoncard.Base) SelectBuilder
	SetFollowMode(*FollowMode) SelectBuilder
	Build() (*jsoncard.Select, error)
}

type selectBuilder struct {
	jsonbuilder.BuilderContext
	base       *jsoncard.Base
	followMode *FollowMode
}

type FollowMode struct {
	Title   string
	Desc    string
	Buttons []string
}

func NewSelectBuilder(ctx jsonbuilder.BuilderContext) SelectBuilder {
	return selectBuilder{BuilderContext: ctx}
}

func (b selectBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) SelectBuilder {
	b.BuilderContext = ctx
	return b
}

func (b selectBuilder) SetBase(base *jsoncard.Base) SelectBuilder {
	b.base = base
	return b
}

func (b selectBuilder) SetFollowMode(followMode *FollowMode) SelectBuilder {
	b.followMode = followMode
	return b
}

func (b selectBuilder) Build() (*jsoncard.Select, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.followMode == nil {
		return nil, errors.Errorf("empty `followMode` field")
	}
	out := &jsoncard.Select{
		Base: b.base,
	}
	out.Title = "提醒：是否需要开启首页推荐的“关注模式”（内测版）？"
	if b.followMode.Title != "" {
		out.Title = b.followMode.Title
	}
	out.Desc = "我们根据你对bilibili推荐的反馈，为你定制了关注模式。开启后，仅为你显示关注UP主更新的视频哦。尝试体验一下？"
	if b.followMode.Desc != "" {
		out.Desc = b.followMode.Desc
	}
	out.LeftButton = &jsoncard.Button{Text: "不参加", Event: "close", Type: appcardmodel.ButtonGrey}
	out.RightButton = &jsoncard.Button{Text: "我想参加", Event: "follow_mode", Type: appcardmodel.ButtonGrey}
	if b.followMode.Buttons != nil {
		out.LeftButton = &jsoncard.Button{Text: b.followMode.Buttons[0], Event: "close", Type: appcardmodel.ButtonGrey}
		out.RightButton = &jsoncard.Button{Text: b.followMode.Buttons[1], Event: "follow_mode", Type: appcardmodel.ButtonGrey}
	}
	return out, nil
}
