package like

// Fate .
type FateShow struct {
	TotalPvCount int64 `json:"total_pv_count"`
	WantCount    int64 `json:"want_count"`
}

// FateData .
type FateData struct {
	TppPv      int64 `json:"tpp_pv"`
	TppWantCnt int64 `json:"tpp_want_cnt"`
	TxPv       int64 `json:"tx_pv"`
	LocalView  int64 `json:"local_view"`
}

// FateData .
type FateSwitch struct {
	TppSuccess int64 `json:"tpp_success"`
	TxSuccess  int64 `json:"tx_success"`
}

// FateConfData .
type FateConfData struct {
	TppMovieID     int64   `json:"tpp_movie_id"`
	TppBackPv      int64   `json:"tpp_back_pv"`
	TppBackWantCnt int64   `json:"tpp_back_want_cnt"`
	TxURL          string  `json:"tx_url"`
	TxBackPv       int64   `json:"tx_back_pv"`
	PvAids         []int64 `json:"pv_aids"`
}
