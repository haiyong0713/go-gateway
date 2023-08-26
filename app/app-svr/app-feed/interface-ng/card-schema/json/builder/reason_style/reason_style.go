package jsonreasonstyle

import (
	"strconv"

	"go-common/library/log"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	appcardai "go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"

	"github.com/pkg/errors"
)

var (
	_inlineGotoSet = sets.NewString("ad_inline_av")
)

func BuildRecommendReasonText(ctx jsonbuilder.BuilderContext, rcmdReason *appcardai.RcmdReason, gotoType string, name string, isAttention bool) (string, string) {
	if rcmdReason == nil {
		if _inlineGotoSet.Has(gotoType) {
			return "", ""
		}
		if isAttention {
			return "已关注", name
		}
		return "", ""
	}
	//nolint:gomnd
	switch rcmdReason.Style {
	case 3:
		if !isAttention {
			return "", ""
		}
		return rcmdReason.Content, name
	case 4:
		// https://info.bilibili.co/pages/viewpage.action?pageId=4551903
		// style 4 动态图文卡片样式，额外添加字段【 "followed_mid" : 123 】，保存的是【召回此动态卡片的关注用户mid】，服务端需要根据此mid拿用户昵称
		if !ctx.IsAttentionTo(rcmdReason.FollowedMid) {
			return "", ""
		}
		return "关注的人赞过", ""
	case 6:
		return rcmdReason.Content, name
	default:
		return rcmdReason.Content, ""
	}
}

func cornerMarkToTopBGStyle(cornerMark int8) int8 {
	//nolint:gomnd
	switch cornerMark {
	case 0, 2:
		return appcardmodel.BgColorOrange
	case 1:
		return appcardmodel.BgColorTransparentOrange
	case 3:
		return appcardmodel.BgTransparentTextOrange
	case 4:
		return appcardmodel.BgColorRed
	case 5:
		return appcardmodel.BgColorFillingOrange
	case 6:
		return appcardmodel.BgColorLumpOrange
	default:
		return appcardmodel.BgColorOrange
	}
}

func cornerMarkToBottomBGStyle(cornerMark int8) int8 {
	//nolint:gomnd
	switch cornerMark {
	case 1:
		return appcardmodel.BgColorTransparentOrange
	case 3:
		return appcardmodel.BgTransparentTextOrange
	case 5:
		return appcardmodel.BgColorFillingOrange
	case 6:
		return appcardmodel.BgColorLumpOrange
	default:
		return appcardmodel.BgColorOrange
	}
}

type cornerMarkCalculator func(cornerMarkContainer *int8) error

// FIXME
func CornerMarkFromAI(ai *appcardai.Item) cornerMarkCalculator {
	return func(cornerMarkContainer *int8) error {
		if ai == nil {
			return errors.Errorf("empty AI input")
		}
		*cornerMarkContainer = ai.CornerMark
		if *cornerMarkContainer != 0 {
			return nil
		}
		if ai.RcmdReason == nil {
			return nil
		}
		*cornerMarkContainer = ai.RcmdReason.CornerMark
		if ai.RcmdReason.Content == "" {
			//nolint:gomnd
			if ai.RcmdReason.Style == 4 {
				return nil
			}
			*cornerMarkContainer = 0
			return nil
		}
		return nil
	}
}

func CornerMarkWithValue(cornerMark int8) cornerMarkCalculator {
	return func(cornerMarkContainer *int8) error {
		*cornerMarkContainer = cornerMark
		return nil
	}
}

func CorverMarkFromContext(ctx cardschema.FeedContext) cornerMarkCalculator {
	return func(cornerMarkContainer *int8) error {
		if ctx.VersionControl().Can("feed.usingNewRcmdReason") {
			*cornerMarkContainer = 5
		}
		if ctx.VersionControl().Can("feed.usingNewRcmdReasonV2") {
			*cornerMarkContainer = 6
		}
		return nil
	}
}

func ConstructTopReasonStyle(text string, cornerMarkFn ...cornerMarkCalculator) *jsoncard.ReasonStyle {
	if text == "" {
		return nil
	}
	cornerMark := int8(0)
	for _, cmFn := range cornerMarkFn {
		if err := cmFn(&cornerMark); err != nil {
			log.Error("Failed to calculate corner mark: %+v", err)
			return nil
		}
	}
	bgStyle := cornerMarkToTopBGStyle(cornerMark)
	return ConstructReasonStyle(bgStyle, text)
}

func ConstructBottomReasonStyle(text string, cornerMarkFn ...cornerMarkCalculator) *jsoncard.ReasonStyle {
	if text == "" {
		return nil
	}
	cornerMark := int8(0)
	for _, cmFn := range cornerMarkFn {
		if err := cmFn(&cornerMark); err != nil {
			log.Error("Failed to calculate corner mark: %+v", err)
			return nil
		}
	}
	bgStyle := cornerMarkToBottomBGStyle(cornerMark)
	return ConstructReasonStyle(bgStyle, text)
}

