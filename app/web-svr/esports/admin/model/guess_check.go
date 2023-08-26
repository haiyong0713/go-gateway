package model

type SeasonSelect struct {
	Options []*SeasonOption `json:"options"`
}

type SeasonOption struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type GuessContestTable struct {
	Items []*GuessContest `json:"items"`
}

type GuessContest struct {
	Id                    int64  `json:"id"`
	Mid                   int64  `json:"mid"`
	ContestInfo           string `json:"contest_info"`
	GuessInfo             string `json:"guess_info"`
	GuessStatus           string `json:"guess_status"`
	JoinStatus            string `json:"join_status"`
	ResultOption          string `json:"result_option"`
	JoinOption            string `json:"join_option"`
	SettlementStatus      string `json:"settlement_status"`
	SettlementStatusCache string `json:"settlement_status_cache"`
	JoinNum               string `json:"join_num"`
	Income                string `json:"income"`
}
