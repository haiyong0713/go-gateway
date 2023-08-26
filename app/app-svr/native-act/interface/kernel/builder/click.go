package builder

import (
	"context"
	"fmt"
	"math"
	"time"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	commscoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
)

type Click struct{}

func (bu Click) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	ckCfg, ok := cfg.(*config.Click)
	if !ok {
		logCfgAssertionError(config.Click{})
		return nil
	}
	items := bu.buildModuleItems(ckCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeClick.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleSetting: &api.Setting{PressSave: ckCfg.PressSave},
		ModuleItems:   items,
		ModuleUkey:    cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Click) After(data *AfterContextData, current *api.Module) bool {
	return true
}

// nolint:gocognit
func (bu Click) buildModuleItems(cfg *config.Click, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	if bu.isLocked(cfg.Unlock, material) {
		return nil
	}
	items := make([]*api.ClickItem, 0, len(cfg.Items))
	for _, v := range cfg.Items {
		if bu.isLocked(v.Unlock, material) {
			continue
		}
		item := &api.ClickItem{}
		switch v.Type {
		case model.ClickTypeRedirect, model.ClickTypeParticipation:
			item.Action = api.ClickItem_ActRedirect
			item.ActionDetail = &api.ClickItem_RedirectAct{
				RedirectAct: &api.ClickActRedirect{Url: v.Url},
			}
		case model.ClickTypeUrlPerDev:
			var url = v.IosUrl
			if ss.IsAndroid() {
				url = v.AndroidUrl
			}
			item.Action = api.ClickItem_ActRedirect
			item.ActionDetail = &api.ClickItem_RedirectAct{
				RedirectAct: &api.ClickActRedirect{Url: url},
			}
		case model.ClickTypeBtnRedirect:
			item.Action = api.ClickItem_ActRedirect
			item.ActionDetail = &api.ClickItem_RedirectAct{
				RedirectAct: &api.ClickActRedirect{Url: v.Url, Image: v.Image},
			}
		case model.ClickTypeDisplayImage:
			item.Action = api.ClickItem_ActImage
			item.ActionDetail = &api.ClickItem_ImageAct{
				ImageAct: &api.ClickActImage{Image: v.Image},
			}
		case model.ClickTypeFollowUser:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeFollowUser, v, func() bool {
				if rel, ok := material.Relations[v.Id]; ok && (rel.Attribute == model.RelationFollow || rel.Attribute == model.RelationFriend) {
					return true
				}
				return false
			})
		case model.ClickTypeFollowEpisode:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeFollowEpisode, v, func() bool {
				if status, ok := material.PgcFollowStatuses[int32(v.Id)]; ok && status.Follow {
					return true
				}
				return false
			})
		case model.ClickTypeFollowComic:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeFollowComic, v, func() bool {
				if info, ok := material.ComicInfos[v.Id]; ok && info.FavStatus == 1 {
					return true
				}
				return false
			})
		case model.ClickTypeReserve:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeReserve, v, func() bool {
				if rly, ok := material.ActRsvFollows[v.Id]; ok && rly.IsFollow {
					return true
				}
				return false
			})
		case model.ClickTypeReceiveAward:
			bu.setReceiveAwardRequestAct(item, v, material.AwardStates[v.Id])
		case model.ClickTypeMallWantGo:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeMallWantGo, v, func() bool {
				if state, ok := material.TicketFavStates[v.Id]; ok && state {
					return true
				}
				return false
			})
		case model.ClickTypeActivity:
			bu.setRequestActOfClickItem(item, api.ClickRequestType_CRTypeActivity, v, func() bool {
				if info, ok := material.ActRelationInfos[v.Id]; ok && info.ReserveItems != nil && info.ReserveItems.State == 1 {
					return true
				}
				return false
			})
		case model.ClickTypeUpReserve:
			info, ok := material.UpRsvInfos[cfg.UpRsvIdsReq][v.Id]
			if !ok {
				continue
			}
			bu.setUpReserveRequestAct(item, v, info)
		case model.ClickTypeRTProgress:
			group, ok := material.ActProgressGroups[v.Id][v.GroupId]
			if !ok {
				continue
			}
			bu.setProgress(item, v, true, func() (curr string, target string) {
				if v.NodeId > 0 {
					for _, node := range group.Nodes {
						if node != nil && node.Nid == v.NodeId {
							return StatString(group.Total), StatString(node.Val)
						}
					}
				}
				return StatString(group.Total), "0"
			})
		case model.ClickTypeNRTProgress:
			switch v.ProgSource {
			case model.ClickNRTProgUserStats:
				group, ok := material.ActProgressGroups[v.Id][v.GroupId]
				if !ok {
					continue
				}
				bu.setProgress(item, v, false, func() (curr string, target string) {
					if v.NodeId > 0 {
						for _, node := range group.Nodes {
							if node != nil && node.Nid == v.NodeId {
								return StatString(group.Total), StatString(node.Val)
							}
						}
					}
					return StatString(group.Total), "0"
				})
			case model.ClickNRTProgActApplyNum:
				bu.setProgress(item, v, false, func() (curr string, target string) {
					rly, ok := material.ActRsvFollows[v.Id]
					if !ok {
						return "0", "0"
					}
					if activitygrpc.GetReserveProgressDimension(v.StatsDimension) == activitygrpc.GetReserveProgressDimension_Rule {
						return StatString(rly.Total), "0"
					}
					if rly.IsFollow {
						return "1", "0"
					}
					return "0", "0"
				})
			case model.ClickNRTProgTaskStats:
				bu.setProgress(item, v, false, func() (curr string, target string) {
					if v.StatsPeriod == "daily" {
						if rly, ok := material.PlatCounterResRlys[v.PlatCounterReqID]; ok && len(rly.CounterList) > 0 {
							return StatString(rly.CounterList[0].Val), "0"
						}
					} else {
						if rly, ok := material.PlatTotalResRlys[v.PlatTotalReqID]; ok {
							return StatString(rly.Total), "0"
						}
					}
					return "0", "0"
				})
			case model.ClickNRTProgLotteryNum:
				bu.setProgress(item, v, false, func() (curr string, target string) {
					if rly, ok := material.LotUnusedRlys[v.LotteryId]; ok {
						return StatString(rly.Times), "0"
					}
					return "0", "0"
				})
			case model.ClickNRTProgScore:
				rly, ok := material.ScoreTargets[v.Id]
				if !ok {
					continue
				}
				bu.setProgress(item, v, false, func() (curr string, target string) {
					return bu.finalScore(rly), "0"
				})
			default:
				continue
			}
		case model.ClickTypeLayerImage:
			layer := &api.ClickActLayer{
				ButtonImage: v.Image,
				Mode:        api.ClickActLayer_LMImage,
			}
			layer.Images = make([]*api.SizeImage, 0, len(v.Images))
			for _, image := range v.Images {
				layer.Images = append(layer.Images, image.ToSizeImage())
			}
			if v.Style == "image" {
				layer.Style = api.ClickActLayer_LTImage
				layer.ImageTitle = v.ImageTitle
			} else {
				layer.Style = api.ClickActLayer_LTColor
				layer.Title = v.Title
				layer.Color = &api.Color{TitleColor: v.TitleColor, TopFontColor: v.TopColor}
			}
			if cfg.PressSave && v.ShareImage != nil && v.ShareImage.Image != "" {
				v.ShareImage.Size = int64(math.Floor(float64(v.ShareImage.Size)/1024 + 0.5))
				layer.ShareImage = v.ShareImage.ToSizeImage()
				layer.Share = &api.Share{ShareOrigin: model.ShareOriginLongPress}
			}
			item.Action = api.ClickItem_ActLayer
			item.ActionDetail = &api.ClickItem_LayerAct{LayerAct: layer}
		case model.ClickTypeLayerRedirect:
			layer := &api.ClickActLayer{
				ButtonImage: v.Image,
				Mode:        api.ClickActLayer_LMRedirect,
				Url:         v.Url,
			}
			if v.Style == "image" {
				layer.Style = api.ClickActLayer_LTImage
				layer.ImageTitle = v.ImageTitle
			} else {
				layer.Style = api.ClickActLayer_LTColor
				layer.Title = v.Title
				layer.Color = &api.Color{FontColor: v.FontColor, TopFontColor: v.TopColor}
			}
			item.Action = api.ClickItem_ActLayer
			item.ActionDetail = &api.ClickItem_LayerAct{LayerAct: layer}
		}
		item.AreaId = v.AreaId
		item.Area = v.Area.ToGrpcArea()
		items = append(items, item)
	}
	moduleItem := &api.ModuleItem{
		CardType: model.CardTypeClick.String(),
		CardDetail: &api.ModuleItem_ClickCard{
			ClickCard: &api.ClickCard{
				BgImage: cfg.BgImage.ToSizeImage(),
				Items:   items,
			},
		},
	}
	return []*api.ModuleItem{moduleItem}
}

