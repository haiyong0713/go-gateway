package note

import (
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/service/api"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cepgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
)

const (
	_tagLen = 4
)

type PageCore struct {
	Id     int64 `json:"id"`
	Status int64 `json:"status"`
	Index  int64 `json:"index"`
}

func ToArcPage(data *arcapi.Page) *PageCore {
	return &PageCore{
		Id:    data.Cid,
		Index: int64(data.Page),
	}
}

func ToEpPage(data *cepgrpc.EpisodeModel) *PageCore {
	return &PageCore{
		Id:    int64(data.Id),
		Index: int64(data.No),
	}
}

func ToTagArray(noteId int64, str string) []*api.NoteTag {
	tagsArr := strings.Split(str, ",")
	res := make([]*api.NoteTag, 0)
	if len(tagsArr) == 0 {
		return res
	}
	for _, val := range tagsArr {
		tagArr := strings.Split(val, "-")
		if len(tagArr) < _tagLen {
			continue
		}
		var (
			cid   int64
			index int64
			sec   int64
			pos   int64
		)
		ok := func() bool {
			var err error
			if cid, err = strconv.ParseInt(tagArr[0], 10, 64); err != nil {
				return false
			}
			if index, err = strconv.ParseInt(tagArr[1], 10, 64); err != nil {
				return false
			}
			if sec, err = strconv.ParseInt(tagArr[2], 10, 64); err != nil {
				return false
			}
			if pos, err = strconv.ParseInt(tagArr[3], 10, 64); err != nil {
				return false
			}
			return true
		}()
		if !ok {
			log.Warn("noteWarn ToTagArray noteId(%d) val(%s) invalid", noteId, val)
			continue
		}
		tag := &api.NoteTag{
			Cid:     cid,
			Index:   index,
			Seconds: sec,
			Pos:     pos,
		}
		res = append(res, tag)
	}
	return res
}
