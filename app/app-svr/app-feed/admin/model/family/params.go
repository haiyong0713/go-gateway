package family

type SearchFamilyReq struct {
	Mid int64 `form:"mid" validate:"required"`
}

type SearchFamilyRly struct {
	List []*SearchFamilyItem `json:"list"`
}

type SearchFamilyItem struct {
	Identity     string         `json:"identity"`
	Mid          int64          `json:"mid"`
	UserName     string         `json:"user_name"`
	RelatedUsers []*RelatedUser `json:"related_users"`
}

type RelatedUser struct {
	ID       int64  `json:"id"`
	Mid      int64  `json:"mid"`
	UserName string `json:"user_name"`
}

type BindListReq struct {
	Mid int64 `form:"mid" validate:"required"`
	Pn  int64 `form:"pn" validate:"min=0" default:"1"`
	Ps  int64 `form:"ps" validate:"min=0" default:"20"`
}

type BindListRly struct {
	Page *Page        `json:"page"`
	List []*FamilyLog `json:"list"`
}

type Page struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}

type UnbindReq struct {
	ID int64 `form:"id" validate:"required"`
}
