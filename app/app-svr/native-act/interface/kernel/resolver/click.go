package resolver

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Click struct{}

func (r Click) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Click{
		BaseCfgManager: config.NewBaseCfg(natModule),
		PressSave:      natModule.IsAttrShareImage() == natpagegrpc.AttrModuleYes,
		Items:          r.clickItems(module.Click),
	}
	if natModule.AvSort == model.TabLockNeed {
		cfg.Unlock = &config.Unlock{
			UnlockCondition: int64(natModule.DySort),
			UnlockTime:      natModule.Stime,
		}
	}
	if natModule.Meta != "" {
		cfg.BgImage = &config.SizeImage{
			Image:  natModule.Meta,
			Height: natModule.Length,
			Width:  natModule.Width,
		}
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

// nolint:gocognit
func (r Click) clickItems(in *natpagegrpc.Click) []*config.ClickItem {
	if in == nil {
		return nil
	}
	items := make([]*config.ClickItem, 0, len(in.Areas))
	for _, area := range in.Areas {
		if area == nil {
			continue
		}
		item := &config.ClickItem{}
		switch area.Type {
		case model.ClickTypeRedirect, model.ClickTypeParticipation:
			item.Url = area.Link
		case model.ClickTypeUrlPerDev:
			item.IosUrl = area.UnfinishedImage
			item.AndroidUrl = area.FinishedImage
		case model.ClickTypeFollowUser, model.ClickTypeFollowEpisode, model.ClickTypeFollowComic, model.ClickTypeReserve,
			model.ClickTypeMallWantGo:
			if area.ForeignID <= 0 {
				continue
			}
			item.Id = area.ForeignID
			item.DoneImage = area.FinishedImage
			item.UndoneImage = area.UnfinishedImage
			item.MsgBoxTip = area.Tip
			if area.Type == model.ClickTypeFollowUser {
				item.MsgBoxTip = "关注"
			}
		case model.ClickTypeReceiveAward, model.ClickTypeUpReserve:
			if area.ForeignID <= 0 {
				continue
			}
			item.Id = area.ForeignID
			item.DoneImage = area.FinishedImage
			item.UndoneImage = area.OptionalImage
			item.DisableImage = area.UnfinishedImage
		case model.ClickTypeBtnRedirect:
			item.Image = area.OptionalImage
			item.Url = area.Link
		case model.ClickTypeDisplayImage:
			item.Image = area.OptionalImage
		case model.ClickTypeRTProgress:
			ext, err := model.UnmarshalClickExtProgress(area.Tip)
			if err != nil || area.ForeignID <= 0 || ext.GroupId <= 0 {
				continue
			}
			item.Id = area.ForeignID
			item.GroupId = ext.GroupId
			item.NodeId = ext.NodeId
			item.DisplayType = ext.DisplayType
			item.FontType = ext.FontType
			item.FontSize = ext.FontSize
			item.FontColor = ext.FontColor
		case model.ClickTypeNRTProgress:
			ext, err := model.UnmarshalClickExtProgress(area.Tip)
			if err != nil {
				continue
			}
			item.ProgSource = ext.PSort
			item.DisplayType = ext.DisplayType
			item.FontType = ext.FontType
			item.FontSize = ext.FontSize
			item.FontColor = ext.FontColor
			switch ext.PSort {
			case model.ClickNRTProgUserStats:
				if area.ForeignID <= 0 || ext.GroupId <= 0 {
					continue
				}
				item.Id = area.ForeignID
				item.GroupId = ext.GroupId
				item.NodeId = ext.NodeId
			case model.ClickNRTProgActApplyNum:
				if area.ForeignID <= 0 {
					continue
				}
				item.Id = area.ForeignID
				if area.FinishedImage != "" {
					item.StatsDimension, _ = strconv.ParseInt(area.FinishedImage, 10, 64)
				}
			case model.ClickNRTProgTaskStats:
				if ext.Counter == "" || ext.Activity == "" {
					continue
				}
				item.Activity = ext.Activity
				item.Counter = ext.Counter
				item.StatsPeriod = ext.StatPc
			case model.ClickNRTProgLotteryNum:
				if ext.LotteryID == "" {
					continue
				}
				item.LotteryId = ext.LotteryID
			case model.ClickNRTProgScore:
				if area.ForeignID <= 0 {
					continue
				}
				item.Id = area.ForeignID
			default:
				continue
			}
		case model.ClickTypeLayerImage:
			if image := r.unmarshalAreaImage(area.UnfinishedImage); image != nil {
				item.Images = append(item.Images, image)
			}
			if image := r.unmarshalAreaImage(area.FinishedImage); image != nil {
				item.Images = append(item.Images, image)
			}
			if image := r.unmarshalAreaImage(area.OptionalImage); image != nil {
				item.Images = append(item.Images, image)
			}
			if ext, err := model.UnmarshalClickExtLayer(area.Ext); err == nil {
				if ext != nil && len(ext.Images) > 0 {
					item.Images = append(item.Images, ext.Images...)
				}
				item.Style = ext.Style
				item.Image = ext.ButtonImage
				item.ShareImage = ext.ShareImage
				item.ImageTitle = ext.LayerImage
			}
			if ext, err := model.UnmarshalClickExtLayerColor(area.Tip); err == nil {
				item.Title = ext.Title
				item.TitleColor = ext.TitleColor
				item.TopColor = ext.TopColor
			}
		case model.ClickTypeLayerRedirect:
			item.Url = area.Link
			if ext, err := model.UnmarshalClickExtLayer(area.Ext); err == nil {
				item.Style = ext.Style
				item.Image = ext.ButtonImage
				item.ImageTitle = ext.LayerImage
			}
			if ext, err := model.UnmarshalClickExtLayerColor(area.Tip); err == nil {
				item.Title = ext.Title
				item.TitleColor = ext.TitleColor
				item.TopColor = ext.TopColor
			}
		case model.ClickTypeActivity:
			if area.ForeignID <= 0 {
				continue
			}
			item.Id = area.ForeignID
			item.DoneImage = area.FinishedImage
			item.UndoneImage = area.UnfinishedImage
			item.MsgBoxTip = area.Tip
		default:
			continue
		}
		item.AreaId = area.ID
		item.Type = area.Type
		item.Area = &config.Area{Height: area.Length, Width: area.Width, X: area.Leftx, Y: area.Lefty}
		if ext, err := model.UnmarshalClickExtCommon(area.Ext); err == nil {
			item.Area.Ukey = ext.Ukey
			if area.Type == model.ClickTypeReserve || area.Type == model.ClickTypeMallWantGo || area.Type == model.ClickTypeActivity {
				item.SyncHover = ext.SynHover
			}
			if ext.DisplayMode == model.TabLockNeed {
				item.Unlock = &config.Unlock{
					UnlockCondition: ext.UnlockCondition,
					UnlockTime:      ext.Stime,
					Sid:             ext.Sid,
					GroupId:         ext.GroupId,
					NodeId:          ext.NodeId,
				}
			}
		}
		items = append(items, item)
	}
	return items
}

func (r Click) setBaseCfg(cfg *config.Click, ss *kernel.Session) {
	var (
		relFids        []int64
		seasonIds      []int32
		comicIds       []int64
		rsvIds         []int64
		awardIds       []int64
		ticketIds      []int64
		actRelationIds []int64
		upRsvIds       []int64
		lotteryIds     []string
		scoreIds       []int64
	)
	for _, item := range cfg.Items {
		switch item.Type {
		case model.ClickTypeFollowUser:
			relFids = append(relFids, item.Id)
		case model.ClickTypeFollowEpisode:
			seasonIds = append(seasonIds, int32(item.Id))
		case model.ClickTypeFollowComic:
			comicIds = append(comicIds, item.Id)
		case model.ClickTypeReserve:
			rsvIds = append(rsvIds, item.Id)
		case model.ClickTypeReceiveAward:
			awardIds = append(awardIds, item.Id)
		case model.ClickTypeMallWantGo:
			ticketIds = append(ticketIds, item.Id)
		case model.ClickTypeActivity:
			actRelationIds = append(actRelationIds, item.Id)
		case model.ClickTypeUpReserve:
			upRsvIds = append(upRsvIds, item.Id)
		case model.ClickTypeRTProgress:
			_, _ = cfg.AddMaterialParam(model.MaterialActProgressGroup, item.Id, []int64{item.GroupId})
		case model.ClickTypeNRTProgress:
			switch item.ProgSource {
			case model.ClickNRTProgUserStats:
				_, _ = cfg.AddMaterialParam(model.MaterialActProgressGroup, item.Id, []int64{item.GroupId})
			case model.ClickNRTProgActApplyNum:
				rsvIds = append(rsvIds, item.Id)
			case model.ClickNRTProgTaskStats:
				if item.StatsPeriod == "daily" {
					item.PlatCounterReqID, _ = cfg.AddMaterialParam(model.MaterialPlatCounterRes, &actplatv2grpc.GetCounterResReq{
						Counter:  item.Counter,
						Activity: item.Activity,
						Mid:      ss.Mid(),
						Time:     time.Now().Unix(),
					})
				} else {
					item.PlatTotalReqID, _ = cfg.AddMaterialParam(model.MaterialPlatTotalRes, &actplatv2grpc.GetTotalResReq{
						Counter:  item.Counter,
						Activity: item.Activity,
						Mid:      ss.Mid(),
					})
				}
			case model.ClickNRTProgLotteryNum:
				lotteryIds = append(lotteryIds, item.LotteryId)
			case model.ClickNRTProgScore:
				scoreIds = append(scoreIds, item.Id)
			}
		}
		if uk := item.Unlock; uk != nil && uk.UnlockCondition == model.TabLockTypeSource {
			_, _ = cfg.AddMaterialParam(model.MaterialActProgressGroup, uk.Sid, []int64{uk.GroupId})
		}
	}
	if ss.Mid() > 0 {
		_, _ = cfg.AddMaterialParam(model.MaterialPgcFollowStatus, seasonIds)
		_, _ = cfg.AddMaterialParam(model.MaterialComicInfo, comicIds)
		_, _ = cfg.AddMaterialParam(model.MaterialActReserveFollow, rsvIds)
		_, _ = cfg.AddMaterialParam(model.MaterialTicketFavState, ticketIds)
		_, _ = cfg.AddMaterialParam(model.MaterialActRelationInfo, actRelationIds)
		_, _ = cfg.AddMaterialParam(model.MaterialLotteryUnused, lotteryIds)
	}
	_, _ = cfg.AddMaterialParam(model.MaterialRelation, relFids)
	_, _ = cfg.AddMaterialParam(model.MaterialActAwardState, awardIds)
	_, _ = cfg.AddMaterialParam(model.MaterialScoreTarget, scoreIds)
	cfg.UpRsvIdsReq, _ = cfg.AddMaterialParam(model.MaterialUpRsvInfo, &kernel.UpRsvIDsReq{IDs: upRsvIds})
}

func (r Click) unmarshalAreaImage(img string) *config.SizeImage {
	if img == "" {
		return nil
	}
	image := &config.SizeImage{}
	if err := json.Unmarshal([]byte(img), image); err == nil && image.Image != "" {
		return image
	}
	return nil
}
