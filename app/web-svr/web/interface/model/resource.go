package model

type ParamConfig struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	Remark     string `json:"remark"`
	Department int64  `json:"department"`
}
