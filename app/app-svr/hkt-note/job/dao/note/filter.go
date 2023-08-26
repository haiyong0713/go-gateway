package note

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/pkg/errors"
)

const (
	_areaNote   = "note"
	_codeOK     = 0
	_filterPath = "/x/internal/filter/v3/hit"
)

func (d *Dao) FilterV3(c context.Context, content string, noteId, mid int64) ([]*note.FilterData, error) {
	res := new(struct {
		Code int                `json:"code"`
		Msg  string             `json:"msg"`
		Data []*note.FilterData `json:"data"`
	})
	params := url.Values{}
	params.Set("area", _areaNote)
	params.Set("msg", content)
	params.Set("mid", strconv.FormatInt(mid, 10))
	requestUrl := fmt.Sprintf("%s%s", d.c.NoteCfg.Host.FilterHost, _filterPath)
	if err := d.client.Post(c, requestUrl, "", params, &res); err != nil { // 连接失败，需要重试
		return nil, errors.Wrapf(err, "FilterV3 client.Do failed ,noteId(%d) param(%v)", noteId, params)
	}
	// request error 无须重试,报警查原因
	if res.Code != _codeOK {
		log.Error("noteError FilterV3 noteId(%d) code err(%+v)", noteId, res)
		return nil, nil
	}
	log.Warn("noteInfo FilterV3 noteId(%d) content(%s) ucc with res(%+v)", noteId, content, res)
	return res.Data, nil
}
