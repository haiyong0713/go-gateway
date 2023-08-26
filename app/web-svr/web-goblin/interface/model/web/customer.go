package web

type Customer struct {
	ID                   int64  `json:"id"`
	CustomerType         int64  `json:"customer_type"`
	BusinessType         int64  `json:"business_type"`
	BusinessName         string `json:"business_name"`
	Logo                 string `json:"logo"`
	Title                string `json:"title"`
	Copywriting          string `json:"copywriting"`
	HighlightTitle       string `json:"highlight_title"`
	Image                string `json:"image"`
	WebUrl               string `json:"web_url"`
	H5Url                string `json:"h5_url"`
	Stime                int64  `json:"stime"`
	Etime                int64  `json:"etime"`
	CustomerRank         int64  `json:"customer_rank"`
	BusinessRank         int64  `json:"-"`
	BusinessCustomerType int64  `json:"-"`
}

type BusinessList struct {
	BusinessID   int64       `json:"business_id"`
	BusinessName string      `json:"business_name"`
	BusinessLogo string      `json:"business_logo"`
	BusinessRank int64       `json:"business_rank"`
	BusinessList []*Customer `json:"list"`
}

type CustomerCenter struct {
	Title        string          `json:"title"`
	List         []*Customer     `json:"list"`
	BusinessList []*BusinessList `json:"business_list"`
}