func ConstructReasonStyle(style int8, text string) *jsoncard.ReasonStyle {
	if text == "" {
		return nil
	}
	out := &jsoncard.ReasonStyle{
		Text: text,
	}
	switch style {
	case appcardmodel.BgColorOrange: //defalut
		// 白天
		out.TextColor = "#FFFFFFFF"
		out.BgColor = "#FFFB9E60"
		out.BorderColor = "#FFFB9E60"
		out.BgStyle = appcardmodel.BgStyleFill
		// 夜间
		out.TextColorNight = "#E5E5E5"
		out.BgColorNight = "#BC7A4F"
		out.BorderColorNight = "#BC7A4F"
	case appcardmodel.BgColorTransparentOrange:
		// 白天
		out.TextColor = "#FFFB9E60"
		out.BorderColor = "#FFFB9E60"
		out.BgStyle = appcardmodel.BgStyleStroke
		// 夜间
		out.TextColorNight = "#BC7A4F"
		out.BorderColorNight = "#BC7A4F"
	case appcardmodel.BgColorBlue:
		out.TextColor = "#FF23ADE5"
		out.BgColor = "#3323ADE5"
		out.BorderColor = "#3323ADE5"
		out.BgStyle = appcardmodel.BgStyleFill
	case appcardmodel.BgColorRed:
		// 白天
		out.TextColor = "#FFFFFF"
		out.BgColor = "#FB7299"
		out.BorderColor = "#FB7299"
		out.BgStyle = appcardmodel.BgStyleFill
		// 夜间
		out.TextColorNight = "#E5E5E5"
		out.BgColorNight = "#BB5B76"
		out.BorderColorNight = "#BB5B76"
	case appcardmodel.BgTransparentTextOrange:
		out.TextColor = "#FFFB9E60"
		out.BgStyle = appcardmodel.BgStyleNoFillAndNoStroke
	case appcardmodel.BgColorPurple:
		out.TextColor = "#FFFFFFFF"
		out.BgColor = "#FF7D75F2"
		out.BorderColor = "#FF7D75F2"
		out.BgStyle = appcardmodel.BgStyleFill
	case appcardmodel.BgColorTransparentRed:
		// 白天
		out.TextColor = "#FB7299"
		out.BorderColor = "#FB7299"
		out.BgStyle = appcardmodel.BgStyleStroke
		// 夜间
		out.TextColorNight = "#BB5B76"
		out.BorderColorNight = "#BB5B76"
	case appcardmodel.BgColorFillingOrange:
		// 白天
		out.TextColor = "#FF6633"
		out.BgColor = "#FFF1ED"
		out.BorderColor = "#FFF1ED"
		out.BgStyle = appcardmodel.BgStyleFill
		// 夜间
		out.TextColorNight = "#BF5330"
		out.BgColorNight = "#BFB5B2"
		out.BorderColorNight = "#BFB5B2"
	case appcardmodel.BgColorYellow:
		// 白天
		out.TextColor = "#FFFFFF"
		out.BgColor = "#FAAB4B"
		out.BorderColor = "#FAAB4B"
		out.BgStyle = appcardmodel.BgStyleFill
		// 夜间
		out.TextColorNight = "#E5E5E5"
		out.BgColorNight = "#BA833F"
		out.BorderColorNight = "#BA833F"
	case appcardmodel.BgColorLumpOrange:
		// 白天
		out.TextColor = "#FF6633"
		out.BgColor = "#FFF1ED"
		out.BorderColor = "#FFF1ED"
		out.BgStyle = appcardmodel.BgStyleFill
		// 夜间
		out.TextColorNight = "#BF5330"
		out.BgColorNight = "#3D2D29"
		out.BorderColorNight = "#3D2D29"
	default:
		log.Warn("Unrecognized reason style: %q", style)
	}
	return out
}

