package popular

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	"github.com/jinzhu/gorm"
)

const (
	//ActionAddCPopular .
	ActionAddCPopular = "ActAddPopularStars"
	//ActionUpCPopular .
	ActionUpCPopular = "ActUpPopularStars"
	//ActionDelCPopular .
	ActionDelCPopular = "ActDelPopularStars"
	//ActionRejCPopular .
	ActionRejCPopular = "ActRejPopularStars"
	//_CardTypeUpRcmdNew 热门新星卡片多视频
	_CardTypeUpRcmdNew = "up_rcmd_new"
	//_CardTypeUpRcmdNewSingle 热门新星卡单视频
	_CardTypeUpRcmdNewSingle = "up_rcmd_new_single"
	//ActionAIAddCPopular .
	ActionAIAddCPopular = "ActAIAddPopularStars"
	//_CardSourceOperate popular stars build by operate
	_CardSourceOperate = 0
	//_CardSourceAI popular stars build by ai
	_CardSourceAI = 1
)

// PopularStarsList channel Popular list
func (s *Service) PopularStarsList(lp *show.PopularStarsLP) (pager *show.PopularStarsPager, err error) {
	pager = &show.PopularStarsPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.PopularStars{})
	query = query.Where("type IN (?)", []string{_CardTypeUpRcmdNew, _CardTypeUpRcmdNewSingle})
	if lp.ID > 0 {
		w["id"] = lp.ID
	}
	if lp.Status > 0 {
		w["status"] = lp.Status
	}
	if lp.Source >= 0 {
		if lp.Source == 0 {
			query = query.Where(map[string]interface{}{"source": _CardSourceOperate})
		} else {
			w["source"] = lp.Source
		}
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.LongTitle != "" {
		query = query.Where("long_title like ?", "%"+lp.LongTitle+"%")
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("popularSvc.PopularStarsList Index count error(%v)", err)
		return
	}
	Populars := make([]*show.PopularStars, 0)
	if err = query.Where(w).Order("`mtime` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&Populars).Error; err != nil {
		log.Error("popularSvc.PopularStarsList First error(%v)", err)
		return
	}
	type ContentAid struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		Bvid  string `json:"bvid,omitempty"`
	}
	for i := 0; i < len(Populars); i++ {
		var str = Populars[i].Content
		var aids []*ContentAid
		if err = json.Unmarshal([]byte(str), &aids); err != nil {
			err = nil
			continue
		}
		for j := 0; j < len(aids); j++ {
			if aids[j].Bvid, err = common.GetBvID(aids[j].ID); err != nil {
				err = nil
				continue
			}
		}
		if bts, err := json.Marshal(aids); err == nil {
			Populars[i].Content = string(bts)
		}
	}
	pager.Item = Populars
	return
}

// AddPopularStars add popular stars
func (s *Service) AddPopularStars(c context.Context, param *show.PopularStarsAP, name string, uid int64) (err error) {
	var (
		popStars *show.PopularStars
	)
	if popStars, err = s.ValidMid(param.Value); err != nil {
		fmt.Println("error")
		return
	}
	if popStars.ID != 0 {
		err = fmt.Errorf("up主ID 已存在")
		return
	}
	type ContentAid struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	cntAidsTmp := []*ContentAid{}
	if err := json.Unmarshal([]byte(param.Content), &cntAidsTmp); err != nil {
		return err
	}
	param.Type = _CardTypeUpRcmdNew
	param.Source = _CardSourceOperate
	param.Status = common.Pass
	if len(cntAidsTmp) == 1 {
		param.Type = _CardTypeUpRcmdNewSingle
	}
	if err = s.showDao.PopularStarsAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopularStars, name, uid, 0, ActionAddCPopular, param); err != nil {
		log.Error("popularSvc.AddPopularStars AddLog error(%v)", err)
		return
	}
	return
}

// UpdatePopularStars update channel Popular
func (s *Service) UpdatePopularStars(c context.Context, param *show.PopularStarsUP, name string, uid int64) (err error) {
	var (
		popStars *show.PopularStars
	)
	if popStars, err = s.ValidMid(param.Value); err != nil {
		return
	}
	if popStars.ID != 0 && popStars.ID != param.ID {
		err = fmt.Errorf("up主ID 已存在")
		return
	}
	type ContentAid struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	cntAidsTmp := []*ContentAid{}
	if err := json.Unmarshal([]byte(param.Content), &cntAidsTmp); err != nil {
		return err
	}
	param.Type = _CardTypeUpRcmdNew
	param.Status = common.Pass
	if len(cntAidsTmp) == 1 {
		param.Type = _CardTypeUpRcmdNewSingle
	}
	if err = s.showDao.PopularStarsUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopularStars, name, uid, 0, ActionUpCPopular, param); err != nil {
		log.Error("popularSvc.UpdatePopularStars AddLog error(%v)", err)
		return
	}
	return
}

