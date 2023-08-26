package model

import (
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	ugcseason "go-gateway/app/app-svr/ugc-season/service/api"
	"go-gateway/pkg/idsafe/bvid"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	webgrpc "git.bilibili.co/bapis/bapis-go/bilibili/web/interface/v1"
	activiegrpc "git.bilibili.co/bapis/bapis-go/manager/service/active"
)

const (
	ActSubTypeBeforeLive           = 1
	ActSubTypeDuringLive           = 2
	ActSubTypeAfterLive            = 3
	_actReserveFrom                = "video_page"
	_actReserveType                = "video_page"
	_actSubBeforeLiveTitle         = "预约"
	_actSubBeforeLiveSelectedTitle = "已预约"
	_actSubDuringLiveTitle         = "去直播"
	_actSubAfterLiveTitle          = "订阅"
	_actSubAfterLiveSelectedTitle  = "已订阅"
)

type ActivitySeasonMem struct {
	// 白名单
	Whitelist        []int64
	ActivityURL      string
	Title            string
	Live             *webgrpc.ActivityLive
	SeasonView       *ActivitySeason
	Theme            *webgrpc.ActivityTheme
	Game             *webgrpc.ActivityGame
	Subscribe        map[int32]*activiegrpc.ReserveMenu
	IsContainedRecom bool
}

type ActivitySeason struct {
	Season   *ugcseason.Season
	Sections []*ActivitySection
}

type ActivitySection struct {
	// 剧集ID
	SeasonID int64
	// 小节ID
	ID int64
	// 小节标题
	Title string
	// 小节类型 0其他 1正片
	Type int64
	// 单集列表
	Episodes []*ActivityEpisode
}

type ActivityEpisode struct {
	*ugcseason.Episode
	ActivityView *ActivityView
}

type ActivityView struct {
	Arc          *webgrpc.Arc
	Pages        []*arcgrpc.Page
	StaffInfo    []*webgrpc.Staff
	RightRelate  *webgrpc.OperationRelate
	BottomRelate *webgrpc.OperationRelate
}

func StaffInfoFromCard(card *accgrpc.Card, staff *arcgrpc.StaffInfo) *webgrpc.Staff {
	item := &webgrpc.Staff{
		Mid:   staff.Mid,
		Title: staff.Title,
		Name:  card.Name,
		Face:  card.Face,
		Vip: &webgrpc.VipInfo{
			Type:       card.Vip.Type,
			Status:     card.Vip.Status,
			VipPayType: card.Vip.VipPayType,
			ThemeType:  card.Vip.ThemeType,
		},
		Official: &webgrpc.OfficialInfo{
			Role:  card.Official.Role,
			Title: card.Official.Title,
			Desc:  card.Official.Desc,
		},
	}
	if staff.StaffAttrVal(arcgrpc.StaffAttrBitAdOrder) == arcgrpc.AttrYes {
		item.LabelStyle = StaffLabelAd
	}
	return item
}

func OperationRelateFromCommRecommends(in *activiegrpc.CommRecommends, title string, aiRelate []*webgrpc.Relate) *webgrpc.OperationRelate {
	out := &webgrpc.OperationRelate{Title: title}
	if in != nil {
		for _, v := range in.CommRecommends {
			if v == nil {
				continue
			}
			out.RelateItem = append(out.RelateItem, &webgrpc.RelateItem{
				Url:   v.CommJumpUrl,
				Cover: v.CommPic,
			})
		}
	}
	for _, val := range aiRelate {
		if val == nil {
			continue
		}
		out.AiRelateItem = append(out.AiRelateItem, val)
	}
	if len(out.RelateItem) == 0 && len(out.AiRelateItem) == 0 {
		return nil
	}
	return out
}

func (out *ActivitySeason) CopyFromUgcSeason(in *ugcseason.View, arcs map[int64]*ActivityView) {
	out.Season = in.Season
	for _, sec := range in.Sections {
		if sec == nil {
			continue
		}
		tmp := &ActivitySection{
			SeasonID: sec.SeasonID,
			ID:       sec.ID,
			Title:    sec.Title,
			Type:     sec.Type,
		}
		for _, ep := range sec.Episodes {
			if ep == nil {
				continue
			}
			arc, ok := arcs[ep.Aid]
			if !ok || arc == nil {
				continue
			}
			tmp.Episodes = append(tmp.Episodes, &ActivityEpisode{
				Episode:      ep,
				ActivityView: arc,
			})
		}
		out.Sections = append(out.Sections, tmp)
	}
}

func FillFromActivityView(res *webgrpc.ActivityView, other *ActivityView) {
	res.Arc = other.Arc
	res.Bvid, _ = bvid.AvToBv(other.Arc.Aid)
	res.Pages = CopyFromArcPageGRPC(other.Pages)
	res.Staff = other.StaffInfo
	res.RightRelate = other.RightRelate
	res.BottomRelate = other.BottomRelate
}

