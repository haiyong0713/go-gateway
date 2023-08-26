package businessConfig

type BusinessConfig struct {
	ID          int64  `json:"id"`
	TreeID      int64  `json:"tre_id"`
	KeyName     string `json:"key_name"`
	Config      string `json:"config"`
	Description string `json:"description"`
	Relations   string `json:"relations"`
}
