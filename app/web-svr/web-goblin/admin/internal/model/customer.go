package model

type GbCustomerBusiness struct {
	ID           int64  `form:"id" json:"id"`
	Business     string `form:"business" validate:"required" json:"business"`
	Logo         string `form:"logo" json:"logo"`
	Rank         string `form:"rank" json:"rank"`
	IsDeleted    int32  `form:"is_deleted" json:"is_deleted"`
	CustomerType int64  `form:"customer_type" json:"customer_type"`
}

type GbCustomerCenters struct {
	ID             int64  `form:"id" json:"id"`
	CustomerType   int64  `form:"customer_type" validate:"required" json:"customer_type"`
	BusinessType   int64  `form:"business_type" json:"business_type"`
	BusinessName   string `form:"business_name" json:"business_name"  gorm:"-"`
	Title          string `form:"title" json:"title"`
	Copywriting    string `form:"copywriting" json:"copywriting"`
	HighlightTitle string `form:"highlight_title" json:"highlight_title"`
	Image          string `form:"image" json:"image"`
	WebUrl         string `form:"web_url" json:"web_url"`
	H5Url          string `form:"h5_url" json:"h5_url"`
	Stime          int64  `form:"stime" json:"stime"`
	Etime          int64  `form:"etime" json:"etime"`
	Rank           string `form:"rank" json:"rank"`
	IsDeleted      int32  `form:"is_deleted" json:"is_deleted"`
}
