package model

import (
	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

// NavResp  struct of nav api response
type NavResp struct {
	IsLogin bool `json:"isLogin"`
	//AccessStatus  int    `json:"accessStatus"`
	//DueRemark     string `json:"dueRemark"`
	EmailVerified int32                     `json:"email_verified"`
	Face          string                    `json:"face"`
	FaceNft       int32                     `json:"face_nft"`
	FaceNftType   pangugsgrpc.NFTRegionType `json:"face_nft_type"`
	LevelInfo     struct {
		Cur     int32       `json:"current_level"`
		Min     int32       `json:"current_min"`
		NowExp  int32       `json:"current_exp"`
		NextExp interface{} `json:"next_exp"`
	} `json:"level_info"`
	Mid            int64               `json:"mid"`
	MobileVerified int32               `json:"mobile_verified"`
	Coins          float64             `json:"money"`
	Moral          float32             `json:"moral"`
	Official       accmdl.OfficialInfo `json:"official"`
	OfficialVerify OfficialVerify      `json:"officialVerify"`
	Pendant        accmdl.PendantInfo  `json:"pendant"`
	Scores         int                 `json:"scores"`
	Uname          string              `json:"uname"`
	// TODO 以后可删除
	/*-----------------------------------------------------------*/
	VipDueDate         int64           `json:"vipDueDate"`
	VipStatus          int32           `json:"vipStatus"`
	VipType            int32           `json:"vipType"`
	VipPayType         int32           `json:"vip_pay_type"`
	VipThemeType       int32           `json:"vip_theme_type"`
	VipLabel           accmdl.VipLabel `json:"vip_label"`
	VipAvatarSubscript int32           `json:"vip_avatar_subscript"`
	VipNicknameColor   string          `json:"vip_nickname_color"`
	/*-----------------------------------------------------------*/
	Vip            accmdl.VipInfo `json:"vip"`
	Wallet         *Wallet        `json:"wallet"`
	HasShop        bool           `json:"has_shop"`
	ShopURL        string         `json:"shop_url"`
	AllowanceCount int64          `json:"allowance_count"`
	AnswerStatus   int32          `json:"answer_status"`
	IsSeniorMember int32          `json:"is_senior_member"`
}

// FailedNavResp struct of failed nav response
type FailedNavResp struct {
	IsLogin bool `json:"isLogin"`
}

// Wallet struct.
type Wallet struct {
	Mid           int64   `json:"mid"`
	BcoinBalance  float32 `json:"bcoin_balance"`
	CouponBalance float32 `json:"coupon_balance"`
	CouponDueTime int64   `json:"coupon_due_time"`
}

// NavStat .
type NavStat struct {
	Following    int64 `json:"following"`
	Follower     int64 `json:"follower"`
	DynamicCount int64 `json:"dynamic_count"`
}
