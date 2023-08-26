package tag

// Tag struct
type Tag struct {
	TagID      int64  `json:"tag_id,omitempty"`
	Name       string `json:"tag_name,omitempty"`
	Cover      string `json:"cover,omitempty"`
	Likes      int64  `json:"likes,omitempty"`
	Hates      int64  `json:"hates,omitempty"`
	Liked      int32  `json:"liked,omitempty"`
	Hated      int32  `json:"hated,omitempty"`
	Attribute  int32  `json:"attribute,omitempty"`
	IsActivity int8   `json:"is_activity,omitempty"`
	URI        string `json:"uri,omitempty"`
	TagType    string `json:"tag_type,omitempty"`
}

// Hot struct
type Hot struct {
	Rid  int16  `json:"rid"`
	Tags []*Tag `json:"tags"`
}

// SubTag struct
type SubTag struct {
	Count   int    `json:"count"`
	SubTags []*Tag `json:"subscribe"`
}

// TIcon struct
type TIcon struct {
	Icon string `json:"icon"`
}
