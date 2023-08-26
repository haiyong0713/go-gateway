package note

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
)

func (d *Dao) NoteSize(c context.Context, noteId int64, mid int64) (*notegrpc.NoteSizeReply, error) {
	req := &notegrpc.NoteSizeReq{
		Mid:    mid,
		NoteId: noteId,
	}
	reply, err := d.noteClient.NoteSize(c, req)
	if err != nil {
		err = errors.Wrapf(err, "NoteSize req(%+v)", req)
		return nil, err
	}
	if reply == nil {
		log.Error("NoteError grpc NoteSize req(%+v) reply empty", req)
		return nil, ecode.NothingFound
	}
	return reply, nil
}

func (d *Dao) NoteInfo(c context.Context, noteId int64, mid int64) (*notegrpc.NoteInfoReply, error) {
	req := &notegrpc.NoteInfoReq{
		Mid:    mid,
		NoteId: noteId,
	}
	reply, err := d.noteClient.NoteInfo(c, req)
	if err != nil {
		err = errors.Wrapf(err, "NoteInfo req(%+v)", req)
		return nil, err
	}
	if reply == nil {
		log.Error("NoteError grpc NoteInfo req(%+v) reply empty", req)
		return nil, ecode.NothingFound
	}
	return reply, nil
}

func (d *Dao) NoteListArc(c context.Context, oid, mid int64, oidType int) (*notegrpc.NoteListInArcReply, error) {
	req := &notegrpc.NoteListInArcReq{
		Mid:     mid,
		Oid:     oid,
		OidType: int64(oidType),
	}
	reply, err := d.noteClient.NoteListInArc(c, req)
	if err != nil {
		err = errors.Wrapf(err, "NoteListArc req(%+v)", req)
		return nil, err
	}
	if reply == nil {
		log.Error("NoteError grpc NoteListArc req(%+v) reply empty", req)
		return nil, ecode.NothingFound
	}
	return reply, nil
}

func (d *Dao) NoteList(c context.Context, mid, pn, ps, oid, oidType, uperMid int64, tp notegrpc.NoteListType) (*notegrpc.NoteListReply, error) {
	req := &notegrpc.NoteListReq{
		Mid:     mid,
		Pn:      pn,
		Ps:      ps,
		Type:    tp,
		Oid:     oid,
		OidType: oidType,
		UperMid: uperMid,
	}
	reply, err := d.noteClient.NoteList(c, req)
	if err != nil {
		err = errors.Wrapf(err, "NoteList req(%+v)", req)
		return nil, err
	}
	if reply == nil {
		log.Error("NoteError grpc NoteList req(%+v) reply empty", req)
		return nil, ecode.NothingFound
	}
	return reply, nil
}

func (d *Dao) SimpleNotes(c context.Context, noteIds []int64, mid int64, tp notegrpc.SimpleNoteType) (map[int64]*notegrpc.SimpleNoteCard, error) {
	req := &notegrpc.SimpleNotesReq{
		NoteIds: noteIds,
		Mid:     mid,
		Tp:      tp,
	}
	reply, err := d.noteClient.SimpleNotes(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "SimpleNotes req(%+v)", req)
	}
	if reply == nil || reply.Items == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "SimpleNotes req(%+v)", req)
	}
	return reply.Items, nil
}

func (d *Dao) NoteCount(c context.Context, mid int64) (*notegrpc.NoteCountReply, error) {
	req := &notegrpc.NoteCountReq{
		Mid: mid,
	}
	reply, err := d.noteClient.NoteCount(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "NoteList req(%+v)", req)
	}
	if reply == nil {
		return nil, errors.Wrapf(ecode.NothingFound, "NoteList req(%+v)", req)
	}
	return reply, nil
}

func (d *Dao) ArcsForbid(c context.Context, aids []int64) (map[int64]bool, error) {
	req := &notegrpc.ArcsForbidReq{
		Aids: aids,
	}
	reply, err := d.noteClient.ArcsForbid(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "ArcsForbid req(%+v)", req)
	}
	if reply == nil || reply.Items == nil {
		return make(map[int64]bool), nil
	}
	return reply.Items, nil
}
