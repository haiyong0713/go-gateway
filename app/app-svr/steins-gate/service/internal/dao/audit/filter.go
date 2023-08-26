package audit

import (
	"context"

	filtergrpc "git.bilibili.co/bapis/bapis-go/filter/service"
)

// MFilterMsg .
func (d *Dao) MFilterMsg(c context.Context, area string, msgs map[string]string) (data *filtergrpc.MFilterReply, err error) {
	return d.filterClient.MFilter(c, &filtergrpc.MFilterReq{Area: area, MsgMap: msgs})
}

// FilterMsg .
func (d *Dao) FilterMsg(c context.Context, area, msg string) (data *filtergrpc.FilterReply, err error) {
	return d.filterClient.Filter(c, &filtergrpc.FilterReq{Area: area, Message: msg})

}
