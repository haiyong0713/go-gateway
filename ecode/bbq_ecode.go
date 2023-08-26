package ecode

import . "go-common/library/ecode"

// bbq ecode interval is [5000000,6000000)
var (
	CheckInviteCodeErr = New(3001016) // 检查邀请码失败(特殊处理请勿修改) todo
	//Common [5000000,5001000)
	Common               = New(5000000)
	TypeDismatch         = New(5000001) // 类型不匹配
	ExternalErr          = New(5000002) // 外部错误
	ReqParamErr          = New(5000003) // 参数错误
	BBQSystemErr         = New(5000004) // 用于一些异常请求
	BBQNoBindPhone       = New(5000005) // 未绑定手机号
	BBQUserBanned        = New(5000006) // 已被封禁，无法进行相关操作，如有疑问可在“设置-吐槽”中进行反馈
	ArchiveDatabusNilErr = New(5000007) // 预发环境不配置稿件databus

	//Search [5001000,5002000)
	SearchCreateIndexErr = New(5001000) // 创建索引失败
	SearchVideoDataErr   = New(5001001) // 获取视频信息失败

	//搜索服务问题
	SearchWordIllegal = New(5001002) // 搜索关键词非法
	SearchError       = New(5001003) // 搜索失败

	//web [5002000,5003000)
	CommentClosed        = New(5002001) // 评论已关闭
	VideoUnExists        = New(5002003) // 视频不存在
	VideoUnReachable     = New(5002004) // 视频不存在，由于状态原因不可访问
	VideoInAudit         = New(5002005) // 视频审核中
	InviteCodeInvalid    = New(5002014) // 无效邀请码
	InviteCodeUsed       = New(5002015) // 邀请码已使用
	CommentForbidden     = New(5002021) // 禁止评论
	CommentTooShort      = New(5002023) // 评论过短
	CommentTooLong       = New(5002024) // 评论过长
	SvNotReachable       = New(5002025)
	NoticeTypeErr        = New(5002026) // 通知类型错误
	CommentForbidLike    = New(5002027) // 禁止赞或踩
	CommentLengthIllegal = New(5002028) // 评论长度不合法
	VideoCanNotChange    = New(5002029) // 视频信息不可更改, 通常由于状态原因

	//video-service [5003000,5004000)
	UnKnownBPS   = New(5003000) // 未知码率
	SyncBVCFail  = New(5003001) // 同步bvc转码失败
	VideoDelFail = New(5003002) // 视频删除失败，不能删除别人的视频

	// UserLike [5005000, 5005100]
	UserLike          = New(5005000) // UserLike [5005000, 5005100]
	AddUserLikeErr    = New(5005001) // 点赞失败
	CancelUserLikeErr = New(5005002) // 取消点赞失败

	// UserInfo [5005100, 5005200]
	UserInfo               = New(5005100) // UserInfo
	BatchUserTooLong       = New(5005101) // 用户批量请求太多
	UPMIDNotExists         = New(5005102) // up主不存在
	GetUserBaseErr         = New(5005103) // 获取用户信息失败
	EditUserBaseErr        = New(5005104) // 更新用户基础信息失败
	UserUnameSpecial       = New(5005105) // 昵称包含特殊字符
	UserUnameLength        = New(5005106) // 昵称长度不符合
	UserUnameExisted       = New(5005107) // 昵称已被占用
	UserUnameFilterErr     = New(5005108) // 昵称包含敏感词
	UserUnamePrefixErr     = New(5005109) // 该昵称无法注册
	UserSexErr             = New(5005110) // 性别只能为0,1,2
	UserBirthFmtErr        = New(5005111) // 出生日期格式错误，示例:19901010
	UserSignatureLengthErr = New(5005112) // 用户签名不能超过60字符
	UserSignatureFmtErr    = New(5005113) // 用户签名格式错误
	UserTeensAuthErr       = New(5005114) // 青少年模式身份验证校验错误
	UserTeensDisableErr    = New(5005115) // 青少年模式未开启
	UserTeensPasswdEqual   = New(5005116) // 新旧密码不能相同
	UserSignatureFilterErr = New(5005117) // 签名包含敏感词

	// UserRelation [5005200, 5005300]
	UserRelation              = New(5005200)
	AddUserFollowErr          = New(5005201) // 关注失败，请稍后重试
	CancelUserFollowErr       = New(5005202) // 取消关注失败，请稍后重试
	UserFollowLimitErr        = New(5005203) // 关注失败，关注已达上限
	FollowMyselfErr           = New(5005204) // 不能关注自己
	UserAlreadyBlackFollowErr = New(5005205) // 关注失败，请将用户移出黑名单后重试
	UserBlackLimitErr         = New(5005206) // 拉黑失败，黑名单已达上限
	UserBlackErr              = New(5005207) // 黑名单请求系统错误
	UserBlackSelfErr          = New(5005208) // 拉黑失败，不能拉黑自己

	// Danmu [5005300, 5005400]
	Danmu           = New(5005300)
	FilterErr       = New(5005301) // 弹幕包含敏感词
	DanmuGetErr     = New(5005302) // 弹幕获取失败
	DanmuPostErr    = New(5005303) // 弹幕发送失败
	DanmuLimitErr   = New(5005304) // 该视频暂时无法发送弹幕
	DanmuUpgradeErr = New(5005305) // 弹幕服务升级，暂时无法发送弹幕

	// Comment [5005400, 5005500]
	Comment                = New(5005400)
	CommentLengthErr       = New(5005401) // 评论需要1-96字
	CommentFilterErr       = New(5005402) // 评论包含敏感词
	CommentMissErr         = New(5005403) // 评论不见了
	CommentBlacklistFilter = New(5005404) // 已被该用户拉黑，无法回复
	CommentAddForbidden    = New(5005405) // 该视频无法评论
	CommentDelForbidden    = New(5005406) // 该评论已删除
	CommentBatchReqTooLong = New(5005407) // 评论批量请求太多
	CommentUpgradeLimit    = New(5005408) // 评论系统升级，暂时无法发送

	// report [5005500, 5005599]
	ReportDanmuError = New(5005501) // 弹幕举报失败

	//Upload [5005600, 5005700]
	Upload       = New(5005600)
	UploadFailed = New(5005601) //上传失败

	// Topic [5005700, 5005800]
	Topic                  = New(5005700)
	TopicReqParamErr       = New(5005701) // 参数错误
	TopicNumTooManyErr     = New(5005702) // 一次性插入db的话题数量太大
	TopicNameLenErr        = New(5005703) // 话题长度太长
	TopicIDErr             = New(5005704) // 话题ID错误
	TopicIDNotFound        = New(5005705) // 话题ID没找到
	TopicStateErr          = New(5005706) // 话题为下架状态
	TopicTooManyInOneVideo = New(5005707) // 一个视频的话题数量太多了
	TopicDescLenErr        = New(5005708) // 话题描述长度太长
	TopicInsertErr         = New(5005709) // 话题插入失败
	TopicVideoStateErr     = New(5005721) // 话题视频状态错误

	//Cms [5005800, 5005900]
	Cms                 = New(5005800)
	CmsDashBoardNoLogin = New(5005801)
	CmsReasonNumErr     = New(5005802) //无效理由值
	CmsInvalidOpt       = New(5005803) //无效操作
	CmsInvalidWorkflow  = New(5005804) //无效工作流
	CmsParamOOR         = New(5005805) //参数长度超出范围
	CmsLockFail         = New(5005806) //抢占分布式锁失败
	CmsExecTaskErr      = New(5005807)
	CmsQueryErr         = New(5005808) //查询外部接口异常
	CmsParmErr          = New(5005809) //参数错误
	CmsNotAuthorized    = New(5005810) //接口无权限
	CmsPushFailed       = New(5005811)
	CmsNoOperator       = New(5005812) //无效操作者

	//Creator [5005901, 5005999]
	CreatorSystem                = New(5005903) //common error, toast '系统繁忙'
	CreatorPublistDelXcodeFailed = New(5005901) //del xcode failed
	CreatorPublistList           = New(5005902) //list err

	//Producer [5006000,5006199]
	ProducerXcodeCallbackFailed = New(5006000) // xcode call back failed

	// Music [5006200, 5006299]
	Music                 = New(5006200)
	MusicUnExists         = New(5006201) // 音乐不存在
	MusicStateForbidden   = New(5006202) // 音乐状态不合法
	MusicPlayInfoErr      = New(5006203) // 音乐播放错误
	MusicCategoryErr      = New(5006204) // 音乐分类无效
	MusicBvcStoreErr      = New(5006205) // 音乐upos多音质存储失败
	MusicBvcPlayURLErr    = New(5006206) // 音乐获取playurl错误
	MusicCategoryExists   = New(5006207) // 音乐分类已存在
	MusicCategoryUnExists = New(5006208) // 该分类下没有这个音乐
	MusicPosChangeErr     = New(5006209) // 挪动音乐位置失败

	//Mode sensitive ecode [5006300, 5006349]
	//ModeSensitiveServiceNotAvailable service not available during sensitive mode
	ModeSensitiveServiceNotAvailable = New(5006300)

	// Favorite [5006400, 506499]
	UserFavorite          = New(5006400)
	AddUserFavoriteErr    = New(5006401) // 收藏失败
	CancelUserFavoriteErr = New(5006402) // 取消收藏失败
	FavoriteTypeErr       = New(5006403) // 收藏类型错误（1.6收藏类型只能为1即视频收藏）
)

// 可以取消点赞的状态
var svLikeCancelAvailableState = map[error]bool{
	VideoUnReachable: true,
	VideoInAudit:     true,
}

// IsCancelSvLikeAvailable 可以取消点赞的状态
func IsCancelSvLikeAvailable(err error) (available bool) {
	_, available = svLikeCancelAvailableState[err]
	return
}
