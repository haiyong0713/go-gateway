package model

// Skin .
type Skin struct {
	ID                  int64   `json:"id"`
	Name                string  `json:"name"`
	Cover               string  `json:"cover"`
	Image               string  `json:"image"`
	TitleTextColor      string  `json:"title_text_color"`
	TitleShadowColor    string  `json:"title_shadow_color"`
	TitleShadowOffsetX  float32 `json:"title_shadow_offset_x"`
	TitleShadowOffsetY  float32 `json:"title_shadow_offset_y"`
	TitleShadowRadius   float32 `json:"title_shadow_radius"`
	ProgressBarColor    string  `json:"progress_bar_color"`
	ProgressShadowColor string  `json:"progress_shadow_color"`
}
