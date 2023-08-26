package feed

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
)

// Archives multi get archives.
func (d *Dao) Archives(c context.Context, aids []int64, ip string) (as map[int64]*api.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	var reply *api.ArcsReply
	arg := &api.ArcsRequest{Aids: aids}
	if reply, err = d.arcGRPC.Arcs(c, arg); err != nil {
		log.Error("d.arcGRPC.Arcs(%v) error(%+v)", arg, err)
		return
	}
	as = reply.Arcs
	return
}
