package operation

import (
	"context"

	opmdl "go-gateway/app/web-svr/web-show/interface/model/operation"
)

// Notice Service
func (s *Service) Notice(c context.Context, arg *opmdl.ArgOp) (res map[string][]*opmdl.Operation) {
	res = s.operation(arg.Tp, arg.Rank, arg.Count)
	return
}
