// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: go-gateway/app/app-svr/archive/middleware/v1/preload.proto

package v1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PlayerArgs struct {
	// 清晰度
	Qn int64 `protobuf:"varint,1,opt,name=qn,proto3" json:"qn,omitempty"`
	// 功能版本号
	Fnver int64 `protobuf:"varint,2,opt,name=fnver,proto3" json:"fnver,omitempty"`
	// 功能标识
	// fnver  | fnval  | 功能
	// 0          0       flv请求，优先返回flv格式视频地址
	// 0          1       flv请求，只返回mp4格式的视频地址
	// 0          16      优先返回DASH-H265视频的JSON内容
	// 0          64      设备支持HDR 视频播放，此位为0，代表不支持HDR，为1，代表支持HDR
	// 0          128     是否需要4k视频，此位为0，代表不需要4k视频，为1，代表需要4k视频
	// 0          256     是否需要杜比音频，此位为0，代表不需要杜比音频，为1，代表需要杜比音频
	// fnval 每位(为1)标识一个功能, 其中HDR/4K位数 与 视频格式位数是可 或 关系，如：80 (01010000) 代表需要请求DASH格式的视频且设备支持HDR
	Fnval int64 `protobuf:"varint,3,opt,name=fnval,proto3" json:"fnval,omitempty"`
	// 返回url是否强制使用域名 1-http域名 2-https域名
	ForceHost int64 `protobuf:"varint,4,opt,name=force_host,json=forceHost,proto3" json:"force_host,omitempty"`
	//是否开启音量均衡，1开启
	VoiceBalance         int64    `protobuf:"varint,5,opt,name=voice_balance,json=voiceBalance,proto3" json:"voice_balance,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PlayerArgs) Reset()         { *m = PlayerArgs{} }
func (m *PlayerArgs) String() string { return proto.CompactTextString(m) }
func (*PlayerArgs) ProtoMessage()    {}
func (*PlayerArgs) Descriptor() ([]byte, []int) {
	return fileDescriptor_b4e52bc9f39ebe46, []int{0}
}
func (m *PlayerArgs) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PlayerArgs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PlayerArgs.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PlayerArgs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlayerArgs.Merge(m, src)
}
func (m *PlayerArgs) XXX_Size() int {
	return m.Size()
}
func (m *PlayerArgs) XXX_DiscardUnknown() {
	xxx_messageInfo_PlayerArgs.DiscardUnknown(m)
}

var xxx_messageInfo_PlayerArgs proto.InternalMessageInfo

func (m *PlayerArgs) GetQn() int64 {
	if m != nil {
		return m.Qn
	}
	return 0
}

func (m *PlayerArgs) GetFnver() int64 {
	if m != nil {
		return m.Fnver
	}
	return 0
}

func (m *PlayerArgs) GetFnval() int64 {
	if m != nil {
		return m.Fnval
	}
	return 0
}

func (m *PlayerArgs) GetForceHost() int64 {
	if m != nil {
		return m.ForceHost
	}
	return 0
}

func (m *PlayerArgs) GetVoiceBalance() int64 {
	if m != nil {
		return m.VoiceBalance
	}
	return 0
}

func init() {
	proto.RegisterType((*PlayerArgs)(nil), "bilibili.app.archive.middleware.v1.PlayerArgs")
}

func init() {
	proto.RegisterFile("go-gateway/app/app-svr/archive/middleware/v1/preload.proto", fileDescriptor_b4e52bc9f39ebe46)
}

var fileDescriptor_b4e52bc9f39ebe46 = []byte{
	// 270 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x71, 0x4a, 0x91, 0xb0, 0x80, 0x21, 0x62, 0x88, 0x90, 0x88, 0x50, 0x59, 0x18, 0xa8,
	0xad, 0x88, 0x8d, 0x2d, 0x99, 0x60, 0x40, 0x8a, 0x18, 0x18, 0x58, 0xaa, 0x8b, 0x73, 0x4d, 0x2d,
	0xb9, 0xb1, 0x6b, 0x47, 0xae, 0xfa, 0x0e, 0x3c, 0x00, 0x33, 0x4f, 0xc3, 0xc8, 0x23, 0xa0, 0xf0,
	0x22, 0xa8, 0x4e, 0x55, 0x46, 0x86, 0x7f, 0xb8, 0xef, 0xbe, 0x7f, 0xb8, 0xa3, 0xf7, 0x8d, 0x9e,
	0x36, 0xd0, 0xe1, 0x1a, 0x36, 0x1c, 0x8c, 0xd9, 0x66, 0xea, 0xbc, 0xe5, 0x60, 0xc5, 0x42, 0x7a,
	0xe4, 0x4b, 0x59, 0xd7, 0x0a, 0xd7, 0x60, 0x91, 0xfb, 0x8c, 0x1b, 0x8b, 0x4a, 0x43, 0xcd, 0x8c,
	0xd5, 0x9d, 0x8e, 0x27, 0x95, 0x54, 0x72, 0x1b, 0x06, 0xc6, 0xb0, 0x5d, 0x83, 0xfd, 0x35, 0x98,
	0xcf, 0x26, 0x6f, 0x84, 0xd2, 0x52, 0xc1, 0x06, 0x6d, 0x6e, 0x1b, 0x17, 0x9f, 0xd1, 0x68, 0xd5,
	0x26, 0xe4, 0x8a, 0xdc, 0x8c, 0x9e, 0xa3, 0x55, 0x1b, 0x9f, 0xd3, 0xf1, 0xbc, 0xf5, 0x68, 0x93,
	0x28, 0xa0, 0x61, 0xd8, 0x51, 0x50, 0xc9, 0x68, 0x4f, 0x41, 0xc5, 0x97, 0x94, 0xce, 0xb5, 0x15,
	0x38, 0x5b, 0x68, 0xd7, 0x25, 0x87, 0x61, 0x75, 0x1c, 0xc8, 0x83, 0x76, 0x5d, 0x7c, 0x4d, 0x4f,
	0xbd, 0x96, 0x02, 0x67, 0x15, 0x28, 0x68, 0x05, 0x26, 0xe3, 0x60, 0x9c, 0x04, 0x58, 0x0c, 0xac,
	0x80, 0xcf, 0x3e, 0x25, 0x5f, 0x7d, 0x4a, 0xbe, 0xfb, 0x94, 0xbc, 0xff, 0xa4, 0x07, 0xf4, 0x56,
	0xe8, 0x25, 0xab, 0xc0, 0x48, 0xc7, 0xfe, 0x3f, 0xa7, 0x24, 0xaf, 0x91, 0xcf, 0x3e, 0xa2, 0x8b,
	0x22, 0x2f, 0x1f, 0x73, 0x63, 0xf2, 0x41, 0x7a, 0xda, 0x3b, 0x2f, 0x59, 0x75, 0x14, 0x9e, 0x73,
	0xf7, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x7b, 0x37, 0x71, 0xfb, 0x5a, 0x01, 0x00, 0x00,
}

func (m *PlayerArgs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PlayerArgs) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PlayerArgs) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.VoiceBalance != 0 {
		i = encodeVarintPreload(dAtA, i, uint64(m.VoiceBalance))
		i--
		dAtA[i] = 0x28
	}
	if m.ForceHost != 0 {
		i = encodeVarintPreload(dAtA, i, uint64(m.ForceHost))
		i--
		dAtA[i] = 0x20
	}
	if m.Fnval != 0 {
		i = encodeVarintPreload(dAtA, i, uint64(m.Fnval))
		i--
		dAtA[i] = 0x18
	}
	if m.Fnver != 0 {
		i = encodeVarintPreload(dAtA, i, uint64(m.Fnver))
		i--
		dAtA[i] = 0x10
	}
	if m.Qn != 0 {
		i = encodeVarintPreload(dAtA, i, uint64(m.Qn))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintPreload(dAtA []byte, offset int, v uint64) int {
	offset -= sovPreload(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PlayerArgs) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Qn != 0 {
		n += 1 + sovPreload(uint64(m.Qn))
	}
	if m.Fnver != 0 {
		n += 1 + sovPreload(uint64(m.Fnver))
	}
	if m.Fnval != 0 {
		n += 1 + sovPreload(uint64(m.Fnval))
	}
	if m.ForceHost != 0 {
		n += 1 + sovPreload(uint64(m.ForceHost))
	}
	if m.VoiceBalance != 0 {
		n += 1 + sovPreload(uint64(m.VoiceBalance))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovPreload(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPreload(x uint64) (n int) {
	return sovPreload(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PlayerArgs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPreload
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
			return fmt.Errorf("proto: PlayerArgs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PlayerArgs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Qn", wireType)
			}
			m.Qn = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Qn |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fnver", wireType)
			}
			m.Fnver = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Fnver |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fnval", wireType)
			}
			m.Fnval = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Fnval |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ForceHost", wireType)
			}
			m.ForceHost = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ForceHost |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VoiceBalance", wireType)
			}
			m.VoiceBalance = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPreload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VoiceBalance |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipPreload(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPreload
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
func skipPreload(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPreload
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
					return 0, ErrIntOverflowPreload
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
					return 0, ErrIntOverflowPreload
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
				return 0, ErrInvalidLengthPreload
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPreload
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPreload
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPreload        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPreload          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPreload = fmt.Errorf("proto: unexpected end of group")
)