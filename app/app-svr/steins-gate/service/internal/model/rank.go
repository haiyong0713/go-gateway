package model

import (
	xtime "go-common/library/time"
)

// RankItem is
type RankItem struct {
	Mid   int64
	Score int32
	MTime xtime.Time
}
