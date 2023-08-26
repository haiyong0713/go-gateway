package api

const (
	// Season Attribute
	AttrSnFinished = uint(0)
	AttrSnYes      = int64(1)
	AttrSnType     = uint(2)
	AttrSnActType  = uint(1)
	//是否设定动态空间禁止项(防刷屏)
	AttrSnNoSpace = uint(4)
	//是否是青少年模式
	AttrSnTeenager = uint(5)

	//是否是付费合集
	SeasonAttrSnPay = uint(6)

	//是否是免费试看
	EpisodeAttrSnFreeWatch = uint(2)
)

// AttrVal get attr val by bit.
func (sn *Season) AttrVal(bit uint) int64 {
	return (sn.Attribute >> bit) & int64(1)
}

func (sArc *Arc) AttrVal(bit uint) int64 {
	return (sArc.Attribute >> bit) & int64(1)
}

// AttrVal get attr val by bit.
func (ep *Episode) AttrVal(bit uint) int64 {
	return (ep.Attribute >> bit) & int64(1)
}
