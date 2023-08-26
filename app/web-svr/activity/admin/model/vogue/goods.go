package vogue

import (
	"strconv"
	"strings"

	"go-common/library/log"
	"go-common/library/time"
)

const (
	GoodsAttrReal    = 0
	GoodsAttrSellOut = 1
)

// ListRsp
type GoodsListRsp struct {
	List []*GoodsData `json:"list"`
}

// GoodsData 商品信息
type GoodsData struct {
	ID      int    `json:"id" gorm:"column:id"`
	Name    string `json:"name" gorm:"column:name"`
	Picture string `json:"picture" gorm:"column:picture"`
	// 商品品类 逗号分割
	TagStr string `json:"tagstr" gorm:"column:type"`
	Tags   []int  `json:"tags"`
	// 需要积分
	Score int `json:"score" gorm:"column:score"`
	// 已送出多少数量
	Send int `json:"send" gorm:"column:send"`
	// 总库存数
	Stock int `json:"stock" gorm:"column:stock"`
	// 剩余库存数 = stock - send
	LeftStock int `json:"left_stock"`
	// 想要人数
	Want int `json:"want" gorm:"column:want"`
	// 属性位，0 - 是否虚拟，1 - 是否领完
	Attr    int `json:"attr" gorm:"column:attr"`
	Type    int `json:"type"`
	SoldOut int `json:"soldout"`
}

// Goods
type Goods struct {
	GoodsData
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// tag string to tags list
func (g *GoodsData) ExtractTags() (extractTags []int, err error) {
	var (
		tag int
	)
	tags := strings.Split(g.TagStr, ",")
	for _, tagStr := range tags {
		if tag, err = strconv.Atoi(tagStr); err != nil {
			log.Error("ExtractTags strconv.Atoi tag err(%v), tag(%s)", err, tagStr)
			return
		}
		extractTags = append(extractTags, tag)
	}
	return
}

func (g *GoodsData) ExtractLeftStock() int {
	return g.Stock - g.Send
}

func (g *GoodsData) AttrVal(bit uint) int {
	return (g.Attr >> bit) & 1
}

// TableName ActVogueGoods def
func (Goods) TableName() string {
	return "act_vogue_goods"
}

// GoodsAddRequest Server层接收到的参数
type GoodsAddRequest struct {
	Name string `form:"name"`
	// 商品类型
	Type int `form:"type"`
	// 商品品类
	Tags    []int  `form:"tags"`
	Picture string `form:"picture"`
	Score   int    `form:"score"`
	Stock   int    `form:"stock"`
}

// GoodsAddParam
type GoodsAddParam struct {
	Name     string `form:"name"`
	Picture  string `form:"picture"`
	AttrReal int    `form:"type"`
	Type     string
	Tags     []string `form:"tags,split"`
	Score    int      `form:"score"`
	Stock    int      `form:"stock"`
}

// tag list to tags string
func (g *GoodsAddParam) TagsToType() string {
	var tags []string
	for _, tag := range g.Tags {
		tags = append(tags, tag)
	}
	return strings.Join(tags, ",")
}

// GoodsModifyParam
type GoodsModifyParam struct {
	GoodsAddParam
	ID int `form:"id"`
}
