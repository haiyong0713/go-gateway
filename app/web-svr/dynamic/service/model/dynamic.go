package model

import "go-gateway/app/app-svr/archive/service/api"

// Page page.
type Page struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Count int `json:"count"`
}

// DynamicArcs3 dynamic archives3.
type DynamicArcs3 struct {
	Page     *Page      `json:"page"`
	Archives []*api.Arc `json:"archives"`
}
