package note

import (
	"context"

	"go-common/library/log"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"
)

const (
	_cidValid   = 0
	_cidInvalid = 1
)

func (d *Dao) ToTags(c context.Context, oid int64, noteId int64, tagStr string, oidType int) (tags []*api.NoteTag, cidCnt int64, err error) {
	tags = note.ToTagArray(noteId, tagStr)
	if len(tags) == 0 {
		return nil, 0, err
	}
	pages, err := func() (map[int64]*note.PageCore, error) {
		if oidType == note.OidTypeUGC {
			return d.ViewPage(c, oid)
		}
		if oidType == note.OidTypeCheese {
			return d.SeasonEps(c, int32(oid))
		}
		return nil, xecode.NoteOidInvalid
	}()
	if err != nil {
		return nil, 0, err
	}
	cidCnt = int64(len(pages))
	for _, tag := range tags {
		if _, ok := pages[tag.Cid]; !ok {
			tag.Status = _cidInvalid
			log.Warn("noteWarn toTags noteId(%d) aid(%d) cid(%d) page not found", noteId, oid, tag.Cid)
			continue
		}
		tag.Status = _cidValid
		tag.Index = pages[tag.Cid].Index
	}
	return tags, cidCnt, nil
}
