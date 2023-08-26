package dynHandler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/log"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/pkg/idsafe/bvid"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

const (
	_handleTypeForward             = "forward"
	_ShowPlayIconKey               = "ShowPlayIcon"
	_moduleDynamicPlayIcon         = "https://i0.hdslb.com/bfs/feed-admin/2269afa7897830b397797ebe5f032b899b405c67.png"
	_moduleDynamicCommonBadgeStyle = 1
)

func (schema *CardSchema) dynCardForward(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.Origin != nil {
		if !dynCtx.Dyn.Origin.Visible {
			dynCtx.Interim.IsPassCard = true // 源卡不可展示
		}
	}
	if dynCtx.Interim.ForwardOrigFaild {
		dynCtx.Interim.IsPassCard = true // 不展示资源失效的转发卡
	}
	// 获取拼接卡片的func handler
	var (
		dyn       = new(dynmdlV2.Dynamic)
		dynCtxTmp = new(dynmdlV2.DynamicContext)
		foldList  *topiccardmodel.DynRawList
	)
	*dyn = *dynCtx.Dyn.Origin
	*dynCtxTmp = *dynCtx
	// 感知转卡信息
	dyn.Forward = dynCtx.Dyn
	// 转发卡拼接
	foldList = schema.ProcListReply(deepCopyDynSchemaCtx(dynCtxTmp, dynSchemaCtx), []*dynmdlV2.Dynamic{dyn}, general, _handleTypeForward)
	if len(foldList.List) == 0 || foldList.List[0].Item == nil {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	var (
		card        = new(dynamicapi.MdlDynForward)
		dynamicItem = foldList.List[0].Item
	)
	card.Item = dynamicItem
	card.Rtype = dynCtx.Dyn.RType
	// 转发卡物料
	// 表情
	for k := range dynCtxTmp.Emoji {
		if dynCtx.Emoji == nil {
			dynCtx.Emoji = make(map[string]struct{})
		}
		dynCtx.Emoji[k] = struct{}{}
	}
	// cv
	for k := range dynCtxTmp.BackfillCvID {
		if dynCtx.BackfillCvID == nil {
			dynCtx.BackfillCvID = make(map[string]struct{})
		}
		dynCtx.BackfillCvID[k] = struct{}{}
	}
	// bv
	for k := range dynCtxTmp.BackfillBvID {
		if dynCtx.BackfillBvID == nil {
			dynCtx.BackfillBvID = make(map[string]struct{})
		}
		dynCtx.BackfillBvID[k] = struct{}{}
	}
	// av
	for k := range dynCtxTmp.BackfillAvID {
		if dynCtx.BackfillAvID == nil {
			dynCtx.BackfillAvID = make(map[string]struct{})
		}
		dynCtx.BackfillAvID[k] = struct{}{}
	}
	// url
	for k := range dynCtxTmp.BackfillDescURL {
		if dynCtx.BackfillDescURL == nil {
			dynCtx.BackfillDescURL = make(map[string]*dynmdlV2.BackfillDescURLItem)
		}
		dynCtx.BackfillDescURL[k] = nil
	}
	// 扩展字段
	dynCtx.Interim.VoteID = dynCtxTmp.Interim.VoteID
	dynCtx.DynamicItem.Extend.OrigDynIdStr = dynamicItem.Extend.DynIdStr
	dynCtx.DynamicItem.Extend.OrigName = dynamicItem.Extend.OrigName     // 转发卡使用内层物料数据
	dynCtx.DynamicItem.Extend.OrigImgUrl = dynamicItem.Extend.OrigImgUrl // 转发卡使用内层物料数据
	dynCtx.DynamicItem.Extend.OrigDesc = dynamicItem.Extend.OrigDesc     // 转发卡使用内层物料数据
	dynCtx.DynamicItem.Extend.OrigDynType = dynamicItem.CardType
	dynCtx.DynamicItem.ItemType = dynamicItem.CardType
	for _, origItem := range dynamicItem.Modules {
		switch origItem.ModuleType {
		case dynamicapi.DynModuleType_module_extend:
			origModule := origItem.ModuleItem.(*dynamicapi.Module_ModuleExtend)
			if origModule.ModuleExtend != nil && len(origModule.ModuleExtend.Extend) > 0 {
				dynCtx.Interim.IsPassExtend = true
			}
		case dynamicapi.DynModuleType_module_additional:
			origModule := origItem.ModuleItem.(*dynamicapi.Module_ModuleAdditional)
			if origModule.ModuleAdditional != nil {
				dynCtx.Interim.IsPassAddition = true
			}
		default:
		}
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type: dynamicapi.ModuleDynamicType_mdl_dyn_forward,
				ModuleItem: &dynamicapi.ModuleDynamic_DynForward{
					DynForward: card,
				},
			},
		},
	})
	return nil
}

