// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: taishan.proto

package taishan

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
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
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type PlayConf struct {
	// true:展示(0也是展示) false:隐藏
	Show                 bool        `protobuf:"varint,1,opt,name=show,proto3" json:"show,omitempty"`
	FieldValue           *FieldValue `protobuf:"bytes,2,opt,name=field_value,json=fieldValue,proto3" json:"field_value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PlayConf) Reset()         { *m = PlayConf{} }
func (m *PlayConf) String() string { return proto.CompactTextString(m) }
func (*PlayConf) ProtoMessage()    {}
func (*PlayConf) Descriptor() ([]byte, []int) {
	return fileDescriptor_493fd784c2a0d75b, []int{0}
}
func (m *PlayConf) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PlayConf) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PlayConf.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PlayConf) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlayConf.Merge(m, src)
}
func (m *PlayConf) XXX_Size() int {
	return m.Size()
}
func (m *PlayConf) XXX_DiscardUnknown() {
	xxx_messageInfo_PlayConf.DiscardUnknown(m)
}

var xxx_messageInfo_PlayConf proto.InternalMessageInfo

func (m *PlayConf) GetShow() bool {
	if m != nil {
		return m.Show
	}
	return false
}

func (m *PlayConf) GetFieldValue() *FieldValue {
	if m != nil {
		return m.FieldValue
	}
	return nil
}

type PlayConfs struct {
	PlayConfs            map[int64]*PlayConf `protobuf:"bytes,1,rep,name=play_confs,json=playConfs,proto3" json:"play_confs,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *PlayConfs) Reset()         { *m = PlayConfs{} }
func (m *PlayConfs) String() string { return proto.CompactTextString(m) }
func (*PlayConfs) ProtoMessage()    {}
func (*PlayConfs) Descriptor() ([]byte, []int) {
	return fileDescriptor_493fd784c2a0d75b, []int{1}
}
func (m *PlayConfs) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PlayConfs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PlayConfs.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PlayConfs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PlayConfs.Merge(m, src)
}
func (m *PlayConfs) XXX_Size() int {
	return m.Size()
}
func (m *PlayConfs) XXX_DiscardUnknown() {
	xxx_messageInfo_PlayConfs.DiscardUnknown(m)
}

var xxx_messageInfo_PlayConfs proto.InternalMessageInfo

func (m *PlayConfs) GetPlayConfs() map[int64]*PlayConf {
	if m != nil {
		return m.PlayConfs
	}
	return nil
}

