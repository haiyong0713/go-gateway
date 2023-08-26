package ranklist

import (
	model "go-gateway/app/app-svr/app-show/interface/model/rank-list"
)

// ListPagination is
type ListPagination struct {
	List []*model.Meta `json:"list"`
	Page struct {
		Total int64 `json:"total"`
		Size  int64 `json:"size"`
		Page  int64 `json:"page"`
	}
}
