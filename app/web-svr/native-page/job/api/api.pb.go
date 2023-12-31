// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api.proto

// package 命名使用 {appid}.{version} 的方式, version 形如 v1, v2 ..

package api

import (
	context "context"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("api.proto", fileDescriptor_00212fb1f9d3bf1c) }

var fileDescriptor_00212fb1f9d3bf1c = []byte{
	// 197 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4c, 0x2c, 0xc8, 0xd4,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0xca, 0x4b, 0x2c, 0xc9, 0x2c, 0x4b, 0xd5, 0x2b, 0x48,
	0x4c, 0x4f, 0xd5, 0xcb, 0xca, 0x4f, 0xd2, 0x2b, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b,
	0x33, 0x94, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf,
	0x4f, 0xcf, 0xd7, 0x07, 0x6b, 0x49, 0x2a, 0x4d, 0x03, 0xf3, 0xc0, 0x1c, 0x30, 0x0b, 0x62, 0x94,
	0x94, 0x74, 0x7a, 0x7e, 0x7e, 0x7a, 0x4e, 0x2a, 0x42, 0x55, 0x6a, 0x6e, 0x41, 0x49, 0x25, 0x54,
	0x52, 0x06, 0x2a, 0x99, 0x58, 0x90, 0xa9, 0x9f, 0x98, 0x97, 0x97, 0x5f, 0x92, 0x58, 0x92, 0x99,
	0x9f, 0x57, 0x0c, 0x91, 0x35, 0x72, 0xe7, 0xe2, 0xf5, 0x03, 0xbb, 0x23, 0x20, 0x31, 0x3d, 0xd5,
	0x2b, 0x3f, 0x49, 0xc8, 0x8c, 0x8b, 0x25, 0x20, 0x33, 0x2f, 0x5d, 0x48, 0x4c, 0x0f, 0xa2, 0x4f,
	0x0f, 0x66, 0xa8, 0x9e, 0x2b, 0xc8, 0x50, 0x29, 0x1c, 0xe2, 0x4e, 0xa2, 0x27, 0x1e, 0xca, 0x31,
	0x9c, 0x78, 0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x51, 0xcc, 0x89,
	0x05, 0x99, 0x49, 0x6c, 0x60, 0x65, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb6, 0x55, 0x33,
	0x32, 0xf9, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NativePageJobClient is the client API for NativePageJob service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NativePageJobClient interface {
	Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
}

type nativePageJobClient struct {
	cc *grpc.ClientConn
}

func NewNativePageJobClient(cc *grpc.ClientConn) NativePageJobClient {
	return &nativePageJobClient{cc}
}

func (c *nativePageJobClient) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/native.page.job.service.v1.NativePageJob/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NativePageJobServer is the server API for NativePageJob service.
type NativePageJobServer interface {
	Ping(context.Context, *empty.Empty) (*empty.Empty, error)
}

// UnimplementedNativePageJobServer can be embedded to have forward compatible implementations.
type UnimplementedNativePageJobServer struct {
}

func (*UnimplementedNativePageJobServer) Ping(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func RegisterNativePageJobServer(s *grpc.Server, srv NativePageJobServer) {
	s.RegisterService(&_NativePageJob_serviceDesc, srv)
}

func _NativePageJob_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NativePageJobServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/native.page.job.service.v1.NativePageJob/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NativePageJobServer).Ping(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _NativePageJob_serviceDesc = grpc.ServiceDesc{
	ServiceName: "native.page.job.service.v1.NativePageJob",
	HandlerType: (*NativePageJobServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _NativePageJob_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