func (bu Click) setRequestActOfClickItem(item *api.ClickItem, reqType api.ClickRequestType, cfg *config.ClickItem, followedFn func() bool) {
	reqAct := &api.ClickActRequest{
		Id:              cfg.Id,
		ReqType:         reqType,
		State:           api.ClickRequestState_CRSUndone,
		SyncHoverButton: cfg.SyncHover,
	}
	if followed := followedFn(); followed {
		reqAct.State = api.ClickRequestState_CRSDone
	}
	reqAct.Details = append(reqAct.Details,
		&api.ClickRequestDetail{
			State:  api.ClickRequestState_CRSDone,
			Params: bu.requestParams(cfg.Id, reqType, true),
			Image:  cfg.DoneImage,
			MessageBox: &api.MessageBox{
				AlertMsg:          fmt.Sprintf("确定取消%s吗？", cfg.MsgBoxTip),
				ConfirmButtonText: fmt.Sprintf("取消%s", cfg.MsgBoxTip),
				CancelButtonText:  "再想想",
				ConfirmMsg:        fmt.Sprintf("取消%s成功", cfg.MsgBoxTip),
				Type:              api.MessageBoxType_Dialog,
			},
		}, &api.ClickRequestDetail{
			State:  api.ClickRequestState_CRSUndone,
			Params: bu.requestParams(cfg.Id, reqType, false),
			Image:  cfg.UndoneImage,
			MessageBox: &api.MessageBox{
				AlertMsg: fmt.Sprintf("%s成功", cfg.MsgBoxTip),
				Type:     api.MessageBoxType_Toast,
			},
		},
	)
	item.Action = api.ClickItem_ActRequest
	item.ActionDetail = &api.ClickItem_RequestAct{RequestAct: reqAct}
}

