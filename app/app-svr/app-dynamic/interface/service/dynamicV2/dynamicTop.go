package dynamicV2

import (
	"context"
	"fmt"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

func (s *Service) top(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_top,
		ModuleItem: &api.Module_ModuleTop{
			ModuleTop: &api.ModuleTop{
				TpList: s.topThreePoint(c, dynCtx, general),
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// nolint:gocognit
func (s *Service) topThreePoint(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) []*api.ThreePointItem {
	const (
		_liveStart = 1
		_liveAv    = 2
	)
	var (
		ext []*api.ThreePointItem
	)
	tmp := &mdlv2.DynamicContext{}
	*tmp = *dynCtx
	switch dynCtx.Dyn.Type {
	case mdlv2.DynTypeAD, mdlv2.DynTypeSubscription:
		// 三点模块不下发
		return nil
	case mdlv2.DynTypeForward:
		// 转发卡
		if dynCtx.Dyn.Origin != nil {
			tmp.Dyn = dynCtx.Dyn.Origin
		}
	case mdlv2.DynTypeDraw, mdlv2.DynTypeWord:
		// 图文按钮
		ext = append(ext, &api.ThreePointItem{
			Type: api.ThreePointType_share,
			Item: &api.ThreePointItem_Share{
				Share: &api.ThreePointShare{
					Icon:        "http://i0.hdslb.com/bfs/feed-admin/361a2b0466d7cd76660f9fc25ca9a62dbf20a969.png",
					Title:       "生成长图",
					ChannelName: "LONG CHART",
				},
			},
		})
	}
	// 预约卡
	if !general.IsPadHD() && !general.IsAndroidHD() && !general.IsPad() && !general.IsOverseas() {
		for _, v := range dynCtx.Dyn.AttachCardInfos {
			// nolint:exhaustive
			switch v.CardType {
			case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
				item := &api.ThreePointShare{
					Title:       s.c.Resource.ReserveShare.Name,
					Icon:        s.c.Resource.ReserveShare.Image,
					ChannelName: s.c.Resource.ReserveShare.Channel,
				}
				up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
				if !ok {
					continue
				}
				if up.UpActVisible == activitygrpc.UpActVisible_OnlyUpVisible {
					continue
				}
				switch upActState(up.State) {
				case _upAudit, _upDelete, _upCancel, _upExpired, _upOnline, _upNotStart:
					continue
				}
				reserve := &api.ShareReserve{
					Title:      up.Title,
					QrCodeIcon: s.c.Resource.ReserveShare.QrCodeIcon,
					QrCodeText: s.c.Resource.ReserveShare.QrCodeText,
					QrCodeUrl:  fmt.Sprintf(s.c.Resource.ReserveShare.QrCodeUrl, dynCtx.Dyn.DynamicID),
				}
				var typeDesc string
				// nolint:exhaustive
				switch up.Type {
				case activitygrpc.UpActReserveRelationType_Archive:
					typeDesc = s.c.Resource.ReserveShare.DescAv
				case activitygrpc.UpActReserveRelationType_Live:
					if up.LivePlanStartTime.Time().Unix() > 0 {
						reserve.Desc = model.UpPubShareDataString(up.LivePlanStartTime.Time())
					}
					typeDesc = s.c.Resource.ReserveShare.DescLive
					if upActState(up.State) == _upOnline {
						var isok bool
						// 直播状态
						if info, ok := dynCtx.ResLiveSessionInfo[up.Oid]; ok {
							live, ok := info.SessionInfoPerLive[up.Oid]
							if ok {
								isok = true
								switch live.Status {
								case _liveStart, _liveAv:
									if live.Status == _liveAv {
										aid, _ := bvid.BvToAv(live.Bvid)
										if _, ok := dynCtx.ResArcs[aid]; !ok {
											isok = false
										}
									}
								default:
									isok = false
								}
							}
						}
						if !isok {
							// 已结束（不可回放）
							continue
						}
					}
				case activitygrpc.UpActReserveRelationType_Premiere:
					continue
				}
				if reserve.Desc != "" {
					reserve.Desc = reserve.Desc + " " + typeDesc
				} else {
					reserve.Desc = typeDesc
				}
				if userInfo, ok := dynCtx.GetUser(up.Upmid); ok {
					reserve.UserInfo = &api.AdditionUserInfo{
						Name: userInfo.Name,
						Face: userInfo.Face,
					}
				}
				item.Reserve = reserve
				ext = append(ext, &api.ThreePointItem{
					Type: api.ThreePointType_share,
					Item: &api.ThreePointItem_Share{
						Share: item,
					},
				})
			}
		}
	}
	// 稍后在看
	if iswait, waitID := s.threePointWait(tmp, general); iswait {
		ext = append(ext, s.tpWait(s.c.Resource.Text.ThreePointWaitAddition, s.c.Resource.Text.ThreePointWaitNotAddition, s.c.Resource.Icon.ThreePointWaitView, waitID))
	}
	// 取消追漫
	if dynCtx.Dyn.IsBatch() {
		ext = append(ext, s.tpBatchCancel(c, dynCtx))
	}
	// 举报
	if isReport, titles, reportMid := s.threePointReport(dynCtx, general); isReport {
		ext = append(ext, s.tpReport(c, dynCtx.Dyn.DynamicID, reportMid, titles))
	}
	// 删除
	if isDel := s.threePointDel(dynCtx, general); isDel {
		ext = append(ext, s.tpDeleteView(dynCtx))
	}
	// 编辑
	if s.isDynEditCapable(c, general) {
		if tpEdit := s.threePointEdit(dynCtx, general); tpEdit != nil {
			ext = append(ext, tpEdit)
		}
	}
	return ext
}

func (s *Service) tpDeleteView(dynCtx *mdlv2.DynamicContext) *api.ThreePointItem {
	item := &api.ThreePointItem{
		Type: api.ThreePointType_delete,
		Item: &api.ThreePointItem_Default{
			Default: &api.ThreePointDefault{
				Icon:  s.c.Resource.Icon.ThreePointDeletedView,
				Title: s.c.Resource.Text.ThreePointDeleted,
				Toast: s.tpDeleteReserveToast(dynCtx),
			},
		},
	}
	return item
}
