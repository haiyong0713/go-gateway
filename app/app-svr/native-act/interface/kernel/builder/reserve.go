package builder

import (
	"context"
	"fmt"
	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	"strconv"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	xtime "go-common/library/time"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
)

type reserveRly struct {
	Sid         int64                                  //UP主预约id
	ChangeType  int64                                  //1:类型A 2:类型C 4:类型CD
	DisplayType int64                                  //1:类型A 2:类型C 3:类型D
	Item        *activitygrpc.UpActReserveRelationInfo //预约信息
	Arc         *arcgrpc.Arc                           //稿件信息
	Live        *roomgategrpc.SessionInfos             //直播信息
	Account     *accountgrpc.Card                      //账号信息
}

type Reserve struct{}

func (bu Reserve) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	rsvCfg, ok := cfg.(*config.Reserve)
	if !ok {
		logCfgAssertionError(config.Reserve{})
		return nil
	}
	items := bu.buildModuleItems(rsvCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeReserve.String(),
		ModuleId:    rsvCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: rsvCfg.BgColor, FontColor: rsvCfg.FontColor, CardBgColor: rsvCfg.CardBgColor},
		ModuleItems: items,
		ModuleUkey:  rsvCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Reserve) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Reserve) buildModuleItems(cfg *config.Reserve, material *kernel.Material, ss *kernel.Session) []*api.ModuleItem {
	rsvRlys := bu.reserveRlys(cfg, material)
	if len(rsvRlys) == 0 {
		return nil
	}
	items := make([]*api.ModuleItem, 0, len(cfg.UpRsvIDs))
	for _, sid := range cfg.UpRsvIDs {
		if _, ok := rsvRlys[sid]; !ok {
			continue
		}
		cd := bu.buildReserveCard(rsvRlys[sid], cfg.DisplayUpFaceName, ss)
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeReserve.String(),
			CardId:     strconv.FormatInt(sid, 10),
			CardDetail: &api.ModuleItem_ReserveCard{ReserveCard: cd},
		})
	}
	return unshiftTitleCard(items, cfg.ImageTitle, cfg.TextTitle, ss.ReqFrom)
}

// nolint: gocognit
func (bu Reserve) reserveRlys(cfg *config.Reserve, material *kernel.Material) map[int64]*reserveRly {
	if len(cfg.UpRsvIDs) == 0 {
		return nil
	}
	upRsvInfos, ok := material.UpRsvInfos[cfg.UpRsvIDsReqID]
	if !ok || len(upRsvInfos) == 0 {
		return nil
	}
	rlys := make(map[int64]*reserveRly, len(cfg.UpRsvIDs))
	nowtime := time.Now().Unix()
	for _, sid := range cfg.UpRsvIDs {
		if val, ok := upRsvInfos[sid]; !ok || val == nil {
			continue
		}
		//话题活动页展示都为客态逻辑
		if upRsvInfos[sid].UpActVisible != activitygrpc.UpActVisible_DefaultVisible {
			continue
		}
		var changeType int64
		switch upRsvInfos[sid].State {
		case activitygrpc.UpActReserveRelationState_UpReserveRelated, activitygrpc.UpActReserveRelationState_UpReserveRelatedOnline:
			changeType = model.ReserveDisplayA
			//nowtime
			if upRsvInfos[sid].Type == activitygrpc.UpActReserveRelationType_Course && int64(upRsvInfos[sid].Etime) < nowtime {
				//预约结束未核销
				changeType = model.ReserveDisplayE
			}
		case activitygrpc.UpActReserveRelationState_UpReserveRelatedWaitCallBack,
			activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackCancel,
			activitygrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone:
			switch upRsvInfos[sid].Type {
			case activitygrpc.UpActReserveRelationType_Archive:
				changeType = model.ReserveDisplayC
			case activitygrpc.UpActReserveRelationType_Live:
				changeType = model.ReserveDisplayLive
			case activitygrpc.UpActReserveRelationType_Course:
				changeType = actmdl.ReserveDisplayC
			default:
				continue
			}
		default: //不认识的类型，不展示对应卡片
			continue
		}
		rlys[sid] = &reserveRly{Sid: sid, ChangeType: changeType, Item: upRsvInfos[sid]}
	}
	rsvRlys := make(map[int64]*reserveRly, len(rlys))
	for sid, rly := range rlys {
		if rly == nil || rly.Item == nil {
			continue
		}
		switch rly.Item.Type {
		case activitygrpc.UpActReserveRelationType_Archive:
			aid, _ := strconv.ParseInt(rly.Item.Oid, 10, 64)
			if arc, ok := material.Arcs[aid]; ok && arc.IsNormal() {
				rly.Arc = arc
			}
		case activitygrpc.UpActReserveRelationType_Live:
			rly.Live = material.RoomSessionInfos[rly.Item.Upmid]
		default:
		}
		if cfg.DisplayUpFaceName {
			if acc, ok := material.AccountCards[rly.Item.Upmid]; ok {
				rly.Account = acc
			}
		}
		rly.DisplayType = rly.ChangeType
		//容错:类型CD 没有获取直播状态不下发
		if rly.ChangeType == model.ReserveDisplayLive {
			if rly.Live == nil || rly.Live.SessionInfoPerLive == nil {
				continue
			}
			si, ok := rly.Live.SessionInfoPerLive[rly.Item.Oid]
			if !ok || si == nil {
				continue
			}
			switch si.Status {
			case model.LiveStatusLiving:
				rly.DisplayType = model.ReserveDisplayC
			case model.LiveStatusEnd:
				rly.DisplayType = model.ReserveDisplayD
			default:
				rly.DisplayType = model.ReserveDisplayE
			}
		}
		rsvRlys[sid] = rly
	}
	return rsvRlys
}

