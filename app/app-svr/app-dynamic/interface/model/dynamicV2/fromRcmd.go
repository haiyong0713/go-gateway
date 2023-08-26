package dynamicV2

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/app-dynamic/interface/api/common"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// https://info.bilibili.co/pages/viewpage.action?pageId=416278182
type RcmdReply struct {
	Code        int             `json:"code"`
	Items       []*RcmdItem     `json:"data"`
	Infoc       *RcmdInfo       `json:"-"`
	UserFeature json.RawMessage `json:"user_feature"`
}

type RcmdItem struct {
	DynamicID  int64           `json:"dynamic_id"`
	UpID       int64           `json:"upid"`
	TrackID    string          `json:"trackid"`
	AvFeature  json.RawMessage `json:"av_feature"`
	ID         int64           `json:"id"`
	Goto       string          `json:"goto"`
	Source     string          `json:"source"`
	FromType   string          `json:"from_type"`
	RcmdReason struct {
		Content    string `json:"content"`     // 学校名称
		ReasonDesc string `json:"reason_desc"` // 推荐理由
	} `json:"rcmd_reason"`
}

type ItemRedis struct {
	ID         int64  `json:"id"`
	RcmdReason string `json:"rcmd_reason"`
}

func (i *RcmdItem) FromItem(r *ItemRedis) {
	i.ID = r.ID
	i.Goto = model.GotoAv
	i.RcmdReason = struct {
		Content    string `json:"content"`
		ReasonDesc string `json:"reason_desc"`
	}{
		Content:    r.RcmdReason,
		ReasonDesc: r.RcmdReason,
	}
}

func (rr *RcmdReply) ToV2CampusWaterFlowItems(dynCtx *DynamicContext, forceVideoHorizontal bool) (ret []*api.CampusWaterFlowItem) {
	if len(rr.Items) <= 0 {
		return nil
	}
	for _, v := range rr.Items {
		var rcmdTmp *api.CampusWaterFlowItem
		switch v.Goto {
		case "av":
			ap, ok := dynCtx.GetArchive(v.ID)
			if !ok || !ap.Arc.IsNormal() {
				continue
			}
			// 付费合集
			if PayAttrVal(ap.Arc) {
				continue
			}
			rcmdArc := &api.WFItemDefault{
				Title: ap.Arc.GetTitle(),
				Cover: ap.Arc.GetPic(),
				BottomLeft_1: &api.CoverIconWithText{
					Icon: api.CoverIcon_cover_icon_play,
					Text: model.UpStatString(int64(ap.Arc.Stat.View), ""),
				},
				BottomLeft_2: &api.CoverIconWithText{
					Icon: api.CoverIcon_cover_icon_danmaku,
					Text: model.UpStatString(int64(ap.Arc.Stat.Danmaku), ""),
				},
				BottomRight_1: &api.CoverIconWithText{
					Icon: api.CoverIcon_cover_icon_none,
					Text: model.VideoDuration(ap.Arc.Duration),
				},
				Uri: model.FillURI(model.GotoAv, strconv.FormatInt(ap.Arc.Aid, 10), model.AvPlayHandlerGRPCV2(ap, ap.Arc.FirstCid, true)),
				RcmdReason: &api.RcmdReason{
					CampusName: v.RcmdReason.Content,
					Style:      api.RcmdReasonStyle_rcmd_reason_style_campus_up,
					UpName:     ap.Arc.GetAuthor().Name,
					RcmdReason: v.RcmdReason.ReasonDesc,
				},
				Annotations: map[string]string{
					"up_mid":     strconv.FormatInt(ap.GetArc().GetAuthor().Mid, 10),
					"aid":        strconv.FormatInt(ap.Arc.Aid, 10),
					"track_id":   v.TrackID,
					"dynamic_id": strconv.FormatInt(v.DynamicID, 10),
				},
			}
			if len(v.RcmdReason.ReasonDesc) > 0 {
				rcmdArc.RcmdReason.Style = api.RcmdReasonStyle_rcmd_reason_style_campus_near_up_mix
			}
			// UGC转PGC逻辑
			if ap.Arc.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && ap.Arc.RedirectURL != "" {
				rcmdArc.Uri = ap.Arc.RedirectURL
			}
			ratio := &common.ItemWHRatio{
				Ratio: common.WHRatio_W_H_RATIO_16_9,
			}
			if !forceVideoHorizontal && dynCtx.IsVerticalArchive(ap.Arc) {
				ratio.Ratio = common.WHRatio_W_H_RATIO_3_4
			}
			rcmdTmp = &api.CampusWaterFlowItem{
				ItemType: api.WFItemType_WATER_FLOW_TYPE_ARCHIVE,
				WhRatio:  ratio,
				FlowItem: &api.CampusWaterFlowItem_ItemDefault{
					ItemDefault: rcmdArc,
				},
			}
		case "dynamic":
			draw, ok := dynCtx.GetResDraw(v.ID)
			if !ok || draw == nil || draw.Item == nil || draw.User == nil || len(draw.Item.Pictures) < 1 {
				continue
			}
			rcmdDraw := &api.WFItemDefault{
				Title: draw.Item.Description,
				Cover: draw.Item.Pictures[0].ImgSrc,
				Uri: model.FillURI(model.GotoDyn, strconv.FormatInt(v.DynamicID, 10),
					model.SuffixHandler(fmt.Sprintf("cardType=%d&rid=%d", DynTypeDraw, v.ID))),
				RcmdReason: &api.RcmdReason{
					CampusName: v.RcmdReason.Content,
					Style:      api.RcmdReasonStyle_rcmd_reason_style_campus_up,
					RcmdReason: v.RcmdReason.ReasonDesc,
					UpName:     draw.User.Name,
				},
				Annotations: map[string]string{
					"up_mid":     strconv.FormatInt(v.UpID, 10),
					"rid":        strconv.FormatInt(v.ID, 10),
					"track_id":   v.TrackID,
					"dynamic_id": strconv.FormatInt(v.DynamicID, 10),
				},
			}
			if len(draw.Item.Title) > 0 {
				rcmdDraw.Title = draw.Item.Title
			}
			if len(v.RcmdReason.ReasonDesc) > 0 {
				rcmdDraw.RcmdReason.Style = api.RcmdReasonStyle_rcmd_reason_style_campus_near_up_mix
			}
			rcmdTmp = &api.CampusWaterFlowItem{
				ItemType: api.WFItemType_WATER_FLOW_TYPE_DYNAMIC,
				WhRatio:  &common.ItemWHRatio{Ratio: common.WHRatio_W_H_RATIO_1_1},
				FlowItem: &api.CampusWaterFlowItem_ItemDefault{
					ItemDefault: rcmdDraw,
				},
			}
		}
		if rcmdTmp != nil {
			ret = append(ret, rcmdTmp)
		}
	}
	return
}
