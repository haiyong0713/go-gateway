package dao

import (
	"context"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
)

const (
	_thirdRecordUri = "/x/third/record"

	DataTypeNatSource     = "native_source"
	ActTypeUpg            = "auto_upgrade_act"
	ActTypeState2Up       = "sp_type_to_up"
	ActTypeState2Operator = "sp_type_to_admin"
)

func (d *Dao) ThirdRecord(c context.Context, dataId int64, dataType, actionType, dataNext, username string) error {
	reqUrl := d.c.Host.ActTmpl + _thirdRecordUri
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("dataId", strconv.FormatInt(dataId, 10))
	params.Set("dataType", dataType)
	params.Set("actionType", actionType)
	params.Set("dataNext", dataNext)
	params.Set("username", username)
	var rly struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := d.client.Post(c, reqUrl, ip, params, &rly); err != nil {
		log.Error("Fail to thirdRecord, params=%+v error=%+v", params.Encode(), err)
		return err
	}
	if rly.Code != ecode.OK.Code() {
		log.Error("Fail to thirdRecord, params=%+v rly=%+v", params.Encode(), rly)
		return errors.Wrap(ecode.Int(rly.Code), reqUrl+"?"+params.Encode())
	}
	return nil
}
