package tag

type Tag struct {
	TagID      int64  `json:"tag_id"`
	Name       string `json:"tag_name"`
	Cover      string `json:"cover"`
	Likes      int64  `json:"likes"`
	Hates      int64  `json:"hates"`
	Liked      int32  `json:"liked"`
	Hated      int32  `json:"hated"`
	Attribute  int32  `json:"attribute"`
	IsActivity int8   `json:"is_activity"`
	URI        string `json:"uri,omitempty"`
	TagType    string `json:"tag_type,omitempty"`
}
