package common

// CardMap .
type CardMap struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// never modify the card number if the card is online
// web search
const (
	//WebSearchSpecialSmall special small card
	WebSearchSpecialSmall = 1
	//WebSearchGame web search game card
	WebSearchGame = 2
	//WebSearchUpUser web search up user
	WebSearchUpUser = 3
	//WebSearchVideoSpecialSmall web search video special small
	WebSearchVideoSpecialSmall = 7
)

var PlatDict = map[int]string{
	2:  "iPad粉",
	12: "OTT",
	20: "iPadHD",
	30: "Web",
}

// web recommand
const (
	//WebRcmdSpecial web recommand special card
	WebRcmdSpecial = 1
	//WebRcmdAV WebRcmd video card
	WebRcmdAV = 2
	//WebRcmdGame WebRcmd game card
	WebRcmdGame = 3
)

// search hot word intervene
const (
	//SeaHotGoToArch goto archiev
	SeaHotGoToArch = 1
	//SeaHotGoToArticle goto article
	SeaHotGoToArticle = 2
	//SeaHotGoToPGC go to pgc
	SeaHotGoToPGC = 3
	//SeaHotGoToURL go to url
	SeaHotGoToURL = 4

	//SeaHotCTDefault search hot card type default
	SeaHotCTDefault = 1
	//SeaHotCTFire card type fire
	SeaHotCTFire = 2
	//SeaHotCTSpeColor special color
	SeaHotCTSpeColor = 3
	//SeaHotCTNewest newest
	SeaHotCTNewest = 4
	//SeaHotCTNewest newest
	SeaHotCTPopular = 5
	//SeaHotCTNewest newest
	SeaHotCTSelfDefine = 6
	//SeaBoxTypeWord search box type word
	SeaBoxTypeWord = 1
	//SeaBoxTypeSpecial type special
	SeaBoxTypeSpecial = 2
	//SeaBoxGoToArch search box goto archive
	SeaBoxGoToArch = 1
	//SeaBoxGoToArticle goto article
	SeaBoxGoToArticle = 2
	//SeaBoxGoToPGC goto pgc
	SeaBoxGoToPGC = 3
	//SeaHotGoToURL go to url
	SeaBoxGoToURL = 4
	//SeaShieldAv search shield goto av
	SeaShieldAv = 1
	//SeaShieldGame goto game
	SeaShieldGame = 2
	//SeaShieldPgc goto pgc
	SeaShieldPgc = 3
	//SeaShieldUp goto user
	SeaShieldUp = 4
	//SeaShieldLive goto live
	SeaShieldLive = 5
	//SeaShieldArt goto article
	SeaShieldArt = 6
	//SeaShieldDync goto dynamic
	SeaShieldDync = 7
	//SeaShieldGoods goto goods
	SeaShieldGoods = 8
	//SeaShieldShow goto vip buy show
	SeaShieldShow = 9
	//SeaShieldComic goto comic
	SeaShieldComic = 10
	//SeaShieldTopic goto topic
	SeaShieldTopic = 11
	//OgvAV ogv av
	OgvAV = 2
	//OgvPgc ogv pgc
	OgvPgc = 3
	//OgvArticle ogv article
	OgvArticle = 4
)

// CardType .
type CardType struct {
	WebSearch        []*CardMap `json:"web_search"`
	WebRcmd          []*CardMap `json:"web_rcmd"`
	WebRcmdCard      []*CardMap `json:"web_rcmd_card"`
	SearchHotGoto    []*CardMap `json:"search_hot"`
	SearchHotType    []*CardMap `json:"search_hot_type"`
	SearchBoxGoto    []*CardMap `json:"search_box"`
	SearchBoxType    []*CardMap `json:"search_box_type"`
	SearchShieldType []*CardMap `json:"search_shield_type"`
}
