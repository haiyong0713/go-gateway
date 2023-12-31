// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package rank is a generated GoMock package.
package rank

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	rank "go-gateway/app/web-svr/activity/interface/model/rank"
	reflect "reflect"
)

// MockDao is a mock of Dao interface
type MockDao struct {
	ctrl     *gomock.Controller
	recorder *MockDaoMockRecorder
}

// MockDaoMockRecorder is the mock recorder for MockDao
type MockDaoMockRecorder struct {
	mock *MockDao
}

// NewMockDao creates a new mock instance
func NewMockDao(ctrl *gomock.Controller) *MockDao {
	mock := &MockDao{ctrl: ctrl}
	mock.recorder = &MockDaoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDao) EXPECT() *MockDaoMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockDao) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockDaoMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDao)(nil).Close))
}

// GetRank mocks base method
func (m *MockDao) GetRank(c context.Context, rankKey string) ([]*rank.Redis, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRank", c, rankKey)
	ret0, _ := ret[0].([]*rank.Redis)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRank indicates an expected call of GetRank
func (mr *MockDaoMockRecorder) GetRank(c, rankKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRank", reflect.TypeOf((*MockDao)(nil).GetRank), c, rankKey)
}

// GetMidRank mocks base method
func (m *MockDao) GetMidRank(c context.Context, rankActivityKey string, mid int64) (*rank.Redis, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMidRank", c, rankActivityKey, mid)
	ret0, _ := ret[0].(*rank.Redis)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMidRank indicates an expected call of GetMidRank
func (mr *MockDaoMockRecorder) GetMidRank(c, rankActivityKey, mid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMidRank", reflect.TypeOf((*MockDao)(nil).GetMidRank), c, rankActivityKey, mid)
}

// Ping mocks base method
func (m *MockDao) Ping(c context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockDaoMockRecorder) Ping(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockDao)(nil).Ping), c)
}
