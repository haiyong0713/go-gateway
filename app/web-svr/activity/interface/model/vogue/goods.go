package model

const (
	GoodsAttrReal    = 0
	GoodsAttrSellOut = 1
)

type Goods struct {
	Id      int64
	Name    string
	Picture string
	Type    string
	Score   int64
	Send    int64
	Stock   int64
	Want    int64
	Attr    int64
}

func (g *Goods) AttrVal(bit uint) int64 {
	return (g.Attr >> bit) & int64(1)
}

func (g *Goods) AttrSet(v int64, bit uint) {
	g.Attr = g.Attr&(^(1 << bit)) | (v << bit)
}
