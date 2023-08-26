package datamanage

type ResDataManageSelect struct {
	Count   int                    `json:"count"`
	List    []map[string]string    `json:"list"`
	Req     *ReqDataManageSelect   `json:"req"`
	Where   map[string]interface{} `json:"where"`
	Columns []string               `json:"columns"`
}
