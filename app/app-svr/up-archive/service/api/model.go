package api

// 各属性地址见 http://syncsvn.bilibili.co/platform/doc/blob/master/archive/field/state.md
// all const
const (
	// open state
	StateOpen    = 0
	StateOrange  = 1
	AccessMember = int32(10000)
	// forbid state
	StateForbidWait       = -1
	StateForbidRecicle    = -2
	StateForbidPolice     = -3
	StateForbidLock       = -4
	StateForbidFixed      = -6
	StateForbidLater      = -7
	StateForbidAdminDelay = -10
	StateForbidXcodeFail  = -16
	StateForbidSubmit     = -30
	StateForbidUserDelay  = -40
	StateForbidUpDelete   = -100
	StateForbidSteins     = -20
	// copyright
	CopyrightUnknow   = int8(0)
	CopyrightOriginal = int8(1)
	CopyrightCopy     = int8(2)

	// attribute yes and no
	AttrYes = int32(1)
	AttrNo  = int32(0)
	// attribute bit
	AttrBitNoRank    = uint(0)
	AttrBitNoDynamic = uint(1)
	AttrBitNoWeb     = uint(2)
	AttrBitNoMobile  = uint(3)
	// AttrBitNoSearch    = uint(4)
	AttrBitOverseaLock = uint(5)
	AttrBitNoRecommend = uint(6)
	AttrBitNoReprint   = uint(7)
	AttrBitHasHD5      = uint(8)
	AttrBitIsPGC       = uint(9)
	AttrBitAllowBp     = uint(10)
	AttrBitIsBangumi   = uint(11)
	AttrBitIsPorder    = uint(12)
	AttrBitLimitArea   = uint(13)
	AttrBitAllowTag    = uint(14)
	// AttrBitIsFromArcApi  = uint(15)
	AttrBitJumpUrl       = uint(16)
	AttrBitIsMovie       = uint(17)
	AttrBitBadgepay      = uint(18)
	AttrBitUGCPay        = uint(22)
	AttrBitHasBGM        = uint(23)
	AttrBitIsCooperation = uint(24)
	AttrBitHasViewpoint  = uint(25)
	AttrBitHasArgument   = uint(26)
	AttrBitUGCPayPreview = uint(27)
	AttrBitSteinsGate    = uint(29)
	AttrBitIsPUGVPay     = uint(30)
	// attribute_v2
	AttrBitV2NoBackground = uint(0)
	AttrBitV2NoPublic     = uint(1)
	AttrBitV2NoSpace      = uint(11)
	// 付费稿件
	AttrBitV2Pay = uint(13)
	// 是否360全景视频
	AttrBitV2Is360 = uint(2)
	//是否云非编稿件
	AttrBitV2BsEditor = uint(3)
	//是否存量导入的小视频
	AttrBitV2IsImportSvideo = uint(4)
	//播放页干净模式
	AttrBitV2CleanMode = uint(5)
	//禁止特别关注push
	AttrBitV2NoFansPush = uint(6)
	//是否开启杜比音效
	AttrBitV2IsDolby = uint(7)
	// 仅收藏可见
	AttrBitV2OnlyFavView = uint(8)
	// 是否活动合集
	AttrBitV2ActSeason = uint(9)
	// staff attribute
	StaffAttrBitAdOrder = uint(0)
)

// IsNormal is
func (a *Arc) IsNormal() bool {
	return a.State >= StateOpen
}

// AttrVal get attr val by bit.
func (a *Arc) AttrVal(bit uint) int32 {
	return (a.Attribute >> bit) & int32(1)
}

// AttrValV2 get attr v2 val by bit.
func (a *Arc) AttrValV2(bit uint) int32 {
	return int32((a.AttributeV2 >> bit) & int64(1))
}

// StaffAttrVal get staff attr val by bit.
func (staff *StaffInfo) StaffAttrVal(bit uint) int32 {
	return int32((staff.Attribute >> bit) & int64(1))
}
