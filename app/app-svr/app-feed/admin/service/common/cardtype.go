package common

import "go-gateway/app/app-svr/app-feed/admin/model/common"

// CardType .
func (s *Service) CardType() (m *common.CardType) {
	//searchWeb web搜索
	searchWeb := []*common.CardMap{
		{
			Name: "特殊小卡",
			ID:   common.WebSearchSpecialSmall,
		},
		{
			Name: "游戏卡片",
			ID:   common.WebSearchGame,
		},
		{
			Name: "UP主卡片",
			ID:   common.WebSearchUpUser,
		},
		{
			Name: "视频模块特殊小卡",
			ID:   common.WebSearchVideoSpecialSmall,
		},
	}
	webRcmd := []*common.CardMap{
		{
			Name: "Web相关推荐 特殊卡片",
			ID:   common.WebRcmdSpecial,
		},
		{
			Name: "Web相关推荐 视频卡片",
			ID:   common.WebRcmdAV,
		},
		{
			Name: "Web相关推荐 游戏卡片",
			ID:   common.WebRcmdGame,
		},
	}
	webRcmdCard := []*common.CardMap{
		{
			Name: "web相关推荐特殊卡片",
			ID:   common.WebRcmdSpecial,
		},
	}
	//搜索热词跳转类型
	searchHotGoto := []*common.CardMap{
		{
			Name: "稿件",
			ID:   common.SeaHotGoToArch,
		},
		{
			Name: "专栏",
			ID:   common.SeaHotGoToArticle,
		},
		{
			Name: "PGC",
			ID:   common.SeaHotGoToPGC,
		},
		{
			Name: "URL",
			ID:   common.SeaHotGoToURL,
		},
	}
	//搜索热词类型
	searchHotType := []*common.CardMap{
		{
			Name: "默认",
			ID:   common.SeaHotCTDefault,
		},
		{
			Name: "小火苗",
			ID:   common.SeaHotCTFire,
		},
		{
			Name: "特殊底色",
			ID:   common.SeaHotCTSpeColor,
		},
		{
			Name: "最新",
			ID:   common.SeaHotCTNewest,
		},
		{
			Name: "最热",
			ID:   common.SeaHotCTPopular,
		},
		{
			Name: "自定义",
			ID:   common.SeaHotCTSelfDefine,
		},
	}
	searchBoxType := []*common.CardMap{
		{
			Name: "搜索词",
			ID:   common.SeaHotGoToArch,
		},
		{
			Name: "特定内容",
			ID:   common.SeaHotGoToArticle,
		},
	}
	searchBoxGoto := []*common.CardMap{
		{
			Name: "稿件",
			ID:   common.SeaBoxGoToArch,
		},
		{
			Name: "专栏",
			ID:   common.SeaBoxGoToArticle,
		},
		{
			Name: "PGC",
			ID:   common.SeaBoxGoToPGC,
		},
		{
			Name: "URL",
			ID:   common.SeaBoxGoToURL,
		},
	}
	searchShieldGoto := []*common.CardMap{
		{
			Name: "稿件",
			ID:   common.SeaShieldAv,
		},
		{
			Name: "游戏",
			ID:   common.SeaShieldGame,
		},
		{
			Name: "pgc",
			ID:   common.SeaShieldPgc,
		},
		{
			Name: "用户",
			ID:   common.SeaShieldUp,
		},
		{
			Name: "直播",
			ID:   common.SeaShieldLive,
		},
		{
			Name: "专栏",
			ID:   common.SeaShieldArt,
		},
		{
			Name: "动态",
			ID:   common.SeaShieldDync,
		},
		{
			Name: "商品",
			ID:   common.SeaShieldGoods,
		},
		{
			Name: "展演",
			ID:   common.SeaShieldShow,
		},
		{
			Name: "漫画",
			ID:   common.SeaShieldComic,
		},
		{
			Name: "话题",
			ID:   common.SeaShieldTopic,
		},
	}
	return &common.CardType{
		WebSearch:        searchWeb,
		WebRcmd:          webRcmd,
		WebRcmdCard:      webRcmdCard,
		SearchHotGoto:    searchHotGoto,
		SearchHotType:    searchHotType,
		SearchBoxType:    searchBoxType,
		SearchBoxGoto:    searchBoxGoto,
		SearchShieldType: searchShieldGoto,
	}
}
