package preheat

type DownInfo struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ImgURL    string `json:"img_url"`
	DownURL   string `json:"down_url"`
	SchemaURL string `json:"schema_url"`
}
