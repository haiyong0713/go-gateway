// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/leelei/go/src/bapis-go/live/xroom/api.pb.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	v1 "git.bilibili.co/bapis/bapis-go/live/xroom"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockRoomClient is a mock of RoomClient interface.
type MockRoomClient struct {
	ctrl     *gomock.Controller
	recorder *MockRoomClientMockRecorder
}

// MockRoomClientMockRecorder is the mock recorder for MockRoomClient.
type MockRoomClientMockRecorder struct {
	mock *MockRoomClient
}

// NewMockRoomClient creates a new mock instance.
func NewMockRoomClient(ctrl *gomock.Controller) *MockRoomClient {
	mock := &MockRoomClient{ctrl: ctrl}
	mock.recorder = &MockRoomClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoomClient) EXPECT() *MockRoomClientMockRecorder {
	return m.recorder
}

// GetMultiple mocks base method.
func (m *MockRoomClient) GetMultiple(ctx context.Context, in *v1.RoomIDsReq, opts ...grpc.CallOption) (*v1.RoomIDsInfosResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMultiple", varargs...)
	ret0, _ := ret[0].(*v1.RoomIDsInfosResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMultiple indicates an expected call of GetMultiple.
func (mr *MockRoomClientMockRecorder) GetMultiple(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMultiple", reflect.TypeOf((*MockRoomClient)(nil).GetMultiple), varargs...)
}

// GetMultipleByUids mocks base method.
func (m *MockRoomClient) GetMultipleByUids(ctx context.Context, in *v1.UIDsReq, opts ...grpc.CallOption) (*v1.UIDsInfosResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetMultipleByUids", varargs...)
	ret0, _ := ret[0].(*v1.UIDsInfosResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMultipleByUids indicates an expected call of GetMultipleByUids.
func (mr *MockRoomClientMockRecorder) GetMultipleByUids(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMultipleByUids", reflect.TypeOf((*MockRoomClient)(nil).GetMultipleByUids), varargs...)
}

// IsAnchor mocks base method.
func (m *MockRoomClient) IsAnchor(ctx context.Context, in *v1.IsAnchorUIDsReq, opts ...grpc.CallOption) (*v1.IsAnchorUIDsResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "IsAnchor", varargs...)
	ret0, _ := ret[0].(*v1.IsAnchorUIDsResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAnchor indicates an expected call of IsAnchor.
func (mr *MockRoomClientMockRecorder) IsAnchor(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAnchor", reflect.TypeOf((*MockRoomClient)(nil).IsAnchor), varargs...)
}

// GetAreaInfo mocks base method.
func (m *MockRoomClient) GetAreaInfo(ctx context.Context, in *v1.AreaInfoReq, opts ...grpc.CallOption) (*v1.AreaInfoResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAreaInfo", varargs...)
	ret0, _ := ret[0].(*v1.AreaInfoResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAreaInfo indicates an expected call of GetAreaInfo.
func (mr *MockRoomClientMockRecorder) GetAreaInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAreaInfo", reflect.TypeOf((*MockRoomClient)(nil).GetAreaInfo), varargs...)
}

// GetPendantByRoomIds mocks base method.
func (m *MockRoomClient) GetPendantByRoomIds(ctx context.Context, in *v1.PendantReq, opts ...grpc.CallOption) (*v1.PendantResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPendantByRoomIds", varargs...)
	ret0, _ := ret[0].(*v1.PendantResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendantByRoomIds indicates an expected call of GetPendantByRoomIds.
func (mr *MockRoomClientMockRecorder) GetPendantByRoomIds(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendantByRoomIds", reflect.TypeOf((*MockRoomClient)(nil).GetPendantByRoomIds), varargs...)
}

// GetStatusInfoByUids mocks base method.
func (m *MockRoomClient) GetStatusInfoByUids(ctx context.Context, in *v1.GetStatusInfoByUidsReq, opts ...grpc.CallOption) (*v1.GetStatusInfoByUidResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetStatusInfoByUids", varargs...)
	ret0, _ := ret[0].(*v1.GetStatusInfoByUidResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatusInfoByUids indicates an expected call of GetStatusInfoByUids.
func (mr *MockRoomClientMockRecorder) GetStatusInfoByUids(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatusInfoByUids", reflect.TypeOf((*MockRoomClient)(nil).GetStatusInfoByUids), varargs...)
}

// GetRoomPlayInfo mocks base method.
func (m *MockRoomClient) GetRoomPlayInfo(ctx context.Context, in *v1.GetRoomPlayReq, opts ...grpc.CallOption) (*v1.GetRoomPlayResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRoomPlayInfo", varargs...)
	ret0, _ := ret[0].(*v1.GetRoomPlayResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRoomPlayInfo indicates an expected call of GetRoomPlayInfo.
func (mr *MockRoomClientMockRecorder) GetRoomPlayInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoomPlayInfo", reflect.TypeOf((*MockRoomClient)(nil).GetRoomPlayInfo), varargs...)
}

// GetLivKeyByRoomId mocks base method.
func (m *MockRoomClient) GetLivKeyByRoomId(ctx context.Context, in *v1.GetLivKeyByRoomIdReq, opts ...grpc.CallOption) (*v1.GetLivKeyByRoomIdResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetLivKeyByRoomId", varargs...)
	ret0, _ := ret[0].(*v1.GetLivKeyByRoomIdResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLivKeyByRoomId indicates an expected call of GetLivKeyByRoomId.
func (mr *MockRoomClientMockRecorder) GetLivKeyByRoomId(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLivKeyByRoomId", reflect.TypeOf((*MockRoomClient)(nil).GetLivKeyByRoomId), varargs...)
}

// GetRecordTranscodeConf mocks base method.
func (m *MockRoomClient) GetRecordTranscodeConf(ctx context.Context, in *v1.GetRecordTranscodeConfReq, opts ...grpc.CallOption) (*v1.GetRecordTranscodeConfResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetRecordTranscodeConf", varargs...)
	ret0, _ := ret[0].(*v1.GetRecordTranscodeConfResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecordTranscodeConf indicates an expected call of GetRecordTranscodeConf.
func (mr *MockRoomClientMockRecorder) GetRecordTranscodeConf(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordTranscodeConf", reflect.TypeOf((*MockRoomClient)(nil).GetRecordTranscodeConf), varargs...)
}

// GetAllLivingRoomsInfo mocks base method.
func (m *MockRoomClient) GetAllLivingRoomsInfo(ctx context.Context, in *v1.AllLivingRoomsInfoReq, opts ...grpc.CallOption) (*v1.AllLivingRoomsInfoResp, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAllLivingRoomsInfo", varargs...)
	ret0, _ := ret[0].(*v1.AllLivingRoomsInfoResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllLivingRoomsInfo indicates an expected call of GetAllLivingRoomsInfo.
func (mr *MockRoomClientMockRecorder) GetAllLivingRoomsInfo(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllLivingRoomsInfo", reflect.TypeOf((*MockRoomClient)(nil).GetAllLivingRoomsInfo), varargs...)
}

// MockRoomServer is a mock of RoomServer interface.
type MockRoomServer struct {
	ctrl     *gomock.Controller
	recorder *MockRoomServerMockRecorder
}

// MockRoomServerMockRecorder is the mock recorder for MockRoomServer.
type MockRoomServerMockRecorder struct {
	mock *MockRoomServer
}

// NewMockRoomServer creates a new mock instance.
func NewMockRoomServer(ctrl *gomock.Controller) *MockRoomServer {
	mock := &MockRoomServer{ctrl: ctrl}
	mock.recorder = &MockRoomServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoomServer) EXPECT() *MockRoomServerMockRecorder {
	return m.recorder
}

// GetMultiple mocks base method.
func (m *MockRoomServer) GetMultiple(arg0 context.Context, arg1 *v1.RoomIDsReq) (*v1.RoomIDsInfosResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMultiple", arg0, arg1)
	ret0, _ := ret[0].(*v1.RoomIDsInfosResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMultiple indicates an expected call of GetMultiple.
func (mr *MockRoomServerMockRecorder) GetMultiple(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMultiple", reflect.TypeOf((*MockRoomServer)(nil).GetMultiple), arg0, arg1)
}

// GetMultipleByUids mocks base method.
func (m *MockRoomServer) GetMultipleByUids(arg0 context.Context, arg1 *v1.UIDsReq) (*v1.UIDsInfosResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMultipleByUids", arg0, arg1)
	ret0, _ := ret[0].(*v1.UIDsInfosResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMultipleByUids indicates an expected call of GetMultipleByUids.
func (mr *MockRoomServerMockRecorder) GetMultipleByUids(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMultipleByUids", reflect.TypeOf((*MockRoomServer)(nil).GetMultipleByUids), arg0, arg1)
}

// IsAnchor mocks base method.
func (m *MockRoomServer) IsAnchor(arg0 context.Context, arg1 *v1.IsAnchorUIDsReq) (*v1.IsAnchorUIDsResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAnchor", arg0, arg1)
	ret0, _ := ret[0].(*v1.IsAnchorUIDsResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAnchor indicates an expected call of IsAnchor.
func (mr *MockRoomServerMockRecorder) IsAnchor(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAnchor", reflect.TypeOf((*MockRoomServer)(nil).IsAnchor), arg0, arg1)
}

// GetAreaInfo mocks base method.
func (m *MockRoomServer) GetAreaInfo(arg0 context.Context, arg1 *v1.AreaInfoReq) (*v1.AreaInfoResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAreaInfo", arg0, arg1)
	ret0, _ := ret[0].(*v1.AreaInfoResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAreaInfo indicates an expected call of GetAreaInfo.
func (mr *MockRoomServerMockRecorder) GetAreaInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAreaInfo", reflect.TypeOf((*MockRoomServer)(nil).GetAreaInfo), arg0, arg1)
}

// GetPendantByRoomIds mocks base method.
func (m *MockRoomServer) GetPendantByRoomIds(arg0 context.Context, arg1 *v1.PendantReq) (*v1.PendantResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPendantByRoomIds", arg0, arg1)
	ret0, _ := ret[0].(*v1.PendantResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPendantByRoomIds indicates an expected call of GetPendantByRoomIds.
func (mr *MockRoomServerMockRecorder) GetPendantByRoomIds(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPendantByRoomIds", reflect.TypeOf((*MockRoomServer)(nil).GetPendantByRoomIds), arg0, arg1)
}

// GetStatusInfoByUids mocks base method.
func (m *MockRoomServer) GetStatusInfoByUids(arg0 context.Context, arg1 *v1.GetStatusInfoByUidsReq) (*v1.GetStatusInfoByUidResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatusInfoByUids", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetStatusInfoByUidResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatusInfoByUids indicates an expected call of GetStatusInfoByUids.
func (mr *MockRoomServerMockRecorder) GetStatusInfoByUids(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatusInfoByUids", reflect.TypeOf((*MockRoomServer)(nil).GetStatusInfoByUids), arg0, arg1)
}

// GetRoomPlayInfo mocks base method.
func (m *MockRoomServer) GetRoomPlayInfo(arg0 context.Context, arg1 *v1.GetRoomPlayReq) (*v1.GetRoomPlayResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRoomPlayInfo", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetRoomPlayResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRoomPlayInfo indicates an expected call of GetRoomPlayInfo.
func (mr *MockRoomServerMockRecorder) GetRoomPlayInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRoomPlayInfo", reflect.TypeOf((*MockRoomServer)(nil).GetRoomPlayInfo), arg0, arg1)
}

// GetLivKeyByRoomId mocks base method.
func (m *MockRoomServer) GetLivKeyByRoomId(arg0 context.Context, arg1 *v1.GetLivKeyByRoomIdReq) (*v1.GetLivKeyByRoomIdResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLivKeyByRoomId", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetLivKeyByRoomIdResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLivKeyByRoomId indicates an expected call of GetLivKeyByRoomId.
func (mr *MockRoomServerMockRecorder) GetLivKeyByRoomId(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLivKeyByRoomId", reflect.TypeOf((*MockRoomServer)(nil).GetLivKeyByRoomId), arg0, arg1)
}

// GetRecordTranscodeConf mocks base method.
func (m *MockRoomServer) GetRecordTranscodeConf(arg0 context.Context, arg1 *v1.GetRecordTranscodeConfReq) (*v1.GetRecordTranscodeConfResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecordTranscodeConf", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetRecordTranscodeConfResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecordTranscodeConf indicates an expected call of GetRecordTranscodeConf.
func (mr *MockRoomServerMockRecorder) GetRecordTranscodeConf(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordTranscodeConf", reflect.TypeOf((*MockRoomServer)(nil).GetRecordTranscodeConf), arg0, arg1)
}

// GetAllLivingRoomsInfo mocks base method.
func (m *MockRoomServer) GetAllLivingRoomsInfo(arg0 context.Context, arg1 *v1.AllLivingRoomsInfoReq) (*v1.AllLivingRoomsInfoResp, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllLivingRoomsInfo", arg0, arg1)
	ret0, _ := ret[0].(*v1.AllLivingRoomsInfoResp)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllLivingRoomsInfo indicates an expected call of GetAllLivingRoomsInfo.
func (mr *MockRoomServerMockRecorder) GetAllLivingRoomsInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllLivingRoomsInfo", reflect.TypeOf((*MockRoomServer)(nil).GetAllLivingRoomsInfo), arg0, arg1)
}
