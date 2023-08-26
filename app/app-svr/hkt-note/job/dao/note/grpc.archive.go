package note

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/model/util"
	"time"

	archive "git.bilibili.co/bapis/bapis-go/archive/service"
	upArc "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

func (d *Dao) ArcPassed(ctx context.Context, mid, lastest int64) ([]int64, error) {
	var (
		err       error
		aids      []int64
		reply     *upArc.ArcPassedReply
		page      = int64(1)
		defaultPs = int64(20)
	)

	for {
		time.Sleep(20 * time.Millisecond)
		reply, err = d.grpc.upArc.ArcPassed(ctx, &upArc.ArcPassedReq{
			Mid: mid,
			Pn:  page,
			Ps:  defaultPs,
		})
		if err != nil {
			log.Errorc(ctx, "d.ArcPassed(%d) error(%v)", mid, err)
			return nil, err
		}

		for _, v := range reply.Archives {
			if int64(v.PubDate) < lastest {
				continue
			}
			aids = append(aids, v.Aid)
		}
		page++

		if len(reply.Archives) < int(defaultPs) {
			break
		}

		if int64(reply.Archives[len(reply.Archives)-1].PubDate) < lastest {
			break
		}
	}
	return aids, nil
}

// BatchArchives gets batch archies.
func (d *Dao) BatchArchives(ctx context.Context, oids []int64) (res map[int64]*archive.Arc, err error) {
	oids = util.Int64ZeroRemoval(oids)
	res = make(map[int64]*archive.Arc, len(oids))
	batchSize := 50
	for len(oids) > 0 {
		if batchSize > len(oids) {
			batchSize = len(oids)
		}
		req := &archive.ArcsRequest{
			Aids: oids[:batchSize],
		}
		reply, err := d.grpc.archive.Arcs(ctx, req)
		if err != nil {
			log.Errorc(ctx, "d.dao.BatchArchives(%+v) error(%+v)", req, err)
			return nil, err
		}
		for id, arc := range reply.Arcs {
			res[id] = arc
		}
		oids = oids[batchSize:]
		time.Sleep(10 * time.Millisecond)
	}
	return
}

func (d *Dao) Arcs(ctx context.Context, aids []int64) (arcsReply *archive.ArcsReply, err error) {
	arcsReply = &archive.ArcsReply{}
	if len(aids) <= 0 {
		return arcsReply, nil
	}
	if arcsReply, err = d.grpc.archive.Arcs(ctx, &archive.ArcsRequest{Aids: aids}); err != nil {
		log.Error("d.Arc aids:%+v d.grpc.ArchiveClient.Arcs err:%+v", aids, err)
		return arcsReply, err
	}
	return arcsReply, nil
}
