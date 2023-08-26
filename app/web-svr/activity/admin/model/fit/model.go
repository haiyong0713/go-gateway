package fit

// PlanRecord act_fit_plan_config计划表字段
type PlanRecord struct {
	PlanTitle   string `json:"plan_title" form:"plan_title" validate:"required" gorm:"plan_title"`
	PlanTags    string `json:"plan_tags" form:"plan_tags" validate:"required" gorm:"plan_tags"`
	BodanId     string `json:"bodan_id" form:"bodan_id" validate:"required" gorm:"bodan_id"`
	PlanView    int64  `json:"plan_view" form:"plan_view" default:"0" gorm:"plan_view"`
	PlanDanmaku int64  `json:"plan_danmaku" form:"plan_danmaku" default:"0" gorm:"paln_danmaku"`
	PlanFav     int64  `json:"plan_fav" form:"plan_fav" default:"0" gorm:"plan_fav"`
	PicCover    string `json:"pic_cover" form:"pic_cover" validate:"required" gorm:"pic_cover"`
	Creator     string `json:"creator" form:"creator" default:"system" gorm:"creator"`
}

// UpdatePlanRecord 更新入参
type UpdatePlanRecord struct {
}
