package jsonsmallcover

import (
	"fmt"
	"strconv"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcard "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	jsoncommon "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/common"
	jsonreasonstyle "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/reason_style"

	article "git.bilibili.co/bapis/bapis-go/article/model"

	"github.com/pkg/errors"
)

//go:generate python3 ../../../../contrib/desc-button-overlapped.py

type V2ArticleBuilder interface {
	Parent() SmallCoverV2BuilderFactory
	SetBase(*jsoncard.Base) V2ArticleBuilder
	SetRcmd(*ai.Item) V2ArticleBuilder
	SetArticle(*article.Meta) V2ArticleBuilder

	Build() (*jsoncard.SmallCoverV2, error)
	WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2ArticleBuilder
}

type v2ArticleBuilder struct {
	jsoncommon.ArticleCommon
	parent  *smallCoverV2BuilderFactory
	base    *jsoncard.Base
	rcmd    *ai.Item
	article *article.Meta
	afterFn []func(*jsoncard.SmallCoverV2)
}

func (b v2ArticleBuilder) Parent() SmallCoverV2BuilderFactory {
	return b.parent
}

func (b v2ArticleBuilder) SetBase(base *jsoncard.Base) V2ArticleBuilder {
	b.base = base
	return b
}

func (b v2ArticleBuilder) SetRcmd(item *ai.Item) V2ArticleBuilder {
	b.rcmd = item
	return b
}

func (b v2ArticleBuilder) SetArticle(in *article.Meta) V2ArticleBuilder {
	b.article = in
	return b
}

func (b v2ArticleBuilder) constructURI() string {
	device := b.parent.BuilderContext.Device()
	return appcardmodel.FillURI(appcardmodel.GotoArticle, device.Plat(), int(device.Build()), strconv.FormatInt(b.article.ID, 10), nil)
}

func (b v2ArticleBuilder) constructArgs() jsoncard.Args {
	args := jsoncard.Args{}
	if b.article.Author != nil {
		args.UpID = b.article.Author.Mid
		args.UpName = b.article.Author.Name
	}
	if len(b.article.Categories) != 0 {
		if b.article.Categories[0] != nil {
			args.Rid = int32(b.article.Categories[0].ID)
			args.Rname = b.article.Categories[0].Name
		}
		if len(b.article.Categories) > 1 {
			if b.article.Categories[1] != nil {
				args.Tid = b.article.Categories[1].ID
				args.Tname = b.article.Categories[1].Name
			}
		}
	}
	return args
}

func (b v2ArticleBuilder) constructDescButton() *jsoncard.Button {
	//nolint:gomnd
	if len(b.article.Categories) < 2 {
		return nil
	}
	name := ""
	if b.article.Categories[0] != nil {
		name = b.article.Categories[0].Name
		if b.article.Categories[1] != nil {
			name = fmt.Sprintf("%s · %s", name, b.article.Categories[1].Name)
		}
	}
	return &jsoncard.Button{
		Type: appcardmodel.ButtonGrey,
		Text: name,
		URI: appcardmodel.FillURI(appcardmodel.GotoArticleTag, 0, 0, "",
			appcardmodel.ArticleTagHandler(b.article.Categories, b.parent.BuilderContext.Device().Plat())),
		Event:   appcardmodel.EventChannelClick,
		EventV2: appcardmodel.EventV2ChannelClick,
	}
}

func (b v2ArticleBuilder) constructThreePoint() *jsoncard.ThreePoint {
	return b.ConstructThreePointFromArticle(b.article)
}

func (b v2ArticleBuilder) constructThreePointV2() []*jsoncard.ThreePointV2 {
	return b.ConstructThreePointV2FromArticle(b.article)
}

func (b v2ArticleBuilder) Build() (*jsoncard.SmallCoverV2, error) {
	if b.base == nil {
		return nil, errors.Errorf("empty `base` field")
	}
	if b.rcmd == nil {
		return nil, errors.Errorf("empty `rcmd` field")
	}
	if b.article == nil {
		return nil, errors.Errorf("empty `article` field")
	}
	if appcard.CheckMidMaxInt32(b.article.Author.Mid) && b.parent.BuilderContext.VersionControl().Can("feed.disableInt64Mid") {
		return nil, errors.Errorf("ignore on maxint32 mid: %d", b.article.Author.Mid)
	}
	if len(b.article.ImageURLs) == 0 {
		return nil, errors.Errorf("empty `ImageURLs`")
	}

	if err := jsonbuilder.NewBaseUpdater(b.parent.BuilderContext, b.base).
		UpdateCover(b.article.ImageURLs[0]).
		UpdateTitle(b.article.Title).
		UpdateArgs(b.constructArgs()).
		UpdateURI(b.constructURI()).
		UpdateThreePoint(b.constructThreePoint()).
		UpdateThreePointV2(b.constructThreePointV2()).Update(); err != nil {
		return nil, err
	}
	out := &jsoncard.SmallCoverV2{}
	out.DescButton = b.constructDescButton()
	if b.article.Stats != nil {
		out.CoverLeftText1 = appcardmodel.StatString(int32(b.article.Stats.View), "")
		out.CoverLeftIcon1 = appcardmodel.IconRead
		out.CoverLeft1ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon1,
			out.CoverLeftText1)
		out.CoverLeftText2 = appcardmodel.StatString(int32(b.article.Stats.Reply), "")
		out.CoverLeftIcon2 = appcardmodel.IconComment
		out.CoverLeft2ContentDescription = appcardmodel.CoverIconContentDescription(out.CoverLeftIcon2,
			out.CoverLeftText2)
	}
	out.Badge = "文章"
	out.BadgeStyle = jsonreasonstyle.ConstructReasonStyle(appcardmodel.BgColorTransparentRed, "文章")

	if b.rcmd.RcmdReason != nil {
		out.DescButton = nil
		out.RcmdReasonStyle = jsonreasonstyle.ConstructTopReasonStyle(b.rcmd.RcmdReason.Content,
			jsonreasonstyle.CornerMarkFromAI(b.rcmd),
			jsonreasonstyle.CorverMarkFromContext(b.parent.BuilderContext))
	}
	out.Base = b.base
	for _, fn := range b.afterFn {
		fn(out)
	}
	return out, nil
}

func (b v2ArticleBuilder) WithAfter(req ...func(*jsoncard.SmallCoverV2)) V2ArticleBuilder {
	b.afterFn = append(b.afterFn, req...)
	return b
}