func deepCopyDynSchemaCtx(dynCtxTmp *dynmdlV2.DynamicContext, dynSchemaCtx *topiccardmodel.DynSchemaCtx) *topiccardmodel.DynSchemaCtx {
	return &topiccardmodel.DynSchemaCtx{
		Ctx:        dynSchemaCtx.Ctx,
		DynCtx:     dynCtxTmp,
		DynCmtMode: dynSchemaCtx.DynCmtMode,
		TopicId:    dynSchemaCtx.TopicId,
		SortBy:     dynSchemaCtx.SortBy,
		Offset:     dynSchemaCtx.Offset,
	}
}

// nolint:gomnd
func (schema *CardSchema) numTransfer(num int) string {
	if num < 10000 {
		return strconv.Itoa(num)
	}
	integer := num / 10000
	decimals := num % 10000
	decimals = decimals / 1000
	return fmt.Sprintf("%d.%d万", integer, decimals)
}

// nolint:gocognit
func (schema *CardSchema) dynCardAv(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	if !ok {
		return nil
	}
	var archive = ap.Arc
	card := &dynamicapi.MdlDynArchive{
		Title:           schema.getTitle(archive.Title, dynCtx),
		Cover:           archive.Pic,
		CoverLeftText_1: topiccardmodel.VideoDuration(archive.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", schema.numTransfer(int(archive.Stat.View))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", schema.numTransfer(int(archive.Stat.Danmaku))),
		Avid:            archive.Aid,
		Cid:             archive.FirstCid,
		MediaType:       dynamicapi.MediaType_MediaTypeUGC,
		Dimension: &dynamicapi.Dimension{
			Height:          archive.Dimension.Height,
			Width:           archive.Dimension.Width,
			Rotate:          archive.Dimension.Rotate,
			ForceHorizontal: true,
		},
		Duration: archive.Duration,
		View:     archive.Stat.View,
	}
	card.Bvid, _ = bvid.AvToBv(archive.Aid)
	var (
		playurl *arcgrpc.PlayerInfo
	)
	if playurl, ok = ap.PlayerInfo[dynCtx.Interim.CID]; !ok {
		if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
			playurl = ap.PlayerInfo[ap.Arc.FirstCid]
		}
	}
	if playurl != nil && playurl.PlayerExtra != nil && playurl.PlayerExtra.Dimension != nil {
		card.Cid = playurl.PlayerExtra.Cid
		card.Dimension.Height = playurl.PlayerExtra.Dimension.Height
		card.Dimension.Width = playurl.PlayerExtra.Dimension.Width
		card.Dimension.Rotate = playurl.PlayerExtra.Dimension.Rotate
	}
	if dynCtx.Dyn.Property != nil && (dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_ARCHIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_PLAY_BACK) {
		// UP主预约是否召回
		card.ReserveType = dynamicapi.ReserveType_reserve_recall
	}
	if g, ok := dynCtx.Grayscale[_ShowPlayIconKey]; ok {
		switch g {
		case 1:
			card.PlayIcon = _moduleDynamicPlayIcon
		}
	}
	card.Uri = topiccardmodel.FillURI(topiccardmodel.GotoAv, strconv.FormatInt(archive.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
	// PGC特殊逻辑
	if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
		card.Uri = archive.RedirectURL
		card.IsPGC = true
		if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; ok && playurl.PlayerExtra != nil && playurl.PlayerExtra.PgcPlayerExtra != nil {
			if playurl.PlayerExtra.PgcPlayerExtra.IsPreview == 1 {
				card.IsPreview = true
			}
			card.EpisodeId = playurl.PlayerExtra.PgcPlayerExtra.EpisodeId
			card.SubType = playurl.PlayerExtra.PgcPlayerExtra.SubType
			card.PgcSeasonId = playurl.PlayerExtra.PgcPlayerExtra.PgcSeasonId
		}
	}
	// 小视频特殊处理
	card.Stype = dynmdlV2.GetArchiveSType(dynCtx.Dyn.SType)
	if card.Stype == dynamicapi.VideoType_video_type_story {
		card.Stype = dynamicapi.VideoType_video_type_dynamic
	}
	if card.Stype == dynamicapi.VideoType_video_type_dynamic {
		card.Title = ""
		card.IsPGC = false
	}
	if archive.Rights.IsCooperation == 1 {
		card.Badge = append(card.Badge, dynmdlV2.CooperationBadge)
	}
	if archive.Rights.UGCPay == 1 {
		card.Badge = append(card.Badge, dynmdlV2.PayBadge)
	}
	if card.Stype == dynamicapi.VideoType_video_type_playback {
		card.Badge = append(card.Badge, dynmdlV2.PlayBackBadge)
	}
	if len(card.Badge) == 0 {
		if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes || general.IsAndroidHD() || general.IsPad() {
			// 新的角标
			if dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.PgcBadge != nil && dynCtx.Dyn.PassThrough.PgcBadge.EpisodeId > 0 {
				if dynCtx.Dyn.PassThrough.PgcBadge.SectionType == 0 {
					// 是否是PGC正片，上报字段
					card.IsFeature = true
					card.IsPGC = true
				}
				if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.PassThrough.PgcBadge.EpisodeId)); ok {
					if pgc.Season != nil && pgc.SectionType == 0 {
						if dynCtx.Dyn.PassThrough.PgcBadge.Show {
							// 追番人数角标
							if pgc.Stat.FollowDesc != "" {
								card.BadgeCategory = append(card.BadgeCategory, dynmdlV2.BadgeStyleFrom(dynmdlV2.BgColorTransparentGray, pgc.Stat.FollowDesc))
							}
							// PGC角标
							if pgc.Season.TypeName != "" {
								card.BadgeCategory = append(card.BadgeCategory, dynmdlV2.BadgeStyleFrom(dynmdlV2.BgColorPink, pgc.Season.TypeName))
							}
						}
					}
				}
			}
		}
	}
	card.CanPlay = dynmdlV2.CanPlay(archive.Rights.Autoplay)
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type: dynamicapi.ModuleDynamicType_mdl_dyn_archive,
				ModuleItem: &dynamicapi.ModuleDynamic_DynArchive{
					DynArchive: card,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) dynCardDraw(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	draw, _ := dynCtx.GetResDraw(dynCtx.Dyn.Rid)
	card := &dynamicapi.MdlDynDraw{
		Uri: dynCtx.Interim.PromoURI,
		Id:  dynCtx.Dyn.Rid,
	}
	for _, pic := range draw.Item.Pictures {
		if pic == nil {
			log.Warn("dynCardDraw module error draw pic miss mid(%v) dynid(%v) type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
			continue
		}
		i := &dynamicapi.MdlDynDrawItem{
			Src:    pic.ImgSrc,
			Width:  pic.ImgWidth,
			Height: pic.ImgHeight,
			Size_:  pic.ImgSize,
		}
		for _, picTag := range pic.ImgTags {
			if picTag == nil {
				log.Warn("dynCardDraw module error draw pic_tag miss mid(%v) dynid(%v) type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
				continue
			}
			switch picTag.Type {
			case dynmdlV2.DrawTagTypeCommon:
				i.Tags = append(i.Tags, &dynamicapi.MdlDynDrawTag{
					Type: dynamicapi.MdlDynDrawTagType_mdl_draw_tag_common,
					Item: &dynamicapi.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Url:         picTag.Url,
					},
				})
			case dynmdlV2.DrawTagTypeGoods:
				i.Tags = append(i.Tags, &dynamicapi.MdlDynDrawTag{
					Type: dynamicapi.MdlDynDrawTagType_mdl_draw_tag_goods,
					Item: &dynamicapi.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Url:         picTag.Url,
						ItemId:      picTag.ItemID,
						Source:      picTag.Source,
						SchemaUrl:   picTag.SchemaURL,
					},
				})
			case dynmdlV2.DrawTagTypeUser:
				i.Tags = append(i.Tags, &dynamicapi.MdlDynDrawTag{
					Type: dynamicapi.MdlDynDrawTagType_mdl_draw_tag_user,
					Item: &dynamicapi.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Mid:         picTag.Mid,
						Url:         topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(picTag.Mid, 10), nil),
					},
				})
			case dynmdlV2.DrawTagTypeTopic:
				var topicURL string
				topicInfos, _ := dynCtx.Dyn.GetTopicInfo()
				for _, topic := range topicInfos {
					if topic != nil && topic.TopicName == picTag.Text {
						topicURL = topic.TopicLink
						break
					}
				}
				i.Tags = append(i.Tags, &dynamicapi.MdlDynDrawTag{
					Type: dynamicapi.MdlDynDrawTagType_mdl_draw_tag_topic,
					Item: &dynamicapi.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Tid:         picTag.Tid,
						Url:         topicURL,
					},
				})
			case dynmdlV2.DrawTagTypeLBS:
				var (
					lbs *dynmdlV2.DrawTagLBS
					uri string
				)
				if err := json.Unmarshal([]byte(picTag.Poi), &lbs); err != nil {
					log.Warn("module error draw pic_tag_slb miss mid(%v) dynid(%v) type(%v) rid(%v) error %v", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid, err)
					continue
				}
				if lbs != nil && lbs.PoiInfo != nil && lbs.PoiInfo.Location != nil {
					uri = fmt.Sprintf(topiccardmodel.LBSURI, lbs.PoiInfo.Poi, lbs.PoiInfo.Type, lbs.PoiInfo.Location.Lat, lbs.PoiInfo.Location.Lng, url.QueryEscape(lbs.PoiInfo.Title), url.QueryEscape(lbs.PoiInfo.Address))
				}
				i.Tags = append(i.Tags, &dynamicapi.MdlDynDrawTag{
					Type: dynamicapi.MdlDynDrawTagType_mdl_draw_tag_lbs,
					Item: &dynamicapi.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Poi:         picTag.Poi,
						Url:         uri,
					},
				})
			default:
				log.Warn("dynCardDraw module error draw pic_tag miss mid(%v) dynid(%v) type(%v) rid(%v) unknow pic tag type %v", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid, picTag.Type)
			}
		}
		card.Items = append(card.Items, i)
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type:       dynamicapi.ModuleDynamicType_mdl_dyn_draw,
				ModuleItem: &dynamicapi.ModuleDynamic_DynDraw{DynDraw: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) dynCardArticle(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	article, _ := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
	card := &dynamicapi.MdlDynArticle{
		Id:         article.ActID,
		Title:      schema.getTitle(article.Title, dynCtx),
		Desc:       article.Summary,
		Covers:     article.ImageURLs,
		Label:      topiccardmodel.StatString(article.Stats.View, "阅读", ""),
		TemplateID: article.TemplateID,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type:       dynamicapi.ModuleDynamicType_mdl_dyn_article,
				ModuleItem: &dynamicapi.ModuleDynamic_DynArticle{DynArticle: card},
			},
		},
	})
	return nil
}

