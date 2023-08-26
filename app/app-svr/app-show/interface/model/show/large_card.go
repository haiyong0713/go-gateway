package show

type LargeCard struct {
	ID        int64  `json:"id" form:"id"`
	Title     string `json:"title" form:"title" validate:"title"`
	CardType  string `json:"card_type" form:"card_type" validate:"card_type"`
	RID       int64  `json:"rid" form:"rid" validate:"rid"`
	WhiteList string `json:"white_list" form:"white_list"`
	Auto      int64  `json:"auto" form:"auto" validate:"auto"`
	Sticky    int64  `json:"sticky" form:"sticky" validate:"sticky"`
}
