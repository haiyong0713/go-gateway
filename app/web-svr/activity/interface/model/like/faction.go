package like

type Faction struct {
	Sid   int64         `json:"sid"`
	Score int64         `json:"score"`
	List  []*FactionAcc `json:"list"`
}

type FactionAcc struct {
	Mid   int64 `json:"mid"`
	Score int64 `json:"score"`
}

type FactionUser struct {
	*SimpleUser
	Score int64 `json:"score"`
}

type FactionRes struct {
	Sid     int64          `json:"sid"`
	Name    string         `json:"name"`
	Score   int64          `json:"score"`
	AccList []*FactionUser `json:"acc_list"`
}
