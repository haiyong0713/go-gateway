package note

import (
	"context"
	"fmt"

	"go-common/library/log"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	bctype "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"
	bcgrpc "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
)

const (
	_targetPath = "bilibili.broadcast.message.note.SyncNotify"
	_roomID     = "note-sync://%d"
)

func (d *Dao) BroadcastSync(c context.Context, noteId int64, hash string) error {
	body, err := types.MarshalAny(&notegrpc.NoteSync{NoteId: noteId, Hash: hash})
	if err != nil {
		return errors.Wrapf(err, "BroadcastSync noteId(%d) hash(%s)", noteId, hash)
	}
	req := &bcgrpc.PushRoomReq{
		Opts: &bctype.PushOptions{
			AckType: bctype.PushOptions_USRE_ACK,
		},
		Msg: &bctype.Message{
			TargetPath: _targetPath,
			Body:       body,
		},
		RoomId: fmt.Sprintf(_roomID, noteId),
		Token:  d.c.NoteCfg.BroadcastToken,
	}
	reply, err := d.syncClient.PushRoom(c, req)
	if err != nil {
		return errors.Wrapf(err, "BroadcastSync req(%+v)", req)
	}
	log.Warn("noteInfo BroadcastSync req(%+v) res(%+v)", req, reply)
	return nil
}
