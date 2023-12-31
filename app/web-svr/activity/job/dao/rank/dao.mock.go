// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package rank is a generated GoMock package.
package rank

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	rank "go-gateway/app/web-svr/activity/job/model/rank"
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

// BatchAddRank mocks base method
func (m *MockDao) BatchAddRank(c context.Context, rank []*rank.DB) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchAddRank", c, rank)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchAddRank indicates an expected call of BatchAddRank
func (mr *MockDaoMockRecorder) BatchAddRank(c, rank interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchAddRank", reflect.TypeOf((*MockDao)(nil).BatchAddRank), c, rank)
}

// SetRank mocks base method
func (m *MockDao) SetRank(c context.Context, rankNameKey string, rankBatch []*rank.Redis) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetRank", c, rankNameKey, rankBatch)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetRank indicates an expected call of SetRank
func (mr *MockDaoMockRecorder) SetRank(c, rankNameKey, rankBatch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetRank", reflect.TypeOf((*MockDao)(nil).SetRank), c, rankNameKey, rankBatch)
}

// GetRank mocks base method
func (m *MockDao) GetRank(c context.Context, rankNameKey string) ([]*rank.Redis, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRank", c, rankNameKey)
	ret0, _ := ret[0].([]*rank.Redis)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRank indicates an expected call of GetRank
func (mr *MockDaoMockRecorder) GetRank(c, rankNameKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRank", reflect.TypeOf((*MockDao)(nil).GetRank), c, rankNameKey)
}

// SetMidRank mocks base method
func (m *MockDao) SetMidRank(c context.Context, rankNameKey string, midRank []*rank.Redis) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMidRank", c, rankNameKey, midRank)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetMidRank indicates an expected call of SetMidRank
func (mr *MockDaoMockRecorder) SetMidRank(c, rankNameKey, midRank interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMidRank", reflect.TypeOf((*MockDao)(nil).SetMidRank), c, rankNameKey, midRank)
}

// GetRankListByBatch mocks base method
func (m *MockDao) GetRankListByBatch(c context.Context, sid, batch int64) ([]*rank.DB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRankListByBatch", c, sid, batch)
	ret0, _ := ret[0].([]*rank.DB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRankListByBatch indicates an expected call of GetRankListByBatch
func (mr *MockDaoMockRecorder) GetRankListByBatch(c, sid, batch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRankListByBatch", reflect.TypeOf((*MockDao)(nil).GetRankListByBatch), c, sid, batch)
}

// GetMemberRankTimes mocks base method
func (m *MockDao) GetMemberRankTimes(c context.Context, sid, startBatch, endBatch int64, mids []int64) ([]*rank.MemberRankTimes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMemberRankTimes", c, sid, startBatch, endBatch, mids)
	ret0, _ := ret[0].([]*rank.MemberRankTimes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMemberRankTimes indicates an expected call of GetMemberRankTimes
func (mr *MockDaoMockRecorder) GetMemberRankTimes(c, sid, startBatch, endBatch, mids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMemberRankTimes", reflect.TypeOf((*MockDao)(nil).GetMemberRankTimes), c, sid, startBatch, endBatch, mids)
}

// GetMemberHighest mocks base method
func (m *MockDao) GetMemberHighest(c context.Context, sid, startBatch, endBatch int64, mids []int64) ([]*rank.MemberRankHighest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMemberHighest", c, sid, startBatch, endBatch, mids)
	ret0, _ := ret[0].([]*rank.MemberRankHighest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMemberHighest indicates an expected call of GetMemberHighest
func (mr *MockDaoMockRecorder) GetMemberHighest(c, sid, startBatch, endBatch, mids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMemberHighest", reflect.TypeOf((*MockDao)(nil).GetMemberHighest), c, sid, startBatch, endBatch, mids)
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
