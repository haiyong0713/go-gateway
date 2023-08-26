package account

import (
	"fmt"
	"strconv"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	shuangqing "git.bilibili.co/bapis/bapis-go/datacenter/shuangqing"
	passportuser "git.bilibili.co/bapis/bapis-go/passport/service/user"
)

const (
	// realname
	RealnameNotVerified = -1 //未实名
	RealnameVerified    = 1  //已实名
)

type StatisticsContent struct {
	Type         string `json:"type"`
	Value        string `json:"value"`
	ValueIOS     string `json:"value_ios,omitempty"`
	ValueAndroid string `json:"value_android,omitempty"`
}

type Statistics struct {
	Index   int               `json:"index"`
	Type    string            `json:"type"`
	Name    string            `json:"name"`
	Usecase string            `json:"usecase"`
	Scene   string            `json:"scene"`
	Status  string            `json:"status"`
	Count   string            `json:"count"`
	Content StatisticsContent `json:"content"`
}

type ExportedStatistics struct {
	Date       int64              `json:"date"`
	Statistics []*Statistics      `json:"statistics"`
	Configs    *StatisticsConfigs `json:"configs"`
}

type StatisticsConfigs struct {
	Info []*StatisticsInfo `json:"info"`
}

type StatisticsInfo struct {
	Type string  `json:"type"`
	Ids  []int64 `json:"ids"`
}

