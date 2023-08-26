package jsonavatar

import (
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	"github.com/pkg/errors"
)

type AvatarBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) AvatarBuilder
	SetAvatarStatus(*jsoncard.AvatarStatus) AvatarBuilder
	Build() (*jsoncard.Avatar, error)
}

type avatarBuilder struct {
	jsonbuilder.BuilderContext
	avatarStatus *jsoncard.AvatarStatus
}

func NewAvatarBuilder(ctx jsonbuilder.BuilderContext) AvatarBuilder {
	return avatarBuilder{BuilderContext: ctx}
}

func (b avatarBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) AvatarBuilder {
	b.BuilderContext = ctx
	return b
}

func (b avatarBuilder) SetAvatarStatus(in *jsoncard.AvatarStatus) AvatarBuilder {
	b.avatarStatus = in
	return b
}

func (b avatarBuilder) Build() (*jsoncard.Avatar, error) {
	if b.avatarStatus == nil {
		return nil, errors.Errorf("empty `avatarStatus` field")
	}
	out := &jsoncard.Avatar{
		Cover:        b.avatarStatus.Cover,
		Text:         b.avatarStatus.Text,
		URI:          appcardmodel.FillURI(b.avatarStatus.Goto, 0, 0, b.avatarStatus.Param, nil),
		Type:         b.avatarStatus.Type,
		Event:        appcardmodel.AvatarEvent[b.avatarStatus.Goto],
		EventV2:      appcardmodel.AvatarEventV2[b.avatarStatus.Goto],
		DefalutCover: b.avatarStatus.DefalutCover,
		FaceNftNew:   b.avatarStatus.FaceNftNew,
	}
	if b.avatarStatus.Goto == appcardmodel.GotoMid {
		out.UpID, _ = strconv.ParseInt(b.avatarStatus.Param, 10, 64)
	}
	return out, nil
}
