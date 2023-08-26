package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	// 稿件第三方推送服务 [153000,153999]
	// 基础错误 [153000, 153099]
	AVBVIDConvertingError  = xecode.New(153000)
	BatchNotFound          = xecode.New(153001)
	BatchDetailNotFound    = xecode.New(153002)
	VendorNotFound         = xecode.New(153003) // 未找到对应稿件推送厂商
	AuthorNotFound         = xecode.New(153004) // 未找到有效推送稿件作者
	AuthorPushNotFound     = xecode.New(153005) // 未找到有效作者推送
	NoValidAuthorsToPush   = xecode.New(153006) // 没有可推送的作者
	AuthorPushExisting     = xecode.New(153007) // 已有生效的对应厂商的作者推送
	BatchTodoAlreadyLocked = xecode.New(153008) // 待推送稿件检查锁已锁

	// 稿件业务校验错误 [153100, 153149]
	ArchiveCannotBeWithDrawn = xecode.New(153100) // 此状态稿件不允许被下架
	ArchiveCannotBePushed    = xecode.New(153101) // 此状态稿件不允许被上架

	// 作者绑定校验错误 [153150, 153199]
	VendorNotAbleToBindAuthor       = xecode.New(153150) // 推送厂商不支持绑定作者
	VendorNotAbleToSyncAuthorStatus = xecode.New(153151) // 推送厂商不支持作者绑定状态回流

	// 外部内容中台交互基础 [153200, 153229]
	PushRequestError    = xecode.New(153200) // 推送参数错误
	SyncRequestError    = xecode.New(153201) // 数据回流请求格式错误
	GetAccessTokenError = xecode.New(153202) // 获取Token达到最大重试次数
	// 腾讯游戏说CMC交互 [153230, 153239]
	QQCMCRequestError    = xecode.New(153230) // 调用游戏说接口失败
	QQCMCArchiveNotFound = xecode.New(153231) // 查询游戏说已推送稿件未找到对应稿件
	// 腾讯TGL交互 [153240, 153249]
	QQTGLRequestError = xecode.New(153240) // 调用TGL接口失败
	// 暴雪内容交互 [153250, 153259]
	BlizzardRequestError = xecode.New(153250) // 调用暴雪接口失败

	// 账户平台交互 [153300, 153319]
	AccountPlatRequestError  = xecode.New(153300) // 账户平台请求错误
	AccountPlatResponseError = xecode.New(153301) // 账户平台请求结果错误
)
