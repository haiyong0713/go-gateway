package archive

// Addit id,aid,source,redirect_url,description
type Addit struct {
	Aid         int64  `json:"aid"`
	Description string `json:"description"`
	DescV2      string `json:"desc_v2"`
	Subtype     int64  `json:"sub_type"`
}

type DescV2FromArchive struct {
	RawText string `json:"raw_text"`
	Type    int32  `json:"type"`
	BizId   string `json:"biz_id"`
}

type DescReply struct {
	Desc   string               `json:"desc"`
	DescV2 []*DescV2FromArchive `json:"desc_v2"`
}

// AttrVal get attribute value.
func (ad *Addit) AttrVal(bit uint) int64 {
	return (ad.Subtype >> bit) & int64(1)
}