// DeletePopularStars delete channel Popular
func (s *Service) DeletePopularStars(id int64, starsType, name string, uid int64) (err error) {
	var (
		popularCards []*Card
		mapPopu      map[string]bool
	)
	w := map[string]interface{}{
		"is_delete": common.NotDeleted,
		"card_type": "up_rcmd_new",
	}
	if err = s.showDao.DB.Model(&Card{}).Where(w).Where("card_value like ?", "%"+fmt.Sprintf("%d", id)+"%").Find(&popularCards).Error; err != nil {
		return
	}
	if len(popularCards) != 0 {
		mapPopu = make(map[string]bool)
		for _, v := range popularCards {
			popIds := strings.Split(v.CardValue, ",")
			for _, popId := range popIds {
				mapPopu[popId] = true
			}
		}
		if _, ok := mapPopu[strconv.Itoa(int(id))]; ok {
			return fmt.Errorf("卡片id(%d)已经配置热门卡片位置推荐，不能被删除", id)
		}
	}
	if err = s.showDao.PopularStarsDelete(id, starsType); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopularStars, name, uid, id, ActionDelCPopular, id); err != nil {
		log.Error("popularSvc.DeletePopularStars AddLog error(%v)", err)
		return
	}
	return
}

// RejectPopularStars reject channel Popular
func (s *Service) RejectPopularStars(id int64, starsType, name string, uid int64) (err error) {
	if err = s.showDao.PopularStarsReject(id, starsType); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopularStars, name, uid, id, ActionRejCPopular, id); err != nil {
		log.Error("popularSvc.DeletePopularStars AddLog error(%v)", err)
		return
	}
	return
}

// ValidMid mid must unique
func (s *Service) ValidMid(mid string) (popStars *show.PopularStars, err error) {
	w := map[string]interface{}{
		"value":   mid,
		"deleted": common.NotDeleted,
	}
	popStars = &show.PopularStars{}
	query := s.showDao.DB.Model(&show.PopularStars{})
	query = query.Where("type IN (?)", []string{_CardTypeUpRcmdNew, _CardTypeUpRcmdNewSingle})
	if err = query.Where(w).Find(&popStars).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		log.Error("popularSvc.ValidMid Find error(%v)", err)
		return
	}
	return
}

// AIAddPopularStars add popular stars
func (s *Service) AIAddPopularStars(c context.Context, values []*show.PopularStarsAP) (err error) {
	if err = util.AddLogs(common.LogPopularStars, "AI", 0, 0, ActionAIAddCPopular, values); err != nil {
		log.Error("popularSvc.AIAddPopularStars AddLog error(%v)", err)
		return
	}
	for _, v := range values {
		var popStars *show.PopularStars
		if popStars, err = s.ValidMid(v.Value); err != nil {
			log.Error("popularSvc.AIAddPopularStars ValidMid value(%v) error(%v)", v, err)
			continue
		}
		if popStars.ID != 0 {
			//运营创建的优先级最高
			if popStars.Source == _CardSourceOperate {
				continue
			}
			//已通过的优先级较高
			if popStars.Status == common.Pass {
				continue
			}
			tmp := &show.PopularStarsUP{
				ID:        popStars.ID,
				Content:   v.Content,
				LongTitle: v.LongTitle,
			}
			if err = s.showDao.PopularStarsUpdate(tmp); err != nil {
				log.Error("popularSvc.AIAddPopularStars PopularStarsUpdate value(%v) error(%v)", tmp, err)
				continue
			}
		} else {
			tmp := &show.PopularStarsAP{
				Type:      _CardTypeUpRcmdNew,
				Source:    _CardSourceAI,
				Status:    common.Verify,
				Value:     v.Value,
				Content:   v.Content,
				LongTitle: v.LongTitle,
			}
			if err = s.showDao.PopularStarsAdd(tmp); err != nil {
				log.Error("popularSvc.AIAddPopularStars PopularStarsAdd value(%v) error(%v)", tmp, err)
				continue
			}
		}
	}
	return
}