func CopyFromArcToWebGRPC(in *arcgrpc.Arc) *webgrpc.Arc {
	return &webgrpc.Arc{
		Aid:         in.Aid,
		Videos:      in.Videos,
		TypeId:      in.TypeID,
		TypeName:    in.TypeName,
		Copyright:   in.Copyright,
		Pic:         in.Pic,
		Title:       in.Title,
		Pubdate:     in.PubDate,
		Ctime:       in.Ctime,
		Desc:        in.Desc,
		State:       in.State,
		Tag:         in.Tag,
		Tags:        in.Tags,
		Duration:    in.Duration,
		MissionId:   in.MissionID,
		OrderId:     in.OrderID,
		RedirectUrl: in.RedirectURL,
		Forward:     in.Forward,
		Rights: webgrpc.Rights{
			Bp:            in.Rights.Bp,
			Elec:          in.Rights.Elec,
			Download:      in.Rights.Download,
			Movie:         in.Rights.Movie,
			Pay:           in.Rights.Pay,
			Hd5:           in.Rights.HD5,
			NoReprint:     in.Rights.NoReprint,
			Autoplay:      in.Rights.Autoplay,
			UgcPay:        in.Rights.UGCPay,
			IsCooperation: in.Rights.IsCooperation,
			UgcPayPreview: in.Rights.UGCPayPreview,
		},
		Author: webgrpc.Author{
			Mid:  in.Author.Mid,
			Name: in.Author.Name,
			Face: in.Author.Face,
		},
		Stat: webgrpc.Stat{
			Aid:     in.Stat.Aid,
			View:    in.Stat.View,
			Danmaku: in.Stat.Danmaku,
			Reply:   in.Stat.Reply,
			Fav:     in.Stat.Fav,
			Coin:    in.Stat.Coin,
			Share:   in.Stat.Share,
			NowRank: in.Stat.NowRank,
			HisRank: in.Stat.HisRank,
			Like:    in.Stat.Like,
			Dislike: in.Stat.DisLike,
		},
		ReportResult: in.ReportResult,
		Dynamic:      in.Dynamic,
		FirstCid:     in.FirstCid,
		Dimension: webgrpc.Dimension{
			Width:  in.Dimension.Width,
			Height: in.Dimension.Height,
			Rotate: in.Dimension.Rotate,
		},
		SeasonId: in.SeasonID,
	}
}

func CopyFromArcPageGRPC(in []*arcgrpc.Page) []*webgrpc.Page {
	var out []*webgrpc.Page
	for _, v := range in {
		if v == nil {
			continue
		}
		out = append(out, &webgrpc.Page{
			Cid:      v.Cid,
			Page:     v.Page,
			From:     v.From,
			Part:     v.Part,
			Duration: v.Duration,
			Vid:      v.Vid,
			Desc:     v.Desc,
			Weblink:  v.WebLink,
			Dimension: webgrpc.Dimension{
				Width:  v.Dimension.Width,
				Height: v.Dimension.Height,
				Rotate: v.Dimension.Rotate,
			},
		})
	}
	return out
}

func CopyFromActivitySection(in []*ActivitySection) []*webgrpc.ActivitySeasonSection {
	var out []*webgrpc.ActivitySeasonSection
	for _, sec := range in {
		if sec == nil {
			continue
		}
		tmp := &webgrpc.ActivitySeasonSection{
			Id:    sec.ID,
			Title: sec.Title,
			Type:  sec.Type,
		}
		for _, ep := range sec.Episodes {
			if ep == nil {
				continue
			}
			bvStr, _ := bvid.AvToBv(ep.Episode.GetAid())
			tmpEp := &webgrpc.ActivityEpisode{
				Id:     ep.Episode.GetID(),
				Aid:    ep.Episode.GetAid(),
				Bvid:   bvStr,
				Cid:    ep.Episode.GetCid(),
				Title:  ep.Episode.GetTitle(),
				Cover:  ep.Episode.GetArc().GetPic(),
				Author: &webgrpc.Author{},
				Rights: &webgrpc.Rights{},
			}
			if ep.ActivityView != nil && ep.ActivityView.Arc != nil {
				tmpEp.Author = &webgrpc.Author{
					Mid:  ep.ActivityView.Arc.Author.Mid,
					Name: ep.ActivityView.Arc.Author.Name,
					Face: ep.ActivityView.Arc.Author.Face,
				}
				tmpEp.Rights = &ep.ActivityView.Arc.Rights
			}
			tmp.Episodes = append(tmp.Episodes, tmpEp)
		}
		out = append(out, tmp)
	}
	return out
}