func (bu Reserve) buildReserveCard(rsvRly *reserveRly, displayUpFaceName bool, ss *kernel.Session) *api.ReserveCard {
	cd := &api.ReserveCard{Sid: rsvRly.Sid}
	if acc := rsvRly.Account; acc != nil && displayUpFaceName {
		cd.Mid = acc.GetMid()
		cd.Name = acc.GetName()
		cd.Face = acc.GetFace()
	}
	if rsvRly.Item == nil {
		return cd
	}
	//跳转空间
	cd.Uri = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", rsvRly.Item.Upmid)
	var hideReserveNum bool
	switch rsvRly.DisplayType {
	case model.ReserveDisplayA:
		switch rsvRly.Item.Type {
		case activitygrpc.UpActReserveRelationType_Archive:
			cd.Title = rsvRly.Item.Title
			cd.Content = "视频预约"
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人预约"
		case activitygrpc.UpActReserveRelationType_Live:
			cd.Title = rsvRly.Item.Title
			cd.Content = fmt.Sprintf("%s 直播", bu.reserveTime(rsvRly.Item.LivePlanStartTime))
			hideReserveNum = !bu.displayReserveNum(rsvRly.Item.Total, rsvRly.Item.ReserveTotalShowLimit, ss.Mid(), rsvRly.Item.Upmid)
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人预约"
		case activitygrpc.UpActReserveRelationType_Course:
			cd.Title = rsvRly.Item.Title
			cd.Content = fmt.Sprintf("%s 开售", bu.reserveTime(rsvRly.Item.LivePlanStartTime))
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人已预约"
		default:
		}
		cd.Button = &api.ReserveButton{
			Goto: api.ReserveGoto_Reserve,
			MessageBox: &api.MessageBox{
				Type:              api.MessageBoxType_Dialog,
				AlertMsg:          "是否确认取消预约？",
				ConfirmButtonText: "确认",
				CancelButtonText:  "取消",
				ConfirmMsg:        "预约成功，会在开始时提醒您",
				CancelMsg:         "已取消预约",
			},
			HasDone:     rsvRly.Item.IsFollow == 1,
			DoneText:    "已预约",
			UndoneText:  "预约",
			Icon:        "http://i0.hdslb.com/bfs/archive/f5b7dae25cce338e339a655ac0e4a7d20d66145c.png",
			IsHighlight: false,
		}
		cd.Button.ReserveParams = bu.reserveParams(rsvRly.Sid, cd.Button.HasDone)
	case model.ReserveDisplayC:
		switch rsvRly.Item.Type {
		case activitygrpc.UpActReserveRelationType_Archive:
			cd.Title = rsvRly.Item.Title
			cd.Content = "视频预约"
			cd.Num = int64(rsvRly.Arc.GetStat().View)
			cd.Subtitle = "观看"
			cd.Button = &api.ReserveButton{Goto: api.ReserveGoto_Redirect, DoneText: "去观看"}
			if rsvRly.Arc != nil {
				cd.Button.Url = appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), strconv.FormatInt(rsvRly.Arc.GetAid(), 10),
					appcardmdl.ArcPlayHandler(rsvRly.Arc, nil, ss.TraceId(), nil, int(ss.RawDevice().Build), ss.RawDevice().RawMobiApp, false))
			}
		case activitygrpc.UpActReserveRelationType_Live:
			cd.Title = rsvRly.Item.Title
			cd.Content = "直播中"
			cd.Subtitle = "人气"
			cd.Button = &api.ReserveButton{Goto: api.ReserveGoto_Redirect, DoneText: "去观看", IsHighlight: true}
			if rsvRly.Live != nil {
				if si, ok := rsvRly.Live.SessionInfoPerLive[rsvRly.Item.Oid]; ok && si != nil {
					cd.Num = si.PopularityCount
					if ws := si.WatchedShow; ws != nil && ws.Switch {
						cd.Num = ws.Num
						cd.Subtitle = "人看过"
					}
				}
				cd.Button.Url = rsvRly.Live.JumpUrl[model.LiveEnteryFrom]
			}
		case activitygrpc.UpActReserveRelationType_Course:
			cd.Title = rsvRly.Item.Title
			cd.Num = rsvRly.Item.OidView
			cd.Subtitle = "人看过"
			cd.Button = &api.ReserveButton{Goto: api.ReserveGoto_Redirect, DoneText: "去观看", IsHighlight: true, Url: rsvRly.Item.BaseJumpUrl}
		default:
		}
	case model.ReserveDisplayD:
		if rsvRly.Item.Type == activitygrpc.UpActReserveRelationType_Live {
			cd.Title = rsvRly.Item.Title
			cd.Content = fmt.Sprintf("%s 直播", bu.reserveTime(rsvRly.Item.LivePlanStartTime))
			// 满足预约数条件
			hideReserveNum = !bu.displayReserveNum(rsvRly.Item.Total, rsvRly.Item.ReserveTotalShowLimit, ss.Mid(), rsvRly.Item.Upmid)
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人预约"
		}
		if si, ok := rsvRly.Live.SessionInfoPerLive[rsvRly.Item.Oid]; ok && si != nil {
			cd.Button = &api.ReserveButton{
				Goto:     api.ReserveGoto_Redirect,
				DoneText: "看回放",
				Url:      appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), si.Bvid, nil),
			}
		}
	case model.ReserveDisplayE:
		if rsvRly.Item.Type == activitygrpc.UpActReserveRelationType_Live {
			cd.Title = rsvRly.Item.Title
			cd.Content = fmt.Sprintf("%s 直播", bu.reserveTime(rsvRly.Item.LivePlanStartTime))
			//满足预约数条件
			hideReserveNum = !bu.displayReserveNum(rsvRly.Item.Total, rsvRly.Item.ReserveTotalShowLimit, ss.Mid(), rsvRly.Item.Upmid)
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人预约"
		} else if rsvRly.Item.Type == activitygrpc.UpActReserveRelationType_Course {
			cd.Title = rsvRly.Item.Title
			cd.Content = fmt.Sprintf("%s 开售", bu.reserveTime(rsvRly.Item.LivePlanStartTime))
			cd.Num = rsvRly.Item.Total
			cd.Subtitle = "人已预约"
		}
		cd.Button = &api.ReserveButton{
			Goto:       api.ReserveGoto_Unable,
			MessageBox: &api.MessageBox{Type: api.MessageBoxType_Toast, AlertMsg: "不在预约时间"},
			DoneText:   "已结束",
		}
	}
	cd.HideReserveNum = hideReserveNum
	return cd
}

func (bu Reserve) reserveTime(in xtime.Time) string {
	nowT := time.Now()
	if in.Time().Format("20060102") == nowT.Format("20060102") {
		return "今天 " + in.Time().Format("15:04")
	}
	atnowT := time.Now().AddDate(0, 0, 1)
	if in.Time().Format("20060102") == atnowT.Format("20060102") {
		return "明天 " + in.Time().Format("15:04")
	}
	if in.Time().Year() == nowT.Year() { //同一个自然年
		return in.Time().Format("01月02日 15:04")
	}
	return in.Time().Format("2006年01月02日 15:04")
}

// 是否展示预约数
func (bu Reserve) displayReserveNum(total, limit, mid, upmid int64) bool {
	// limit无限制 或者 发起人mid为0
	if limit <= 0 || upmid == 0 {
		return true
	}
	if mid == upmid {
		return true
	}
	if total >= limit {
		return true
	}
	return false
}

func (bu Reserve) reserveParams(sid int64, isFollowed bool) string {
	action := api.ActionType_Do
	if isFollowed {
		action = api.ActionType_Undo
	}
	return passthrough.Marshal(&api.ReserveParams{Action: action, Sid: sid})
}
