package feature

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/log"

	degrademdl "go-gateway/app/app-svr/feature/service/model/degrade"
)

const (
	_selectTvSwtich = "SELECT id,brand,chid,model,sys_version,config FROM switch_tv WHERE deleted=0 ORDER BY id desc"
)

func (d *Dao) TvSwitch(c context.Context) (map[string][]*degrademdl.TvSwitch, map[string]*degrademdl.TvSwitch, error) {
	rows, err := d.db.Query(c, _selectTvSwtich)
	if err != nil {
		return nil, nil, err
	}
	res := make(map[string][]*degrademdl.TvSwitch)
	resKeym := make(map[string]*degrademdl.TvSwitch)
	defer rows.Close()
	for rows.Next() {
		var (
			re         = new(degrademdl.TvSwitch)
			sysVersion string
			configStr  string
		)
		if err = rows.Scan(&re.ID, &re.Brand, &re.Chil, &re.Model, &sysVersion, &configStr); err != nil {
			return nil, nil, err
		}
		if re.Brand == "" && re.Chil == "" && re.Model == "" {
			log.Warn("TvSwtichs get empty brand chil model(%+v)", re)
			continue
		}
		var (
			tmpSysVersion *degrademdl.TvSwitchSysVersion
			tmpConfig     []string
		)
		if err = json.Unmarshal([]byte(sysVersion), &tmpSysVersion); err != nil {
			log.Error("%v", err)
			continue
		}
		re.SysVersion = tmpSysVersion
		var key = fmt.Sprintf("%v_%v_%v", re.Brand, re.Chil, re.Model)
		if _, ok := resKeym[key]; !ok {
			resKeym[key] = re
		}
		if tmpConfig = strings.Split(configStr, ","); len(tmpConfig) < 1 {
			log.Warn("TvSwtichs get config(%+v) len 0", re)
			continue
		}
		for _, config := range tmpConfig {
			res[config] = append(res[config], re)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return res, resKeym, nil
}
