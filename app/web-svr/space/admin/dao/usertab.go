package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
)

const (
	_tableusertab = "space_usertab"
	dateFormat    = "2006-01-02 15:04:05"
	// status
	ONLINE          = 1
	OFFLINE         = 0
	LastTime        = 2147454847
	Sentinel        = 93600
	_offlineUsertab = "/x/admin/native_page/native/ts/space/offline"
	NativeType      = "napage"
	UpTabType       = "up_act"
)

/*
func (d *Dao) NoticeUpTab(c context.Context, mid, pageID int64) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("page_id", strconv.FormatInt(pageID, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.http.Post(c, d.onlineUpSpaceURL, "", params, &res); err != nil {
		log.Error("dao.NoticeUpTab d.client.Post(%s) error(%+v)", d.onlineUpSpaceURL +"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("NoticeUpTab url(%s) res code(%d)", d.onlineUpSpaceURL+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
	}
	return
}
*/

// SpaceUserTab Add
func (d *Dao) SpaceUserTabAdd(arg *model.UserTabReq) (err error) {
	var limitsStr []byte
	limitsStr, err = json.Marshal(arg.Limits)
	if err != nil {
		log.Error("dao.SpaceUserTabAdd Marshal fail, value(%v) error(%v)", arg, err)
		return
	}
	arg.LimitsStr = string(limitsStr)

	if err = d.DB.Table(_tableusertab).Create(arg).Error; err != nil {
		log.Error("dao.SpaceUserTabAdd value(%v) error(%v)", arg, err)
		return
	}
	return
}

// SpaceUserTab Modify
func (d *Dao) SpaceUserTabModify(arg *model.UserTabReq) (err error) {
	var limitsStr []byte
	limitsStr, err = json.Marshal(arg.Limits)
	if err != nil {
		log.Error("dao.SpaceUserTabModify Marshal fail, value(%v) error(%v)", arg, err)
		return
	}
	attrToUpdates := map[string]interface{}{
		"mid":        arg.Mid,
		"tab_type":   arg.TabType,
		"tab_name":   arg.TabName,
		"tab_cont":   arg.TabCont,
		"tab_order":  arg.TabOrder,
		"stime":      arg.Stime,
		"etime":      arg.Etime,
		"is_default": arg.IsDefault,
		"limits":     string(limitsStr),
		"h5_link":    arg.H5Link,
	}
	if err = d.DB.Table(_tableusertab).Where("id = ?", arg.ID).Updates(attrToUpdates).Error; err != nil {
		log.Error("dao.SpaceUserTabModify value(%v) error(%v)", arg, err)
		return
	}
	return
}

