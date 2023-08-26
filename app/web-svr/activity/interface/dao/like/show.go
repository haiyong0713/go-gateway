package like

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	_ticketAddWishURI  = "/api/ticket/api/ticket/user/addWish"
	_ticketFavCountURI = "/api/ticket/project/favCount"
	_ticketAddFavInner = "/api/ticket/user/addfavinner"
)

func (d *Dao) TicketAddWish(c context.Context, ticketID int64, ck string) (err error) {
	params := url.Values{}
	params.Set("item_id", strconv.FormatInt(ticketID, 10))
	var req *http.Request
	if req, err = d.client.NewRequest(http.MethodPost, d.ticketAddWishURL, metadata.String(c, metadata.RemoteIP), params); err != nil {
		return
	}
	req.Header.Set("Cookie", ck)
	var res struct {
		Errno int    `json:"errno"`
		Msg   string `json:"msg"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "TicketAddWish d.client.Do ticketID:%d", ticketID)
		return
	}
	if res.Errno != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Errno), "TicketAddWish ticketID:%d msg:%s", ticketID, res.Msg)
	}
	return
}

func (d *Dao) TicketAddFavInner(c context.Context, mallID int64, mid int64) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("item_id", strconv.FormatInt(mallID, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))

	req, _ := d.addFavInnerClient.NewRequest(http.MethodPost, d.ticketAddFavInnerURL, metadata.String(c, metadata.RemoteIP), params)

	var res struct {
		Errno int      `json:"errno"`
		Msg   string   `json:"msg"`
		Data  struct{} `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "TicketAddFavInner d.client.Do mallID:%d", mallID)
		return
	}

	if res.Errno != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Errno), "TicketAddFavInner mallID:%d msg:%s", mallID, res.Msg)
	}
	return
}

func (d *Dao) TicketFavCount(c context.Context, ticketID int64) (count int64, err error) {
	params := url.Values{}
	params.Set("item_id", strconv.FormatInt(ticketID, 10))
	var res struct {
		Errno int    `json:"errno"`
		Msg   string `json:"msg"`
		Data  struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.ticketFavCountURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "TicketFavCount d.client.Get ticketID:%d", ticketID)
		return
	}
	if res.Errno != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Errno), "TicketFavCount ticketID:%d", ticketID)
		return
	}
	count = res.Data.Count
	return
}
