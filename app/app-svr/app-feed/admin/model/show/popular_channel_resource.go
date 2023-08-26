package show

// PopChannelResource
type PopChannelResource struct {
	ID            int64 `json:"id" form:"id"`
	RID           int64 `json:"rid" form:"rid" validate:"rid" gorm:"column:rid"`
	TagID         int64 `json:"tag_id" form:"tag_id" validate:"tag_id"`
	TopEntranceId int64 `json:"top_entrance_id" form:"top_entrance_id" validate:"top_entrance_id"`
	Deleted       int   `json:"deleted" form:"deleted" validate:"required"`
	State         int   `json:"state" form:"state" validate:"state"`
}

// PopChannelResourceAD
type PopChannelResourceAD struct {
	RID           int64 `json:"rid" form:"rid" validate:"rid" gorm:"column:rid"`
	TagID         int64 `json:"tag_id" form:"tag_id" validate:"tag_id"`
	TopEntranceId int64 `json:"top_entrance_id" form:"top_entrance_id" validate:"top_entrance_id"`
	Deleted       int   `json:"deleted" form:"deleted" validate:"required"`
	State         int   `json:"state" form:"state" validate:"state"`
}

// TableName .
func (a PopChannelResourceAD) TableName() string {
	return "popular_channel_resource"
}

// TableName .
func (a PopChannelResource) TableName() string {
	return "popular_channel_resource"
}

type PopAIChannelResource struct {
	RID      int64  `json:"id" form:"id" validate:"id"`
	TagId    int64  `json:"tag_id" form:"tag_id" validate:"tag_id"`
	Goto     string `json:"goto"`
	FromType string `json:"from_type"`
}

type PopularCard struct {
	Type     string `json:"type"`
	Value    int64  `json:"value"`
	FromType string `json:"from_type"`
	TagIdStr string `json:"tag_id"`
	TagId    []int64
}
