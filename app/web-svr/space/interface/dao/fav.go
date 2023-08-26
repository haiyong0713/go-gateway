package dao

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_favArchiveURI = "/x/internal/v2/fav/video"
	_favAlbumURI   = "/userext/v1/Fav/getMyFav"
)

// LiveFavCount get live(vc or album) favorite count.
func (d *Dao) LiveFavCount(c context.Context, mid int64, favType int) (count int, err error) {
	var (
		req *http.Request
		rs  struct {
			Code int `json:"code"`
			Data struct {
				PageInfo struct {
					Page      int `json:"page"`
					PageSize  int `json:"page_size"`
					TotalPage int `json:"total_page"`
					Count     int `json:"count"`
				} `json:"pageinfo"`
			} `json:"data"`
		}
		ip = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("biz_type", strconv.Itoa(favType))
	if req, err = d.httpR.NewRequest("GET", d.favAlbumURL, ip, params); err != nil {
		log.Error("d.httpR.NewRequest %s error(%v)", d.favAlbumURL, err)
		return
	}
	req.Header.Set("X-BILILIVE-UID", strconv.FormatInt(mid, 10))
	if err = d.httpR.Do(c, req, &rs); err != nil {
		log.Error("d.httpR.Get(%s,%d) error(%v)", d.favAlbumURL, mid, err)
		return
	}
	if rs.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s,%d) code(%d)", d.favAlbumURL, mid, rs.Code)
		err = ecode.Int(rs.Code)
		return
	}
	count = rs.Data.PageInfo.Count
	return
}

// FavArchive fav archive.
func (d *Dao) FavArchive(c context.Context, mid int64, arg *model.FavArcArg) (res *model.SearchArchive, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	if mid > 0 {
		params.Set("mid", strconv.FormatInt(mid, 10))
	}
	params.Set("vmid", strconv.FormatInt(arg.Vmid, 10))
	params.Set("fid", strconv.FormatInt(arg.Fid, 10))
	if arg.Tid > 0 {
		params.Set("tid", strconv.FormatInt(arg.Tid, 10))
	}
	if arg.Keyword != "" {
		params.Set("keyword", arg.Keyword)
	}
	if arg.Order != "" {
		params.Set("order", arg.Order)
	}
	params.Set("pn", strconv.Itoa(arg.Pn))
	params.Set("ps", strconv.Itoa(arg.Ps))
	var rs struct {
		Code int                  `json:"code"`
		Data *model.SearchArchive `json:"data"`
	}
	if err = d.httpR.Get(c, d.favArcURL, ip, params, &rs); err != nil {
		log.Error("d.http.Get(%s,%d) error(%v)", d.favArcURL, mid, err)
		return
	}
	if rs.Code != ecode.OK.Code() {
		log.Error("d.http.Get(%s,%d) code(%d)", d.favArcURL, mid, rs.Code)
		err = ecode.Int(rs.Code)
		return
	}
	res = rs.Data
	return
}