func WithReasonStyleV2(ai *appcardai.Item, device cardschema.Device) func(*jsoncard.ReasonStyle) error {
	return func(reasonStyle *jsoncard.ReasonStyle) error {
		if ai.RcmdReason == nil {
			return errors.Errorf("empty `rcmd_reason` field")
		}
		if ai.RcmdReason.JumpGoto == "" {
			return errors.Errorf("empty `jumpgoto` field")
		}

		cornerMark := ai.CornerMark
		if ai.RcmdReason.Content != "" {
			cornerMark = ai.RcmdReason.CornerMark
		}
		jumpGoto := appcardmodel.Gt(ai.RcmdReason.JumpGoto)
		//nolint:gomnd
		switch cornerMark {
		case 0, 2:
			//nolint:exhaustive
			switch jumpGoto {
			case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/6983dc5b73d32a8241421a7b25f78c855b8e0362.png"
			case appcardmodel.GotoPlaylist:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/b985518f35ad2b12eec34d4b3b6dca33df2b85a2.png"
			case appcardmodel.GotoTag:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/c735ef9f33feb19f52bce852355e72a6b367c466.png"
			case appcardmodel.GotoHotPage:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/e257e216b965905774b1ef9d526dfd2d8dae02f2.png"
			}
		case 1:
			//nolint:exhaustive
			switch jumpGoto {
			case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/8ba6d17e066f6ad3497e071abe654615fb073726.png"
			case appcardmodel.GotoPlaylist:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/e0bd607cb58289fb32866e19ec207efae23657de.png"
			case appcardmodel.GotoTag:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/0d1185d6ceca3de0e0a99bdce0479838265adcc3.png"
			case appcardmodel.GotoHotPage:
				reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/d0f105544b1b0df77e8795ae297a72e627642b43.png"
			}
		}
		var urlExtraFn func(string) string
		//nolint:exhaustive
		switch jumpGoto {
		case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
			jumpGoto = appcardmodel.GotoAvConverge
			urlExtraFn = appcardmodel.TrackIDHandler(ai.TrackID, ai, device.Plat(), int(device.Build()))
		}
		reasonStyle.Event = appcardmodel.EventButtonClick
		reasonStyle.EventV2 = appcardmodel.EventV2ButtonClick
		reasonStyle.URI = appcardmodel.FillURI(jumpGoto, 0, 0, strconv.FormatInt(ai.RcmdReason.JumpID, 10), urlExtraFn)
		return nil
	}
}

func WithReasonStyleV3(ai *appcardai.Item, device cardschema.Device) func(*jsoncard.ReasonStyle) error {
	return func(reasonStyle *jsoncard.ReasonStyle) error {
		if ai.RcmdReason == nil {
			return errors.Errorf("empty `rcmd_reason` field")
		}
		if ai.RcmdReason.JumpGoto == "" {
			return errors.Errorf("empty `jumpgoto` field")
		}

		_rightIconOrange := int8(1) // 带icon的推荐理由最右边的小箭头、展示橙色
		reasonStyle.RightIconType = _rightIconOrange
		jumpGoto := appcardmodel.Gt(ai.RcmdReason.JumpGoto)
		//nolint:exhaustive
		switch jumpGoto {
		case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/c705fe617230bcfce0234c8a7323deb68350a209.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/6632d96c3504047c04a0118b5e6bae95241f7680.png"
		case appcardmodel.GotoPlaylist:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/87182dd087b82d2e07928046d5a78311982ce04d.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/14e626a61674081aabf8e1ea77f2c9122077951a.png"
		case appcardmodel.GotoTag:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/5df817cb2ca2c325534d940c3d1381bf6ac78258.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/1a3f717a993ae671b31c18b58c6d20a305097141.png"
		case appcardmodel.GotoHotPage:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/5c34304eff82a432c8320a67d2977fece9d500a5.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/c303e397d707c84e736c31e8871f57206c7778b0.png"
		}
		var urlExtraFn func(string) string
		//nolint:exhaustive
		switch jumpGoto {
		case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
			jumpGoto = appcardmodel.GotoAvConverge
			urlExtraFn = appcardmodel.TrackIDHandler(ai.TrackID, ai, device.Plat(), int(device.Build()))
		}
		reasonStyle.Event = appcardmodel.EventButtonClick
		reasonStyle.EventV2 = appcardmodel.EventV2ButtonClick
		reasonStyle.URI = appcardmodel.FillURI(jumpGoto, 0, 0, strconv.FormatInt(ai.RcmdReason.JumpID, 10), urlExtraFn)

		return nil
	}
}

