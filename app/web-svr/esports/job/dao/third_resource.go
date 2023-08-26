package dao

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
)

const (
	_sqlThirdDataSave     = "REPLACE INTO `es_third_resource_data` (`resource_id`, `resource_data`) VALUES (?, ?)"
	ResourceIDTournament  = "score:tournament:%s"
	ResourceIDRoundInfo   = "score:round:info:%s"
	ResourceIDRoundTree   = "score:round:tree:%s:%s"
	ResourceIDRankingData = "score:ranking:data:%s"
)

func (d *Dao) SaveThirdResourceData(c context.Context, resourceID string, resourceData interface{}) (err error) {
	var b []byte
	b, err = json.Marshal(resourceData)
	if err != nil {
		err = errors.Wrapf(err, "SaveThirdResourceData json.Marshal error")
		return
	}
	if _, err = d.db.Exec(c, _sqlThirdDataSave, resourceID, string(b)); err != nil {
		err = errors.Wrapf(err, "SaveThirdResourceData db.Exec error")
	}
	return
}
