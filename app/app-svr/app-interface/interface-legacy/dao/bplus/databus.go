package bplus

import (
	"context"
	"strconv"

	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	"github.com/pkg/errors"
)

// NotifyContribute .
//
//nolint:bilirailguncheck
func (d *Dao) NotifyContribute(c context.Context, vmid int64, attrs *space.Attrs, ctime xtime.Time, isCooperation, isComic bool) (err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	value := struct {
		Vmid          int64        `json:"vmid"`
		Attrs         *space.Attrs `json:"attrs"`
		CTime         xtime.Time   `json:"ctime"`
		IP            string       `json:"ip"`
		IsCooperation bool         `json:"is_cooperation"`
		IsComic       bool         `json:"is_comic"`
	}{vmid, attrs, ctime, ip, isCooperation, isComic}
	if err = d.pub.Send(c, strconv.FormatInt(vmid, 10), value); err != nil {
		err = errors.Wrapf(err, "%v", value)
	}
	return
}
