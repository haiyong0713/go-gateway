package model

type S10RankDataInterventionReq struct {
	TournamentID      string `json:"tournament_id" form:"tournament_id"`
	CurrentRound      string `json:"current_round" form:"current_round"`
	FinalistRound     string `json:"finalist_round" form:"finalist_round"`
	FinalistH5Pic     string `json:"finalist_h_5_pic" form:"finalist_h_5_pic"`
	FinalistWebPic    string `json:"finalist_web_pic" form:"finalist_web_pic"`
	FinalRound        string `json:"final_round" form:"final_round"`
	FinalH5Pic        string `json:"final_h_5_pic" form:"final_h_5_pic"`
	FinalWebPic       string `json:"final_web_pic" form:"final_web_pic"`
	UpdatePic         int    `json:"update_pic" form:"update_pic"`
	PromoteNum        int    `json:"promote_num" form:"promote_num"`
	EliminateNum      int    `json:"eliminate_num" form:"eliminate_num"`
	FinalPromoteNum   int    `json:"final_promote_num" form:"final_promote_num"`
	FinalEliminateNum int    `json:"final_eliminate_num" form:"final_eliminate_num"`
}

type S10RankingInterventionData struct {
	TournamentID      string                            `json:"tournament_id" form:"tournament_id"`
	CurrentRound      string                            `json:"current_round" form:"current_round"`
	FinalistRound     string                            `json:"finalist_round" form:"finalist_round"`
	PromoteNum        int                               `json:"promote_num" form:"promote_num"`
	EliminateNum      int                               `json:"eliminate_num" form:"eliminate_num"`
	FinalPromoteNum   int                               `json:"final_promote_num" form:"final_promote_num"`
	FinalEliminateNum int                               `json:"final_eliminate_num" form:"final_eliminate_num"`
	RoundInfo         []S10RankingInterventionRoundInfo `json:"round_info" form:"round_info"`
}

type S10RankingInterventionRoundInfo struct {
	RoundID string `json:"round_id"`
	H5Pic   string `json:"h_5_pic"`
	WebPic  string `json:"web_pic"`
}
