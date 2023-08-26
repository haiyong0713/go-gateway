package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	TabTimeLimit = xecode.New(78043) //线上已存在配置

	EntryOfflineBeforeDelete    = xecode.New(77901) // 删除入口前需要下线
	EntryTimeSettingError       = xecode.New(77902) // 时间设定错误
	EntryIsOffline              = xecode.New(77903) // 该入口尚未上线，请先上线
	EntryDeleteOnlineStateError = xecode.New(77904) // 线上状态不得删除
	EntryOnlineLimit            = xecode.New(77905) // 线上存在生效中的入口配置，有时间配置冲突
	EntryParamsError            = xecode.New(77906) // 入口配置传入参数错误

	// 天马业务弹窗 [77910-77919]
	PopupConfigParameterError     = xecode.New(77910) // 配置参数不正确
	PopupConfigConflictTime       = xecode.New(77911) // 该生效时段内已存在配置
	PopupConfigBuildsParsingError = xecode.New(77912) // 版本限制JSON格式错误
	PopupConfigNotFound           = xecode.New(77913) // 未找到对应弹窗配置

	// 人群包 [77920-77929]
	BGroupIDError     = xecode.New(77920) // 人群包ID不正确
	BGroupAddError    = xecode.New(77921) // 添加人群包失败（提示：配置描述不可重复）
	BGroupUpdateError = xecode.New(77922) // 更新人群包失败

	// 运营主题配置 [77930-77939]
	SkinConfigNotFound                    = xecode.New(77930) // 未找到对应的主题配置
	SkinResourceNotFound                  = xecode.New(77931) // 未找到对应的主题资源
	SkinOnlineConfigResourceNotModifiable = xecode.New(77932) // 上线后不可修改主题ID

	// 品专卡黑名单 [77940-77949]
	PageInvalid                = xecode.New(77940) // 分页参数错误
	BrandBlacklistQueryInvalid = xecode.New(77941) // 非法品专卡黑名单词
	BrandBlacklistQueryExists  = xecode.New(77942) // 已有生效中配置
	BrandBlacklistNotFound     = xecode.New(77943) // 未找到对应的品专卡黑名单

	// 品牌闪屏 [77950-77969]
	SplashScreenCategoryExists  = xecode.New(77950) // 同名分类已存在
	SplashScreenConfigNotExists = xecode.New(77951) // 对应配置不存在

	// resource卡片 [77970-77999]
	Navigation2ndEmpty   = xecode.New(77970) // 导航卡二级目录不能为空
	Navigation3rdEmpty   = xecode.New(77971) // 导航卡三级目录不能为空
	CardNotFound         = xecode.New(77972) // 卡片不存在
	Navigation2ndExceeds = xecode.New(77973) // 导航卡二级目录已达上限
	Navigation3rdExceeds = xecode.New(77974) // 导航卡三级目录已达上限
	InvalidResourceId    = xecode.New(77975) // 非法资源ID

	// 版头 [175000,175999]
	FrontPageConfigNotFound     = xecode.New(175000) // 对应资源位版头配置未找到
	FrontPageConfigDuplicated   = xecode.New(175001) // 同一个分区不能有相同时间段的版头
	FrontPageCacheError         = xecode.New(175002) // 获取版头缓存时错误
	FrontPageCacheSaveError     = xecode.New(175003) // 更新版头缓存时错误
	FrontPageLocationParseError = xecode.New(175004) // 判断IP区域策略错误
	FrontPageNotEditable        = xecode.New(175005) // 兜底配置不允许操作

	// 忘记密码申诉
	PwdAppealProcessed    = xecode.New(77980) //该条申诉已处理
	PwdAppealSendFail     = xecode.New(77981) //短信发送失败
	PwdAppealUpdateDBFail = xecode.New(77982) //数据更新失败
	PwdAppealSmsCfgError  = xecode.New(77983) //短信模板错误
	PwdAppealPendingExist = xecode.New(77984) //您已提交过信息，请耐心等待处理结果

	TagNotExist = xecode.New(16001) // tag不存在

	AppSelectedIDExist  = xecode.New(78020) // 热门精选ID已存在
	AppSelectedNotValid = xecode.New(78005) // 热门精选不可通过，前三位未有寄语

	MemberNotExist = xecode.New(40061) // 用户不存在
)
