package note

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
)

type Dao struct {
	// grpc
	noteGRPC notegrpc.HktNoteClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.noteGRPC, err = notegrpc.NewClient(c.NoteClient); err != nil {
		panic(fmt.Sprintf("notegrpc NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) NoteCount(c context.Context, mid int64) (int64, error) {
	reply, err := d.noteGRPC.NoteCount(c, &notegrpc.NoteCountReq{Mid: mid})
	if err != nil {
		return 0, errors.Wrapf(err, "NoteCount mid(%d)", mid)
	}
	if reply == nil {
		return 0, errors.Wrapf(ecode.NothingFound, "NoteCount mid(%d)", mid)
	}
	return reply.NoteCount, nil
}
