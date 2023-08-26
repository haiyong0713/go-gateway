package model

import (
	"go-common/library/time"
)

// archive action .
const (
	Insert = "insert"
	Update = "update"
	Delete = "del"
	// archive flag
	CopyrightOriginal = int8(1)
	// fail redis key
	FailList = "dyreg_fail_list"
)

// AllRegKey all region key .
type AllRegKey struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// ResKey .
type ResKey struct {
	Reskey string
	Start  int
	End    int
}

// ArcMsg archive .
type ArcMsg struct {
	Action string      `json:"action"`
	Table  string      `json:"table"`
	New    *ArchiveSub `json:"new"`
	Old    *ArchiveSub `json:"old"`
}

// ArchiveSub archive .
type ArchiveSub struct {
	Aid       int64  `json:"aid"`
	PubTime   string `json:"pubtime"`
	State     int    `json:"state"`
	Typeid    int32  `json:"typeid"`
	Copyright int8   `json:"copyright"`
	Attribute int32  `json:"attribute"`
}

// CanPlay def.
func (v *ArchiveSub) CanPlay() bool {
	return v.State >= 0 || v.State == -6
}

// IsOriginArc .
func (v *ArchiveSub) IsOriginArc() bool {
	return v.Copyright == 1
}

// ActAid aid and action .
type ActAid struct {
	Aid    int64  `json:"aid"`
	Action string `json:"action"`
	TypeID int32  `json:"type_id"`
}

// RegionArc RegionArc
type RegionArc struct {
	Aid       int64
	Attribute int32
	Copyright int8
	PubDate   time.Time
	TypeID    int32
	State     int
}
