// Code generated by MockGen. DO NOT EDIT.
// Source: dao.go

// Package mock_dao is a generated GoMock package.
package mock_dao

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	api "go-gateway/app/api-gateway/api-manager/api"
	model "go-gateway/app/api-gateway/api-manager/internal/model"
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

// Ping mocks base method
func (m *MockDao) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockDaoMockRecorder) Ping(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockDao)(nil).Ping), ctx)
}

// AddApi mocks base method
func (m *MockDao) AddApi(c context.Context, apis *model.ApiRawInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddApi", c, apis)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddApi indicates an expected call of AddApi
func (mr *MockDaoMockRecorder) AddApi(c, apis interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddApi", reflect.TypeOf((*MockDao)(nil).AddApi), c, apis)
}

// GetHttpApis mocks base method
func (m *MockDao) GetHttpApis(c context.Context) ([]*model.ApiRawInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHttpApis", c)
	ret0, _ := ret[0].([]*model.ApiRawInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHttpApis indicates an expected call of GetHttpApis
func (mr *MockDaoMockRecorder) GetHttpApis(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHttpApis", reflect.TypeOf((*MockDao)(nil).GetHttpApis), c)
}

// GetGrpcApis mocks base method
func (m *MockDao) GetGrpcApis(c context.Context, discoveryID string) ([]*model.ApiRawInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGrpcApis", c, discoveryID)
	ret0, _ := ret[0].([]*model.ApiRawInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGrpcApis indicates an expected call of GetGrpcApis
func (mr *MockDaoMockRecorder) GetGrpcApis(c, discoveryID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGrpcApis", reflect.TypeOf((*MockDao)(nil).GetGrpcApis), c, discoveryID)
}

// UpApi mocks base method
func (m *MockDao) UpApi(c context.Context, id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpApi", c, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpApi indicates an expected call of UpApi
func (mr *MockDaoMockRecorder) UpApi(c, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpApi", reflect.TypeOf((*MockDao)(nil).UpApi), c, id)
}

// GetHttpApisByPath mocks base method
func (m *MockDao) GetHttpApisByPath(c context.Context, paths []string) (map[string]*api.ApiInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHttpApisByPath", c, paths)
	ret0, _ := ret[0].(map[string]*api.ApiInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHttpApisByPath indicates an expected call of GetHttpApisByPath
func (mr *MockDaoMockRecorder) GetHttpApisByPath(c, paths interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHttpApisByPath", reflect.TypeOf((*MockDao)(nil).GetHttpApisByPath), c, paths)
}

// GetServiceName mocks base method
func (m *MockDao) GetServiceName(c context.Context, discoveryIDs []string) (map[string][]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetServiceName", c, discoveryIDs)
	ret0, _ := ret[0].(map[string][]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetServiceName indicates an expected call of GetServiceName
func (mr *MockDaoMockRecorder) GetServiceName(c, discoveryIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetServiceName", reflect.TypeOf((*MockDao)(nil).GetServiceName), c, discoveryIDs)
}

// AddProto mocks base method
func (m *MockDao) AddProto(c context.Context, pros *model.ProtoInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddProto", c, pros)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddProto indicates an expected call of AddProto
func (mr *MockDaoMockRecorder) AddProto(c, pros interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddProto", reflect.TypeOf((*MockDao)(nil).AddProto), c, pros)
}

// GetAllProtos mocks base method
func (m *MockDao) GetAllProtos(c context.Context) ([]*model.ProtoInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllProtos", c)
	ret0, _ := ret[0].([]*model.ProtoInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllProtos indicates an expected call of GetAllProtos
func (mr *MockDaoMockRecorder) GetAllProtos(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllProtos", reflect.TypeOf((*MockDao)(nil).GetAllProtos), c)
}

// GetProto mocks base method
func (m *MockDao) GetProto(c context.Context, discoveryID string) ([]*model.ProtoInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProto", c, discoveryID)
	ret0, _ := ret[0].([]*model.ProtoInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProto indicates an expected call of GetProto
func (mr *MockDaoMockRecorder) GetProto(c, discoveryID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProto", reflect.TypeOf((*MockDao)(nil).GetProto), c, discoveryID)
}

// GetProtoByDis mocks base method
func (m *MockDao) GetProtoByDis(c context.Context, discoveryIDs []string) (map[string]*api.ApiInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProtoByDis", c, discoveryIDs)
	ret0, _ := ret[0].(map[string]*api.ApiInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProtoByDis indicates an expected call of GetProtoByDis
func (mr *MockDaoMockRecorder) GetProtoByDis(c, discoveryIDs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProtoByDis", reflect.TypeOf((*MockDao)(nil).GetProtoByDis), c, discoveryIDs)
}