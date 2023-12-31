// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package rank is a generated GoMock package.
package rank

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	sql "go-common/library/database/sql"
	rank "go-gateway/app/web-svr/activity/job/model/rank_v3"
	reflect "reflect"
	time "time"
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

// BeginTran mocks base method
func (m *MockDao) BeginTran(c context.Context) (*sql.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginTran", c)
	ret0, _ := ret[0].(*sql.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTran indicates an expected call of BeginTran
func (mr *MockDaoMockRecorder) BeginTran(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTran", reflect.TypeOf((*MockDao)(nil).BeginTran), c)
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

// GetBaseOnline mocks base method
func (m *MockDao) GetBaseOnline(c context.Context, now time.Time) ([]*rank.Base, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBaseOnline", c, now)
	ret0, _ := ret[0].([]*rank.Base)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBaseOnline indicates an expected call of GetBaseOnline
func (mr *MockDaoMockRecorder) GetBaseOnline(c, now interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBaseOnline", reflect.TypeOf((*MockDao)(nil).GetBaseOnline), c, now)
}

// GetRuleOnline mocks base method
func (m *MockDao) GetRuleOnline(c context.Context, baseID []int64, now time.Time) ([]*rank.Rule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRuleOnline", c, baseID, now)
	ret0, _ := ret[0].([]*rank.Rule)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRuleOnline indicates an expected call of GetRuleOnline
func (mr *MockDaoMockRecorder) GetRuleOnline(c, baseID, now interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRuleOnline", reflect.TypeOf((*MockDao)(nil).GetRuleOnline), c, baseID, now)
}

// GetRankLog mocks base method
func (m *MockDao) GetRankLog(c context.Context, rankID []int64, thisDate string) ([]*rank.Log, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRankLog", c, rankID, thisDate)
	ret0, _ := ret[0].([]*rank.Log)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRankLog indicates an expected call of GetRankLog
func (mr *MockDaoMockRecorder) GetRankLog(c, rankID, thisDate interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRankLog", reflect.TypeOf((*MockDao)(nil).GetRankLog), c, rankID, thisDate)
}

// InsertRankLog mocks base method
func (m *MockDao) InsertRankLog(c context.Context, rank []*rank.Log) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertRankLog", c, rank)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertRankLog indicates an expected call of InsertRankLog
func (mr *MockDaoMockRecorder) InsertRankLog(c, rank interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertRankLog", reflect.TypeOf((*MockDao)(nil).InsertRankLog), c, rank)
}
