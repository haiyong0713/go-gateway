package page

type ResPageList struct {
	List     []*ActPage `json:"list"`
	Count    int        `json:"count"`
	Page     int        `json:"page"`
	PageSize int        `json:"pagesize"`
}
