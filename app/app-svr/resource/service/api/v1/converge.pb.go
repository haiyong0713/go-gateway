// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: converge.proto

// use {app_id}.{version} as package name

package v1

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ConvergeCard struct {
	Id                   int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	ReType               int32    `protobuf:"varint,2,opt,name=re_type,json=reType,proto3" json:"re_type,omitempty"`
	ReValue              string   `protobuf:"bytes,3,opt,name=re_value,json=reValue,proto3" json:"re_value,omitempty"`
	Title                string   `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	Cover                string   `protobuf:"bytes,5,opt,name=cover,proto3" json:"cover,omitempty"`
	Content              []byte   `protobuf:"bytes,6,opt,name=content,proto3" json:"content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConvergeCard) Reset()         { *m = ConvergeCard{} }
func (m *ConvergeCard) String() string { return proto.CompactTextString(m) }
func (*ConvergeCard) ProtoMessage()    {}
func (*ConvergeCard) Descriptor() ([]byte, []int) {
	return fileDescriptor_6b7178c32cc8a67c, []int{0}
}
func (m *ConvergeCard) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ConvergeCard) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ConvergeCard.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ConvergeCard) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConvergeCard.Merge(m, src)
}
func (m *ConvergeCard) XXX_Size() int {
	return m.Size()
}
func (m *ConvergeCard) XXX_DiscardUnknown() {
	xxx_messageInfo_ConvergeCard.DiscardUnknown(m)
}

var xxx_messageInfo_ConvergeCard proto.InternalMessageInfo

func (m *ConvergeCard) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *ConvergeCard) GetReType() int32 {
	if m != nil {
		return m.ReType
	}
	return 0
}

func (m *ConvergeCard) GetReValue() string {
	if m != nil {
		return m.ReValue
	}
	return ""
}

func (m *ConvergeCard) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *ConvergeCard) GetCover() string {
	if m != nil {
		return m.Cover
	}
	return ""
}

func (m *ConvergeCard) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

