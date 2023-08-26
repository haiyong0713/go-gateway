package bws

type User struct {
	ID  int64
	Mid int64
	Key string
}

type LotteryUser struct {
	Bid  int64 `json:"bid"`
	Mid  int64 `json:"mid"`
	Rank int   `json:"rank"`
}
