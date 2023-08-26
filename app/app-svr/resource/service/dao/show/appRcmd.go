package show

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_appSpecialCard = "SELECT id,title,`desc`,cover,re_type,re_value,corner,card,scover,gifcover,bgcover,reason," +
		"tab_uri,power_pic_sun,power_pic_night,`size`,width,height,url From special_card WHERE id in (%s)"
	_appRcmdOnline = "SELECT DISTINCT(card_value) FROM tianma_recom where rec_type='app_related' and card_type='app_rcmd_special'" +
		" and deleted=0 and stime <=? and etime >= ?"
	_appRcmdOnlineOld = "SELECT DISTINCT(param) from app_rcmd_pos WHERE `goto`='special' and stime <= ? and etime >= ? "
	_appRcmdRelatePgc = "SELECT id,card_type,card_value,rec_reason,position,plat_ver,pgc_ids FROM " +
		"tianma_recom WHERE rec_type='app_related' and card_type='app_rcmd_special' and deleted=0 and pgc_ids != '' and `state`=1 " +
		"and stime < ? and etime > ?"
)

func (d *Dao) AppSpecialCard(c context.Context) (rcs []*pb2.AppSpecialCard, err error) {
	//获取app相关推荐再投特殊卡ID
	specailCardIds, err := d.GetAppRcmdOnlineSpecailCardId(c)
	if err != nil {
		log.Error("dao.AppSpecialCard GetAppRcmdOnlineSpecailCardId err(%+v)", err)
		return
	}

	//当前无在投特殊卡，直接返回
	if len(specailCardIds) == 0 {
		return
	}

	var placeholders []string
	for i := 0; i < len(specailCardIds); i++ {
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf(_appSpecialCard, strings.Join(placeholders, ","))
	rows, err := d.dbMgr.Query(c, sql, specailCardIds...)

	if err != nil {
		log.Error("dao.AppSpecialCard query error (%+v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rc := &pb2.AppSpecialCard{}
		if err = rows.Scan(&rc.Id, &rc.Title, &rc.Desc, &rc.Cover, &rc.ReType, &rc.ReValue, &rc.Corner, &rc.Card,
			&rc.Scover, &rc.Gifcover, &rc.Bgcover, &rc.Reason, &rc.TabUri, &rc.PowerPicSun, &rc.PowerPicNight,
			&rc.Size_, &rc.Width, &rc.Height, &rc.Url); err != nil {
			log.Error("dao.AppSpecialCard rows.Scan err (%+v)", err)
			return
		}
		rcs = append(rcs, rc)
	}
	err = rows.Err()
	return
}

func (d *Dao) GetAppRcmdOnlineSpecailCardId(c context.Context) (ids []interface{}, err error) {
	ids = make([]interface{}, 0)
	idsMap := make(map[string]string)
	stime := time.Now().Add(time.Hour * time.Duration(d.c.ResourceParam.AppSpecailCardTimeSize))  //当前时间前 +2小时
	etime := time.Now().Add(time.Hour * time.Duration(-d.c.ResourceParam.AppSpecailCardTimeSize)) //当前时间 -2小时

	//重构后表中再投特殊卡ID
	rows, err := d.dbMgr.Query(c, _appRcmdOnline, stime, etime)
	if err != nil {
		log.Error("dao.GetAppRcmdOnlineSpecailCardId appRcmdOnline query error (%+v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			log.Error("dao.GetAppRcmdOnlineSpecailCardId appRcmdOnline rows.Scan err (%+v)", err)
			return
		}
		ids = append(ids, id)
		idsMap[id] = id
	}
	err = rows.Err()
	if err != nil {
		log.Error("dao.GetAppRcmdOnlineSpecailCardId appRcmdOnline rows.Err(%+v)", err)
	}

	//重构前表再投特殊卡ID
	rowsOld, err := d.db.Query(c, _appRcmdOnlineOld, stime, etime)
	if err != nil {
		log.Error("dao.GetAppRcmdOnlineSpecailCardId appRcmdOnlineOld query error (%+v)", err)
		return
	}
	defer rowsOld.Close()
	for rowsOld.Next() {
		var id string
		if err = rowsOld.Scan(&id); err != nil {
			log.Error("dao.GetAppRcmdOnlineSpecailCardId  appRcmdOnlineOld rows.Scan err (%+v)", err)
			return
		}
		//过滤掉重复的ID
		if _, ok := idsMap[id]; !ok {
			ids = append(ids, id)
		}
	}
	err = rows.Err()
	return
}

func (d *Dao) AppRcmdRelatePgc(c context.Context, now time.Time) (appRcmdList []*model.AppRcmd, err error) {
	rows, err := d.dbMgr.Query(c, _appRcmdRelatePgc, now, now)
	if err != nil {
		log.Error("dao.AppRcmdRelatePgc query error (%+v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rc := &model.AppRcmd{}
		var platVerStr string
		if err = rows.Scan(&rc.ID, &rc.CardType, &rc.CardValue, &rc.RecReason, &rc.Position, &platVerStr, &rc.PgcIds); err != nil {
			log.Error("dao.AppRcmdRelatePgc rows err (%+v)", err)
			return
		}
		if platVerStr != "" {
			var PlatVer []*model.PlatVer
			if err = json.Unmarshal([]byte(platVerStr), &PlatVer); err != nil {
				log.Error("dao.AppRcmdRelatePgc json.Unmarshal(%s) error(%v)", platVerStr, err)
				continue
			}
			vm := make(map[int8][]*model.PlatVer, len(PlatVer))
			for _, v := range PlatVer {
				vm[v.Plat] = append(vm[v.Plat], v)
			}
			rc.PlatVer = vm
		}
		appRcmdList = append(appRcmdList, rc)
	}
	err = rows.Err()
	return
}