func (bu Click) setReceiveAwardRequestAct(item *api.ClickItem, cfg *config.ClickItem, stateRly *activitygrpc.AwardSubjectStateReply) {
	reqAct := &api.ClickActRequest{
		Id:      cfg.Id,
		ReqType: api.ClickRequestType_CRTypeReceiveAward,
	}
	state := model.AwardStateNoQualify
	if stateRly != nil {
		state = int(stateRly.State)
	}
	switch state {
	case model.AwardStateNotReceive:
		reqAct.State = api.ClickRequestState_CRSUndone
	case model.AwardStateReceived:
		reqAct.State = api.ClickRequestState_CRSDone
	default:
		reqAct.State = api.ClickRequestState_CRSDisable
	}
	reqAct.Details = append(reqAct.Details,
		&api.ClickRequestDetail{
			State:      api.ClickRequestState_CRSDone,
			Params:     bu.requestParams(cfg.Id, api.ClickRequestType_CRTypeReceiveAward, true),
			Image:      cfg.DoneImage,
			MessageBox: &api.MessageBox{AlertMsg: "你已经领取了该奖励", Type: api.MessageBoxType_Toast},
		}, &api.ClickRequestDetail{
			State:      api.ClickRequestState_CRSUndone,
			Params:     bu.requestParams(cfg.Id, api.ClickRequestType_CRTypeReceiveAward, false),
			Image:      cfg.UndoneImage,
			MessageBox: &api.MessageBox{AlertMsg: "奖励领取成功", Type: api.MessageBoxType_Toast},
		}, &api.ClickRequestDetail{
			State:      api.ClickRequestState_CRSDisable,
			Image:      cfg.DisableImage,
			MessageBox: &api.MessageBox{AlertMsg: "暂无领取资格", Type: api.MessageBoxType_Toast},
		},
	)
	item.Action = api.ClickItem_ActRequest
	item.ActionDetail = &api.ClickItem_RequestAct{RequestAct: reqAct}
}

