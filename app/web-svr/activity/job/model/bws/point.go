package bws

type RechargeAward struct {
	Recharge []*Recharge `json:"recharge"`
}

type Recharge struct {
	Unlock []*Award `json:"unlock"`
}

type Award struct {
	ID     int64 `json:"id"`
	Amount int   `json:"amount"`
}
