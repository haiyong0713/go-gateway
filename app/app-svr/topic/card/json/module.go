package jsonwebcard

import account "git.bilibili.co/bapis/bapis-go/account/service"

type Modules struct {
	ModuleTag         *ModuleTag         `json:"module_tag,omitempty"`
	ModuleAuthor      *ModuleAuthor      `json:"module_author,omitempty"`
	ModuleDynamic     *ModuleDynamic     `json:"module_dynamic,omitempty"`
	ModuleDispute     *ModuleDispute     `json:"module_dispute,omitempty"`
	ModuleStat        *ModuleStat        `json:"module_stat,omitempty"`
	ModuleInteraction *ModuleInteraction `json:"module_interaction,omitempty"`
	ModuleShareInfo   *ModuleShareInfo   `json:"module_share_info,omitempty"`
	ModuleMore        *ModuleMore        `json:"module_more,omitempty"`
}

type ModuleAuthor struct {
	AuthorType AuthorType `json:"type"`
	UserInfo
	Following bool   `json:"following"`
	PubTime   string `json:"pub_time"`
	PubAction string `json:"pub_action"`
	PubTs     int64  `json:"pub_ts"`
	// 装扮
	Pendant struct {
		Expire            int64  `json:"expire,omitempty"`
		Pid               int64  `json:"pid,omitempty"`
		Name              string `json:"name,omitempty"`
		Image             string `json:"image,omitempty"`
		ImageEnhance      string `json:"image_enhance,omitempty"`
		ImageEnhanceFrame string `json:"image_enhance_frame,omitempty"`
	} `json:"pendant"`
	// 会员信息
	Vip struct {
		Type               int32            `json:"type,omitempty"`
		VipStatus          int32            `json:"status,omitempty"`
		Label              account.VipLabel `json:"label,omitempty"`
		DueDate            int64            `json:"due_date,omitempty"`
		Role               int64            `json:"role,omitempty"`
		VipPayType         int32            `json:"vip_pay_type,omitempty"`
		ThemeType          int32            `json:"themeType,omitempty"`
		AvatarSubscript    int32            `json:"avatar_subscript,omitempty"`
		AvatarSubscriptUrl string           `json:"avatar_subscript_url,omitempty"`
		NicknameColor      string           `json:"nickname_color,omitempty"`
	} `json:"vip,omitempty"`
	// 认证信息
	OfficialVerify struct {
		Type int32  `json:"type,omitempty"`
		Desc string `json:"desc,omitempty"`
	} `json:"official_verify,omitempty"`
	// 装饰信息
	Decorate *Decorate `json:"decorate,omitempty"`
	IsTop    bool      `json:"is_top"` // 是否置顶
}

type Decorate struct {
	Id          int64  `json:"id"`
	Type        int64  `json:"type"`
	Name        string `json:"name"`
	CardUrl     string `json:"card_url"`
	JumpUrl     string `json:"jump_url"`
	DecorateFan struct {
		IsFan  bool   `json:"is_fan"`
		Color  string `json:"color"`
		NumStr string `json:"num_str"`
		Number int32  `json:"number"`
	} `json:"fan,omitempty"`
}

type ModuleDynamic struct {
	Desc       *DynDesc      `json:"desc,omitempty"`
	Major      DynMajor      `json:"major,omitempty"`
	Additional DynAdditional `json:"additional,omitempty"`
	Topic      *DynTopic     `json:"topic,omitempty"`
}

type DynDesc struct {
	Text         string          `json:"text,omitempty"`
	RichTextNode []*RichTextNode `json:"rich_text_nodes,omitempty"`
}

type RichTextNode struct {
	Text              string             `json:"text,omitempty"`
	OrigText          string             `json:"orig_text,omitempty"` // 原文字
	DescItemType      RichTextNodeType   `json:"type,omitempty"`
	JumpUrl           string             `json:"jump_url,omitempty"` // 点击跳转
	IconUrl           string             `json:"icon_url,omitempty"`
	IconName          string             `json:"icon_name,omitempty"`
	Rid               string             `json:"rid,omitempty"`
	RichTextNodeGood  *RichTextNodeGood  `json:"good,omitempty"`
	RichTextNodeEmoji *RichTextNodeEmoji `json:"emoji,omitempty"`
}

type RichTextNodeGood struct {
	Type    int64  `json:"type,omitempty"`
	Text    string `json:"text,omitempty"`
	JumpUrl string `json:"jump_url,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type RichTextNodeEmoji struct {
	Type    int64  `json:"type,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url,omitempty"` // emoji的url
}

type DynTopic struct {
	TopicId   int64  `json:"topic_id,omitempty"`
	TopicName string `json:"topic_name,omitempty"`
	JumpUrl   string `json:"jump_url,omitempty"`
}

type ModuleDispute struct {
	Title   string `json:"title,omitempty"`
	Desc    string `json:"desc,omitempty"`
	JumpUrl string `json:"jump_url,omitempty"`
}

type ModuleInteraction struct {
	Like    *InteractiveItem `json:"like,omitempty"`
	Comment *InteractiveItem `json:"comment,omitempty"`
}

type InteractiveItem struct {
	JumpUrl string   `json:"jump_url,omitempty"`
	Desc    *DynDesc `json:"desc,omitempty"`
}

type ModuleShareInfo struct {
	Title        string          `json:"title,omitempty"`
	ShareChannel []*ShareChannel `json:"share_channels,omitempty"`
	ShareOrigin  string          `json:"share_origin,omitempty"`
	Oid          string          `json:"oid,omitempty"`
	Sid          string          `json:"sid,omitempty"`
}

type ShareChannel struct {
	Name    string `json:"name,omitempty"`
	Image   string `json:"image,omitempty"`
	Channel string `json:"channel,omitempty"`
}

type ModuleStat struct {
	Forward *MdlStatItem `json:"forward,omitempty"`
	Comment *MdlStatItem `json:"comment,omitempty"`
	Like    *MdlStatItem `json:"like,omitempty"`
}

type MdlStatItem struct {
	Count     int64  `json:"count"`
	Forbidden bool   `json:"forbidden"`
	Text      string `json:"text"`
	Status    string `json:"status,omitempty"`
}

// web卡三点逻辑
type ModuleMore struct {
	ThreePointItems []*ThreePointItems `json:"three_point_items,omitempty"`
}

type ThreePointItems struct {
	Type  ThreePointType `json:"type,omitempty"`
	Label string         `json:"label,omitempty"`
}

type ModuleTag struct {
	Text string `json:"text,omitempty"`
}
