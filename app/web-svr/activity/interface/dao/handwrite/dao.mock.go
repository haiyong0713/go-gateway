// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package handwrite is a generated GoMock package.
package handwrite

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	handwrite "go-gateway/app/web-svr/activity/interface/model/handwrite"
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

// GetMidAward mocks base method
func (m *MockDao) GetMidAward(c context.Context, mid int64) (*handwrite.MidAward, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMidAward", c, mid)
	ret0, _ := ret[0].(*handwrite.MidAward)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMidAward indicates an expected call of GetMidAward
func (mr *MockDaoMockRecorder) GetMidAward(c, mid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMidAward", reflect.TypeOf((*MockDao)(nil).GetMidAward), c, mid)
}

// GetAwardCount mocks base method
func (m *MockDao) GetAwardCount(c context.Context) (*handwrite.AwardCount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAwardCount", c)
	ret0, _ := ret[0].(*handwrite.AwardCount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAwardCount indicates an expected call of GetAwardCount
func (mr *MockDaoMockRecorder) GetAwardCount(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAwardCount", reflect.TypeOf((*MockDao)(nil).GetAwardCount), c)
}

// AddTimeLock mocks base method
func (m *MockDao) AddTimeLock(c context.Context, mid int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTimeLock", c, mid)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTimeLock indicates an expected call of AddTimeLock
func (mr *MockDaoMockRecorder) AddTimeLock(c, mid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTimeLock", reflect.TypeOf((*MockDao)(nil).AddTimeLock), c, mid)
}

// AddTimesRecord mocks base method
func (m *MockDao) AddTimesRecord(c context.Context, mid int64, day string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddTimesRecord", c, mid, day)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddTimesRecord indicates an expected call of AddTimesRecord
func (mr *MockDaoMockRecorder) AddTimesRecord(c, mid, day interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddTimesRecord", reflect.TypeOf((*MockDao)(nil).AddTimesRecord), c, mid, day)
}

// GetAddTimesRecord mocks base method
func (m *MockDao) GetAddTimesRecord(c context.Context, mid int64, day string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAddTimesRecord", c, mid, day)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAddTimesRecord indicates an expected call of GetAddTimesRecord
func (mr *MockDaoMockRecorder) GetAddTimesRecord(c, mid, day interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAddTimesRecord", reflect.TypeOf((*MockDao)(nil).GetAddTimesRecord), c, mid, day)
}