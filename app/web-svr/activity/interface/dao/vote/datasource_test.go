package vote

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	mockDSSize = 2000
)

var mockDS *mockDataSource

type mockItem struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (i *mockItem) GetName() string {
	return i.Name
}

func (i *mockItem) GetId() int64 {
	return i.Id
}

func (i *mockItem) GetSearchField1() string {
	return i.Name
}

func (i *mockItem) GetSearchField2() string {
	return ""
}

func (i *mockItem) GetSearchField3() string {
	return ""
}

type mockDataSource struct {
}

func (m *mockDataSource) ListAllItems(ctx context.Context, sourceId int64) (res []DataSourceItem, err error) {
	res = make([]DataSourceItem, 0)
	for i := int64(mockDSSize); i >= 1; i-- {
		res = append(res, &mockItem{
			Id:   i,
			Name: fmt.Sprintf("test-%v", i),
		})
	}
	return
}

func (m *mockDataSource) IsItemExists(ctx context.Context, itemId int64) (exist bool, err error) {
	exist = itemId <= mockDSSize
	return
}

func (m *mockDataSource) NewEmptyItem() DataSourceItem {
	return &mockItem{}
}

func init() {
	testDao.datasourceMap["TEST"] = mockDS
}

func TestDataSource(t *testing.T) {
	activityId := int64(1)
	ctx := context.Background()
	Convey("DataSource", t, func() {
		Convey("RefreshVoteActivityDSItems", func() {
			err := testDao.RefreshVoteActivityDSItems(ctx, activityId)
			So(err, ShouldBeNil)
		})
	})
}
