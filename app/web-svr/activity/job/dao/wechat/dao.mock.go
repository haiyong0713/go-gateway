// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package wechat is a generated GoMock package.
package wechat

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
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

// SendWeChat mocks base method
func (m *MockDao) SendWeChat(c context.Context, publicKey, title, msg, user string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendWeChat", c, publicKey, title, msg, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendWeChat indicates an expected call of SendWeChat
func (mr *MockDaoMockRecorder) SendWeChat(c, publicKey, title, msg, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendWeChat", reflect.TypeOf((*MockDao)(nil).SendWeChat), c, publicKey, title, msg, user)
}
