package ecode

import xecode "go-common/library/ecode"

var (
	ElecDenied           = xecode.New(62001) // 不需要展示充电信息
	ArchiveDenied        = xecode.New(62002) // 稿件不可见
	ArchivePass          = xecode.New(62003) // 稿件已审核通过，等待发布中
	ArchiveChecking      = xecode.New(62004) // 视频正在审核中，请耐心等待～
	ArchiveNotLogin      = xecode.New(62005) // 视频不见了？你可以试试登录！
	HelpListError        = xecode.New(62006) // 智齿列表结果错误
	HelpDetailError      = xecode.New(62007) // 智齿详情结果错误
	HelpSearchError      = xecode.New(62008) // 智齿搜索结果错误
	ArcAppealLimit       = xecode.New(62009) // 短时间内请勿重复投诉相同稿件
	CardNothingFound     = xecode.New(62010) // 没有新内容
	FeedbackBodyTooLarge = xecode.New(18002) // 上传的文件太大
	TagIsSealing         = xecode.New(16025) // tag已经被封印了~
	// 忘记密码申诉
	AppealExist       = xecode.New(181000) //您已提交过信息，请耐心等待处理结果
	AppealWrongMobile = xecode.New(181001) //请填写正确的手机号
	AppealUpgradeApp  = xecode.New(181002) //请升级APP版本
	// 验证码
	CaptchaIpLimit    = xecode.New(181010) //今日验证码发送次数已达上限
	CaptchaFrequently = xecode.New(181011) //操作过于频繁，请稍后再试
	CaptchaWrong      = xecode.New(181012) //验证码错误
	CaptchaInvalid    = xecode.New(181013) //验证码失效，请重新获取
)
