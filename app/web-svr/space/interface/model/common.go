package model

import api "git.bilibili.co/bapis/bapis-go/up-archive/service"

const (
	FilterLevel = 30
	FilterArea  = "space"
)

// AttrVal get attr val by bit.
func AttrVal(a *api.Arc, bit uint) int32 {
	return (a.Attribute >> bit) & int32(1)
}
