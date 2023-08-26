package show

// PopChannelTag
type PopChannelTag struct {
	ID            int64 `json:"id" form:"id"`
	TagID         int64 `json:"tag_id" form:"tag_id" validate:"required"`
	TopEntranceId int64 `json:"top_entrance_id" form:"top_entrance_id" validate:"top_entrance_id"`
	Deleted       int   `json:"deleted" form:"deleted" validate:"required"`
}

// PopChannelTagAD
type PopChannelTagAD struct {
	TagID         int64 `json:"tag_id" form:"tag_id" validate:"required"`
	TopEntranceId int64 `json:"top_entrance_id" form:"top_entrance_id" validate:"top_entrance_id"`
	Deleted       int   `json:"deleted" form:"deleted" validate:"required"`
}

// TableName .
func (a PopChannelTagAD) TableName() string {
	return "popular_channel_tag"
}

// TableName .
func (a PopChannelTag) TableName() string {
	return "popular_channel_tag"
}