type ConvergeCardReply struct {
	List                 []*ConvergeCard `protobuf:"bytes,1,rep,name=list,proto3" json:"list,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ConvergeCardReply) Reset()         { *m = ConvergeCardReply{} }
func (m *ConvergeCardReply) String() string { return proto.CompactTextString(m) }
func (*ConvergeCardReply) ProtoMessage()    {}
func (*ConvergeCardReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_6b7178c32cc8a67c, []int{1}
}
func (m *ConvergeCardReply) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ConvergeCardReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ConvergeCardReply.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ConvergeCardReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConvergeCardReply.Merge(m, src)
}
func (m *ConvergeCardReply) XXX_Size() int {
	return m.Size()
}
func (m *ConvergeCardReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ConvergeCardReply.DiscardUnknown(m)
}

var xxx_messageInfo_ConvergeCardReply proto.InternalMessageInfo

func (m *ConvergeCardReply) GetList() []*ConvergeCard {
	if m != nil {
		return m.List
	}
	return nil
}

func init() {
	proto.RegisterType((*ConvergeCard)(nil), "resource.service.v1.ConvergeCard")
	proto.RegisterType((*ConvergeCardReply)(nil), "resource.service.v1.ConvergeCardReply")
}

func init() { proto.RegisterFile("converge.proto", fileDescriptor_6b7178c32cc8a67c) }

var fileDescriptor_6b7178c32cc8a67c = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x40, 0x99, 0xa4, 0x49, 0x74, 0x2c, 0x45, 0x57, 0xc1, 0xf5, 0x12, 0x62, 0x4f, 0x7b, 0x5a,
	0xa8, 0xe2, 0x0f, 0xd8, 0x9b, 0xc7, 0x45, 0x3c, 0x78, 0x29, 0x75, 0x33, 0xc8, 0x42, 0xc8, 0x86,
	0xc9, 0x76, 0x21, 0x7f, 0xe2, 0x27, 0x79, 0xf4, 0x13, 0x24, 0x5f, 0x22, 0x4d, 0x15, 0x14, 0x3c,
	0xbe, 0xc7, 0x1b, 0x98, 0x19, 0x5c, 0x58, 0xdf, 0x46, 0xe2, 0x57, 0xd2, 0x1d, 0xfb, 0xe0, 0xc5,
	0x39, 0x53, 0xef, 0x77, 0x6c, 0x49, 0xf7, 0xc4, 0xd1, 0x59, 0xd2, 0x71, 0xb5, 0x7c, 0x03, 0x9c,
	0xaf, 0xbf, 0xbb, 0xf5, 0x96, 0x6b, 0xb1, 0xc0, 0xc4, 0xd5, 0x12, 0x2a, 0x50, 0xa9, 0x49, 0x5c,
	0x2d, 0x2e, 0xb1, 0x60, 0xda, 0x84, 0xa1, 0x23, 0x99, 0x54, 0xa0, 0x32, 0x93, 0x33, 0x3d, 0x0e,
	0x1d, 0x89, 0x2b, 0x3c, 0x62, 0xda, 0xc4, 0x6d, 0xb3, 0x23, 0x99, 0x56, 0xa0, 0x8e, 0x4d, 0xc1,
	0xf4, 0xb4, 0x47, 0x71, 0x81, 0x59, 0x70, 0xa1, 0x21, 0x39, 0x9b, 0xfc, 0x01, 0xf6, 0xd6, 0xfa,
	0x48, 0x2c, 0xb3, 0x83, 0x9d, 0x40, 0x48, 0x2c, 0xac, 0x6f, 0x03, 0xb5, 0x41, 0xe6, 0x15, 0xa8,
	0xb9, 0xf9, 0xc1, 0xe5, 0x03, 0x9e, 0xfd, 0xde, 0xcc, 0x50, 0xd7, 0x0c, 0xe2, 0x0e, 0x67, 0x8d,
	0xeb, 0x83, 0x84, 0x2a, 0x55, 0x27, 0x37, 0xd7, 0xfa, 0x9f, 0x9b, 0xf4, 0x9f, 0xa9, 0x29, 0xbf,
	0x3f, 0x7d, 0x1f, 0x4b, 0xf8, 0x18, 0x4b, 0xf8, 0x1c, 0x4b, 0x78, 0x4e, 0xe2, 0xea, 0x25, 0x9f,
	0x9e, 0x72, 0xfb, 0x15, 0x00, 0x00, 0xff, 0xff, 0x52, 0xcc, 0x6e, 0x3e, 0x26, 0x01, 0x00, 0x00,
}

func (m *ConvergeCard) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConvergeCard) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ConvergeCard) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Content) > 0 {
		i -= len(m.Content)
		copy(dAtA[i:], m.Content)
		i = encodeVarintConverge(dAtA, i, uint64(len(m.Content)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Cover) > 0 {
		i -= len(m.Cover)
		copy(dAtA[i:], m.Cover)
		i = encodeVarintConverge(dAtA, i, uint64(len(m.Cover)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintConverge(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ReValue) > 0 {
		i -= len(m.ReValue)
		copy(dAtA[i:], m.ReValue)
		i = encodeVarintConverge(dAtA, i, uint64(len(m.ReValue)))
		i--
		dAtA[i] = 0x1a
	}
	if m.ReType != 0 {
		i = encodeVarintConverge(dAtA, i, uint64(m.ReType))
		i--
		dAtA[i] = 0x10
	}
	if m.Id != 0 {
		i = encodeVarintConverge(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ConvergeCardReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ConvergeCardReply) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ConvergeCardReply) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.List) > 0 {
		for iNdEx := len(m.List) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.List[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintConverge(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintConverge(dAtA []byte, offset int, v uint64) int {
	offset -= sovConverge(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ConvergeCard) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovConverge(uint64(m.Id))
	}
	if m.ReType != 0 {
		n += 1 + sovConverge(uint64(m.ReType))
	}
	l = len(m.ReValue)
	if l > 0 {
		n += 1 + l + sovConverge(uint64(l))
	}
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovConverge(uint64(l))
	}
	l = len(m.Cover)
	if l > 0 {
		n += 1 + l + sovConverge(uint64(l))
	}
	l = len(m.Content)
	if l > 0 {
		n += 1 + l + sovConverge(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *ConvergeCardReply) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.List) > 0 {
		for _, e := range m.List {
			l = e.Size()
			n += 1 + l + sovConverge(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovConverge(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozConverge(x uint64) (n int) {
	return sovConverge(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ConvergeCard) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConverge
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
			return fmt.Errorf("proto: ConvergeCard: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConvergeCard: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReType", wireType)
			}
			m.ReType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReType |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReValue", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
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
				return ErrInvalidLengthConverge
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConverge
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReValue = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
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
				return ErrInvalidLengthConverge
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConverge
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cover", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
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
				return ErrInvalidLengthConverge
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConverge
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cover = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Content", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthConverge
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthConverge
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Content = append(m.Content[:0], dAtA[iNdEx:postIndex]...)
			if m.Content == nil {
				m.Content = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConverge(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthConverge
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthConverge
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
func (m *ConvergeCardReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConverge
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
			return fmt.Errorf("proto: ConvergeCardReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ConvergeCardReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field List", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConverge
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConverge
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConverge
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.List = append(m.List, &ConvergeCard{})
			if err := m.List[len(m.List)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConverge(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthConverge
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthConverge
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
func skipConverge(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConverge
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
					return 0, ErrIntOverflowConverge
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
					return 0, ErrIntOverflowConverge
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
				return 0, ErrInvalidLengthConverge
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupConverge
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthConverge
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthConverge        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConverge          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupConverge = fmt.Errorf("proto: unexpected end of group")
)
