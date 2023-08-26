package ecode

import xecode "go-common/library/ecode"

var (
	ChNameToLong         = xecode.New(53001) // 频道名字数超过限制啦
	ChIntroToLong        = xecode.New(53002) // 频道简介字数超过限制啦
	ChMaxArcCount        = xecode.New(53003) // 本频道里的视频已经满啦
	ChMaxCount           = xecode.New(53004) // 你创建的频道已经满额了哦
	ChFakeAid            = xecode.New(53005) // 频道内有失效视频了哦
	ChAidsExist          = xecode.New(53006) // 你提交的视频已失效或者频道里已经有了哦
	ChNameExist          = xecode.New(53007) // 频道名称已经存在了哦
	ChNoArcs             = xecode.New(53008) // 频道内没有视频
	ChNoArc              = xecode.New(53009) // 频道内没有该视频
	ChNameBanned         = xecode.New(53010) // 频道名称有敏感词，请重新编写
	ChIntroBanned        = xecode.New(53011) // 频道简介有敏感词，请重新编写
	SpaceNoShop          = xecode.New(53012) // 非营业中商户号
	SpaceNoPrivacy       = xecode.New(53013) // 用户隐私设置未公开
	SpaceFakeAid         = xecode.New(53014) // 该稿件已失效
	TopReasonLong        = xecode.New(53015) // 置顶理由字数超过限制啦
	SpaceNoTopArc        = xecode.New(53016) // 没有置顶视频
	SpaceNotAuthor       = xecode.New(53017) // 只能操作自己的稿件
	SpaceTextBanned      = xecode.New(53018) // 提交文本有敏感词
	SpaceMpMaxCount      = xecode.New(53019) // 代表作已达上限
	SpaceMpExist         = xecode.New(53020) // 代表作内已有该视频
	SpaceMpNoArc         = xecode.New(53021) // 代表作内没有该视频
	SpacePayUGV          = xecode.New(53022) // 不支持付费稿件
	SpaceBanUser         = xecode.New(53023) // 账号封禁中，修改失败
	SpaceBanFilter       = xecode.New(53024) // 编辑内容命中敏感词
	NotAllowedArc        = xecode.New(53025) // 不被允许的稿件
	CutCoverFail         = xecode.New(53026) // 封面裁剪失败
	ForbitTop            = xecode.New(53028) // 稿件状态变更，暂不支持置顶
	ForbitMasterpiece    = xecode.New(53029) // 稿件状态变更，暂不支持设置为代表作
	ChForbitModify       = xecode.New(53030) // 频道升级中，暂不支持修改
	ChArcStaff           = xecode.New(53031) // 联合投稿，主投稿人才能放入频道
	SysNoticeConflict    = xecode.New(53032) // PR公告冲突
	TopPhotoRequestError = xecode.New(53033) // 头图审核参数错误
	TopPhotoNoReason     = xecode.New(53034) // 头图审核缺少理由
	TopPhotoNotFound     = xecode.New(53035) // 未找到对应头图记录
	SpaceTPPicLarge      = xecode.New(53036) // 图片过大
	SpaceVIPError        = xecode.New(53037) // vip状态异常
	SpaceTPPicError      = xecode.New(53038) // 仅支持png、jpg格式的图片上传
	SpaceTPMallError     = xecode.New(53039) // 头图不存在
	SpaceSetTPError      = xecode.New(53040) // 所要设置的头图已被下架
	SpacePayTPError      = xecode.New(53041) // 想用必须先买哦
	SpaceSizError        = xecode.New(53042) // 图片宽度至少为1280，高度至少为200
	SpaceBase64Error     = xecode.New(53043) // base64解析失败
	PhotoMallEmptyError  = xecode.New(53044) // 商城正在装修中
)