func (d *Dao) SpaceUserTabOnline(id int64, arg *model.UserTabReq) (err error) {
	up := map[string]interface{}{
		"online": arg.Online,
		"stime":  arg.Stime,
		"etime":  arg.Etime,
	}
	if err = d.DB.Table(_tableusertab).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.SpaceUserTabOnline id(%d) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) SpaceUserTabDelete(id int64, t int) (err error) {
	up := map[string]interface{}{
		"deleted": model.Deleted,
	}
	//nolint:gomnd
	if t == 2 {
		if err = d.DB.Table(_tableusertab).Where("deleted=0").Update(up).Error; err != nil {
			log.Error("dao.SpaceUserTabDelte id(%d) error(%v)", id, err)
			return
		}
		return
	}
	if err = d.DB.Table(_tableusertab).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.SpaceUserTabDelte id(%d) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) SpaceUserTabList(arg *model.UserTabListReq) (list []*model.UserTabListReply, count int, err error) {
	w := map[string]interface{}{
		"deleted": model.NotDelete,
	}
	query := d.DB.Table(_tableusertab)
	if arg.TabType != 0 {
		query = query.Where("tab_type = ? ", arg.TabType)
	}
	if arg.Mid != 0 {
		query = query.Where("mid = ?", arg.Mid)
	}
	if arg.Online != -1 {
		query = query.Where("online = ?", arg.Online)
	}
	if err = query.Where(w).Count(&count).Error; err != nil {
		log.Error("dao.SpaceUserTabList arg(%v) count error(%v)", arg, err)
		return
	}
	if count == 0 {
		list = make([]*model.UserTabListReply, 0)
		return
	}
	list = make([]*model.UserTabListReply, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((arg.Pn - 1) * arg.Ps).Limit(arg.Ps).Find(&list).Error; err != nil {
		log.Error("dao.SpaceUserTabList query(%v) error(%v)", arg, err)
		return
	}
	for _, item := range list {
		if item.Mid != 0 {
			var midinfo *midrpc.CardReply
			if midinfo, err = d.MidInfoReply(context.Background(), item.Mid); err != nil {
				log.Error("dao.SpaceUserTabList get mid(%v) error(%v)", item.Mid, err)
				return
			}
			item.MidName = midinfo.Card.Name
			item.Official = midinfo.Card.Official.Role
		}

		if item.Etime == LastTime {
			item.Etime = 0
		}

		item.Limits = make([]*model.Limit, 0)
		if err = json.Unmarshal([]byte(item.LimitsStr), &item.Limits); err != nil {
			log.Error("dao.SpaceUserTabList Unmarshal limit(%v) error(%v)", item.LimitsStr, err)
			return
		}

	}
	return
}

func (d *Dao) ValidUserTabFindByMid(mid int64) (ret *model.UserTabReq, err error) {
	ret = &model.UserTabReq{}
	if err = d.DB.Table(_tableusertab).Where("mid=? and online=1 and deleted=0", mid).First(ret).Error; err != nil {
		log.Error("dao.SpaceUserTabFindByMid find(%v) error(%v)", ret, err)
		return
	}
	return
}

func (d *Dao) EtimeUserTabFindByMid(arg *model.UserTabReq) (ret bool, err error) {
	var (
		tmp []*model.UserTabReq
	)
	tmp = make([]*model.UserTabReq, 0)
	if err = d.DB.Table(_tableusertab).Where("mid=? and deleted=0", arg.Mid).Order("`etime` DESC").Find(&tmp).Error; err != nil {
		log.Error("dao.SpaceUserTabFindByMid find(%v) error(%v)", ret, err)
		return
	}
	if len(tmp) == 0 {
		return true, nil
	}
	tmp = append(tmp, &model.UserTabReq{
		ID:    -1,
		Etime: Sentinel,
	})
	for _, row := range tmp {
		if arg.ID == row.ID {
			continue
		}
		if arg.Stime > row.Etime {
			return true, nil
		} else if arg.Etime < row.Stime {
			continue
		} else if arg.Etime > row.Stime {
			return false, nil
		}
	}
	return false, nil
}

func (d *Dao) SpaceUserTabFindById(id int64) (ret *model.UserTabReq, err error) {
	ret = &model.UserTabReq{}
	if err = d.DB.Table(_tableusertab).Where("id = ?", id).First(ret).Error; err != nil {
		log.Error("dao.SpaceUserTabFindByMid find(%v) error(%v)", ret, err)
		return
	}
	return
}

func (d *Dao) MidInfoReply(c context.Context, mid int64) (res *midrpc.CardReply, err error) {
	var midinfo *midrpc.CardReply
	arg := &midrpc.MidReq{
		Mid: mid,
	}
	midinfo, err = d.midClient.Card3(c, arg)
	if err != nil {
		err = fmt.Errorf("Get MidInfo error")
		log.Error("MidInfoReply req(%v) err(%v)", mid, err)
		return nil, err
	}
	if midinfo == nil {
		err = fmt.Errorf("无效Mid(%v)", mid)
		return nil, err
	}
	res = midinfo
	return
}

func (d *Dao) FindUserTabByTime(state int) (ret []*model.UserTabReq, err error) {
	now := time.Now().Format(dateFormat)
	query := d.DB.Table(_tableusertab)
	ret = make([]*model.UserTabReq, 0)
	if state == OFFLINE {
		if err = query.Where("etime < ? and online=1 and deleted=0", now).Find(&ret).Error; err != nil {
			log.Error("dao.FindUserTabByTime.Find error(%v)", err)
			return
		}
	}
	if state == ONLINE {
		if err = query.Where("stime < ? and etime > ? and online=0 and deleted=0", now, now).Find(&ret).Error; err != nil {
			log.Error("dao.FindUserTabByTime.Find error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) NoticeUserTab(mid, pageId int64, tabType string) (err error) {
	var (
		ret = struct {
			Code int    `json:"Code"`
			Msg  string `json:"message"`
		}{}
	)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("page_id", strconv.FormatInt(pageId, 10))
	params.Set("tab_type", tabType)

	if err = d.http.Post(context.Background(), d.usertabURL, "", params, &ret); err != nil {
		log.Error("dao.NoticeUserTab d.http.Post url(%s) error(%+v)", d.usertabURL, err)
		return
	}
	if ret.Code != 0 {
		err = ecode.Error(-400, "发送失败")
		log.Error("dao.NoticeUserTab url(%s) error code(%+v)", d.usertabURL, ret.Code)
		return
	}
	return
}
