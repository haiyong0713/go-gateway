package jsonsmallcover

import (
	"hash/crc32"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"

	"github.com/pkg/errors"
)

type V4UpdateBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V4UpdateBuilder
	SetBase(*jsoncard.Base) V4UpdateBuilder
	SetRcmd(*ai.Item) V4UpdateBuilder
	SetUpdate(*bangumi.Update) V4UpdateBuilder
	Build() (*jsoncard.SmallCoverV4, error)
}

type v4UpdateBuilder struct {
	jsonbuilder.BuilderContext
	base   *jsoncard.Base
	rcmd   *ai.Item
	update *bangumi.Update
}

func NewV4UpdateBuilder(ctx jsonbuilder.BuilderContext) V4UpdateBuilder {
	return v4UpdateBuilder{BuilderContext: ctx}
}

func (b v4UpdateBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V4UpdateBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v4UpdateBuilder) SetBase(base *jsoncard.Base) V4UpdateBuilder {
	b.base = base
	return b
}

func (b v4UpdateBuilder) SetRcmd(in *ai.Item) V4UpdateBuilder {
	b.rcmd = in
	return b
}

func (b v4UpdateBuilder) SetUpdate(update *bangumi.Update) V4UpdateBuilder {
	b.update = update
	return b
}

func (b v4UpdateBuilder) Build() (*jsoncard.SmallCoverV4, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.update == nil {
		return nil, errors.Errorf("empty `remind` field")
	}
	if b.update.Updates == 0 {
		return nil, errors.Errorf("remind updates is 0")
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.update.SquareCover).
		UpdateTitle(b.constructTitle()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV4{
		Base: b.base,
		Desc: b.update.Title,
	}
	out.TitleRightPic, out.TitleRightText = b.constructTitleRight()
	return out, nil
}

func (b v4UpdateBuilder) constructTitle() string {
	title := "你的追番更新啦"
	return title + emojiMap[crc32.ChecksumIEEE([]byte(b.rcmd.TrackID))%4]
}

func (b v4UpdateBuilder) constructTitleRight() (appcardmodel.Icon, string) {
	updates := b.update.Updates
	//nolint:gomnd
	if updates > 99 {
		return appcardmodel.IconBomb, strconv.Itoa(99)
	}
	return appcardmodel.IconTV, strconv.Itoa(updates)
}
