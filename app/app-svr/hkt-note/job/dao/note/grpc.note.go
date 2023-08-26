package note

import (
	"context"
	"go-common/library/log"
	note "go-gateway/app/app-svr/hkt-note/service/api"
)

func (d *Dao) ArcNotesCount(ctx context.Context, oid, oType int64) (int64, error) {
	reply, err := d.grpc.note.ArcNotesCount(ctx, &note.ArcNotesCountReq{
		Oid:     oid,
		OidType: oType,
	})
	if err != nil {
		log.Errorc(ctx, "d.ArcNotesCount(%d %d) error(%v)", oid, oType, err)
		return 0, err
	}
	return reply.NotesCount, nil
}