type FieldValue struct {
	// Types that are valid to be assigned to Value:
	//	*FieldValue_Switch
	Value                isFieldValue_Value `protobuf_oneof:"value"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *FieldValue) Reset()         { *m = FieldValue{} }
func (m *FieldValue) String() string { return proto.CompactTextString(m) }
func (*FieldValue) ProtoMessage()    {}
func (*FieldValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_493fd784c2a0d75b, []int{2}
}
func (m *FieldValue) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FieldValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FieldValue.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FieldValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FieldValue.Merge(m, src)
}
func (m *FieldValue) XXX_Size() int {
	return m.Size()
}
func (m *FieldValue) XXX_DiscardUnknown() {
	xxx_messageInfo_FieldValue.DiscardUnknown(m)
}

var xxx_messageInfo_FieldValue proto.InternalMessageInfo

type isFieldValue_Value interface {
	isFieldValue_Value()
	MarshalTo([]byte) (int, error)
	Size() int
}

type FieldValue_Switch struct {
	Switch bool `protobuf:"varint,1,opt,name=switch,proto3,oneof"`
}

func (*FieldValue_Switch) isFieldValue_Value() {}

func (m *FieldValue) GetValue() isFieldValue_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *FieldValue) GetSwitch() bool {
	if x, ok := m.GetValue().(*FieldValue_Switch); ok {
		return x.Switch
	}
	return false
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*FieldValue) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _FieldValue_OneofMarshaler, _FieldValue_OneofUnmarshaler, _FieldValue_OneofSizer, []interface{}{
		(*FieldValue_Switch)(nil),
	}
}

func _FieldValue_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*FieldValue)
	// value
	switch x := m.Value.(type) {
	case *FieldValue_Switch:
		t := uint64(0)
		if x.Switch {
			t = 1
		}
		_ = b.EncodeVarint(1<<3 | proto.WireVarint)
		_ = b.EncodeVarint(t)
	case nil:
	default:
		return fmt.Errorf("FieldValue.Value has unexpected type %T", x)
	}
	return nil
}

func _FieldValue_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*FieldValue)
	switch tag {
	case 1: // value.switch
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.Value = &FieldValue_Switch{x != 0}
		return true, err
	default:
		return false, nil
	}
}

func _FieldValue_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*FieldValue)
	// value
	switch x := m.Value.(type) {
	case *FieldValue_Switch:
		n += 1 // tag and wire
		n += 1
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

func init() {
	proto.RegisterType((*PlayConf)(nil), "PlayConf")
	proto.RegisterType((*PlayConfs)(nil), "PlayConfs")
	proto.RegisterMapType((map[int64]*PlayConf)(nil), "PlayConfs.PlayConfsEntry")
	proto.RegisterType((*FieldValue)(nil), "FieldValue")
}

func init() { proto.RegisterFile("taishan.proto", fileDescriptor_493fd784c2a0d75b) }

var fileDescriptor_493fd784c2a0d75b = []byte{
	// 262 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x49, 0xcc, 0x2c,
	0xce, 0x48, 0xcc, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x97, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9, 0x28,
	0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf, 0xd7, 0x07, 0x0b, 0x27, 0x95, 0xa6,
	0x81, 0x79, 0x60, 0x0e, 0x98, 0x05, 0x51, 0xae, 0xe4, 0xc3, 0xc5, 0x11, 0x90, 0x93, 0x58, 0xe9,
	0x9c, 0x9f, 0x97, 0x26, 0x24, 0xc4, 0xc5, 0x52, 0x9c, 0x91, 0x5f, 0x2e, 0xc1, 0xa8, 0xc0, 0xa8,
	0xc1, 0x11, 0x04, 0x66, 0x0b, 0xe9, 0x70, 0x71, 0xa7, 0x65, 0xa6, 0xe6, 0xa4, 0xc4, 0x97, 0x25,
	0xe6, 0x94, 0xa6, 0x4a, 0x30, 0x29, 0x30, 0x6a, 0x70, 0x1b, 0x71, 0xeb, 0xb9, 0x81, 0xc4, 0xc2,
	0x40, 0x42, 0x41, 0x5c, 0x69, 0x70, 0xb6, 0x52, 0x1f, 0x23, 0x17, 0x27, 0xcc, 0xb8, 0x62, 0x21,
	0x0b, 0x2e, 0xae, 0x82, 0x9c, 0xc4, 0xca, 0xf8, 0x64, 0x10, 0x4f, 0x82, 0x51, 0x81, 0x59, 0x83,
	0xdb, 0x48, 0x52, 0x0f, 0x2e, 0x8f, 0x60, 0xb9, 0xe6, 0x95, 0x14, 0x55, 0x06, 0x71, 0x16, 0xc0,
	0xf8, 0x52, 0xee, 0x5c, 0x7c, 0xa8, 0x92, 0x42, 0x02, 0x5c, 0xcc, 0xd9, 0xa9, 0x95, 0x60, 0xa7,
	0x31, 0x07, 0x81, 0x98, 0x42, 0xf2, 0x5c, 0xac, 0xc8, 0x6e, 0xe2, 0x84, 0x1b, 0x17, 0x04, 0x11,
	0xb7, 0x62, 0xb2, 0x60, 0x54, 0xd2, 0xe7, 0xe2, 0x42, 0x38, 0x55, 0x48, 0x82, 0x8b, 0xad, 0xb8,
	0x3c, 0xb3, 0x24, 0x39, 0x03, 0xe2, 0x45, 0x0f, 0x86, 0x20, 0x28, 0xdf, 0x89, 0x1d, 0x6a, 0x98,
	0x93, 0xe8, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0x18, 0xc5,
	0x0e, 0x0d, 0xdb, 0x24, 0x36, 0x70, 0x68, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x91, 0xf9,
	0xf8, 0x79, 0x6d, 0x01, 0x00, 0x00,
}

func (m *PlayConf) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PlayConf) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Show {
		dAtA[i] = 0x8
		i++
		if m.Show {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.FieldValue != nil {
		dAtA[i] = 0x12
		i++
		i = encodeVarintTaishan(dAtA, i, uint64(m.FieldValue.Size()))
		n1, err := m.FieldValue.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n1
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *PlayConfs) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PlayConfs) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.PlayConfs) > 0 {
		for k := range m.PlayConfs {
			dAtA[i] = 0xa
			i++
			v := m.PlayConfs[k]
			msgSize := 0
			if v != nil {
				msgSize = v.Size()
				msgSize += 1 + sovTaishan(uint64(msgSize))
			}
			mapSize := 1 + sovTaishan(uint64(k)) + msgSize
			i = encodeVarintTaishan(dAtA, i, uint64(mapSize))
			dAtA[i] = 0x8
			i++
			i = encodeVarintTaishan(dAtA, i, uint64(k))
			if v != nil {
				dAtA[i] = 0x12
				i++
				i = encodeVarintTaishan(dAtA, i, uint64(v.Size()))
				n2, err := v.MarshalTo(dAtA[i:])
				if err != nil {
					return 0, err
				}
				i += n2
			}
		}
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *FieldValue) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FieldValue) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value != nil {
		nn3, err := m.Value.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += nn3
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *FieldValue_Switch) MarshalTo(dAtA []byte) (int, error) {
	i := 0
	dAtA[i] = 0x8
	i++
	if m.Switch {
		dAtA[i] = 1
	} else {
		dAtA[i] = 0
	}
	i++
	return i, nil
}
func encodeVarintTaishan(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *PlayConf) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Show {
		n += 2
	}
	if m.FieldValue != nil {
		l = m.FieldValue.Size()
		n += 1 + l + sovTaishan(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *PlayConfs) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.PlayConfs) > 0 {
		for k, v := range m.PlayConfs {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.Size()
				l += 1 + sovTaishan(uint64(l))
			}
			mapEntrySize := 1 + sovTaishan(uint64(k)) + l
			n += mapEntrySize + 1 + sovTaishan(uint64(mapEntrySize))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *FieldValue) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Value != nil {
		n += m.Value.Size()
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *FieldValue_Switch) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	n += 2
	return n
}

func sovTaishan(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTaishan(x uint64) (n int) {
	return sovTaishan(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PlayConf) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaishan
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
			return fmt.Errorf("proto: PlayConf: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PlayConf: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Show", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaishan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Show = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FieldValue", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaishan
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
				return ErrInvalidLengthTaishan
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaishan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.FieldValue == nil {
				m.FieldValue = &FieldValue{}
			}
			if err := m.FieldValue.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaishan(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTaishan
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTaishan
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
func (m *PlayConfs) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaishan
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
			return fmt.Errorf("proto: PlayConfs: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PlayConfs: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PlayConfs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaishan
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
				return ErrInvalidLengthTaishan
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTaishan
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PlayConfs == nil {
				m.PlayConfs = make(map[int64]*PlayConf)
			}
			var mapkey int64
			var mapvalue *PlayConf
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTaishan
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
				if fieldNum == 1 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTaishan
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapkey |= int64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTaishan
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthTaishan
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthTaishan
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &PlayConf{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipTaishan(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthTaishan
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.PlayConfs[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTaishan(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTaishan
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTaishan
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
func (m *FieldValue) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTaishan
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
			return fmt.Errorf("proto: FieldValue: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FieldValue: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Switch", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTaishan
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			b := bool(v != 0)
			m.Value = &FieldValue_Switch{b}
		default:
			iNdEx = preIndex
			skippy, err := skipTaishan(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTaishan
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTaishan
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
func skipTaishan(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTaishan
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
					return 0, ErrIntOverflowTaishan
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTaishan
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
				return 0, ErrInvalidLengthTaishan
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthTaishan
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTaishan
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipTaishan(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthTaishan
				}
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthTaishan = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTaishan   = fmt.Errorf("proto: integer overflow")
)
