// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: example/laser_cmd.proto

package laser

import (
	context "context"
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	proto "github.com/gogo/protobuf/proto"
	types "github.com/gogo/protobuf/types"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// 服务端下发Laser事件
type LaserEventResp struct {
	// 任务id
	Taskid int64 `protobuf:"varint,1,opt,name=taskid,proto3" json:"taskid,omitempty"`
	// 指令名
	Action string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	// 指令参数json字符串
	Params               string   `protobuf:"bytes,3,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LaserEventResp) Reset()         { *m = LaserEventResp{} }
func (m *LaserEventResp) String() string { return proto.CompactTextString(m) }
func (*LaserEventResp) ProtoMessage()    {}
func (*LaserEventResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_e6b768a3923bb6f7, []int{0}
}
func (m *LaserEventResp) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *LaserEventResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_LaserEventResp.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *LaserEventResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LaserEventResp.Merge(m, src)
}
func (m *LaserEventResp) XXX_Size() int {
	return m.Size()
}
func (m *LaserEventResp) XXX_DiscardUnknown() {
	xxx_messageInfo_LaserEventResp.DiscardUnknown(m)
}

var xxx_messageInfo_LaserEventResp proto.InternalMessageInfo

func (m *LaserEventResp) GetTaskid() int64 {
	if m != nil {
		return m.Taskid
	}
	return 0
}

func (m *LaserEventResp) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *LaserEventResp) GetParams() string {
	if m != nil {
		return m.Params
	}
	return ""
}

func init() {
	proto.RegisterType((*LaserEventResp)(nil), "bilibili.broadcast.v2.LaserEventResp")
}

func init() { proto.RegisterFile("example/laser_cmd.proto", fileDescriptor_e6b768a3923bb6f7) }

var fileDescriptor_e6b768a3923bb6f7 = []byte{
	// 275 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4f, 0xad, 0x48, 0xcc,
	0x2d, 0xc8, 0x49, 0xd5, 0xcf, 0x49, 0x2c, 0x4e, 0x2d, 0x8a, 0x4f, 0xce, 0x4d, 0xd1, 0x2b, 0x28,
	0xca, 0x2f, 0xc9, 0x17, 0x12, 0x4d, 0xca, 0xcc, 0xc9, 0x04, 0x61, 0xbd, 0xa4, 0xa2, 0xfc, 0xc4,
	0x94, 0xe4, 0xc4, 0xe2, 0x12, 0xbd, 0x32, 0x23, 0x29, 0xe9, 0xf4, 0xfc, 0xfc, 0xf4, 0x9c, 0x54,
	0x7d, 0xb0, 0xa2, 0xa4, 0xd2, 0x34, 0xfd, 0xd4, 0xdc, 0x82, 0x92, 0x4a, 0x88, 0x1e, 0xa5, 0x08,
	0x2e, 0x3e, 0x1f, 0x90, 0x31, 0xae, 0x65, 0xa9, 0x79, 0x25, 0x41, 0xa9, 0xc5, 0x05, 0x42, 0x62,
	0x5c, 0x6c, 0x25, 0x89, 0xc5, 0xd9, 0x99, 0x29, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0xcc, 0x41, 0x50,
	0x1e, 0x48, 0x3c, 0x31, 0xb9, 0x24, 0x33, 0x3f, 0x4f, 0x82, 0x49, 0x81, 0x51, 0x83, 0x33, 0x08,
	0xca, 0x03, 0x89, 0x17, 0x24, 0x16, 0x25, 0xe6, 0x16, 0x4b, 0x30, 0x43, 0xc4, 0x21, 0x3c, 0xa3,
	0x08, 0x2e, 0x76, 0xb0, 0xc9, 0x61, 0x46, 0x42, 0xbe, 0x5c, 0x5c, 0xe1, 0x89, 0x25, 0xc9, 0x19,
	0x60, 0x4b, 0x84, 0xc4, 0xf4, 0x20, 0x0e, 0xd2, 0x83, 0x39, 0x48, 0xcf, 0x15, 0xe4, 0x20, 0x29,
	0x55, 0x3d, 0xac, 0xee, 0xd7, 0x43, 0x75, 0x9f, 0x01, 0xa3, 0x53, 0xd5, 0x89, 0x47, 0x72, 0x8c,
	0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0xc8, 0x25, 0x9f, 0x9c, 0x9f, 0xab, 0x97, 0x94,
	0x58, 0x90, 0x59, 0x8c, 0x5d, 0x7f, 0x00, 0x63, 0x94, 0x79, 0x7a, 0x66, 0x09, 0x42, 0x32, 0x39,
	0x5f, 0x1f, 0xac, 0x1c, 0x42, 0xea, 0xa6, 0xe7, 0xeb, 0xc3, 0xa4, 0xf4, 0xe1, 0xfa, 0xf4, 0xcb,
	0x8c, 0xac, 0xcb, 0x8c, 0x16, 0x31, 0xf1, 0x3b, 0x39, 0x06, 0x78, 0x3a, 0xc1, 0x44, 0xc3, 0x8c,
	0x92, 0xd8, 0xc0, 0x8e, 0x36, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x2d, 0xcc, 0xfa, 0x0e, 0x85,
	0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// LaserV2Client is the client API for LaserV2 service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type LaserV2Client interface {
	// 监听Laser事件
	WatchEvent(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (LaserV2_WatchEventClient, error)
}

type laserV2Client struct {
	cc *grpc.ClientConn
}

func NewLaserV2Client(cc *grpc.ClientConn) LaserV2Client {
	return &laserV2Client{cc}
}

func (c *laserV2Client) WatchEvent(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (LaserV2_WatchEventClient, error) {
	stream, err := c.cc.NewStream(ctx, &_LaserV2_serviceDesc.Streams[0], "/bilibili.broadcast.v2.LaserV2/WatchEvent", opts...)
	if err != nil {
		return nil, err
	}
	x := &laserV2WatchEventClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type LaserV2_WatchEventClient interface {
	Recv() (*LaserEventResp, error)
	grpc.ClientStream
}

type laserV2WatchEventClient struct {
	grpc.ClientStream
}

func (x *laserV2WatchEventClient) Recv() (*LaserEventResp, error) {
	m := new(LaserEventResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// LaserV2Server is the server API for LaserV2 service.
type LaserV2Server interface {
	// 监听Laser事件
	WatchEvent(*types.Empty, LaserV2_WatchEventServer) error
}

// UnimplementedLaserV2Server can be embedded to have forward compatible implementations.
type UnimplementedLaserV2Server struct {
}

func (*UnimplementedLaserV2Server) WatchEvent(req *types.Empty, srv LaserV2_WatchEventServer) error {
	return status.Errorf(codes.Unimplemented, "method WatchEvent not implemented")
}

func RegisterLaserV2Server(s *grpc.Server, srv LaserV2Server) {
	s.RegisterService(&_LaserV2_serviceDesc, srv)
}

func _LaserV2_WatchEvent_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(types.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LaserV2Server).WatchEvent(m, &laserV2WatchEventServer{stream})
}

type LaserV2_WatchEventServer interface {
	Send(*LaserEventResp) error
	grpc.ServerStream
}

type laserV2WatchEventServer struct {
	grpc.ServerStream
}

func (x *laserV2WatchEventServer) Send(m *LaserEventResp) error {
	return x.ServerStream.SendMsg(m)
}

var _LaserV2_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bilibili.broadcast.v2.LaserV2",
	HandlerType: (*LaserV2Server)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "WatchEvent",
			Handler:       _LaserV2_WatchEvent_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "example/laser_cmd.proto",
}

func (m *LaserEventResp) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *LaserEventResp) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *LaserEventResp) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Params) > 0 {
		i -= len(m.Params)
		copy(dAtA[i:], m.Params)
		i = encodeVarintLaserCmd(dAtA, i, uint64(len(m.Params)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Action) > 0 {
		i -= len(m.Action)
		copy(dAtA[i:], m.Action)
		i = encodeVarintLaserCmd(dAtA, i, uint64(len(m.Action)))
		i--
		dAtA[i] = 0x12
	}
	if m.Taskid != 0 {
		i = encodeVarintLaserCmd(dAtA, i, uint64(m.Taskid))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintLaserCmd(dAtA []byte, offset int, v uint64) int {
	offset -= sovLaserCmd(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *LaserEventResp) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Taskid != 0 {
		n += 1 + sovLaserCmd(uint64(m.Taskid))
	}
	l = len(m.Action)
	if l > 0 {
		n += 1 + l + sovLaserCmd(uint64(l))
	}
	l = len(m.Params)
	if l > 0 {
		n += 1 + l + sovLaserCmd(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovLaserCmd(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLaserCmd(x uint64) (n int) {
	return sovLaserCmd(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *LaserEventResp) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLaserCmd
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: LaserEventResp: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LaserEventResp: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Taskid", wireType)
			}
			m.Taskid = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLaserCmd
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Taskid |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Action", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLaserCmd
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthLaserCmd
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLaserCmd
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Action = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLaserCmd
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthLaserCmd
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthLaserCmd
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Params = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLaserCmd(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLaserCmd
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLaserCmd(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLaserCmd
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLaserCmd
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLaserCmd
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthLaserCmd
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLaserCmd
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLaserCmd
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLaserCmd        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLaserCmd          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLaserCmd = fmt.Errorf("proto: unexpected end of group")
)
