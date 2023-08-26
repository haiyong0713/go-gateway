package jsoncommon

import (
	"fmt"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
)

type ThreePoint struct{}

const (
	_noSeason                      = 1
	_region                        = 2
	_channel                       = 3
	_upper                         = 4
	_dislikeToast                  = "将减少相似内容推荐"
	_feedbackToast                 = "将优化首页此类内容"
	_watched                       = 10
	_moreContent                   = 12
	_repeatedRcmd                  = 13
	_addMoreContentAndRepeatedRcmd = 1
)

var (
	_defaultFeedbacks = []*jsoncard.DislikeReason{
		{ID: 1, Name: "恐怖血腥", Toast: _feedbackToast},
		{ID: 2, Name: "色情低俗", Toast: _feedbackToast},
		{ID: 3, Name: "封面恶心", Toast: _feedbackToast},
		{ID: 4, Name: "标题党/封面党", Toast: _feedbackToast},
	}
	_ogvFeedbacks = []*jsoncard.DislikeReason{
		{ID: 3, Name: "封面恶心", Toast: _feedbackToast},
		{ID: 4, Name: "标题党/封面党", Toast: _feedbackToast},
	}
	_noReasonDislike = []*jsoncard.DislikeReason{
		{ID: _noSeason, Name: "不感兴趣", Toast: _dislikeToast},
	}
	_watchedDislike = []*jsoncard.DislikeReason{
		{ID: _watched, Name: "看过了", Toast: _dislikeToast},
		{ID: _noSeason, Name: "不感兴趣", Toast: _dislikeToast},
	}
)

func constructArchiveDislikeReasons(args *jsoncard.Args, avDislikeInfo int8) []*jsoncard.DislikeReason {
	out := make([]*jsoncard.DislikeReason, 0, 4)
	if args.UpName != "" {
		out = append(out, &jsoncard.DislikeReason{
			ID:    _upper,
			Name:  fmt.Sprintf("UP主:%s", args.UpName),
			Toast: _dislikeToast,
		})
	}
	if args.Rname != "" {
		out = append(out, &jsoncard.DislikeReason{
			ID:    _region,
			Name:  fmt.Sprintf("分区:%s", args.Rname),
			Toast: _dislikeToast,
		})
	}
	if args.Tname != "" {
		out = append(out, &jsoncard.DislikeReason{
			ID:    _channel,
			Name:  fmt.Sprintf("频道:%s", args.Tname),
			Toast: _dislikeToast,
		})
	}
	if avDislikeInfo == _addMoreContentAndRepeatedRcmd {
		out = append(out, &jsoncard.DislikeReason{
			ID:    _moreContent,
			Name:  "此类内容过多",
			Toast: _dislikeToast,
		}, &jsoncard.DislikeReason{
			ID:    _repeatedRcmd,
			Name:  "推荐过",
			Toast: _dislikeToast,
		})
	}
	out = append(out, &jsoncard.DislikeReason{
		ID:    _noSeason,
		Name:  "不感兴趣",
		Toast: _dislikeToast,
	})
	return out
}

func (ThreePoint) ConstructArchvieThreePoint(args *jsoncard.Args, avDislikeInfo int8) *jsoncard.ThreePoint {
	out := &jsoncard.ThreePoint{}
	out.DislikeReasons = constructArchiveDislikeReasons(args, avDislikeInfo)
	out.Feedbacks = _defaultFeedbacks
	out.WatchLater = 1
	return out
}

