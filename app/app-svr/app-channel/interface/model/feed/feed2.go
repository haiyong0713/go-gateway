package feed

import (
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
)

type Show2 struct {
	Topic cardm.Handler   `json:"topic,omitempty"`
	Feed  []cardm.Handler `json:"feed"`
}

type Tab struct {
	Items []cardm.Handler `json:"items"`
}
