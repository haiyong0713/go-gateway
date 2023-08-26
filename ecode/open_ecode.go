package ecode

import . "go-common/library/ecode"

// common  ecode
// 开放平台 2000000~2999999
// 票务的code码 2000000~2099999
var (
	//销售-营销
	TicketUnKnown               = New(2000000) //未知错误
	TicketParamInvalid          = New(2000001) //参数错误
	TicketRecordDupli           = New(2000002) //重复插入
	TicketRecordLost            = New(2000003) //数据不存在
	TicketPromotionLost         = New(2000004) //活动不存在
	TicketPromotionEnd          = New(2000005) //活动结束
	TicketPromotionRepeatJoin   = New(2000006) //活动重复参加
	TicketPromotionGroupLost    = New(2000007) //拼团不存在
	TicketPromotionGroupFull    = New(2000008) //拼团人数已满
	TicketPromotionGroupNotFull = New(2000009) //拼团人数未满
	TicketPromotionOrderLost    = New(2000010) //拼团订单不存在
	TicketPromoExistSameTime    = New(2000011) //同时间段存在已上架拼团活动
	TicketAddPromoOrderFail     = New(2000012) //添加活动订单失败
	TicketAddPromoGroupFail     = New(2000013) //添加拼团 团订单失败
	TicketPromoGroupEnd         = New(2000014) //拼团 团订单已失效
	TicketUpdatePromoOrderFail  = New(2000015) //更新拼团订单失败
	TicketUpdatePromoGroupFail  = New(2000016) //更新拼团 团订单失败
	IllegalPromoOperate         = New(2000017) //拼团 不支持的操作类型
	PromoStatusChanged          = New(2000018) //拼团状态无法变更
	TicketPromoGroupStatusErr   = New(2000019) //拼团状态不对
	TicketPromoOrderTypeErr     = New(2000020) //订单类型不对
	PromoEditNotALlowed         = New(2000021) //不可编辑
	PromoEditFieldNotALlowed    = New(2000022) //不可编辑部分字段
	PromoExists                 = New(2000023) //拼团已存在
	PromoExtremeItemLost        = New(2000024) //项目id不存在

	//销售-交易
	TicketGetOidFail   = New(2000101) //获取订单号失败
	TicketExceedLimit  = New(2000102) //超过购买限制
	TicketParamMissed  = New(2000103) //信息不完整
	TicketSaleNotStart = New(2000104) //没开售
	TicketSaleEnd      = New(2000105) //已结束
	TicketNoPriv       = New(2000106) //无权操作
	TicketNotLogin     = New(2000107) //用户未登录
	TicketPriceChanged = New(2000108) //价格变化
	TicketUseCouponErr = New(2000109) //使用优惠券失败
	TicketNotSale      = New(2000110) //不可售
	TicketLevelLess    = New(2000111) //等级不够
	TicketVIPLess      = New(2000112) //会员身份不够

	//销售-库存
	TicketStockLack        = New(2000201) //库存不足
	TicketStockLogNotFound = New(2000202) //没有库存操作记录
	TicketStockUpdateFail  = New(2000203) //库存更新失败
	TicketStockNotFound    = New(2000204) //查询库存失败
	TicketItemSkuNotFound  = New(2000205) //查询项目SKU失败

	//番剧推荐
	SugEsSearchErr    = New(2002000) //es搜索错误
	SugSearchTypeErr  = New(2002001) //搜索类型错误
	SugOpTypeErr      = New(2002002) //操作类型错误
	SugOpErr          = New(2002003) //add or del match fail
	SugItemNone       = New(2002004) //商品不存在
	SugSeasonNone     = New(2002005) //番剧不存在
	SugSeasonNoPermit = New(2002006) //无权限操作

	//防刷工具
	ParamInvalid          = New(2001000) //参数错误
	UpdateError           = New(2001002) //更新失败
	QusbNotFound          = New(2001003) //找不到题库
	QusIDInvalid          = New(2001005) //题目id错误
	BankUsing             = New(2001007) //题目正在使用
	BindBankNotFound      = New(2001009) //未找到题库绑定关系
	AnswerError           = New(2001010) //答案错误
	GetQusBankInfoCache   = New(2001011) //获取题库缓存失败
	GetComponentTimesErr  = New(2001012) //获取组件缓存失败
	SetComponentTimesErr  = New(2001013) //设置答题次数缓存失败
	SetComponentIDErr     = New(2001014) //设置组件缓存失败
	GetComponentIDErr     = New(2001015) //获取组件ID缓存失败
	SameCompentErr        = New(2001016) //相同组件
	GetQusIDsErr          = New(2001017) //获取题目失败
	AnswerPoiError        = New(2001018) //答案错误
	NotEnoughQuestion     = New(2001019) //部分题库不足3题，无法绑定，请绑定别的题库，或者修改题库
	AntiSalesTimeErr      = New(2001020) //售卖时间有错
	AntiIPChangeLimit     = New(2001021) //用户IP变更
	AntiLimitNumUpper     = New(2001022) //次数达到上限
	AntiCheckVoucherErr   = New(2001023) //用户凭证验证失败
	AntiValidateFailed    = New(2001024) //验证失败
	AntiGeetestCountUpper = New(2001025) //极验总数达到上线
	AntiCustomerErr       = New(2001026) //业务方错误
	AntiBlackErr          = New(2001027) //黑名单用户
	AntiItemDetailErr     = New(2001028) //项目详情出错
	AntiVerifyErr         = New(2001029) //验证出错

	//项目
	TicketCannotDelTk      = New(2004000) //无法删除票价
	TicketDelTkFailed      = New(2004001) //删除票价失败
	TicketLkTkNotFound     = New(2004002) //关联票种不存在
	TicketLkTkTypeNotFound = New(2004003) //关联票种类型不存在
	TicketLkScNotFound     = New(2004004) //关联场次不存在
	TicketCannotDelSc      = New(2004005) //无法删除场次
	TicketLkScTimeNotFound = New(2004006) //关联的场次时间不存在
	TicketPidIsEmpty       = New(2004007) //项目id为空
	TicketMainInfoTooLarge = New(2004008) //项目版本详情信息量过大
	TicketDelTkExFailed    = New(2004009) //删除票价额外信息失败
	TicketAddVersionFailed = New(2004010) //添加版本信息失败
	TicketAddVerExtFailed  = New(2004011) //添加版本详情失败
	TicketBannerIDEmpty    = New(2004012) //BannerID为空
	TicketVerCannotEdit    = New(2004013) //版本不可编辑
	TicketVerCannotReview  = New(2004014) //无法审核 非待审核版本
	TicketAddTagFailed     = New(2004015) //添加项目标签失败

	//权限 2005000～2005199
	AuthErr = New(2005000) //权限验证失败

	//策略 2100001~2100999
	ABMatchErr   = New(2100001) //数据平台ABTest接口返回错误
	ParamIllegal = New(2100002)

	//验证小工具 2005200～2005699
	SMSPrepareErr       = New(2005200) // 短信初始化失败
	SMSVoucherErr       = New(2005201) // 短信凭证错误
	SMSVerifyTypeErr    = New(2005202) // 短信认证类型错误
	SMSBusinessErr      = New(2005203) // 短信认证业务方错误
	SMSPhoneErr         = New(2005204) // 手机号错误
	SMSCallbackErr      = New(2005205) // 回调接口错误
	VoucherErr          = New(2005206) // 凭证解密错误
	SMSCodeOverdueErr   = New(2005207) // 短信验证码过期
	SMSVerifyErr        = New(2005208) // 短信验证失败
	SMSSendErr          = New(2005209) // 短信发送失败
	RecaptchaPrepareErr = New(2005251) //google初始化失败
	RecaptchaCheckErr   = New(2005252) //google验证失败
	GeetestPrepareErr   = New(2005271) //geetest初始化失败
	GeetestCheckErr     = New(2005272) //geetest验证失败

	//数据平台
	InvalidSkuIDErr  = New(2006001) // 无效的SKUID
	NotLoginErr      = New(2006002) // 需要登录
	QwLoginParamsErr = New(2006003) // 企微免登陆参数错误
	QwLoginCacheErr  = New(2006004) // 企微cache出错
	QwInvalidSignErr = New(2006005) // 无效的sign

	//端监控
	DuplicateSubEvent = New(2007001) // 重复的sub_event
	DuplicateGroup    = New(2007002) // 重复的group
	DuplicateLogConf  = New(2007003) // 重复的数据流配置
	DuplicateLogID    = New(2007004) // 重复的LogID配置
	StartLaterThanEnd = New(2007005) // 请求的开始时间晚于结束时间
	DetailNotEmpty    = New(2007006) // detail不能为空
)
