package component

import (
	"go-common/library/log/infoc.v2"
)

var (
	DWInfo infoc.Infoc
)

func InitDWRelations() (err error) {
	DWInfo, err = infoc.New(nil)
	return
}
