package show

// PopTopPhoto
type PopTopPhoto struct {
	ID         int64  `json:"id" form:"id"`
	TopPhoto   string `json:"top_photo" form:"top_photo" validate:"top_photo" gorm:"column:top_photo"`
	LocationId int64  `json:"location_id" form:"location_id" validate:"location_id"`
	Deleted    int    `json:"deleted" form:"deleted" validate:"required"`
}

// PopTopPhotoAD
type PopTopPhotoAD struct {
	TopPhoto   string `json:"top_photo" form:"top_photo" validate:"top_photo" gorm:"column:top_photo"`
	LocationId int64  `json:"location_id" form:"location_id" validate:"location_id"`
	Deleted    int    `json:"deleted" form:"deleted" validate:"required"`
}

// TableName .
func (a PopTopPhotoAD) TableName() string {
	return "popular_top_photo"
}

// TableName .
func (a PopTopPhoto) TableName() string {
	return "popular_top_photo"
}