func (schema *CardSchema) dynCardPGC(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	if !ok {
		dynCtx.Interim.IsPassCard = true
		log.Error("dynCardPGC dynCtx.GetResPGC is nil general=%+v, DynamicID=%d", general, dynCtx.Dyn.DynamicID)
		return nil
	}
	card := &dynamicapi.MdlDynPGC{
		Title:           schema.getTitle(pgc.CardShowTitle, dynCtx),
		Cover:           pgc.Cover,
		Uri:             pgc.Url,
		CoverLeftText_1: topiccardmodel.VideoDuration(pgc.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", schema.numTransfer(int(pgc.Stat.Play))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", schema.numTransfer(int(pgc.Stat.Danmaku))),
		Cid:             pgc.Cid,
		Epid:            int64(pgc.EpisodeId),
		Aid:             pgc.Aid,
		MediaType:       dynamicapi.MediaType_MediaTypePGC,
		IsPreview:       dynmdlV2.Int32ToBool(pgc.IsPreview),
		Dimension: &dynamicapi.Dimension{
			Height: int64(pgc.Dimension.Height),
			Width:  int64(pgc.Dimension.Width),
			Rotate: int64(pgc.Dimension.Rotate),
		},
		Duration: pgc.Duration,
		SubType:  dynCtx.Dyn.GetPGCSubType(),
		CanPlay:  pgc.PlayerInfo != nil,
	}
	if pgc.Season != nil {
		card.Season = &dynamicapi.PGCSeason{
			IsFinish: int32(pgc.Season.IsFinish),
			Title:    pgc.Season.Title,
			Type:     int32(pgc.Season.Type),
		}
		card.SeasonId = int64(pgc.Season.SeasonId)
		if pgc.SectionType == 0 {
			// 是否是PGC正片，上报字段
			card.IsFeature = true
			if dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.PgcBadge != nil && dynCtx.Dyn.PassThrough.PgcBadge.Show {
				// 追番人数角标
				if pgc.Stat.FollowDesc != "" {
					card.BadgeCategory = append(card.BadgeCategory, dynmdlV2.BadgeStyleFrom(dynmdlV2.BgColorTransparentGray, pgc.Stat.FollowDesc))
				}
				// PGC角标
				if pgc.Season.TypeName != "" {
					card.BadgeCategory = append(card.BadgeCategory, dynmdlV2.BadgeStyleFrom(dynmdlV2.BgColorPink, pgc.Season.TypeName))
				}
			}
		}
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type: dynamicapi.ModuleDynamicType_mdl_dyn_pgc,
				ModuleItem: &dynamicapi.ModuleDynamic_DynPgc{
					DynPgc: card,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) dynCardCommon(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
	if !ok {
		dynCtx.Interim.IsPassCard = true
		log.Error("dynCardCommon dynCtx.GetResCommon is nil general=%+v, Rid=%d", general, dynCtx.Dyn.Rid)
		return nil
	}
	card := &dynamicapi.MdlDynCommon{
		Oid:      common.Sketch.BizID,
		Uri:      common.Sketch.TagURL,
		Title:    schema.getTitle(common.Sketch.Title, dynCtx),
		Desc:     common.Sketch.DescText,
		Cover:    common.Sketch.CoverURL,
		Label:    common.Sketch.Text,
		BizType:  int32(common.Sketch.BizType),
		SketchID: common.Sketch.SketchID,
		Style:    dynamicapi.MdlDynCommonType_mdl_dyn_common_vertica,
	}
	if dynCtx.Dyn.IsCommonSquare() {
		card.Style = dynamicapi.MdlDynCommonType_mdl_dyn_common_square
	}
	var tags []*dynmdlV2.DynamicCommonCardTags
	if err := json.Unmarshal(common.Sketch.Tags, &tags); err != nil {
		log.Warn("module err tags(%+v) dynid(%v) dynCardCommon error %v", common.Sketch.Tags, dynCtx.Dyn.DynamicID, err)
	}
	for _, tag := range tags {
		if tag == nil || tag.Name == "" {
			continue
		}
		if !strings.Contains(tag.Color, "#") {
			tag.Color = fmt.Sprintf("#%s", tag.Color)
		}
		card.Badge = append(card.Badge, &dynamicapi.VideoBadge{
			Text:             tag.Name,
			TextColor:        "#FFFFFF",
			TextColorNight:   "#FFFFFF",
			BgColor:          tag.Color,
			BgColorNight:     tag.Color,
			BorderColor:      tag.Color,
			BorderColorNight: tag.Color,
			BgStyle:          _moduleDynamicCommonBadgeStyle,
		})
	}
	// 按钮
	if button := common.Sketch.Button; button != nil {
		if button.JumpStyle != nil {
			card.Button = &dynamicapi.AdditionalButton{
				Type:    dynamicapi.AddButtonType_bt_jump,
				JumpUrl: button.JumpURL,
				JumpStyle: &dynamicapi.AdditionalButtonStyle{
					Icon:    button.JumpStyle.Icon,
					Text:    button.JumpStyle.Text,
					BgStyle: dynamicapi.AddButtonBgStyle_fill,
					Disable: dynamicapi.DisableState_highlight,
				},
			}
			if button.Status == dynmdlV2.AttachButtonStatusCheck {
				card.Button = &dynamicapi.AdditionalButton{
					Type:   dynamicapi.AddButtonType_bt_button,
					Status: dynamicapi.AdditionalButtonStatus_check,
					Check: &dynamicapi.AdditionalButtonStyle{
						Icon:    button.JumpStyle.Icon,
						Text:    button.JumpStyle.Text,
						BgStyle: dynamicapi.AddButtonBgStyle_gray,
						Disable: dynamicapi.DisableState_gary,
					},
				}
			}
		}
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_dynamic,
		ModuleItem: &dynamicapi.Module_ModuleDynamic{
			ModuleDynamic: &dynamicapi.ModuleDynamic{
				Type:       dynamicapi.ModuleDynamicType_mdl_dyn_common,
				ModuleItem: &dynamicapi.ModuleDynamic_DynCommon{DynCommon: card},
			},
		},
	})
	return nil
}