func CopyFromPcPlay(in *activiegrpc.PCActivePlay) *webgrpc.ActivityTheme {
	if in == nil {
		return nil
	}
	return &webgrpc.ActivityTheme{
		BaseColor:                 in.PcColor.GetBaseColor(),
		LoadingBgColor:            in.PcColor.GetLoadingBgColor(),
		OperatedBgColor:           in.PcColor.GetOperatedBgColor(),
		DefaultElementColor:       in.PcColor.GetDefaultElementColor(),
		HoverElementColor:         in.PcColor.GetHoverElementColor(),
		SelectedElementColor:      in.PcColor.GetSelectedElementColor(),
		BaseFontColor:             in.PcColor.GetBaseFontColor(),
		InfoFontColor:             in.PcColor.GetInfoFontColor(),
		MaskBgColor:               in.PcColor.GetMaskBgColor(),
		PageBgColor:               in.PcColor.GetPageBgColor(),
		CenterLogoImg:             in.PcRes.GetCenterLogoImg(),
		PageBgImg:                 in.PcRes.GetPageBgImg(),
		Decorations_2233Img:       in.PcRes.GetDecorations_2233Img(),
		MainBannerBgImg:           in.PcRes.GetMainBannerBgImg(),
		MainBannerTitleImg:        in.PcRes.GetMainBannerTitleImg(),
		LikeAnimationImg:          in.PcRes.GetLikeAnimationImg(),
		ComboLikeImg:              in.PcRes.GetComboLikeImg(),
		ComboCoinImg:              in.PcRes.GetComboCoinImg(),
		ComboFavImg:               in.PcRes.GetComboFavImg(),
		ArrowBtnImg:               in.PcRes.GetArrowBtnImg(),
		ShareIconBgImg:            in.PcRes.GetShareIconBgImg(),
		LiveListLocationImg:       in.PcRes.GetLiveListLocationImg(),
		LiveListLocationImgActive: in.PcRes.GetLiveListLocationImgActive(),
		PlayerLoadingImg:          in.PcRes.GetPlayerLoadingImg(),
		ShareImg:                  in.PcRes.GetShareImg(),
		KvColor:                   in.GetKvColor(),
	}
}

func FillSubscribe(actSub *webgrpc.ActivitySubscribe, menu *activiegrpc.ReserveMenu, subType int, seasonID int64, activityURL string) {
	if menu == nil {
		switch subType {
		case ActSubTypeBeforeLive, ActSubTypeAfterLive:
			actSub.ButtonTitle = _actSubAfterLiveTitle
			actSub.ButtonSelectedTitle = _actSubAfterLiveSelectedTitle
			actSub.OrderType = webgrpc.OrderType_TypeFavSeason
			actSub.Param = &webgrpc.ActivitySubscribe_FavParam{FavParam: &webgrpc.FavSeasonParam{SeasonId: seasonID}}
		case ActSubTypeDuringLive:
			actSub.ButtonTitle = _actSubDuringLiveTitle
			actSub.ButtonSelectedTitle = _actSubDuringLiveTitle
			actSub.OrderType = webgrpc.OrderType_TypeClick
			actSub.Param = &webgrpc.ActivitySubscribe_JumpParam{JumpParam: &webgrpc.JumpURLParam{JumpUrl: activityURL}}
		}
		return
	}
	switch subType {
	case ActSubTypeBeforeLive:
		actSub.ButtonTitle = func() string {
			if menu.UnclickedText != "" {
				return menu.UnclickedText
			}
			return _actSubBeforeLiveTitle
		}()
		actSub.ButtonSelectedTitle = func() string {
			if menu.ClickedText != "" {
				return menu.ClickedText
			}
			return _actSubBeforeLiveSelectedTitle
		}()
		actSub.OrderType = webgrpc.OrderType_TypeOrderActivity
		actSub.Param = &webgrpc.ActivitySubscribe_ReserveParam{ReserveParam: &webgrpc.ReserveActivityParam{
			ReserveId: menu.MenuId,
			From:      _actReserveFrom,
			Type:      _actReserveType,
			Oid:       seasonID,
		}}
	case ActSubTypeDuringLive:
		actSub.ButtonTitle = func() string {
			if menu.UnclickedText != "" {
				return menu.UnclickedText
			}
			return _actSubDuringLiveTitle
		}()
		actSub.ButtonSelectedTitle = func() string {
			if menu.UnclickedText != "" {
				return menu.UnclickedText
			}
			return _actSubDuringLiveTitle
		}()
		actSub.OrderType = webgrpc.OrderType_TypeClick
		actSub.Param = &webgrpc.ActivitySubscribe_JumpParam{JumpParam: &webgrpc.JumpURLParam{JumpUrl: activityURL}}
	case ActSubTypeAfterLive:
		actSub.ButtonTitle = func() string {
			if menu.UnclickedText != "" {
				return menu.UnclickedText
			}
			return _actSubAfterLiveTitle
		}()
		actSub.ButtonSelectedTitle = func() string {
			if menu.ClickedText != "" {
				return menu.ClickedText
			}
			return _actSubAfterLiveSelectedTitle
		}()
		actSub.OrderType = webgrpc.OrderType_TypeFavSeason
		actSub.Param = &webgrpc.ActivitySubscribe_FavParam{FavParam: &webgrpc.FavSeasonParam{SeasonId: seasonID}}
	}
}
