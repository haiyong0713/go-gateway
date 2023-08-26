package adapters

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/dao/vote"
)

const (
	mockDSItemSize = 100
)

var MockDS = &mockDataSource{}

type mockDataSourceItem struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (i *mockDataSourceItem) GetName() string {
	return i.Name
}

func (i *mockDataSourceItem) GetId() int64 {
	return i.Id
}

func (i *mockDataSourceItem) GetSearchField1() string {
	return ""
}

func (i *mockDataSourceItem) GetSearchField2() string {
	return ""
}

func (i *mockDataSourceItem) GetSearchField3() string {
	return ""
}

type mockDataSource struct {
}

func (m *mockDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []vote.DataSourceItem, err error) {
	res = make([]vote.DataSourceItem, 0)
	for i := int64(mockDSItemSize); i >= 1; i-- {
		res = append(res, &mockDataSourceItem{
			Id:   i,
			Name: fmt.Sprintf("test-%v", i),
		})
	}
	return
}

func (m *mockDataSource) IsItemExists(ctx context.Context, itemId int64) (exist bool, err error) {
	exist = itemId <= mockDSItemSize
	return
}

func (m *mockDataSource) NewEmptyItem() vote.DataSourceItem {
	return &mockDataSourceItem{}
}
