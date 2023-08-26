package rename

type Rename struct {
	//业务类型 1-abtest 2-多人群包
	Type     int64             `json:"type"`
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	TabNames map[string]string `json:"tab_names"`
}