func (ThreePoint) ConstructArchvieThreePointV2(ctx cardschema.FeedContext, args *jsoncard.Args, opts ...ThreePointOption) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	cfg := &threePointConfig{}
	cfg.Apply(opts...)
	if cfg.watchLater {
		out = append(out, &jsoncard.ThreePointV2{
			Title: "添加至稍后再看",
			Type:  appcardmodel.ThreePointWatchLater,
			Icon:  appcardmodel.IconWatchLater,
		})
	}
	if cfg.switchColumn {
		out = append(out, constructSwitchColumnThreePoint(appcardmodel.ColumnStatus(ctx.IndexParam().Column())))
	}
	if ctx.VersionControl().Can("feed.enableThreePointV2Feedback") {
		out = append(out, &jsoncard.ThreePointV2{
			Title:    "反馈",
			Subtitle: "(选择后将优化首页此类内容)",
			Reasons:  _defaultFeedbacks,
			Type:     appcardmodel.ThreePointFeedback,
		})
	}
	dislikeSubTitle, _, dislikeTitle := dislikeText(ctx)
	reasons := constructArchiveDislikeReasons(args, cfg.avDislikeInfo)
	replaceDislikeReason(ctx, reasons, cfg)
	out = append(out, &jsoncard.ThreePointV2{
		Title:    dislikeTitle,
		Subtitle: dislikeSubTitle,
		Reasons:  reasons,
		Type:     appcardmodel.ThreePointDislike,
	})
	return out
}

func constructSwitchColumnThreePoint(column appcardmodel.ColumnStatus) *jsoncard.ThreePointV2 {
	title := ""
	type_ := ""
	toast := ""
	subTitle := "(首页模式)"
	icon := ""
	switch column {
	case appcardmodel.ColumnSvrDouble, appcardmodel.ColumnDefault, appcardmodel.ColumnUserDouble:
		title = "切换至单列"
		type_ = appcardmodel.ThreePointSwitchToSingle
		toast = "已成功切换至单列模式"
		icon = appcardmodel.IconSwitchToSingle
	case appcardmodel.ColumnSvrSingle, appcardmodel.ColumnUserSingle:
		title = "切换至双列"
		type_ = appcardmodel.ThreePointSwitchToDouble
		toast = "已成功切换至双列模式"
		icon = appcardmodel.IconSwitchToDouble
	default:
	}
	return &jsoncard.ThreePointV2{
		Title:    title,
		Type:     type_,
		Toast:    toast,
		Subtitle: subTitle,
		Icon:     icon,
	}
}

func (ThreePoint) ConstructDefaultThreePoint() *jsoncard.ThreePoint {
	out := &jsoncard.ThreePoint{}
	out.DislikeReasons = []*jsoncard.DislikeReason{
		{ID: _noSeason, Name: "不感兴趣", Toast: _dislikeToast},
	}
	return out
}

func (ThreePoint) ConstructDefaultThreePointV2(ctx cardschema.FeedContext, switchColumn bool) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	if switchColumn {
		out = append(out, constructSwitchColumnThreePoint(appcardmodel.ColumnStatus(ctx.IndexParam().Column())))
	}

	out = append(out, &jsoncard.ThreePointV2{
		Title: "不感兴趣",
		Type:  appcardmodel.ThreePointDislike,
		ID:    _noSeason,
	})
	return out
}

func (ThreePoint) ConstructDefaultThreePointV2Legacy(ctx cardschema.FeedContext, switchColumn bool) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	if switchColumn {
		out = append(out, constructSwitchColumnThreePoint(appcardmodel.ColumnStatus(ctx.IndexParam().Column())))
	}
	_, dislikeReasonToast, _ := dislikeText(ctx)
	out = append(out, &jsoncard.ThreePointV2{
		Reasons: []*jsoncard.DislikeReason{
			{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast},
		},
		Type: appcardmodel.ThreePointDislike,
	})
	return out
}

func (ThreePoint) ConstructOGVThreePointV2(ctx cardschema.FeedContext, switchColumn bool, enableFeedback bool,
	enableWatched bool) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	if switchColumn {
		out = append(out, constructSwitchColumnThreePoint(appcardmodel.ColumnStatus(ctx.IndexParam().Column())))
	}
	dislikeSubTitle, _, dislikeTitle := dislikeText(ctx)
	if enableFeedback {
		out = append(out, &jsoncard.ThreePointV2{
			Title:    "反馈",
			Subtitle: "(选择后将优化首页此类内容)",
			Reasons:  _ogvFeedbacks,
			Type:     appcardmodel.ThreePointFeedback,
		})
		reason := _noReasonDislike
		if enableWatched {
			reason = _watchedDislike
		}
		replaceDislikeReason(ctx, reason, &threePointConfig{})
		out = append(out, &jsoncard.ThreePointV2{
			Title:    dislikeTitle,
			Subtitle: dislikeSubTitle,
			Reasons:  reason,
			Type:     appcardmodel.ThreePointDislike,
		})
		return out
	}
	out = append(out, &jsoncard.ThreePointV2{
		Title: "不感兴趣",
		Type:  appcardmodel.ThreePointDislike,
		ID:    _noSeason,
	})
	return out
}