func ResolveShuangQiongStats(in *shuangqing.ShuangQing, card *accgrpc.Card, ppUser *passportuser.UserDetailReply) []*Statistics {
	out := []*Statistics{}
	out = append(out, &Statistics{
		Index:   1,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "uid",
		Usecase: "x",
		Scene:   "用户注册账号",
		Status:  "已收集",
		// Count:   strconv.FormatInt(in.GetUid1(), 10),
		Count: strconv.FormatInt(0, 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: fmt.Sprintf("%d", card.Mid),
		},
	})
	out = append(out, &Statistics{
		Index:   2,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "uid",
		Usecase: "展示用户ID",
		Scene:   "用户注册账号、客服用户反馈展示、社区互动",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetUid2(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: fmt.Sprintf("%d", card.Mid),
		},
	})
	out = append(out, &Statistics{
		Index:   3,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "昵称",
		Usecase: "完善网络身份标识、展示昵称",
		Scene:   "注册账号、社区互动",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetNickName(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: card.Name,
		},
	})
	out = append(out, &Statistics{
		Index:   4,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "头像",
		Usecase: "完善网络身份标识、展示头像",
		Scene:   "社区互动",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetAvatar(), 10),
		Content: StatisticsContent{
			Type:  "image",
			Value: card.Face,
		},
	})
	out = append(out, &Statistics{
		Index:   5,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "签名",
		Usecase: "完善网络身份标识、展示签名",
		Scene:   "社区互动",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetSign(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: card.Sign,
		},
	})
	out = append(out, &Statistics{
		Index:   6,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "第三方账号ID",
		Usecase: "账号快捷登录",
		Scene:   "使用微信、QQ、微博ID进行快捷登录",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetThirdPartyParticipantAccountId(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "https://passport.bilibili.com/account/mobile/security",
		},
	})
	out = append(out, &Statistics{
		Index:   7,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "电子邮箱地址",
		Usecase: "提供账号服务",
		Scene:   "完善资料、账号找回、账号申诉、版权保护计划",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetEmail(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: ppUser.HideEmail,
		},
	})
	out = append(out, &Statistics{
		Index:   33,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "电子邮箱的登录次数",
		Usecase: "快捷登录",
		Scene:   "邮箱登录",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetEmailLoginNum(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: ppUser.HideEmail,
		},
	})
	out = append(out, &Statistics{
		Index:   34,
		Type:    "用户网络身份标识和鉴权信息",
		Name:    "手机号的登录次数",
		Usecase: "快捷登录",
		Scene:   "手机号码快捷登录",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetPhoneLoginNum(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: ppUser.HideTel,
		},
	})

	out = append(out, &Statistics{
		Index:   8,
		Type:    "用户基本信息",
		Name:    "电话号码",
		Usecase: "注册账号",
		Scene:   "用户注册账号、账号登录、完善收货信息、身份验证",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetPhone(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: ppUser.HideTel,
		},
	})
	out = append(out, &Statistics{
		Index:   9,
		Type:    "用户基本信息",
		Name:    "住址",
		Usecase: "配送发货",
		Scene:   "完善收货信息",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetAddress(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "https://passport.bilibili.com/account/mobile/security",
		},
	})
	out = append(out, &Statistics{
		Index:   10,
		Type:    "用户基本信息",
		Name:    "生日",
		Usecase: "生日开屏祝福",
		Scene:   "编辑资料",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetBirthday(), 10),
		Content: StatisticsContent{
			Type:         "link",
			ValueIOS:     "bilibili://user_center/edit_profile",
			ValueAndroid: "activity://personinfo/info",
		},
	})
	out = append(out, &Statistics{
		Index:   11,
		Type:    "用户基本信息",
		Name:    "性别",
		Usecase: "展示性别",
		Scene:   "编辑资料",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetGender(), 10),
		Content: StatisticsContent{
			Type:         "link",
			ValueIOS:     "bilibili://user_center/edit_profile",
			ValueAndroid: "activity://personinfo/info",
		},
	})

	out = append(out, &Statistics{
		Index:   12,
		Type:    "用户身份证明",
		Name:    "个人姓名",
		Usecase: "身份认证",
		Scene:   "实名认证、完善收货信息、报税",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetName(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://user_center/auth/realname_v2?source_event=1",
		},
	})
	out = append(out, &Statistics{
		Index:   13,
		Type:    "用户身份证明",
		Name:    "证件号码",
		Usecase: "身份认证",
		Scene:   "实名认证、完善收货信息、报税、青少年保护计划",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetIdNumber(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://user_center/auth/realname_v2?source_event=1",
		},
	})
	out = append(out, &Statistics{
		Index:   14,
		Type:    "用户身份证明",
		Name:    "证件照片",
		Usecase: "身份认证",
		Scene:   "实名认证人工渠道、跨境购",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetIdPhoto(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://user_center/auth/realname_v2?source_event=1",
		},
	})
	out = append(out, &Statistics{
		Index:   15,
		Type:    "用户身份证明",
		Name:    "其他证件",
		Usecase: "身份认证",
		Scene:   "实名认证人工渠道、跨境购",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetOtherDocuments(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://user_center/auth/realname_v2?source_event=1",
		},
	})

	out = append(out, &Statistics{
		Index:   16,
		Type:    "个人财产信息",
		Name:    "交易信息",
		Usecase: "完成交易、保障交易安全",
		Scene:   "查询交易信息、提供客户服务、购买商品和服务",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetFlowRecord(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "https://pay.bilibili.com/pay-v2/all-bill?noTitleBar=1&native.theme=1",
		},
	})
	out = append(out, &Statistics{
		Index:   17,
		Type:    "个人财产信息",
		Name:    "支付账号",
		Usecase: "绑定用户收款以便打款",
		Scene:   "提现、报税、购买商品和服务",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetBankCard(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://bilipay/mine_wallet",
		},
	})

	out = append(out, &Statistics{
		Index:   18,
		Type:    "内容制作与发布",
		Name:    "用户上传的音视频",
		Usecase: "提供音视频、图文的展示和播放服务",
		Scene:   "稿件管理、展示和播放、动态",
		Status:  "已收集",
		Count:   "/",
		Content: StatisticsContent{
			Type:  "link",
			Value: fmt.Sprintf("bilibili://space/%d", card.Mid),
		},
	})
	// out = append(out, &Statistics{
	// 	Type:    "内容制作与发布",
	// 	Name:    "用户上传的文字",
	// 	Usecase: "提供社区互动服务",
	// 	Scene:   "评论、弹幕",
	// 	Status:  "已收集",
	// 	Count:   in.GetUploadedByUser2(),
	// })
	// out = append(out, &Statistics{
	// 	Type:    "内容制作与发布",
	// 	Name:    "用户自主上传的音视频、文字等",
	// 	Usecase: "AI分析视频信息帮助用户智能填写稿件信息",
	// 	Scene:   "投稿",
	// 	Status:  "已收集",
	// 	Count:   in.GetUploadedByUser3(),
	// })
	// out = append(out, &Statistics{
	// 	Type:    "内容制作与发布",
	// 	Name:    "用户上传的音视频、图文稿件",
	// 	Usecase: "监测站外侵权",
	// 	Scene:   "版权保护计划",
	// 	Status:  "已收集",
	// 	Count:   in.GetUploadedByUser4(),
	// })
	out = append(out, &Statistics{
		Index:   24,
		Type:    "内容制作与发布",
		Name:    "用户上传的图文",
		Usecase: "提供音视频、图文的展示和播放服务",
		Scene:   "动态、专栏",
		Status:  "已收集",
		Count:   "/",
		Content: StatisticsContent{
			Type:  "link",
			Value: fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", card.Mid),
		},
	})

	// out = append(out, &Statistics{
	// 	Type:    "互动与服务",
	// 	Name:    "用户视频播放记录",
	// 	Usecase: "提升个性化体验",
	// 	Scene:   "定制化内容展示",
	// 	Status:  "已收集",
	// 	Count:   in.GetUserPlaybackRecord(),
	// 	Content: StatisticsContent{
	// 		Type:         "link",
	// 		ValueAndroid: "bilibili://history",
	// 		ValueIOS:     "bilibili://user_center/history",
	// 	},
	// })
	// out = append(out, &Statistics{
	// 	Type:    "互动与服务",
	// 	Name:    "内容分享",
	// 	Usecase: "内容分享",
	// 	Scene:   "分享视频、动态等",
	// 	Status:  "已收集",
	// 	Count:   in.GetUserShare(),
	// })
	out = append(out, &Statistics{
		Index:   25,
		Type:    "互动与服务",
		Name:    "会员购订单信息",
		Usecase: "完成会员购交易、保障交易安全",
		Scene:   "查询交易信息、提供客户服务、购买商品和服务",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetOrder(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "bilibili://mall/order/list",
		},
	})
	out = append(out, &Statistics{
		Index:   26,
		Type:    "互动与服务",
		Name:    "历史记录",
		Usecase: "提供信息查询服务",
		Scene:   "定制化内容展示、历史记录查询",
		Status:  "已收集",
		Count:   "/",
		Content: StatisticsContent{
			Type:         "link",
			ValueAndroid: "bilibili://history",
			ValueIOS:     "bilibili://user_center/history",
		},
	})
	out = append(out, &Statistics{
		Index:   27,
		Type:    "互动与服务",
		Name:    "当前网络状态",
		Usecase: "用于判断是否要自动播放、流量提醒",
		Scene:   "视频播放展示",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetTrafficOrWifi(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: "/",
		},
	})
	out = append(out, &Statistics{
		Index:   28,
		Type:    "互动与服务",
		Name:    "设备信息",
		Usecase: "提高服务安全性、安全风控、视频播放",
		Scene:   "定制化内容展示、视频播放展示、账号风控",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetDeviceAttributeInfo(), 10),
		Content: StatisticsContent{
			Type:  "text",
			Value: "可在手机系统设置关于本机查看详情",
		},
	})
	// out = append(out, &Statistics{
	// 	Type:    "互动与服务",
	// 	Name:    "mac",
	// 	Usecase: "提高服务安全性、安全风控、视频播放、提升个性化体验",
	// 	Scene:   "定制化内容展示、视频播放展示、账号风控、个性化广告推送",
	// 	Status:  "已收集",
	// 	Count:   in.GetMac(),
	// })
	// out = append(out, &Statistics{
	// 	Type:    "互动与服务",
	// 	Name:    "imei",
	// 	Usecase: "提高服务安全性、安全风控、视频播放、提升个性化体验",
	// 	Scene:   "定制化内容展示、视频播放展示、账号风控、个性化广告推送",
	// 	Status:  "已收集",
	// 	Count:   in.GetImei(),
	// })
	out = append(out, &Statistics{
		Index:   31,
		Type:    "互动与服务",
		Name:    "关注",
		Usecase: "关注管理、更新内容提示",
		Scene:   "定制化内容展示、动态、关注列表",
		Status:  "已收集",
		Count:   strconv.FormatInt(in.GetAttention(), 10),
		Content: StatisticsContent{
			Type:  "link",
			Value: "https://space.bilibili.com/h5/follow",
		},
	})
	// out = append(out, &Statistics{
	// 	Type:    "互动与服务",
	// 	Name:    "剪切板内容",
	// 	Usecase: "分享的短链进行自动跳转及口令跳转",
	// 	Scene:   "读取被分享的短链进行自动跳转及口令跳转",
	// 	Status:  "已收集",
	// 	Count:   in.GetClipboardContent(),
	// 	Content: StatisticsContent{
	// 		Type:  "text",
	// 		Value: "一次性消费完既删除",
	// 	},
	// })

	// out = append(out, &Statistics{
	// 	Type:    "用户网络身份标识和鉴权信息",
	// 	Name:    "电子邮箱的登录次数",
	// 	Usecase: "快捷登录",
	// 	Scene:   "邮箱登录",
	// 	Status:  "已收集",
	// 	Count:   strconv.FormatInt(in.GetEmailLoginNum(), 10),
	// 	Content: StatisticsContent{
	// 		Type:  "text",
	// 		Value: ppUser.HideEmail,
	// 	},
	// })
	// out = append(out, &Statistics{
	// 	Type:    "用户网络身份标识和鉴权信息",
	// 	Name:    "手机号的登录次数",
	// 	Usecase: "快捷登录",
	// 	Scene:   "手机号码快捷登录",
	// 	Status:  "已收集",
	// 	Count:   strconv.FormatInt(in.GetPhoneLoginNum(), 10),
	// 	Content: StatisticsContent{
	// 		Type:  "text",
	// 		Value: ppUser.HideTel,
	// 	},
	// })

	for _, v := range out {
		if v.Content.Type == "" {
			v.Content.Type = "text"
		}
	}
	return out
}

type NFTSettingButtonReq struct {
	Mid     int64
	MobiApp string
	SLocale string `form:"s_locale"`
	CLocale string `form:"c_locale"`
}

type NFTSettingButtonReply struct {
	//按钮文案
	Text string `json:"text,omitempty"`
	//跳转链接
	Url string `json:"url,omitempty"`
}
