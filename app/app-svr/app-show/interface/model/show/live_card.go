package show

type LiveCard struct {
	ID    int64  `json:"id" form:"id"`
	RID   int64  `json:"rid" form:"rid" validate:"rid"`
	Cover string `json:"cover" form:"cover" validate:"cover"`
}
