package view

import (
	"fmt"
	"strconv"

	resApiV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/game"
)

var (
	badgeMap = map[string]*viewApi.PowerIconStyle{
		"1_1": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/XD3YCJIhhv.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/iTh2CQg1YX.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"1_2": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/GByBp5LMWF.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/3cNU6BZ1aL.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_3": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/MxbANa3J72.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/B2Oksm4t9V.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_4": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/MB4IQgpoA8.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/zrM6QfENFF.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_5": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/gJN43PzdI1.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/heoPGI2Wvt.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_6": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/LqXv6ZZN86.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/avdT6Nqdke.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_7": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/O1qfv4jvk6.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/FOLDEjPnDw.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_8": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/vXYa3ICiOb.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/S4Ihi6v77I.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_9": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/9LIbd98muA.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/HZpncjmheP.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"1_10": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mqqaJmcqBA.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/f6PS2KMtxV.png",
			IconWidth:    117,
			IconHeight:   22,
		},
		"5_1": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/lOa1gNvG3f.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/id3IYpMSzE.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"5_2": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/ucDkinU4hl.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/2WeyBDzGBj.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_3": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Y1imhOGTIH.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xDGL1iQbBM.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_4": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Yt0jQfwp9Y.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/0olaKlOpmA.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_5": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/1R1eBplOgJ.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/UR8yIgnhqT.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_6": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/5XdKov3Cu4.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/9bSe9TWdse.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_7": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/ZXvhyYQt2y.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/znce4KWeVi.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_8": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/rzYDmuK9PW.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/8h3HD3vk88.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_9": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/tcJWViZgln.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mBdoBHwRcd.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"5_10": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/7AhMGAmm8V.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/1fCID70ZIH.png",
			IconWidth:    117,
			IconHeight:   22,
		},
		"6_1": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/pjt3KPuZAG.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PptWhh0ZRN.png",
			IconWidth:    111,
			IconHeight:   22,
		},
		"6_2": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/k80MJ43lQ8.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/uWrNpSae20.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_3": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/IMTymWX39V.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/4UtEAgPfvZ.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_4": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xxdBzJoyLz.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/nD1PXLgghH.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_5": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PhYWAFMLCh.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/PCxfqIEjbI.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_6": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/q4C37JsUPV.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/E4OGn0gYHd.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_7": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/86Tf5sSqEO.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/mAAfPptBrH.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_8": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/oHECVTSzv0.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/5YG0UyAF1c.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_9": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/xnnDThMC5R.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/NcRnHgA3FV.png",
			IconWidth:    113,
			IconHeight:   22,
		},
		"6_10": {
			IconUrl:      "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/Jh577OveiI.png",
			IconNightUrl: "https://i0.hdslb.com/bfs/activity-plat/static/20220314/0977767b2e79d8ad0a36a731068a83d7/nNiQxtpU8V.png",
			IconWidth:    117,
			IconHeight:   22,
		},
	}
)