func (bu Click) setUpReserveRequestAct(item *api.ClickItem, cfg *config.ClickItem, rsvInfo *activitygrpc.UpActReserveRelationInfo) {
	reqAct := &api.ClickActRequest{
		Id:      cfg.Id,
		ReqType: api.ClickRequestType_CRTypeUpReserve,
	}
	reqAct.State = func() api.ClickRequestState {
		if rsvInfo.UpActVisible == activitygrpc.UpActVisible_DefaultVisible &&
			(rsvInfo.State == activitygrpc.UpActReserveRelationState_UpReserveRelated || rsvInfo.State == activitygrpc.UpActReserveRelationState_UpReserveRelatedOnline) {
			if rsvInfo.IsFollow == 1 {
				return api.ClickRequestState_CRSDone
			}
			return api.ClickRequestState_CRSUndone
		}
		return api.ClickRequestState_CRSDisable
	}()
	reqAct.Details = append(reqAct.Details,
		&api.ClickRequestDetail{
			State:  api.ClickRequestState_CRSDone,
			Params: bu.requestParams(cfg.Id, api.ClickRequestType_CRTypeUpReserve, true),
			Image:  cfg.DoneImage,
		}, &api.ClickRequestDetail{
			State:  api.ClickRequestState_CRSUndone,
			Params: bu.requestParams(cfg.Id, api.ClickRequestType_CRTypeUpReserve, false),
			Image:  cfg.UndoneImage,
		}, &api.ClickRequestDetail{
			State:      api.ClickRequestState_CRSDisable,
			Image:      cfg.DisableImage,
			MessageBox: &api.MessageBox{AlertMsg: "不可预约", Type: api.MessageBoxType_Toast},
		},
	)
	item.Action = api.ClickItem_ActRequest
	item.ActionDetail = &api.ClickItem_RequestAct{RequestAct: reqAct}
}

func (bu Click) setProgress(item *api.ClickItem, cfg *config.ClickItem, isRT bool, numFn func() (curr string, target string)) {
	fontType := api.FontType_FontTypeNormal
	if cfg.FontType == "bold" {
		fontType = api.FontType_FontTypeBold
	}
	displayMode := api.ClickActProgress_DisplayCurr
	if cfg.DisplayType == "num_and_target" {
		displayMode = api.ClickActProgress_DisplayCurrTarget
	}
	currNum, targetNum := numFn()
	progress := &api.ClickActProgress{
		Color:       &api.Color{FontColor: cfg.FontColor},
		FontType:    fontType,
		FontSize:    cfg.FontSize,
		DisplayMode: displayMode,
		CurrentNum:  currNum,
		TargetNum:   targetNum,
	}
	if isRT {
		item.Action = api.ClickItem_ActRTProgress
		item.ActionDetail = &api.ClickItem_RtProgressAct{RtProgressAct: progress}
	} else {
		item.Action = api.ClickItem_ActNRTProgress
		item.ActionDetail = &api.ClickItem_NrtProgressAct{NrtProgressAct: progress}
	}
}

func (bu Click) isLocked(cfg *config.Unlock, material *kernel.Material) bool {
	if cfg == nil {
		return false
	}
	switch cfg.UnlockCondition {
	case model.TabLockTypePass:
		return false
	case model.TabLockTypeTime:
		return cfg.UnlockTime > time.Now().Unix()
	case model.TabLockTypeSource:
		group, ok := material.ActProgressGroups[cfg.Sid][cfg.GroupId]
		if !ok || len(group.Nodes) == 0 {
			return true
		}
		for _, node := range group.Nodes {
			if cfg.NodeId == node.Nid {
				return node.Val > group.Total
			}
		}
	}
	return true
}

func (bu Click) requestParams(id int64, reqType api.ClickRequestType, followed bool) string {
	action := api.ActionType_Do
	if followed {
		action = api.ActionType_Undo
	}
	return passthrough.Marshal(&api.ClickRequestParams{Action: action, Id: id, ReqType: reqType})
}

func (bu Click) finalScore(target *commscoregrpc.ScoreTarget) string {
	if target.GetShowFlag() == 1 {
		return "暂无评分"
	}
	if fs := target.GetFixScore(); fs != "" && fs != "0" && fs != "0.0" {
		return target.GetFixScore()
	}
	if target.GetTargetScore() == "0" || target.GetTargetScore() == "0.0" {
		return "暂无评分"
	}
	return target.GetTargetScore()
}
