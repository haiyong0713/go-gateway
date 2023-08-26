package model

// Kratos hello kratos.
type Kratos struct {
	Hello string
}

type Article struct {
	ID      int64
	Content string
	Author  string
}

const (
	// 冻结状态
	FreezeFalse = int64(0)
	FreezeTrue  = int64(1)

	IsDeletedTrue  = int64(1)
	IsDeletedFalse = int64(0)

	//积分赛
	SeriesTypPoint = 1
	//淘汰赛
	SeriesTypKnockout = 2
)
