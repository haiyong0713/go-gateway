// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: go-gateway/app/app-svr/app-card/interface/model/card/proto/card.proto

package api

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

type Card struct {
	// Types that are valid to be assigned to Item:
	//	*Card_SmallCoverV5
	//	*Card_LargeCoverV1
	//	*Card_ThreeItemAllV2
	//	*Card_ThreeItemV1
	//	*Card_HotTopic
	//	*Card_ThreeItemHV5
	//	*Card_MiddleCoverV3
	//	*Card_LargeCoverV4
	//	*Card_PopularTopEntrance
	//	*Card_RcmdOneItem
	//	*Card_SmallCoverV5Ad
	Item                 isCard_Item `protobuf_oneof:"item"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Card) Reset()         { *m = Card{} }
func (m *Card) String() string { return proto.CompactTextString(m) }
func (*Card) ProtoMessage()    {}
func (*Card) Descriptor() ([]byte, []int) {
	return fileDescriptor_2f71e4c6ba655635, []int{0}
}
func (m *Card) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Card) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Card.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Card) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Card.Merge(m, src)
}
func (m *Card) XXX_Size() int {
	return m.Size()
}
func (m *Card) XXX_DiscardUnknown() {
	xxx_messageInfo_Card.DiscardUnknown(m)
}

var xxx_messageInfo_Card proto.InternalMessageInfo

type isCard_Item interface {
	isCard_Item()
	MarshalTo([]byte) (int, error)
	Size() int
}

type Card_SmallCoverV5 struct {
	SmallCoverV5 *SmallCoverV5 `protobuf:"bytes,1,opt,name=small_cover_v5,json=smallCoverV5,proto3,oneof" json:"small_cover_v5,omitempty"`
}
type Card_LargeCoverV1 struct {
	LargeCoverV1 *LargeCoverV1 `protobuf:"bytes,2,opt,name=large_cover_v1,json=largeCoverV1,proto3,oneof" json:"large_cover_v1,omitempty"`
}
type Card_ThreeItemAllV2 struct {
	ThreeItemAllV2 *ThreeItemAllV2 `protobuf:"bytes,3,opt,name=three_item_all_v2,json=threeItemAllV2,proto3,oneof" json:"three_item_all_v2,omitempty"`
}
type Card_ThreeItemV1 struct {
	ThreeItemV1 *ThreeItemV1 `protobuf:"bytes,4,opt,name=three_item_v1,json=threeItemV1,proto3,oneof" json:"three_item_v1,omitempty"`
}
type Card_HotTopic struct {
	HotTopic *HotTopic `protobuf:"bytes,5,opt,name=hot_topic,json=hotTopic,proto3,oneof" json:"hot_topic,omitempty"`
}
type Card_ThreeItemHV5 struct {
	ThreeItemHV5 *DynamicHot `protobuf:"bytes,6,opt,name=three_item_h_v5,json=threeItemHV5,proto3,oneof" json:"three_item_h_v5,omitempty"`
}
type Card_MiddleCoverV3 struct {
	MiddleCoverV3 *MiddleCoverV3 `protobuf:"bytes,7,opt,name=middle_cover_v3,json=middleCoverV3,proto3,oneof" json:"middle_cover_v3,omitempty"`
}
type Card_LargeCoverV4 struct {
	LargeCoverV4 *LargeCoverV4 `protobuf:"bytes,8,opt,name=large_cover_v4,json=largeCoverV4,proto3,oneof" json:"large_cover_v4,omitempty"`
}
type Card_PopularTopEntrance struct {
	PopularTopEntrance *PopularTopEntrance `protobuf:"bytes,9,opt,name=popular_top_entrance,json=popularTopEntrance,proto3,oneof" json:"popular_top_entrance,omitempty"`
}
type Card_RcmdOneItem struct {
	RcmdOneItem *RcmdOneItem `protobuf:"bytes,10,opt,name=rcmd_one_item,json=rcmdOneItem,proto3,oneof" json:"rcmd_one_item,omitempty"`
}
type Card_SmallCoverV5Ad struct {
	SmallCoverV5Ad *SmallCoverV5Ad `protobuf:"bytes,11,opt,name=small_cover_v5_ad,json=smallCoverV5Ad,proto3,oneof" json:"small_cover_v5_ad,omitempty"`
}

func (*Card_SmallCoverV5) isCard_Item()       {}
func (*Card_LargeCoverV1) isCard_Item()       {}
func (*Card_ThreeItemAllV2) isCard_Item()     {}
func (*Card_ThreeItemV1) isCard_Item()        {}
func (*Card_HotTopic) isCard_Item()           {}
func (*Card_ThreeItemHV5) isCard_Item()       {}
func (*Card_MiddleCoverV3) isCard_Item()      {}
func (*Card_LargeCoverV4) isCard_Item()       {}
func (*Card_PopularTopEntrance) isCard_Item() {}
func (*Card_RcmdOneItem) isCard_Item()        {}
func (*Card_SmallCoverV5Ad) isCard_Item()     {}

func (m *Card) GetItem() isCard_Item {
	if m != nil {
		return m.Item
	}
	return nil
}

func (m *Card) GetSmallCoverV5() *SmallCoverV5 {
	if x, ok := m.GetItem().(*Card_SmallCoverV5); ok {
		return x.SmallCoverV5
	}
	return nil
}

func (m *Card) GetLargeCoverV1() *LargeCoverV1 {
	if x, ok := m.GetItem().(*Card_LargeCoverV1); ok {
		return x.LargeCoverV1
	}
	return nil
}

func (m *Card) GetThreeItemAllV2() *ThreeItemAllV2 {
	if x, ok := m.GetItem().(*Card_ThreeItemAllV2); ok {
		return x.ThreeItemAllV2
	}
	return nil
}

func (m *Card) GetThreeItemV1() *ThreeItemV1 {
	if x, ok := m.GetItem().(*Card_ThreeItemV1); ok {
		return x.ThreeItemV1
	}
	return nil
}

func (m *Card) GetHotTopic() *HotTopic {
	if x, ok := m.GetItem().(*Card_HotTopic); ok {
		return x.HotTopic
	}
	return nil
}

func (m *Card) GetThreeItemHV5() *DynamicHot {
	if x, ok := m.GetItem().(*Card_ThreeItemHV5); ok {
		return x.ThreeItemHV5
	}
	return nil
}

func (m *Card) GetMiddleCoverV3() *MiddleCoverV3 {
	if x, ok := m.GetItem().(*Card_MiddleCoverV3); ok {
		return x.MiddleCoverV3
	}
	return nil
}

func (m *Card) GetLargeCoverV4() *LargeCoverV4 {
	if x, ok := m.GetItem().(*Card_LargeCoverV4); ok {
		return x.LargeCoverV4
	}
	return nil
}

func (m *Card) GetPopularTopEntrance() *PopularTopEntrance {
	if x, ok := m.GetItem().(*Card_PopularTopEntrance); ok {
		return x.PopularTopEntrance
	}
	return nil
}

func (m *Card) GetRcmdOneItem() *RcmdOneItem {
	if x, ok := m.GetItem().(*Card_RcmdOneItem); ok {
		return x.RcmdOneItem
	}
	return nil
}

func (m *Card) GetSmallCoverV5Ad() *SmallCoverV5Ad {
	if x, ok := m.GetItem().(*Card_SmallCoverV5Ad); ok {
		return x.SmallCoverV5Ad
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Card) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Card_SmallCoverV5)(nil),
		(*Card_LargeCoverV1)(nil),
		(*Card_ThreeItemAllV2)(nil),
		(*Card_ThreeItemV1)(nil),
		(*Card_HotTopic)(nil),
		(*Card_ThreeItemHV5)(nil),
		(*Card_MiddleCoverV3)(nil),
		(*Card_LargeCoverV4)(nil),
		(*Card_PopularTopEntrance)(nil),
		(*Card_RcmdOneItem)(nil),
		(*Card_SmallCoverV5Ad)(nil),
	}
}

func init() {
	proto.RegisterType((*Card)(nil), "bilibili.app.card.v1.Card")
}

func init() {
	proto.RegisterFile("go-gateway/app/app-svr/app-card/interface/model/card/proto/card.proto", fileDescriptor_2f71e4c6ba655635)
}

var fileDescriptor_2f71e4c6ba655635 = []byte{
	// 507 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x93, 0xcf, 0x6e, 0xd3, 0x4e,
	0x10, 0xc7, 0xed, 0x36, 0xbf, 0xfc, 0xda, 0x0d, 0x4d, 0x85, 0xd5, 0xc3, 0x8a, 0x83, 0x55, 0x0a,
	0x87, 0x5e, 0xea, 0x28, 0xff, 0x8e, 0x1c, 0x92, 0x52, 0xd5, 0x41, 0x54, 0x04, 0x13, 0xe5, 0x80,
	0x90, 0xac, 0x8d, 0xbd, 0x24, 0x2b, 0xed, 0x7a, 0x57, 0xeb, 0xc5, 0xa8, 0x6f, 0xc2, 0x99, 0xa7,
	0xe1, 0xc8, 0x23, 0x40, 0x78, 0x11, 0x34, 0xeb, 0xd2, 0x3a, 0xd4, 0x0a, 0x48, 0x1c, 0x36, 0xd9,
	0x19, 0x7f, 0xf7, 0xa3, 0xd1, 0x77, 0x66, 0xd0, 0xc5, 0x52, 0x9e, 0x2d, 0x89, 0xa1, 0x1f, 0xc9,
	0x75, 0x87, 0x28, 0x05, 0xe7, 0x2c, 0x2f, 0xb4, 0xfd, 0x4f, 0x88, 0x4e, 0x3b, 0x2c, 0x33, 0x54,
	0xbf, 0x27, 0x09, 0xed, 0x08, 0x99, 0x52, 0xde, 0xb1, 0x49, 0xa5, 0xa5, 0x91, 0xf6, 0x1a, 0xd8,
	0xab, 0x77, 0xb4, 0x60, 0x9c, 0xc1, 0x09, 0x88, 0x52, 0x81, 0xfd, 0x50, 0x74, 0x1f, 0x5d, 0xfe,
	0x03, 0x3c, 0x67, 0xd9, 0x92, 0xd3, 0x12, 0x7f, 0xf2, 0xbd, 0x89, 0x1a, 0xe7, 0x44, 0xa7, 0xde,
	0x0b, 0xd4, 0xce, 0x05, 0xe1, 0x3c, 0x4e, 0x64, 0x41, 0x75, 0x5c, 0x0c, 0xb1, 0x7b, 0xec, 0x9e,
	0xb6, 0x7a, 0x27, 0x41, 0x5d, 0x01, 0xc1, 0x1b, 0xd0, 0x9e, 0x83, 0x74, 0x3e, 0x0c, 0x9d, 0xe8,
	0x41, 0x5e, 0x89, 0x81, 0xc5, 0x89, 0x5e, 0xd2, 0x5f, 0xac, 0x2e, 0xde, 0xd9, 0xc6, 0x7a, 0x09,
	0xda, 0xf2, 0x6d, 0x17, 0x58, 0xbc, 0x12, 0x7b, 0xaf, 0xd1, 0x43, 0xb3, 0xd2, 0x94, 0xc6, 0xcc,
	0x50, 0x11, 0x43, 0x81, 0x45, 0x0f, 0xef, 0x5a, 0xdc, 0xd3, 0x7a, 0xdc, 0x0c, 0xe4, 0x13, 0x43,
	0xc5, 0x88, 0xf3, 0x79, 0x2f, 0x74, 0xa2, 0xb6, 0xd9, 0xc8, 0x78, 0x97, 0xe8, 0xa0, 0x82, 0x2c,
	0xba, 0xb8, 0x61, 0x71, 0x8f, 0xff, 0x80, 0xb3, 0xc5, 0xb5, 0xcc, 0x5d, 0xe8, 0x3d, 0x43, 0xfb,
	0x2b, 0x69, 0x62, 0x23, 0x15, 0x4b, 0xf0, 0x7f, 0x16, 0xe2, 0xd7, 0x43, 0x42, 0x69, 0x66, 0xa0,
	0x0a, 0x9d, 0x68, 0x6f, 0x75, 0x73, 0xf7, 0x26, 0xe8, 0xb0, 0x52, 0xc7, 0x0a, 0x3c, 0x6f, 0x5a,
	0xc8, 0x71, 0x3d, 0xe4, 0xf9, 0x75, 0x46, 0x04, 0x4b, 0x42, 0x69, 0xc0, 0xa5, 0xdb, 0x42, 0xc2,
	0xf9, 0xd0, 0xbb, 0x42, 0x87, 0x82, 0xa5, 0x29, 0xbf, 0xb5, 0xbc, 0x8f, 0xff, 0xb7, 0xa8, 0x27,
	0xf5, 0xa8, 0x2b, 0x2b, 0x2e, 0x3d, 0xee, 0x87, 0x4e, 0x74, 0x20, 0xaa, 0x89, 0x7b, 0x0d, 0x1c,
	0xe0, 0xbd, 0xbf, 0x6c, 0xe0, 0xe0, 0xb7, 0x06, 0x0e, 0xbc, 0x77, 0xe8, 0x48, 0x49, 0xf5, 0x81,
	0x13, 0x0d, 0x46, 0xc5, 0x34, 0x33, 0x9a, 0x64, 0x09, 0xc5, 0xfb, 0x96, 0x78, 0x5a, 0x4f, 0x9c,
	0x96, 0x2f, 0x66, 0x52, 0x5d, 0xdc, 0xe8, 0x43, 0x27, 0xf2, 0xd4, 0xbd, 0x2c, 0xf4, 0x52, 0x27,
	0x22, 0x8d, 0x65, 0x56, 0xda, 0x88, 0xd1, 0xb6, 0x5e, 0x46, 0x89, 0x48, 0x5f, 0x65, 0xa5, 0x6b,
	0x4e, 0xd4, 0xd2, 0x77, 0x21, 0xcc, 0xd9, 0xe6, 0xfc, 0xc7, 0x24, 0xc5, 0xad, 0x6d, 0x73, 0x56,
	0x5d, 0x81, 0x51, 0x0a, 0x73, 0x96, 0x6f, 0x64, 0xc6, 0x4d, 0xd4, 0x80, 0x92, 0xc6, 0xe1, 0x97,
	0xb5, 0xef, 0x7e, 0x5d, 0xfb, 0xee, 0xb7, 0xb5, 0xef, 0x7e, 0xfa, 0xe1, 0x3b, 0xc8, 0x4f, 0xa4,
	0x08, 0x16, 0x44, 0xb1, 0xbc, 0x16, 0x3d, 0x75, 0xdf, 0xee, 0x12, 0xc5, 0x3e, 0xef, 0xb4, 0xc7,
	0xa3, 0xe9, 0x04, 0xd6, 0x73, 0xde, 0x85, 0xdf, 0x45, 0xd3, 0x2e, 0x6d, 0xff, 0x67, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x18, 0x38, 0xba, 0x6a, 0x5c, 0x04, 0x00, 0x00,
}

func (m *Card) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Card) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if m.Item != nil {
		{
			size := m.Item.Size()
			i -= size
			if _, err := m.Item.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
		}
	}
	return len(dAtA) - i, nil
}

func (m *Card_SmallCoverV5) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_SmallCoverV5) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.SmallCoverV5 != nil {
		{
			size, err := m.SmallCoverV5.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}
func (m *Card_LargeCoverV1) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_LargeCoverV1) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.LargeCoverV1 != nil {
		{
			size, err := m.LargeCoverV1.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	return len(dAtA) - i, nil
}
func (m *Card_ThreeItemAllV2) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_ThreeItemAllV2) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ThreeItemAllV2 != nil {
		{
			size, err := m.ThreeItemAllV2.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	return len(dAtA) - i, nil
}
func (m *Card_ThreeItemV1) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_ThreeItemV1) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ThreeItemV1 != nil {
		{
			size, err := m.ThreeItemV1.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	return len(dAtA) - i, nil
}
func (m *Card_HotTopic) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_HotTopic) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.HotTopic != nil {
		{
			size, err := m.HotTopic.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	return len(dAtA) - i, nil
}
func (m *Card_ThreeItemHV5) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_ThreeItemHV5) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.ThreeItemHV5 != nil {
		{
			size, err := m.ThreeItemHV5.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x32
	}
	return len(dAtA) - i, nil
}
func (m *Card_MiddleCoverV3) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_MiddleCoverV3) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.MiddleCoverV3 != nil {
		{
			size, err := m.MiddleCoverV3.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x3a
	}
	return len(dAtA) - i, nil
}
func (m *Card_LargeCoverV4) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_LargeCoverV4) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.LargeCoverV4 != nil {
		{
			size, err := m.LargeCoverV4.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x42
	}
	return len(dAtA) - i, nil
}
func (m *Card_PopularTopEntrance) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_PopularTopEntrance) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.PopularTopEntrance != nil {
		{
			size, err := m.PopularTopEntrance.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x4a
	}
	return len(dAtA) - i, nil
}
func (m *Card_RcmdOneItem) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_RcmdOneItem) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.RcmdOneItem != nil {
		{
			size, err := m.RcmdOneItem.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x52
	}
	return len(dAtA) - i, nil
}
func (m *Card_SmallCoverV5Ad) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Card_SmallCoverV5Ad) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.SmallCoverV5Ad != nil {
		{
			size, err := m.SmallCoverV5Ad.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintCard(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x5a
	}
	return len(dAtA) - i, nil
}
func encodeVarintCard(dAtA []byte, offset int, v uint64) int {
	offset -= sovCard(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Card) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Item != nil {
		n += m.Item.Size()
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *Card_SmallCoverV5) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SmallCoverV5 != nil {
		l = m.SmallCoverV5.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_LargeCoverV1) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LargeCoverV1 != nil {
		l = m.LargeCoverV1.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_ThreeItemAllV2) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ThreeItemAllV2 != nil {
		l = m.ThreeItemAllV2.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_ThreeItemV1) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ThreeItemV1 != nil {
		l = m.ThreeItemV1.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_HotTopic) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.HotTopic != nil {
		l = m.HotTopic.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_ThreeItemHV5) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ThreeItemHV5 != nil {
		l = m.ThreeItemHV5.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_MiddleCoverV3) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MiddleCoverV3 != nil {
		l = m.MiddleCoverV3.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_LargeCoverV4) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.LargeCoverV4 != nil {
		l = m.LargeCoverV4.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_PopularTopEntrance) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PopularTopEntrance != nil {
		l = m.PopularTopEntrance.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_RcmdOneItem) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.RcmdOneItem != nil {
		l = m.RcmdOneItem.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}
func (m *Card_SmallCoverV5Ad) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SmallCoverV5Ad != nil {
		l = m.SmallCoverV5Ad.Size()
		n += 1 + l + sovCard(uint64(l))
	}
	return n
}

func sovCard(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozCard(x uint64) (n int) {
	return sovCard(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Card) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowCard
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
			return fmt.Errorf("proto: Card: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Card: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SmallCoverV5", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &SmallCoverV5{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_SmallCoverV5{v}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LargeCoverV1", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &LargeCoverV1{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_LargeCoverV1{v}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ThreeItemAllV2", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &ThreeItemAllV2{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_ThreeItemAllV2{v}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ThreeItemV1", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &ThreeItemV1{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_ThreeItemV1{v}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HotTopic", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &HotTopic{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_HotTopic{v}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ThreeItemHV5", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &DynamicHot{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_ThreeItemHV5{v}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MiddleCoverV3", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &MiddleCoverV3{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_MiddleCoverV3{v}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LargeCoverV4", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &LargeCoverV4{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_LargeCoverV4{v}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PopularTopEntrance", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &PopularTopEntrance{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_PopularTopEntrance{v}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RcmdOneItem", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &RcmdOneItem{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_RcmdOneItem{v}
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SmallCoverV5Ad", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowCard
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
				return ErrInvalidLengthCard
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthCard
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &SmallCoverV5Ad{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.Item = &Card_SmallCoverV5Ad{v}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipCard(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthCard
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
func skipCard(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowCard
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
					return 0, ErrIntOverflowCard
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
					return 0, ErrIntOverflowCard
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
				return 0, ErrInvalidLengthCard
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupCard
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthCard
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthCard        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowCard          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupCard = fmt.Errorf("proto: unexpected end of group")
)