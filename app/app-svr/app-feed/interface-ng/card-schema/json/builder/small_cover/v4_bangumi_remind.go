package jsonsmallcover

import (
	"hash/crc32"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"

	"github.com/pkg/errors"
)

type V4RemindBuilder interface {
	ReplaceContext(jsonbuilder.BuilderContext) V4RemindBuilder
	SetBase(*jsoncard.Base) V4RemindBuilder
	SetRcmd(*ai.Item) V4RemindBuilder
	SetRemind(*bangumi.Remind) V4RemindBuilder
	Build() (*jsoncard.SmallCoverV4, error)
}

type v4RemindBuilder struct {
	jsonbuilder.BuilderContext
	jsoncommon.BangumiNotify
	base   *jsoncard.Base
	rcmd   *ai.Item
	remind *bangumi.Remind
}

var emojiMap = map[uint32]string{
	0: "(´∀｀*)ｳﾌﾌ",
	1: "ヾ( ・∀・)ﾉ",
	2: "(｀･ω･´)ゞ",
	3: "(・∀・)ｲｲ!!",
}

func NewV4RemindBuilder(ctx jsonbuilder.BuilderContext) V4RemindBuilder {
	return v4RemindBuilder{BuilderContext: ctx}
}

func (b v4RemindBuilder) ReplaceContext(ctx jsonbuilder.BuilderContext) V4RemindBuilder {
	b.BuilderContext = ctx
	return b
}

func (b v4RemindBuilder) SetBase(base *jsoncard.Base) V4RemindBuilder {
	b.base = base
	return b
}

func (b v4RemindBuilder) SetRcmd(in *ai.Item) V4RemindBuilder {
	b.rcmd = in
	return b
}

func (b v4RemindBuilder) SetRemind(remind *bangumi.Remind) V4RemindBuilder {
	b.remind = remind
	return b
}

func (b v4RemindBuilder) Build() (*jsoncard.SmallCoverV4, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.remind == nil {
		return nil, errors.Errorf("empty `remind` field")
	}
	if b.remind.Updates == 0 {
		return nil, errors.Errorf("remind updates is 0")
	}
	if len(b.remind.List) == 0 {
		return nil, errors.Errorf("empty `remind` list")
	}
	if err := jsonbuilder.NewBaseUpdater(b.BuilderContext, b.base).
		UpdateCover(b.constructCover()).
		UpdateTitle(b.constructTitle()).
		UpdateURI(b.constructURI()).
		Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV4{
		Base: b.base,
		Desc: b.remind.List[0].UpdateDesc,
	}
	for _, v := range b.remind.List {
		out.SeasonId = append(out.SeasonId, v.SeasonId)
		out.Epid = append(out.Epid, v.Epid)
	}
	out.TitleRightPic, out.TitleRightText = b.constructTitleRight()
	return out, nil
}

func (b v4RemindBuilder) constructTitle() string {
	title := b.remind.List[0].UpdateTitle
	return title + emojiMap[crc32.ChecksumIEEE([]byte(b.rcmd.TrackID))%4]
}

func (b v4RemindBuilder) constructCover() string {
	return b.ConstructRemindCover(b.remind)
}

func (b v4RemindBuilder) constructURI() string {
	return b.ConstructRemindURI(b.remind)
}

func (b v4RemindBuilder) constructTitleRight() (appcardmodel.Icon, string) {
	updates := b.remind.Updates
	//nolint:gomnd
	if updates > 99 {
		return appcardmodel.IconBomb, strconv.Itoa(99)
	}
	return appcardmodel.IconTV, strconv.Itoa(updates)
}
