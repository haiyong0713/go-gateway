package egg

import (
	"go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

var (
	//NotDelete egg not deleted
	NotDelete uint8
	//Delete egg deleted
	Delete uint8 = 1
	//Publish egg publish
	Publish uint8 = 1
	//NotPublish egg not publish
	NotPublish uint8
	//OffLint egg off line
	OffLint uint8 = 2
	//EggVideo video egg
	EggVideo = 1
	//EggJump jusmp egg
	EggJump = 2
	//EggPic pic egg
	EggTypePic = 3
)

// Obj add egg object
type Obj struct {
	Query     []string  `json:"query" form:"query,split" validate:"required"`
	Stime     time.Time `json:"stime" form:"stime"`
	Etime     time.Time `json:"etime" form:"etime"`
	Type      int       `json:"type" form:"type" validate:"required"`
	ReType    int64     `json:"re_type" form:"re_type"`
	ReValue   string    `json:"re_value" form:"re_value"`
	ShowCount int       `json:"show_count" form:"show_count" validate:"required"`
	Plat      string    `json:"plat" form:"plat" validate:"required"`
	// v5.59新增字段
	PreTime          time.Time `json:"pre_time" form:"pre_time"`
	Mids             string    `json:"mid" form:"mids"`
	MaskTransparency int       `json:"mask_transparency" form:"mask_transparency"`
	MaskColor        string    `json:"mask_color" form:"mask_color"`
	Pic              string    `json:"pic" form:"pic"`
}

// ObjUpdate add egg object
type ObjUpdate struct {
	ID        uint      `form:"id" validate:"required"`
	Query     []string  `json:"query" form:"query,split" validate:"required"`
	Stime     time.Time `json:"stime" form:"stime"`
	Etime     time.Time `json:"etime" form:"etime"`
	ReType    int64     `json:"re_type" form:"re_type"`
	ReValue   string    `json:"re_value" form:"re_value"`
	ShowCount int       `json:"show_count" form:"show_count" validate:"required"`
	Type      int       `json:"type" form:"type" validate:"required"`
	Plat      string    `json:"plat" form:"plat" validate:"required"`
	// v5.59新增字段
	Publish          uint8
	PreTime          time.Time `json:"pre_time" form:"pre_time"`
	Mids             string    `json:"mid" form:"mids"`
	MaskTransparency int       `json:"mask_transparency" form:"mask_transparency"`
	MaskColor        string    `json:"mask_color" form:"mask_color"`
	Pic              string    `json:"pic" form:"pic"`
}

// Plat egg plat
type Plat struct {
	EggID      uint   `json:"egg_id"`
	Plat       uint8  `json:"plat"`
	Conditions string `json:"conditions"`
	Build      string `json:"build"`
	URL        string `json:"url"`
	Md5        string `json:"md5"`
	Size       uint   `json:"size"`
	Deleted    uint8  `json:"deleted"`
}

// Query egg query
type Query struct {
	EggID   uint
	Word    string
	STime   time.Time
	ETime   time.Time
	Deleted uint8
}

// Egg egg
type Egg struct {
	ID        uint
	Stime     time.Time
	Etime     time.Time
	ShowCount int
	Type      int   `json:"type" form:"type" validate:"required"`
	UID       int64 `gorm:"column:uid"`
	Publish   uint8
	Person    string
	Delete    uint8
	ReType    int64  `form:"re_type" gorm:"column:re_type" json:"re_type"`
	ReValue   string `form:"re_value" gorm:"column:re_value" json:"re_value"`
	// v5.59新增字段
	PreTime          time.Time `form:"pre_time" gorm:"column:pre_time" json:"pre_time"`
	Mids             string    `form:"mids" gorm:"column:mids" json:"mids"`
	MaskTransparency int       `form:"mask_transparency" gorm:"column:mask_transparency" json:"mask_transparency"`
	MaskColor        string    `form:"mask_color" gorm:"column:mask_color" json:"mask_color"`
}

// IndexParam Index egg index param
type IndexParam struct {
	ID     string `json:"id" form:"id"`         // ID
	Stime  string `json:"stime" form:"stime"`   // 开始时间
	Etime  string `json:"etime" form:"etime"`   // 结束时间
	Person string `json:"person" form:"person"` // 创建人
	Word   string `json:"word" form:"word"`     // 关键词
	Type   int    `json:"type" form:"type" validate:"required"`
	Ps     int    `json:"ps" form:"ps" default:"20"` // 分页大小
	Pn     int    `json:"pn" form:"pn" default:"1"`  // 第几个分页
}

// Index egg index
type Index struct {
	ID        uint      `json:"id"`
	Words     string    `json:"words"`
	Stime     time.Time `json:"stime"`
	Etime     time.Time `json:"etime"`
	Plat      []Plat    `json:"plat"`
	ShowCount int       `json:"show_count"`
	Type      int64     `json:"type"`
	ReType    int64     `json:"re_type"`
	ReValue   string    `json:"re_value"`
	Publish   uint8     `json:"publish"`
	Person    string    `json:"person"`
	// v5.59新增
	PreTime          time.Time `json:"pre_time"`
	Mids             string    `json:"mids"`
	MaskTransparency int       `json:"mask_transparency"`
	MaskColor        string    `json:"mask_color"`
	Pic              EggPic    `json:"pic"`
}

// IndexPager return values
type IndexPager struct {
	Item []*Index    `json:"item"`
	Page common.Page `json:"page"`
}

// SearchEgg for searching
type SearchEgg struct {
	ID      uint           `json:"id"`
	Words   []string       `json:"query_list"`
	Stime   time.Time      `json:"stime"`
	Etime   time.Time      `json:"etime"`
	Plat    map[uint8]Plat `json:"plat"`
	Type    int            `json:"type"`
	ReType  int64          `json:"re_type"`
	ReValue string         `json:"re_value"`
	//Plat      []Plat         `json:"plat"`
	ShowCount int   `json:"show_count"`
	Publish   uint8 `json:"publish"`
	// 559版本优化字段
	PreTime          time.Time `json:"pre_time"`
	Mids             string    `json:"mids"`
	MaskTransparency int       `json:"mask_transparency"`
	MaskColor        string    `json:"mask_color"`
	Pic              EggPic    `json:"pic"`
}

// SearchEggWeb for searching
type SearchEggWeb struct {
	ID    uint             `json:"id"`
	Words []string         `json:"query_list"`
	Stime time.Time        `json:"stime"`
	Etime time.Time        `json:"etime"`
	Plat  map[uint8][]Plat `json:"plat"`
	//Plat      []Plat         `json:"plat"`
	ShowCount int   `json:"show_count"`
	Publish   uint8 `json:"publish"`
}

type EggPic struct {
	EggID    uint   `json:"egg_id"`
	PicType  int    `json:"pic_type"`
	ShowTime int    `json:"show_time"`
	URL      string `json:"url"`
	Md5      string `json:"md5"`
	Size     int    `json:"size"`
	Deleted  uint8  `json:"deleted"`
}

// TableName Egg
func (a SearchEggWeb) TableName() string {
	return "egg"
}

// TableName Egg
func (a Egg) TableName() string {
	return "egg"
}

// TableName Egg plat
func (a Plat) TableName() string {
	return "egg_plat"
}

// TableName Egg query
func (a Query) TableName() string {
	return "egg_query"
}

// TableName Egg
func (a Index) TableName() string {
	return "egg"
}

// TableName Egg
func (a SearchEgg) TableName() string {
	return "egg"
}

// TableName Egg pic
func (a EggPic) TableName() string {
	return "egg_pic"
}