func WithReasonStyleV4(ai *appcardai.Item, device cardschema.Device) func(*jsoncard.ReasonStyle) error {
	return func(reasonStyle *jsoncard.ReasonStyle) error {
		if ai.RcmdReason == nil {
			return errors.Errorf("empty `rcmd_reason` field")
		}
		if ai.RcmdReason.JumpGoto == "" {
			return errors.Errorf("empty `jumpgoto` field")
		}

		_rightIconOrange := int8(1) // 带icon的推荐理由最右边的小箭头、展示橙色
		reasonStyle.RightIconType = _rightIconOrange
		//nolint:exhaustive
		switch appcardmodel.Gt(ai.RcmdReason.JumpGoto) {
		case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/09fea8d3ed60aae6f5f7e8147fce12379dbff726.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/d4c711c39d5c95b29f5730389536952bdb2b7e9c.png"
		case appcardmodel.GotoPlaylist:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/b568780ef7310490d16a2c7a257f7542a98e59f0.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/e1a96822f3b6908253761fbdbadb65d354b26ebd.png"
		case appcardmodel.GotoTag:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/703c23ab27abf70bd03167a9606c9ba1e9895caa.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/32499393ba3be99d342811bdf8ff597ac128fc77.png"
		case appcardmodel.GotoHotPage:
			reasonStyle.IconURL = "https://i0.hdslb.com/bfs/archive/c9ca993374ecef309c63044e1cd135977fb3ae88.png"
			reasonStyle.IconURLNight = "https://i0.hdslb.com/bfs/archive/615ddf4dd574ab54d0cdb2b07de6fddad887b277.png"
		}
		jumpGoto := appcardmodel.Gt(ai.RcmdReason.JumpGoto)
		var urlExtraFn func(string) string
		//nolint:exhaustive
		switch jumpGoto {
		case appcardmodel.GotoAvConverge, appcardmodel.GotoMultilayerConverge:
			jumpGoto = appcardmodel.GotoAvConverge
			urlExtraFn = appcardmodel.TrackIDHandler(ai.TrackID, ai, device.Plat(), int(device.Build()))
		}
		reasonStyle.Event = appcardmodel.EventButtonClick
		reasonStyle.EventV2 = appcardmodel.EventV2ButtonClick
		reasonStyle.URI = appcardmodel.FillURI(jumpGoto, 0, 0, strconv.FormatInt(ai.RcmdReason.JumpID, 10), urlExtraFn)
		return nil
	}
}

func constructReasonStyleCustomized(text string, extraFn func(*jsoncard.ReasonStyle) error, cornerMarkFn ...cornerMarkCalculator) *jsoncard.ReasonStyle {
	reasonStyle := ConstructTopReasonStyle(text, cornerMarkFn...)
	if reasonStyle == nil {
		return nil
	}
	if extraFn == nil {
		return reasonStyle
	}
	if err := extraFn(reasonStyle); err != nil {
		log.Error("Failed to restruct reason style: %+v", err)
		return nil
	}
	return reasonStyle
}

func ConstructReasonStyleV2(ctx cardschema.FeedContext, text string, ai *appcardai.Item) *jsoncard.ReasonStyle {
	return constructReasonStyleCustomized(text, WithReasonStyleV2(ai, ctx.Device()), CornerMarkFromAI(ai))
}

func ConstructReasonStyleV3(ctx cardschema.FeedContext, text string, ai *appcardai.Item) *jsoncard.ReasonStyle {
	return constructReasonStyleCustomized(text, WithReasonStyleV3(ai, ctx.Device()), CornerMarkWithValue(5))
}

func ConstructReasonStyleV4(ctx cardschema.FeedContext, text string, ai *appcardai.Item) *jsoncard.ReasonStyle {
	return constructReasonStyleCustomized(text, WithReasonStyleV4(ai, ctx.Device()), CornerMarkWithValue(6))
}

func ConstructIconBadgeStyleFromLive(live *live.Room, style int8) *jsoncard.ReasonStyle {
	out := &jsoncard.ReasonStyle{}
	if live.PendentRu == "" || live.PendentRuPic == "" {
		return out
	}
	out.Text = live.PendentRu
	out.TextColor = "#FFFFFFFF"
	out.IconBGURL = live.PendentRuPic
	return out
}

func BuildTopBottomRecommendReasonText(ctx jsonbuilder.BuilderContext, rcmdReason *appcardai.RcmdReason, gotoType string, isAttention bool) (string, string) {
	if rcmdReason == nil {
		if _inlineGotoSet.Has(gotoType) {
			return "", ""
		}
		if isAttention {
			return "", "已关注"
		}
		return "", ""
	}
	//nolint:gomnd
	switch rcmdReason.Style {
	case 3:
		if !isAttention {
			return "", ""
		}
		return "", rcmdReason.Content
	case 4:
		if !ctx.IsAttentionTo(rcmdReason.FollowedMid) {
			return "", ""
		}
		return "关注的人赞过", ""
	case 5:
		return "", rcmdReason.Content
	default:
		return rcmdReason.Content, ""
	}
}

func BuildInlineReasonText(rcmdReason *appcardai.RcmdReason, name string, isAttention bool, enableSingleRcmdReason bool) (string, string) {
	if rcmdReason == nil || !enableSingleRcmdReason {
		return "", ""
	}
	//nolint:gomnd
	switch rcmdReason.Style {
	case 3:
		if !isAttention {
			return "", ""
		}
		return rcmdReason.Content, name
	case 6:
		return rcmdReason.Content, name
	default:
		return rcmdReason.Content, ""
	}
}