func (r *Relate) FromGameCard(rec *NewRelateRec, gameInfo *game.Game, from string, plat int8, build int, ms map[int64]*resApiV2.Material, mobiApp string, feedStyle string, powerBadgeSwitch bool) {
	r.Goto = model.GotoGame
	r.From = from
	r.Badge = "游戏"
	r.NewCard = 1
	r.URI = model.FillURI(r.Goto, gameInfo.GameLink, nil)
	r.Param = strconv.FormatInt(gameInfo.GameBaseID, 10)
	r.Button = &viewApi.Button{Title: "进入", Uri: r.URI}

	//游戏标题
	r.Title = getGameTitle(plat, build, gameInfo.GameName, rec.Title)
	//游戏封面
	r.Cover = gameInfo.Cover
	//游戏图标
	r.Pic = gameInfo.GameIcon
	//游戏状态
	r.ReserveStatus = int64(gameInfo.GameStatusV2)
	r.ReserveStatusText = getGameReserveStatusText(gameInfo.GameStatusV2)
	//游戏预约/下载人数文案
	r.Reserve = getGameReserve(gameInfo.GameStatusV2, gameInfo.BookNum, gameInfo.DownloadNum, mobiApp, plat, build, feedStyle)
	//游戏评分
	r.Rating = float64(gameInfo.Grade)
	//游戏分类
	r.TagName = gameInfo.GameTags
	//榜单信息
	if gameInfo.RankInfo != nil {
		r.RankInfo = gameInfo.RankInfo
	}
	//fix: ios73版本bug.
	// nolint:gomnd
	versionMatch := mobiApp == "iphone" && plat == model.PlatIPhone && build >= 67300000
	if versionMatch {
		r.RatingCount = 1
	}

	//游戏礼包
	if gameInfo.GiftTitle != "" && gameInfo.GiftURL != "" {
		r.PackInfo = &viewApi.PackInfo{
			Title: gameInfo.GiftTitle,
			Uri:   gameInfo.GiftURL,
		}
	}
	//游戏公告
	if gameInfo.NoticeTitle != "" && gameInfo.Notice != "" {
		r.Notice = &viewApi.Notice{
			Title: gameInfo.NoticeTitle,
			Desc:  gameInfo.Notice,
		}
	}
	//粉标
	r.BadgeStyle = reasonStyle(model.BgColorTransparentRed, "游戏")
	//强化角标
	if powerBadgeSwitch {
		r.PowerIconStyle = buildPowerIconStyle(gameInfo.RankType, gameInfo.GameRank)
	}

	//以下推荐配置
	r.CoverGif = rec.CoverGif
	//推荐理由
	if rec.RcmdReason != nil && rec.RcmdReason.Content != "" {
		r.RcmdReason = rec.RcmdReason.Content
	}
	if rec.UniqueId != "" {
		r.UniqueId = rec.UniqueId
	}
	//运营卡物料id
	if rec.MaterialId > 0 && ms != nil {
		if _, ok := ms[rec.MaterialId]; ok {
			if ms[rec.MaterialId].Title != "" && ms[rec.MaterialId].Cover != "" && ms[rec.MaterialId].Desc != "" {
				r.MaterialId = rec.MaterialId
				r.Title = ms[rec.MaterialId].Title
				r.Pic = ms[rec.MaterialId].Cover
				r.Desc = ms[rec.MaterialId].Desc
			}
		}
	}
	//运营卡召回源
	r.FromSourceType = rec.FromSourceType
	//运营卡召回源id list
	r.FromSourceId = rec.FromSourceId
}

//nolint:gomnd
func getGameReserveStatusText(gameStatus int32) string {
	//游戏状态为下载：立即下载
	//游戏状态为预约：立即预约
	//游戏状态为其它：点击查看
	switch gameStatus {
	case 2:
		return "立即下载"
	case 1:
		return "立即预约"
	default:
		return "点击查看"
	}
}

func getGameTitle(plat int8, build int, gameName, recTitle string) string {
	if gameName == "" {
		return recTitle
	}
	if plat == model.PlatIPhone && build > 8740 || plat == model.PlatAndroid && build > 5455000 {
		return gameName
	}
	return "相关游戏：" + gameName
}

//nolint:gomnd
func getGameReserve(gameStatus int32, bookNum int64, downloadNum int64, mobiApp string, plat int8, build int, feedStyle string) string {
	var reserve string
	switch gameStatus {
	case 2:
		//仅针对v2双列实验小卡下发下载量
		versionMatch := (mobiApp == "android" && build >= 6730000) ||
			(mobiApp == "iphone" && plat == model.PlatIPhone && build >= 67300000)
		if feedStyle == "v2" && versionMatch {
			if downloadNum < 10000 {
				reserve = strconv.FormatInt(downloadNum, 10) + "下载"
			} else {
				reserve = strconv.FormatFloat(float64(downloadNum)/10000, 'f', 1, 64) + "万下载"
			}
		}
	case 1:
		if bookNum < 10000 {
			reserve = strconv.FormatInt(bookNum, 10) + "人预约"
		} else {
			reserve = strconv.FormatFloat(float64(bookNum)/10000, 'f', 1, 64) + "万人预约"
		}
	}
	return reserve
}

func buildPowerIconStyle(rankType int8, gameRank int8) *viewApi.PowerIconStyle {
	badge, ok := badgeMap[fmt.Sprintf("%d_%d", rankType, gameRank)]
	if !ok {
		return nil
	}
	return &viewApi.PowerIconStyle{
		IconUrl:      badge.IconUrl,
		IconNightUrl: badge.IconNightUrl,
		IconHeight:   badge.IconHeight,
		IconWidth:    badge.IconWidth,
	}
}
