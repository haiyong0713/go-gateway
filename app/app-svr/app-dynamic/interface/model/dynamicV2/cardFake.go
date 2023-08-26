package dynamicV2

import dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"

type FakeDynamicContent struct {
	DynamicID  string               `json:"dynamic_id"`
	AtUids     string               `json:"at_uids"`
	Type       string               `json:"type"`
	VoteID     string               `json:"vote_id"`
	CoverURL   string               `json:"coverUrl"`
	Images     []*FakeDynamicImages `json:"images"`
	Ctrls      []*Ctrl              `json:"ctrls"`
	Content    string               `json:"content"`
	Emojis     []*EmojiItem         `json:"emojis"`
	Duration   string               `json:"duration"` //稿件总时长 单位=秒
	AttachAvID string               `json:"attach_avid"`
	Extend     string               `json:"extend"`
}

type FakeExtend struct {
	Lottery        *dyncommongrpc.ExtLottery    `json:"lott_cfg"`
	Vote           *dyncommongrpc.ExtVote       `json:"vote_cfg"`
	Goods          *dyncommongrpc.ExtOpenGoods  `json:"open_goods_cfg"`
	LBS            *dyncommongrpc.ExtLbs        `json:"lbs_cfg"`
	FlagCfg        *dyncommongrpc.ExtFlagCfg    `json:"flag_cfg"`
	BottomBusiness *dyncommongrpc.ExtBottom     `json:"bottom"`
	ReserveCfg     *dyncommongrpc.ExtReserveCfg `json:"reserve"`
}

type FakeDynamicImages struct {
	ImgHeight int64   `json:"img_height"`
	ImgSize   float32 `json:"img_size"`
	ImgSrc    string  `json:"img_src"`
	ImgWidth  int64   `json:"img_width"`
}

type AppletLabel struct {
	Icon        string `json:"icon"`
	JumpText    string `json:"jump_text"`
	ProgramText string `json:"program_text"`
}

type VoteResule struct {
	Info *struct {
		VoteID    int64  `json:"vote_id"`
		Title     string `json:"title"`
		Desc      string `json:"desc"`
		Type      int32  `json:"type"`
		ChoiceCnt int32  `json:"choice_cnt"`
		UID       int    `json:"uid"`
		Endtime   int64  `json:"endtime"`
		Status    int32  `json:"status"`
		Cnt       int64  `json:"cnt"`
		Options   []*struct {
			Idx    int32  `json:"idx"`
			Desc   string `json:"desc"`
			Cnt    int32  `json:"cnt"`
			BtnStr string `json:"btn_str"`
			Title  string `json:"title"`
			ImgURL string `json:"img_url"`
		} `json:"options"`
		OptionsCnt   int32  `json:"options_cnt"`
		Face         string `json:"face"`
		Name         string `json:"name"`
		BizType      int32  `json:"biz_type"`
		ImgURL       string `json:"img_url"`
		DefaultShare int    `json:"default_share"`
	} `json:"info"`
	MyVotes []int32 `json:"my_votes"`
}
