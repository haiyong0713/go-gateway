package tianma

type PosRecItem struct {
	Id               int64  `gorm:"id" json:"id" form:"id"`
	FileStatus       int    `gorm:"file_status" json:"file_status" form:"file_status"`
	FilePath         string `gorm:"file_path" json:"file_path" form:"file_path"`
	FileRows         int64  `gorm:"file_rows" json:"file_rows" form:"file_rows"`
	FileTypeAddition int    `gorm:"file_type_addition" json:"file_type_addition" form:"file_type_addition"`
	FilePathAddition string `gorm:"file_path_addition" json:"file_path_addition" form:"file_path_addition"`
}

func (*PosRecItem) TableName() string {
	return "pos_rec"
}
