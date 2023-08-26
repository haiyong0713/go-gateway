package jsonwebcard

type DynAdditional interface {
	GetAdditionalType() AdditionalType
}

type AdditionalGoods struct {
	AdditionalType AdditionalType `json:"type,omitempty"`
	Goods          *Goods         `json:"goods,omitempty"`
}

type Goods struct {
	HeadText string       `json:"head_text,omitempty"`
	HeadIcon string       `json:"head_icon,omitempty"`
	JumpUrl  string       `json:"jump_url,omitempty"`
	Items    []*GoodsItem `json:"items,omitempty"`
}

type GoodsItem struct {
	Cover    string `json:"cover,omitempty"`
	Name     string `json:"name,omitempty"`
	Brief    string `json:"brief,omitempty"`
	Price    string `json:"price,omitempty"`
	JumpUrl  string `json:"jump_url,omitempty"`
	JumpDesc string `json:"jump_desc,omitempty"`
	Id       int64  `json:"id,omitempty"`
}

type AdditionalVote struct {
	AdditionalType AdditionalType `json:"type,omitempty"`
	Vote           *Vote          `json:"vote,omitempty"`
}

type Vote struct {
	VoteId       int64  `json:"vote_id,omitempty"`
	Title        string `json:"title,omitempty"`
	ChoiceCnt    int32  `json:"choice_cnt,omitempty"`
	DefaultShare int32  `json:"default_share,omitempty"`
	Desc         string `json:"desc,omitempty"`
	EndTime      int64  `json:"end_time,omitempty"`
	JoinNum      int64  `json:"join_num,omitempty"`
	Status       int32  `json:"status,omitempty"`
	Type         int32  `json:"type,omitempty"`
	Uid          int64  `json:"uid,omitempty"`
}

func NewAdditionalVote() DynAdditional {
	return AdditionalVote{AdditionalType: AdditionalTypeVote}
}

func (add AdditionalVote) GetAdditionalType() AdditionalType {
	return add.AdditionalType
}

func (add AdditionalGoods) GetAdditionalType() AdditionalType {
	return add.AdditionalType
}

type AdditionalReserve struct {
	AdditionalType AdditionalType `json:"type,omitempty"`
	Reserve        *Reserve       `json:"reserve,omitempty"`
}

type Reserve struct {
	Title         string         `json:"title,omitempty"`
	Desc1         *Desc1         `json:"desc1,omitempty"`
	Desc2         *Desc2         `json:"desc2,omitempty"`
	Desc3         *Desc3         `json:"desc3,omitempty"`
	JumpUrl       string         `json:"jump_url,omitempty"`
	ReserveButton *ReserveButton `json:"button,omitempty"`
	Rid           int64          `json:"rid,omitempty"`
	ReserveTotal  int64          `json:"reserve_total"`
	State         int64          `json:"state"`
	Stype         int64          `json:"stype"`
	UpMid         int64          `json:"up_mid,omitempty"`
}

type Desc1 struct {
	Text  string `json:"text,omitempty"`
	Style int64  `json:"style,omitempty"`
}

type Desc2 struct {
	Text    string `json:"text,omitempty"`
	Style   int64  `json:"style,omitempty"`
	Visible bool   `json:"visible,omitempty"`
}

type Desc3 struct {
	IconUrl string `json:"icon_url,omitempty"`
	Text    string `json:"text,omitempty"`
	Style   int64  `json:"style,omitempty"`
	JumpUrl string `json:"jump_url,omitempty"`
}

type ReserveButton struct {
	Type      int64         `json:"type,omitempty"`
	Status    int64         `json:"status,omitempty"`
	Check     *ReserveCheck `json:"check,omitempty"`
	UnCheck   *ReserveCheck `json:"uncheck,omitempty"`
	JumpStyle *ReserveCheck `json:"jump_style,omitempty"`
	JumpUrl   string        `json:"jump_url,omitempty"`
}

type ReserveCheck struct {
	IconUrl string `json:"icon_url,omitempty"`
	Text    string `json:"text,omitempty"`
	Disable int64  `json:"disable,omitempty"`
	Toast   string `json:"toast,omitempty"`
}

func NewAdditionalReserveNull() DynAdditional {
	const (
		_reserveNullState = 1
	)
	return AdditionalReserve{AdditionalType: AdditionalTypeReserve, Reserve: &Reserve{State: _reserveNullState}}
}

func (add AdditionalReserve) GetAdditionalType() AdditionalType {
	return add.AdditionalType
}
