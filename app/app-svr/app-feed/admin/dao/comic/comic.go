package comic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/comic"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_comicInner = "/twirp/comic.v0.Comic/ComicInfo"
	_comicInfos = "/twirp/comic.v0.Comic/GetComicInfos"
)

// ComicTitle .
func (d *Dao) ComicTitle(c context.Context, id int64) (title string, err error) {
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	res := new(comic.Comics)
	if err = d.comicHTTPClient.Post(c, d.c.Host.ComicInner+_comicInner, "", params, &res); err != nil {
		log.Error("ComicTitle Req(%d) error(%v) res(%+v)", id, err, res)
		return "", fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.c.Host.ComicInner+_comicInner+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("ComicTitle Req(%d) error(%v) res(%+v)", id, err, res)
		return "", fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Comic, d.c.Host.ComicInner+_comicInner+"?"+params.Encode())
	}
	if res.Data == nil {
		return "", fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Comic, d.c.Host.ComicInner+_comicInner+"?"+params.Encode())
	}
	return res.Data.Title, nil
}

// ComicInfo .
func (d *Dao) ComicInfo(c context.Context, id int64) (data []*comic.ComicInfo, err error) {
	params := url.Values{}
	params.Set("ids", strconv.FormatInt(id, 10))
	res := new(comic.ComicInfos)
	if err = d.comicHTTPClient.Post(c, d.c.Host.ComicInner+_comicInfos, "", params, &res); err != nil {
		log.Error("ComicInfo Req(%d) error(%v) res(%+v)", id, err, res)
		return nil, fmt.Errorf(util.ErrorNetFmts, util.ErrorNet, d.c.Host.ComicInner+_comicInfos+"?"+params.Encode(), err.Error())
	}
	if res.Code != ecode.OK.Code() {
		log.Error("ComicInfo Req(%d) error(%v) res(%+v)", id, err, res)
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Comic, d.c.Host.ComicInner+_comicInner+"?"+params.Encode())
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf(util.ErrorRes, util.ErrorDataNull, d.c.UserFeed.Comic, d.c.Host.ComicInner+_comicInner+"?"+params.Encode())
	}
	return res.Data, nil
}
