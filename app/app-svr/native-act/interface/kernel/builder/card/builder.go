package card

import (
	"go-gateway/app/app-svr/native-act/interface/api"
)

type Builder interface {
	Build() *api.ModuleItem
}