func (ThreePoint) ConstructOGVThreePointV2Legacy(ctx cardschema.FeedContext, switchColumn bool, enableFeedback bool,
	enableWatched bool) []*jsoncard.ThreePointV2 {
	out := []*jsoncard.ThreePointV2{}
	if switchColumn {
		out = append(out, constructSwitchColumnThreePoint(appcardmodel.ColumnStatus(ctx.IndexParam().Column())))
	}
	dislikeSubTitle, dislikeReasonToast, dislikeTitle := dislikeText(ctx)
	if enableFeedback {
		out = append(out, &jsoncard.ThreePointV2{
			Title:    "反馈",
			Subtitle: "(选择后将优化首页此类内容)",
			Reasons:  _ogvFeedbacks,
			Type:     appcardmodel.ThreePointFeedback,
		})
		reason := _noReasonDislike
		if enableWatched {
			reason = _watchedDislike
		}
		replaceDislikeReason(ctx, reason, &threePointConfig{})
		out = append(out, &jsoncard.ThreePointV2{
			Title:    dislikeTitle,
			Subtitle: dislikeSubTitle,
			Reasons:  reason,
			Type:     appcardmodel.ThreePointDislike,
		})
		return out
	}
	out = append(out, &jsoncard.ThreePointV2{
		Reasons: []*jsoncard.DislikeReason{
			{ID: _noSeason, Name: "不感兴趣", Toast: dislikeReasonToast},
		},
		Type: appcardmodel.ThreePointDislike,
	})
	return out
}

func dislikeText(ctx cardschema.FeedContext) (string, string, string) {
	dislikeSubTitle := "(选择后将减少相似内容推荐)"
	dislikeReasonToast := _dislikeToast
	dislikeTitle := "不感兴趣"
	if ctx.FeatureGates().FeatureEnabled(cardschema.FeatureCloseRcmd) {
		dislikeSubTitle = ""
		dislikeReasonToast = ""
	}
	if ctx.FeatureGates().FeatureEnabled(cardschema.FeatureDislikeText) {
		dislikeTitle = "我不想看"
	}
	return dislikeSubTitle, dislikeReasonToast, dislikeTitle
}

func replaceDislikeReason(ctx cardschema.FeedContext, reasons []*jsoncard.DislikeReason, cfg *threePointConfig) {
	card.ReplaceStoryDislikeReason(reasons, cfg.item)
	if !ctx.FeatureGates().FeatureEnabled(cardschema.FeatureCloseRcmd) {
		return
	}
	for _, reason := range reasons {
		reason.Toast = "将在开启个性化推荐后生效"
	}
}

type threePointConfig struct {
	watchLater    bool
	switchColumn  bool
	avDislikeInfo int8
	item          *ai.Item
}

func (f *threePointConfig) Apply(opts ...ThreePointOption) {
	for _, opt := range opts {
		opt(f)
	}
}

type ThreePointOption func(*threePointConfig)

func WatchLater(in bool) ThreePointOption {
	return func(cfg *threePointConfig) {
		cfg.watchLater = in
	}
}

func SwitchColumn(in bool) ThreePointOption {
	return func(cfg *threePointConfig) {
		cfg.switchColumn = in
	}
}

func AvDislikeInfo(in int8) ThreePointOption {
	return func(cfg *threePointConfig) {
		cfg.avDislikeInfo = in
	}
}

func Item(in *ai.Item) ThreePointOption {
	return func(cfg *threePointConfig) {
		cfg.item = in
	}
}
