package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	ActivityNotExist                       = xecode.New(75001) // 活动不存在
	ActivityNotStart                       = xecode.New(75002) // 活动没有开始
	ActivityOverEnd                        = xecode.New(75003) // 活动已结束
	ActivityHaveGuess                      = xecode.New(75004) // 活动已竞猜
	ActivityNotEnoughCoin                  = xecode.New(75005) // 硬币不足
	ActivityOverCoin                       = xecode.New(75006) // 超额投币
	ActivityServerTimeout                  = xecode.New(75007) // 服务超时
	ActivityKeyNotExists                   = xecode.New(75008) // key不存在
	ActivityKeyBindAlready                 = xecode.New(75009) // key已绑定
	ActivityMidBindAlready                 = xecode.New(75010) // 用户已绑定
	ActivityNotBind                        = xecode.New(75011) // 用户未绑定
	ActivityIDNotExists                    = xecode.New(75012) // ID不存在
	ActivityAwardAlready                   = xecode.New(75013) // 奖品已兑换
	ActivityNoAward                        = xecode.New(75014) // 没有奖品
	ActivityNotOwner                       = xecode.New(75015) // 不是该point点Owner
	ActivityHasUnlock                      = xecode.New(75016) // 该point点已解锁
	ActivityGameResult                     = xecode.New(75017) // 请选择游戏结果
	ActivityUserAchieveFail                = xecode.New(75018) // 用户成就记录获取失败
	ActivityUserPointFail                  = xecode.New(75019) // 用户point记录获取失败
	ActivityAchieveFail                    = xecode.New(75020) // 成就列表获取失败
	ActivityPointFail                      = xecode.New(75021) // point列表获取失败
	ActivityNoAchieve                      = xecode.New(75022) // 没有成就
	ActivityKeyFail                        = xecode.New(75023) // key获取失败
	ActivityMidFail                        = xecode.New(75024) // mid获取失败
	ActivityNotAdmin                       = xecode.New(75025) // 非管理员登录
	ActivityAddAchieveFail                 = xecode.New(75026) // 获得成就失败
	ActivityLackHp                         = xecode.New(75027) // 您的能量不足
	ActivityMaxHp                          = xecode.New(75028) // 您的HP已满
	ActivityNotAwardAdmin                  = xecode.New(75029) // 您不是奖品兑换管理员
	ActivityNotLotteryAdmin                = xecode.New(75030) // 您不是抽奖管理员
	ActivityLotteryFail                    = xecode.New(75031) // 未中奖
	ActivityUnlockFail                     = xecode.New(75032) // 解锁失败
	ActivityNotLotteryAchieve              = xecode.New(75033) // 该成就不支持抽奖
	ActivityLikeIPFrequence                = xecode.New(75034) //点赞活动ip访问过于频繁-访问过快
	ActivityLikeScoreLower                 = xecode.New(75035) //账户score过低不支持点赞-异常账号!
	ActivityLikeLevelLimit                 = xecode.New(75036) //-用户等级不够!
	ActivityLikeMemberLimit                = xecode.New(75037) //-用户注册不足7天！
	ActivityLikeNotStart                   = xecode.New(75038) //-评分未开始!
	ActivityLikeHasEnd                     = xecode.New(75039) //-评分已结束!
	ActivityLikeHasLike                    = xecode.New(75040) //-已点赞过!
	ActivityLikeHasGrade                   = xecode.New(75041) //-已评过分!
	ActivityLikeRegisterLimit              = xecode.New(75042) //-晚于活动限制注册时间!
	ActivityLikeHasVote                    = xecode.New(75043) //-已投过票!
	ActivityHasOffLine                     = xecode.New(75044) //-活动已经下线
	ActivityLikeHasOffLine                 = xecode.New(75045) //-活动稿件已经下线
	ActivityLikeBeforeRegister             = xecode.New(75046) //-早于活动限制注册时间!
	ActivityHasMissionGroup                = xecode.New(75047) //-已发起过活动!
	ActivityMGNotYourself                  = xecode.New(75048) //-不支持给自己助力哦!
	ActivityMissionNotStart                = xecode.New(75049) //-助攻未开始!
	ActivityMissionHasEnd                  = xecode.New(75050) //-助攻已结束!
	ActivityHasMission                     = xecode.New(75051) //-已助攻过!
	ActivityOverMissionLimit               = xecode.New(75052) //-超出可助攻上限!
	ActivityHasAward                       = xecode.New(75053) //-重复领取
	ActivityNotAward                       = xecode.New(75054) //-非法领取
	ActivityTelValid                       = xecode.New(75055) //-未绑定有效手机号码
	ActivityOverDailyScore                 = xecode.New(75056) //-超过单日投票上线
	ActivityBnjTimeCancel                  = xecode.New(75057) // 倒计时活动取消
	ActivityBnjResetCD                     = xecode.New(75058) // 倒计时重置CD
	ActivityBnjTimeFinish                  = xecode.New(75059) // 倒计时已完成
	ActivityBnjNotSub                      = xecode.New(75060) // 未预约拜年祭活动
	ActivityBnjSubLow                      = xecode.New(75061) // 该宝箱预约人数未达成
	ActivityBnjHasReward                   = xecode.New(75062) // 该宝箱奖励已被领取
	ActivityBnjRewardFail                  = xecode.New(75063) // 宝箱奖励领取失败
	ActivityRewardConfErr                  = xecode.New(75064) // 宝箱奖励配置错误
	ActivityMemberBlocked                  = xecode.New(75065) //-封禁用户无法操作
	ActivityKfcHasUsed                     = xecode.New(75066) //-code已经被使用
	ActivityKfcNotExist                    = xecode.New(75067) //-code不存在
	ActivityKfcNotGiveOut                  = xecode.New(75068) //-code未发放
	ActivityKfcSqlError                    = xecode.New(75069) //-发生未知错误
	ActivityUpScoreLower                   = xecode.New(75070) //投稿侧--账户score过低-异常账号!
	ActivityUpRegisterLimit                = xecode.New(75071) //投稿侧--晚于活动限制注册时间!
	ActivityUpBeforeRegister               = xecode.New(75072) //投稿侧-早于活动限制注册时间!
	ActivityUpLevelLimit                   = xecode.New(75073) //投稿侧--用户等级不够!
	ActivityUpFanLimit                     = xecode.New(75074) //投稿侧--粉丝数不符合条件!
	ActivityUpVipLimit                     = xecode.New(75075) //投稿侧--不是大会员!
	ActivityUpYearVipLimit                 = xecode.New(75076) //投稿侧--不是年度大会员!
	ActivityRepeatSubmit                   = xecode.New(75077) //投稿侧--重复参加活动!
	ActivityBodyTooLarge                   = xecode.New(75078) //-上传的文件太大
	ActivityTaskNotStart                   = xecode.New(75079) //任务未开始
	ActivityTaskOverEnd                    = xecode.New(75080) //任务已结束
	ActivityTaskHasLed                     = xecode.New(75081) //任务已领取
	ActivityTaskNotLed                     = xecode.New(75082) //任务未领取
	ActivityTaskHasFinish                  = xecode.New(75083) //任务已完成
	ActivityTaskNotFinish                  = xecode.New(75084) //任务未完成
	ActivityTaskNoAward                    = xecode.New(75085) //任务未完成
	ActivityTaskHadAward                   = xecode.New(75086) //任务奖励已领取
	ActivityCurrLackAmount                 = xecode.New(75087) //货币不够
	ActivityNotJoin                        = xecode.New(75088) //活动未参加
	ActivityOverLikeLimit                  = xecode.New(75089) //-票数已用完，无法投票
	ActivityQuestionCD                     = xecode.New(75090) //开启新答题冷却中
	ActivityQuestionNo                     = xecode.New(75091) //题池内容有误
	ActivityQuestionNotStart               = xecode.New(75092) //题池未开启
	ActivityQuestionFinish                 = xecode.New(75093) //题池已完成
	NativePageOffline                      = xecode.New(75094) //-活动已经下线
	ActivityOverLotteryMax                 = xecode.New(75095) //抽奖次数已用完，无法抽奖
	ActivityUnlocked                       = xecode.New(75096) // 材料未解锁
	ActivityNotReceived                    = xecode.New(75097) // 材料未领取
	ActivityHasReceived                    = xecode.New(75098) // 材料已领取
	ActivityRedDotClearFail                = xecode.New(75099) // 材料红点清除失败
	ActivityGameFinish                     = xecode.New(75100) // 游戏已结束
	ActivityReserveCancelForbidden         = xecode.New(75101) // 活动不能取消预约
	ActivityQuestionLimit                  = xecode.New(75111) // 您今日的答题机会已用完
	ActGuessNotExist                       = xecode.New(75200) // 竞猜不存在
	ActGuessOverEnd                        = xecode.New(75201) // 竞猜已结束
	ActGuessFail                           = xecode.New(75202) // 竞猜失败
	ActGuessDisabled                       = xecode.New(75203) // 您无法参与竞猜哦~
	ActGuessOverMax                        = xecode.New(75204) // 竞猜超出上限
	ActGuessesFail                         = xecode.New(75205) // 竞猜列表出错
	ActGuessDataFail                       = xecode.New(75206) // 用户数据出错
	ActUserGuessAlready                    = xecode.New(75207) // 您已经投注了哦~
	ActGuessResFail                        = xecode.New(75208) // 设置竞猜结果失败
	ActGuessDelFail                        = xecode.New(75209) // 竞猜已结算,不能删除
	ActGuessCoinFail                       = xecode.New(75210) // 账户可用余额不足
	ActGuessAddMax                         = xecode.New(75211) // 添加竞猜组超限
	ActivitySignNotOpen                    = xecode.New(75112) // 签到冷却中
	ActivitySignNotEnough                  = xecode.New(75113) // 签到能量不足
	ActivityFrequence                      = xecode.New(75114) // 操作过快
	ActivityFileTypeFail                   = xecode.New(75115) //-文件格式不正确
	ActivitySuitsFail                      = xecode.New(75116) // 领取挂件失败
	ActivitySubjectTagDup                  = xecode.New(75117) // tag同步异常
	ActivityPollNotExist                   = xecode.New(75300) // 投票话题不存在
	ActivityLackOfPollVote                 = xecode.New(75301) // 投票选择不足
	ActivityPollOptionNotExist             = xecode.New(75302) // 投票选项不存在
	ActivityPollVoteInvalid                = xecode.New(75303) // 投票不合法
	ActivityPollAlreadyVoted               = xecode.New(75304) // 已经投过票了
	ActivityPollExceededDailyChance        = xecode.New(75305) // 到达每日投票上限
	ActivityPollEnd                        = xecode.New(75306) // 投票已经结束
	ActivityLikeNotOwner                   = xecode.New(75307) // 不是该作品的作者
	ActivityTaskPreNotCheck                = xecode.New(75308) // 任务前置条件未达成
	ActivityTaskAwardFailed                = xecode.New(75309) // 任务奖励领取失败
	ActivityNoIdentification               = xecode.New(75310) // 账号未经过实名认证
	ActivityIsINOrCheck                    = xecode.New(75311) // 有其余活动正在审核或进行中
	ActivityTelNotPassCheck                = xecode.New(75312) // 手机号未通过新账号认证
	ActivityLotteryTimesNotEnough          = xecode.New(75313) // 还不能增加抽奖次数
	ActivityLotteryIPFrequence             = xecode.New(75400) // 活动ip访问过于频繁
	ActivityLotteryRegisterEarlyLimit      = xecode.New(75401) // 抽奖活动晚于活动限制的注册时间!
	ActivityLotteryRegisterLastLimit       = xecode.New(75402) // 抽奖活动早于活动限制的注册时间!
	ActivityLotteryLevelLimit              = xecode.New(75403) // 抽奖活动等级不满足限制
	ActivityIdentificationValid            = xecode.New(75404) // 用户账号未实名认证
	ActivityLotteryAddTimesLimit           = xecode.New(75405) // 获得的抽奖次数已达到上限
	ActivityNotConfig                      = xecode.New(75406) // 活动未配置
	ActivityAddrHasAdd                     = xecode.New(75407) // 活动地址已添加
	ActivityAddrAddFail                    = xecode.New(75408) // 会员购地址添加失败
	ActivityLotteryTimesFail               = xecode.New(75409) // 抽奖次数扣除失败
	ActivityIPIllegal                      = xecode.New(75410) // 非法IP
	ActivityNotVip                         = xecode.New(75411) // 不是vip不能参与抽奖
	ActivityNotMonthVip                    = xecode.New(75412) // 不是月度vip不能参与抽奖
	ActivityNotYearVip                     = xecode.New(75413) // 不是年度vip不能参与抽奖
	ActivityCoinNotEnough                  = xecode.New(75414) // 硬币不足够消耗
	ActivityNoTimes                        = xecode.New(75415) // 没有抽奖次数
	ActivityAddrNotAdd                     = xecode.New(75416) // 抽奖地址未添加
	ActivityInLottery                      = xecode.New(75417) // 抽奖中
	ActivityLotteryUserUnusual             = xecode.New(75418) // 抽奖账号异常
	ActivityLotteryErr                     = xecode.New(75419) // 抽奖出错
	ActivityHadLottery                     = xecode.New(75420) // 已参加过抽奖
	ActivityAppstoreModelNameValid         = xecode.New(75500) // 机型不符合
	ActivityAppstoreIsReceived             = xecode.New(75501) // 已经领过奖励
	ActivityDecideRiskErr                  = xecode.New(75502) // 风控拦截
	ActivityAppstoreVipBatchNotEnoughErr   = xecode.New(75503) // 奖励已领完
	ActivityAppstoreEnd                    = xecode.New(75504) // 活动已结束
	ActivityAppstoreNotStart               = xecode.New(75505) // 活动未开始
	ActivityTokenError                     = xecode.New(75600) // token错误
	ActivityGoodsNoFind                    = xecode.New(75601) // 商品不存在
	ActivityAddrNotNeed                    = xecode.New(75602) // 无需填写地址
	ActivityRapid                          = xecode.New(75603) // 操作过快
	ActivitySelected                       = xecode.New(75604) // 奖品已选择
	ActivityCashed                         = xecode.New(75605) // 奖品已兑换
	ActivityInsufficient                   = xecode.New(75606) // 积分不足
	ActivityShortage                       = xecode.New(75607) // 奖品不足
	ActivityVogueNotAward                  = xecode.New(75608) // 您的账号异常，暂无法兑换，详情请咨询客服~
	ActivityVogueTelValid                  = xecode.New(75609) // 提交失败，请完成手机号绑定~
	ActivityVogueBlocked                   = xecode.New(75610) // 提交失败，封禁用户无法参与活动
	ActivityNtUserLimit                    = xecode.New(75611) // up发起活动：用户不合法
	ActivityNtNoBind                       = xecode.New(75612) //暂不支持编辑
	ActivityReserveFirst                   = xecode.New(75613) // 用户未参与活动
	TaskCanNotFinish                       = xecode.New(75614) // 无法完成任务
	BwsOnlinePrintPieceNumErr              = xecode.New(75620) // 解锁需碎片数量错误
	BwsOnlinePrintPieceLow                 = xecode.New(75621) // %s碎片数量不够
	BwsOnlinePrintHadUnlock                = xecode.New(75622) // 图鉴已解锁
	BwsOnlineDressNotHave                  = xecode.New(75623) // 装扮%s未获得
	BwsOnlineDressNotExist                 = xecode.New(75624) // 装扮%d不存在
	BwsOnlineDressPosRepeat                = xecode.New(75625) // 装备部位重复
	BwsOnlineCoinLow                       = xecode.New(75626) // 乐园币不够
	BwsOnlineAwardAll                      = xecode.New(75627) // 已获取礼包所有奖励
	BwsOnlinePackageNoAward                = xecode.New(75628) // 礼包内无奖品
	BwsOnlineNotReward                     = xecode.New(75629) // 未拥有该奖品
	BwsOnlineAwardUsed                     = xecode.New(75630) // 该奖品已使用
	BwsOnlineEnergyLow                     = xecode.New(75631) // 能量不够
	BwsOnlineTimeUsed                      = xecode.New(75632) // 次数已耗尽
	BwsOnlinePrintUnlockFail               = xecode.New(75633) // 图鉴解锁失败
	BwsOnlineNotRewardType                 = xecode.New(75634) // 不支持的线上兑奖类型
	BwsOnlineTicketServerErr               = xecode.New(75635) // 请求票务服务异常
	BwsOnlineTicketInfoNotMatch            = xecode.New(75636) // 票务身份信息校验不通过
	BwsOnlineInterReserveFailed            = xecode.New(75637) // 互动预约失败
	BwsOnlineNotBindTicket                 = xecode.New(75638) // 需先绑定门票信息，才可参与哔哩乐园活动
	BwsOnlineIdRepeateBind                 = xecode.New(75639) // 证件号已经被绑定
	BwsOnlineMidHasBind                    = xecode.New(75642) // 当前账号已经被绑定
	BwsOnlineTicketInfoNotFind             = xecode.New(75643) // 当前证件下，未查询到购票信息
	ActivityBrandMidErr                    = xecode.New(75701) // 获取用户信息失败
	ActivityBrandAwardOnceErr              = xecode.New(75702) // 只能领取一次哦
	ActivityBrandRiskErr                   = xecode.New(75703) // 该手机号之前输过了哦
	ActivityBrandCouponErr                 = xecode.New(75704) // 发送优惠券失败
	ActivityQPSLimitErr                    = xecode.New(75705) // 服务开小差了哦
	ActivityReserveOfNoCancel              = xecode.New(75706) // 您已经预约了～
	ActivityWriteHandArchiveErr            = xecode.New(75710) // 手书活动稿件信息获取失败
	ActivityWriteHandAwardErr              = xecode.New(75711) // 手书活动获奖情况存储失败
	ActivityWriteHandRankErr               = xecode.New(75712) // 手书活动排名存储失败
	ActivityWriteHandFansErr               = xecode.New(75713) // 查询粉丝数失败
	ActivityWriteHandMemberErr             = xecode.New(75714) // 查询用户信息失败
	ActivityWriteHandActivityMemberErr     = xecode.New(75715) // 手书活动参与人数获取失败
	ActivityWriteHandMemberInfoErr         = xecode.New(75716) // 获取用户信息失败
	ActivityWriteHandAddtimesTooFastErr    = xecode.New(75717) // 增加获奖次数太频繁
	ActivityWriteHandTelValid              = xecode.New(75718) // 提交失败，请完成手机号绑定~
	ActivityWriteHandBlocked               = xecode.New(75719) // 提交失败，封禁用户无法参与活动
	ActivityWriteHandGetCoinErr            = xecode.New(75720) // 获取投币数据有误，服务开小差了哦
	ActivityWriteHandCoinNotEnoughErr      = xecode.New(75721) // 投币数不够哦
	ActivityRemixMidInfoErr                = xecode.New(75722) // 获取鬼畜用户信息失败
	ActivityRemixArchiveInfoErr            = xecode.New(75723) // 稿件信息获取失败
	ActivityRemixSidErr                    = xecode.New(75724) // sid有误
	ActivityRemixCountErr                  = xecode.New(75725) // 获取人数有误
	ActivityStarBeforeErr                  = xecode.New(75726) // 很抱歉，之前您已经成功受邀，本次邀请失败~
	ActivityStarAlreadyErr                 = xecode.New(75727) // 恭喜您已成功受邀，请继续你的UP主成长任务吧！
	ActivityStarLimitErr                   = xecode.New(75728) // 对方邀请人数已达上限，进入活动首页直接入驻吧！
	ActivityStarSelfErr                    = xecode.New(75729) // 很抱歉，不能邀请自己！
	ActivityStarNotVErr                    = xecode.New(75737) // 很遗憾，只有未发布过稿件大v才能申请入驻~前往开启你的新星推荐官任务吧！
	BwsPageOverTime                        = xecode.New(75730) // 页面已过期
	BwsHasVote                             = xecode.New(75731) // 该场次已经投票
	BwsNoMainTask                          = xecode.New(75732) // 没配置主线任务
	BwsNoAward                             = xecode.New(75733) // 没在线奖品
	BwsAwardNoStock                        = xecode.New(75734) // 抽中奖励无库存
	BawAwardNeedCheck                      = xecode.New(75735) // 实物奖励需用户先确认
	ActivityAwardNotExpected               = xecode.New(75736) // 奖励无法对应
	ActivityLotteryGiftFoundNoType         = xecode.New(75738) // 没有找到相应的奖品类型
	ActivityLotteryRiskInfo                = xecode.New(75739) // 风险用户
	ActivityLotteryTimesTypeError          = xecode.New(75740) // 次数类型有误
	ActivityLotteryNotMonthVip             = xecode.New(75741) // 不是月度vip
	ActivityLotteryNotAnnualVip            = xecode.New(75742) // 不是年度vip
	ActivityLotteryVip                     = xecode.New(75743) // 是大会员
	ActivityLotteryNotNewVip               = xecode.New(75744) // 不是新大会员
	ActivityLotteryNotOldVip               = xecode.New(75745) // 不是老大会员
	ActivityLotteryMemberGroupVipTypeError = xecode.New(75746) // 次数类型有误
	ActivityLotteryMemberGroupStructError  = xecode.New(75747) // 用户组格式有误
	ActivityLotteryValidIP                 = xecode.New(75748) // 不是有效IP地址
	ActivityLotteryGiftParamsError         = xecode.New(75749) // gift params 参数错误
	ActivityLotteryMemberNotNewError       = xecode.New(75750) // 不是新用户
	ActivityLotteryMemberNotOldError       = xecode.New(75751) // 不是老用户
	ActivityLotteryIsInternalError         = xecode.New(75752) // 内部抽奖不能http调用

	ActivityLotteryMemberGroupNotReserveError    = xecode.New(75753) // 抽奖用户没预约
	ActivityLotteryMemberGroupNotCartoonNewError = xecode.New(75754) // 抽奖用户非漫画新用户

	ActivityLotteryMemberInfoError = xecode.New(75755) // 用户信息获取失败
	ActivityLotteryGiftNostoreErr  = xecode.New(75756) // 商品无库存
	ActivityLikeNotEnoughErr       = xecode.New(75757) // 点赞数不足
	ActivityLikeGetErr             = xecode.New(75758) // 点赞数获取失败
	ActivityArticleDayAlreadyErr   = xecode.New(75759) // 你已报名成功，无需重复报名（￣▽￣）

	ActivityPointDetailFetchFailed = xecode.New(75774) // 积分明细获取失败
	ActivityLotteryInfoGetFail     = xecode.New(75775) // 抽奖信息获取失败
	ActivityLotteryGiftGetFail     = xecode.New(75776) // 商品领取失败
	ActivityLotteryGiftReceived    = xecode.New(75777) // 商品已经领取过
	ActivityLotteryNotLucky        = xecode.New(75778) // 很遗憾，未中奖
	ActivityMatchStageEnd          = xecode.New(75779) // 赛事阶段结束
	ActivityExchangePointFail      = xecode.New(75780) // 活动太火爆
	ActivityMatchExchangedPoint    = xecode.New(75781) // 该阶段已经兑换过
	ActivityGoodsOverTimes         = xecode.New(75782) // 商品兑换超出次数限制
	ActivityGoodsEnd               = xecode.New(75783) // 商品兑换结束
	ActivityKeyExists              = xecode.New(75784) // key存在
	ActivityGoodsExpired           = xecode.New(75785) //商品领取时间已过期
	ActivityAlreadySigned          = xecode.New(75786) // 已经签过到了
	ActivityPointGetFail           = xecode.New(75787) // 积分获取失败
	ActivityTasksProgressGetFail   = xecode.New(75788) // 任务进度获取失败
	ActivityUerInBlackList         = xecode.New(75789) // 账号存在风险
	ActivityGoodsNotStart          = xecode.New(75790) // 该商品兑换未开始
	ActivityLotteryNotStart        = xecode.New(75791) // 开奖时间未到
	ActivityMatchStageNotStart     = xecode.New(75792) // 赛事阶段未开始

	ActivityGoodsNoStoreErr = xecode.New(75793) // 商品无库存
	ActivityGoodsNoExist    = xecode.New(75794) //商品不存在
	ActivitySidError        = xecode.New(75795) // 无效的活动id

	ActivityDataPackageFail      = xecode.New(75796) //免流包兑换失败
	ActivityDataPackageExchanged = xecode.New(75797) //已经兑换过免流包
	ActivityCallServiceFail      = xecode.New(75798) //商品出货失败，请联系客服

	ActivityFunnyLikeGetErr       = xecode.New(75810) // 点赞数获取失败
	ActivityFunnyLikeNotEnoughErr = xecode.New(75811) // 点赞次数不足10次
	ActivityFunnyTelValid         = xecode.New(75812) // 提交失败，请完成手机号绑定~
	ActivityFunnyBlocked          = xecode.New(75813) // 提交失败，封禁用户无法参与活动
	ActivityFunnyAddTimesLimit    = xecode.New(75814) // 每日增加抽奖次数达到上限
	ActivityFunnyAddTimesErr      = xecode.New(75815) // 每日增加抽奖次数请求失败

	ActivityDoRelationErr = xecode.New(75809) // doRelation失败

	ActivityLikeWithSidCvidParamErr = xecode.New(75816) // 参数校验失败
	ActivityLikeWithSidCvidNetErr   = xecode.New(75817) // 网络请求失败 请稍后再试
	ActivityLikeWithSidCvidDataErr  = xecode.New(75818) // 请输入正确CV号
	ActivityRelationIDNoExistErr    = xecode.New(75819) // 活动聚合平台ID不存在
	ActivityRelationParamsErr       = xecode.New(75823) // 参数缺失

	ActivityLotteryGiftErr   = xecode.New(75821) // 奖品信息有误
	RelationReserveCancelErr = xecode.New(75822) // relationReserve取消预约失败

	SystemGetWXAccessTokenErr    = xecode.New(75850) // token获取失败
	SystemGetWXUserIDFailed      = xecode.New(75851) // userID获取失败
	SystemGetDBUserInfoFailed    = xecode.New(75852) // userInfo信息获取失败
	SystemUpdateWXUserInfoFailed = xecode.New(75853) // userInfo更新失败
	SystemCreateWXUserInfoFailed = xecode.New(75854) // userInfo创建失败
	SystemGetWXJSAPITicket       = xecode.New(75855) // JSAPITicket获取失败
	SystemGetOAAccessTokenErr    = xecode.New(75856) // token获取失败
	SystemFromParamsErr          = xecode.New(75857) // 请求来源异常
	SystemNoUserErr              = xecode.New(75858) // 用户不存在
	SystemNetWorkBuzyErr         = xecode.New(75859) // 网络繁忙，请稍后再试
	SystemNoTokenErr             = xecode.New(75860) // 缺少token参数
	SystemNoUserInOAErr          = xecode.New(75861) // OA系统中用户不存在该用户，请联系管理员

	SystemNoActivityErr       = xecode.New(75862) // 活动不存在
	SystemActivityNotStartErr = xecode.New(75863) // 活动未开始
	SystemActivityIsEndErr    = xecode.New(75864) // 活动已结束

	SystemActivitySignedErr        = xecode.New(75865) // 您已经签到过
	SystemActivityVotedErr         = xecode.New(75866) // 您已经投票过
	SystemActivitySignErr          = xecode.New(75867) // 签到失败
	SystemActivityConfigErr        = xecode.New(75868) // 配置信息错误
	SystemNotIn2021PartyMembersErr = xecode.New(75869) // 不在年会名单中，请联系管理员
	SystemActivityParamsErr        = xecode.New(75870) // 参数不正确
	SystemActivityVoteErr          = xecode.New(75871) // 投票失败

	CreateUpActReserveCancelErr                   = xecode.New(75875) // 预约审核中不允许撤销
	CreateUpActReserveExistErr                    = xecode.New(75876) // 流程中存在其他预约
	CreateUpActReserveNotInWhiteListErr           = xecode.New(75877) // 抱歉，不在白名单内
	CreateUpActReserveTitleEmptyErr               = xecode.New(75878) // 标题不能为空
	CreateUpActReserveTitleIllegalErr             = xecode.New(75879) // 标题内容含有敏感词汇
	CreateUpActReserveStimeIllegalErr             = xecode.New(75880) // 预约活动开始时间不合法
	CreateUpActReserveTimeIllegalErr              = xecode.New(75881) // 预约活动开始时间不能晚于结束时间
	CreateUpActReserveLiveTimeIllegalErr1         = xecode.New(75882) // 发起直播预约开播时间不可早于当前时间后5分钟
	CreateUpActReserveLiveTimeIllegalErr2         = xecode.New(75883) // 发起直播预约开播时间最长不能超过6个月
	CreateUpActReserveLiveTimeIllegalErr3         = xecode.New(75884) // 直播预约开播时间不可为空
	CreateUpActReserveVerification4CancelStateErr = xecode.New(75890) // 可以被忽略的核销错误
	CreateUpActReserveExistDynamicID              = xecode.New(75891) // 动态回调已经绑定过动态id了（为了动态做幂等）

	ActivityAnswerHpOver = xecode.New(75910) //HP用户用完
	ActivityAlreadyShare = xecode.New(75911) //HP已分享
	ActivityNotPendant   = xecode.New(75912) //没有答对100题
	ActivityAnswerRepeat = xecode.New(75913) //不可以重复答题

	ActivityLotteryNetWorkError   = xecode.New(75920) // grpc调用失败
	ActivityLotteryNoPayError     = xecode.New(75921) // 暂无资格
	ActivityLotteryPayJoinedError = xecode.New(75922) // 已参与
	ActivitySelectionAddErr       = xecode.New(75923) // 年度动画评选提交失败
	ActivitySelectionJoinErr      = xecode.New(75924) // 账号注册时间不满足要求！
	ActivitySelectionOneErr       = xecode.New(75925) // 请至少填写一项
	ActivityUpFollowErr           = xecode.New(75927) // 没有关注UP主

	ActivityVoteRepeatErr       = xecode.New(75926) // 当前维度已投票
	ActivityVoteRiskErr         = xecode.New(75928) // 账号异常
	ActivityVoteCheckErr        = xecode.New(75929) // 页面维护中，预计03:00前更新完成，请耐心等待，感谢您的理解和支持~
	ActivityLotteryOrderNoErr   = xecode.New(75930) // orderno中不能包含@字符
	ActivityLotteryRiskErr      = xecode.New(75931) // 依赖方错误
	ActivityLotteryDuplicateErr = xecode.New(75932) // 重复抽奖

	ActivityTunnelGroupErr        = xecode.New(75936) // 预约人群包出错
	ActivityGetAccRelationGRPCErr = xecode.New(75820) // GRPC请求失败
	ActivityRiskOverseaErr        = xecode.New(75933) // 风控海外账号异常
	ActivityRiskTelErr            = xecode.New(75934) // 风控账号手机异常
	ActivityRiskRejectErr         = xecode.New(75935) // 风控账号异常

	ActivityOrderNoFindErr          = xecode.New(75937) // 未找到相应抽奖记录
	ActivityBwsMidErr               = xecode.New(75940) // 获取用户信息失败
	ActivityBwsHeartErr             = xecode.New(75941) // 体力值不够
	ActivityBwsGameErr              = xecode.New(75942) // 游戏未配置
	ActivityBwsStockErr             = xecode.New(75943) // 未配置实物奖品
	ActivityBwsVipKeyErr            = xecode.New(75944) // 未找到vipKey
	ActivityBwsVipKeyAlreadyBindErr = xecode.New(75945) // vipKeY已绑定
	ActivityBwsVipMidAlreadyBindErr = xecode.New(75946) // vip用户已经绑定
	ActivityBwsDuplicateErr         = xecode.New(75947) // 重复添加

	ActivityTaskNotExist = xecode.New(75950) //任务不存在~

	BNJNoEnoughCommitTimes  = xecode.New(75960) // 拜年纪AR当天提交次数已到达上限
	BNJInvalidRequestID     = xecode.New(75961) // 不合法的请求标识
	BNJTooManyUser          = xecode.New(75962) // 当前参与人数太多啦，请稍后再试哦～
	BNJExamInvalidCommit    = xecode.New(75963) // 请在规定时间内及有效题库中答题哦～
	BNJUserNotPaid          = xecode.New(75964) //您还不是付费用户哦~
	BNJNoEnoughCoupon2Draw  = xecode.New(75965) // 您的抽奖次数不够了哦～
	BNJExamDuplicatedCommit = xecode.New(75966) // 请不要重复答题哦～
	BNJLiveDrawNotStart     = xecode.New(75967) // 抽奖将于19：30开始～
	BNJDrawNothing          = xecode.New(75968) // 抽到了空气，不要灰心哦～
	BNJLiveDrawEnd          = xecode.New(75969) // 抽奖已结束～

	RewardsAwardAlreadySent                    = xecode.New(75971) //活动奖励已发放~
	RewardsAwardSendFail                       = xecode.New(75972) //活动奖励发放失败~
	ActivityWinterAlreadyErr                   = xecode.New(75973) // 重复添加活动
	ActivityWinterNoPayErr                     = xecode.New(75974) // 付费课程不正确
	ActivityWinterJoinErr                      = xecode.New(75975) // 服务出错，请重试
	ActivityLotteryMemberGroupMemberLevelError = xecode.New(75976) // 用户等级不达标
	RankConfigError                            = xecode.New(75977) // 排行榜配置错误

	SpringFestivalCardsErr              = xecode.New(75978) // 卡片有误
	SpringFestivalComposeCardStoreErr   = xecode.New(75980) // 卡片库存不足
	SpringFestivalComposeCardErr        = xecode.New(75981) // 合成卡失败
	SpringFestivalJoinErr               = xecode.New(75982) // 加入失败
	SpringFestivalGetInviter            = xecode.New(75983) // 获取邀请者错误
	SpringFestivalInviterTokenErr       = xecode.New(75984) // 无法获取到邀请者信息
	SpringFestivalInviterAlreadyBindErr = xecode.New(75985) // 已经接受过邀请
	SpringFestivalInviterAlreadyJoinErr = xecode.New(75986) // 已经加入活动
	SpringFestivalNotJoinErr            = xecode.New(75987) // 没有加入活动
	SpringFestivalCanInviteSelfErr      = xecode.New(75988) // 不能邀请自己
	SpringFestivalCardStoreErr          = xecode.New(75989) // 卡库存不足
	SpringFestivalSendCardErr           = xecode.New(75990) // 分享卡失败
	SpringFestivalCardAlreadyErr        = xecode.New(75991) // 卡片已经被领取
	SpringFestivalGetCardErr            = xecode.New(75992) // 领取卡失败
	SpringFestivalCantGetCardErr        = xecode.New(75993) // 不能领取自己的卡
	SpringFestivalCantGetTokenMidErr    = xecode.New(75994) // 获取用户信息失败
	SpringFestivalTaskErr               = xecode.New(75995) // 任务无法完成
	SpringFestivalGetCardMaxErr         = xecode.New(75996) // 已达上限，不可领取
	SpringFestivalRiskMemberErr         = xecode.New(75997) // 风控用户，无法操作
	SpringFestivalAlreadyDonatedErr     = xecode.New(75998) // 卡片已经被领取
	SpringFestivalTooFastErr            = xecode.New(75999) // 操作频繁
	ActivityNativePageError             = xecode.New(75118) // 请记得创建同名native话题页
	ActivityDomainAddError              = xecode.New(75119) // 添加活动域名失败
	ActivityDomainConflictError         = xecode.New(75120) // 当前自定义域名已存在

	ActivityVoteRankExpired                 = xecode.New(75121) //该榜单已过期, 请刷新页面查看最新榜单
	ActivityVoteNotFound                    = xecode.New(75122) //该投票活动不存在
	ActivityVoteItemNotFound                = xecode.New(75123) //投票选项不存在
	ActivityVoteOverLimit                   = xecode.New(75124) //投票次数不足
	ActivityVoteNotStarted                  = xecode.New(75125) //投票活动未开始
	ActivityVoteFinished                    = xecode.New(75126) //投票活动已结束
	ActivityVoteItemVoted                   = xecode.New(75127) //该选项已投过票~
	ActivityVoteError                       = xecode.New(75128) //投票失败
	ActivityVoteExceed                      = xecode.New(75129) //当日投票总次数超过上限
	ActivityVoteSourceTypeUnknown           = xecode.New(75130) //投票数据源类型未知
	ActivityVoteDSGNotFound                 = xecode.New(75131) //该投票活动数据组不存在
	ActivityVoteRuleNotConfig               = xecode.New(75132) //该活动投票规则未配置
	ActivityVoteNoHistory                   = xecode.New(75133) //取消投票失败, 当日未找到对应投票记录
	ActivityVoteRefreshAndRenewDSGItemsFail = xecode.New(75134) //投票数据源刷新失败, 并且延长当前版本有效期失败. 请立即处理
	ActivityVoteRefreshDSGItemsFail         = xecode.New(75135) //投票数据源刷新失败, 已延长当前版本缓存有效期. 请处理
	ActivityVoteRuleConfigError             = xecode.New(75136) //投票开始/结束时间配置错误

	FitActivityUserNotJoin      = xecode.New(75137) //健身打卡用户未参与
	GetPlanListErr              = xecode.New(75144) //获取系列计划列表出错
	GetPlanDetailErr            = xecode.New(75139) //获取计划详情出错
	StringToInt64Err            = xecode.New(75140) //类型转换出错
	GetHotVideosByTagErr        = xecode.New(75141) //获取热门标签视频列表出错
	PlanCardVideosEmpty         = xecode.New(75142) //系列计划里的播单视频为空
	TianMaCardNotReady          = xecode.New(75143) //天马卡新建后状态未ready
	BwsNotReserveError          = xecode.New(75146) // 未预约
	BwsNotReserveDuplicateError = xecode.New(75147) // 重复核销

	ActivityVoteRejectErr           = xecode.New(75150) // 投票风控账号异常
	MissionActivityNotStartError    = xecode.New(75151) // 活动尚未开始，无法领取奖品
	MissionActivityEndError         = xecode.New(75152) // 活动已结束，无法领取奖品
	MissionActivityReceiveTimeError = xecode.New(75153) // 不在奖品领取时间内，无法领取奖品~
	MissionActivityNoStockError     = xecode.New(75154) // 来晚了，奖品已被领完~

	SummerCampUserCourseInfoErr = xecode.New(75155) //暑期夏令营获取用户课程信息db出错
	SCUserNotJoinCourse         = xecode.New(75156) //用户未参与该课程
	SCcourseNotExit             = xecode.New(75157) //查找的课程播单找不到
	SCUserNoRight               = xecode.New(75158) //用户没有权限查看
	SCTaskErr                   = xecode.New(75159) //任务相关查询出错
	SCActivityIdErr             = xecode.New(75160) //非本次活动
	SCTaskSetErr                = xecode.New(75161) //任务写入出错
	SCStartCourseDBErr          = xecode.New(75162) //任务课程加入失败
	AwardStockErr               = xecode.New(75163) //库存不足
	UserHasExchanged            = xecode.New(75164) //用户今日已经兑换
	PointNotEnough              = xecode.New(75165) //积分不足
	ExchangeErr                 = xecode.New(75166) //兑换出错
	RewardListErr               = xecode.New(75167) //获取奖品列表出错
	PointHisListErr             = xecode.New(75168) //获取积分列表出错
	RiskUser                    = xecode.New(75169) //很抱歉您的账号目前存在风险，无法参与兑奖
	StockErr                    = xecode.New(75170) //库存调用出错
	StockNumErr                 = xecode.New(75171) //库存数量配置错误

	BwsNotReserveIndexError        = xecode.New(75148)  // 预约id不正确
	DynamicErrDynamicRemoved       = xecode.New(500404) // 动态服务
	BMLGuessNotRightError          = xecode.New(75249)  // 谜语不正确，请重试
	BMLGuessRepeateDrawError       = xecode.New(75250)  // 这个谜语已经被其他人猜到了
	BMLCommonGuessNotcompleteError = xecode.New(75251)  // 需先完成第一题谜语才能猜joker key
	BMLRepeateGuessError           = xecode.New(75252)  // 重复猜谜语
	BMLGuessRewardSendOutError     = xecode.New(75253)  // 发完了
	BMLGuessJokerKeyOverLimitError = xecode.New(75254)  // joker key 猜谜次数超过上限

	StockServerNoStockInCycleError   = xecode.New(75255) // 当前周期库存已经使用完
	StockServerUserStockUsedUpError  = xecode.New(75256) // 超出用户领取数量（当前周期）限制
	StockServerInvalidStockTimeError = xecode.New(75257) // 不在有效时间内
	StockServerConsumerFailedError   = xecode.New(75258) // 扣减库存失败
	StockServerConfInvalidError      = xecode.New(75259) // 库存配置不正确
	StockServerNoStockError          = xecode.New(75260) // 无剩余库存

	UpActReserveCreateDynamicTimeOut      = xecode.New(75261) // up主预约定时发布创建动态超时
	UpActReserveCreateDynamicReplyCodeErr = xecode.New(75262) // up主预约定时发布创建动态错误码非0

)