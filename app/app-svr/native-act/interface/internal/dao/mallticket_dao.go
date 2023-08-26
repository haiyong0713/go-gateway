package dao

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/utils/collection"
)

const (
	_ticketFavStatusUri = "/api/ticket/user/favstatusinner"
	_ticketAddFavUri    = "/api/ticket/user/addfavinner"
	_ticketDelFavUri    = "/api/ticket/user/delfavinner"
)

type mallticketDao struct {
	host       string
	httpClient *bm.Client
}

func (d *mallticketDao) FavStatuses(c context.Context, itemIds []int64, mid int64) (map[int64]bool, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("item_id", collection.JoinSliceInt(itemIds, ","))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Result map[int64]bool `json:"result"`
		} `json:"data"`
	}
	if err := d.httpClient.Get(c, d.host+_ticketFavStatusUri, ip, params, &res); err != nil {
		log.Error("Fail to request mallticket.FavStatus, params=%+v error=%+v", params.Encode(), err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_ticketFavStatusUri+"?"+params.Encode())
		log.Error("Fail to request mallticket.FavStatus, error=%+v", err)
		return nil, err
	}
	statuses := make(map[int64]bool, len(res.Data.Result))
	for id, status := range res.Data.Result {
		if id <= 0 {
			continue
		}
		statuses[id] = status
	}
	return statuses, nil
}

func (d *mallticketDao) AddFav(c context.Context, itemId, mid int64) error {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("item_id", strconv.FormatInt(itemId, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpClient.Post(c, d.host+_ticketAddFavUri, ip, params, &res); err != nil {
		log.Error("Fail to request mallticket.AddFav, params=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_ticketAddFavUri+"?"+params.Encode())
		log.Error("Fail to request mallticket.AddFav, error=%+v", err)
		return err
	}
	return nil
}

func (d *mallticketDao) DelFav(c context.Context, itemId, mid int64) error {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("item_id", strconv.FormatInt(itemId, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpClient.Post(c, d.host+_ticketDelFavUri, ip, params, &res); err != nil {
		log.Error("Fail to request mallticket.DelFav, params=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_ticketDelFavUri+"?"+params.Encode())
		log.Error("Fail to request mallticket.DelFav, error=%+v", err)
		return err
	}
	return nil
}
