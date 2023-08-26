package like

import (
	xtime "go-common/library/time"
)

type BigVUser struct {
	ID         int64
	Mid        int64
	InviterMid int64
	Ctime      xtime.Time
	IsIdentity int64
}
