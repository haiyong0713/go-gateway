package system

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/admin/model/system"
	"strings"
)

const (
	_systemActStateSQL = "update system_activity set state = ? where id = ?"
)

func (d *Dao) ClearVipList(ctx context.Context, aid int64) (err error) {
	if err = d.DB.Where("aid = ?", aid).Delete(&model.SystemSignStatistics{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "ClearVipList Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

func (d *Dao) InsertSignVipList(ctx context.Context, aid int64, uids []string) (err error) {
	insert := make([]model.SystemSignStatistics, 0)
	for _, uid := range uids {
		insert = append(insert, model.SystemSignStatistics{AID: aid, UID: uid})
	}
	sql := "insert into system_activity_sign_statistics (aid,uid) values "
	length := len(insert)
	values := ""
	var args []interface{}
	for k, v := range insert {
		if k == length-1 {
			values += "(?, ?)"
		} else {
			values += "(?, ?),"
		}
		args = append(args, v.AID, v.UID)
	}
	sql += values
	if err = d.DB.Exec(sql, args...).Error; err != nil {
		log.Errorc(ctx, "InsertSignVipList Error aid(%v) uids(%v) err(%v)", aid, uids, err)
		return
	}
	return
}

// 获取所有签到人员列表
func (d *Dao) GetSignUserList(ctx context.Context, aid int64, page int64, size int64) (res []*model.SystemSignUser, count int64, err error) {
	SQL := d.DB.Where("aid = ?", aid).Order("ctime desc").Offset((page - 1) * size).Limit(size).Find(&res)
	if size == 0 {
		SQL = d.DB.Where("aid = ?", aid).Find(&res)
	}
	if err = SQL.Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetSignUserList Find Error aid(%v) err(%v)", aid, err)
		return
	}
	var list []*model.SystemSignUser
	if err = d.DB.Where("aid = ?", aid).Find(&list).Count(&count).Error; err != nil {
		log.Errorc(ctx, "GetSignUserList Count Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 获取白名单所有签到人员列表
func (d *Dao) GetSignVipUserList(ctx context.Context, aid int64, page int64, size int64) (res []*model.SystemSignStatisticsList, count int64, err error) {
	SQL := d.DB.Where("aid = ?", aid).Order("id asc").Offset((page - 1) * size).Limit(size).Find(&res)
	if size == 0 {
		SQL = d.DB.Where("aid = ?", aid).Find(&res)
	}
	if err = SQL.Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetSignVipUserList Find Error aid(%v) err(%v)", aid, err)
		return
	}
	var list []*model.SystemSignStatisticsList
	if err = d.DB.Where("aid = ?", aid).Find(&list).Count(&count).Error; err != nil {
		log.Errorc(ctx, "GetSignVipUserList Count Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 获取指定用户签到状态
func (d *Dao) GetSignVipUserStateList(ctx context.Context, aid int64, uids []string) (res []*model.SystemSignUser, err error) {
	res = make([]*model.SystemSignUser, 0)
	if len(uids) == 0 {
		return
	}
	where := fmt.Sprintf("uid IN (%s)", strings.Join(uids, ","))
	if err = d.DB.Where("aid = ?", aid).Where(where).Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetSignVipUserStateList Find Error aid(%v) uids(%v) err(%v)", aid, uids, err)
		return
	}
	return
}

// 手动补签
func (d *Dao) SignUser(ctx context.Context, aid int64, uid string) (err error) {
	res := new(model.SystemSignUser)
	if err = d.DB.Where("aid = ?", aid).Where("uid = ?", uid).Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "SignUser Find Error aid(%v) uid(%v) err(%v)", aid, uid, err)
		return
	}
	// 已经签到过 返回
	if res.ID > 0 {
		return
	}
	// 没有签到过 创建签到记录
	create := &model.SystemSignUser{UID: uid, AID: aid}
	if err = d.DB.Create(create).Error; err != nil {
		log.Errorc(ctx, "SignUser Create Error aid(%v) uid(%v) err(%v)", aid, uid, err)
		return
	}
	return
}

func (d *Dao) SystemActAdd(ctx context.Context, args *model.SystemActAddArgs) (lastID int64, err error) {
	SQL := d.DB.Model(&model.SystemAct{}).Create(args)
	if err = SQL.Error; err != nil {
		log.Errorc(ctx, "SystemActAdd Error args(%v) err(%v)", args, err)
		return
	}
	lastID = args.ID
	return
}

func (d *Dao) SystemActEdit(ctx context.Context, args *model.SystemActEditArgs) (err error) {
	SQL := d.DB.Model(&model.SystemAct{}).Where("id = ?", args.ID).Update(args)
	if err = SQL.Error; err != nil {
		log.Errorc(ctx, "SystemActEdit Error args(%v) err(%v)", args, err)
		return
	}
	return
}

func (d *Dao) SystemActState(ctx context.Context, id int64, state int64) (err error) {
	if err = d.DB.Exec(_systemActStateSQL, state, id).Error; err != nil {
		log.Errorc(ctx, "SystemActState Error id(%v) state(%v) err(%v)", id, state, err)
		return
	}
	return
}

func (d *Dao) SystemActInfo(ctx context.Context, id int64) (res *model.SystemActInfo, err error) {
	where := fmt.Sprintf("state IN (%d,%d)", model.SystemActStateNormal, model.SystemActStateOffline)
	res = new(model.SystemActInfo)
	if err = d.DB.Model(&model.SystemAct{}).Where("id = ?", id).Where(where).Find(&res).Error; err != nil {
		log.Errorc(ctx, "SystemActInfo Error id(%v) err(%v)", id, err)
		res = nil
		return
	}
	return
}

func (d *Dao) SystemActList(ctx context.Context, query string, page int64, size int64) (res []*model.SystemActInfo, count int64, err error) {
	where := fmt.Sprintf("state IN (%d,%d)", model.SystemActStateNormal, model.SystemActStateOffline)
	res = make([]*model.SystemActInfo, 0)

	base := d.DB.Model(&model.SystemAct{})
	if query != "" {
		base = base.Where("name like ?", "%"+query+"%")
	}

	if err = base.Count(&count).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "SystemActList Count Error query(%v) page(%v) size(%v) err(%v)", query, page, size, err)
		return
	}

	sql := base.Where(where).Order("id desc").Offset((page - 1) * size).Limit(size)
	if size == 0 {
		sql = base.Where(where).Order("id desc")
	}

	if err = sql.Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "SystemActList List Error query(%v) page(%v) size(%v) err(%v)", query, page, size, err)
		return
	}
	return
}

func (d *Dao) ClearVipSeatList(ctx context.Context, aid int64) (err error) {
	if err = d.DB.Where("aid = ?", aid).Delete(&model.SystemActSeat{}).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "ClearVipList Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

func (d *Dao) InsertSignVipSeatList(ctx context.Context, aid int64, seats []*model.UIDSeat) (err error) {
	insert := make([]*model.UIDSeat, 0)
	for _, seat := range seats {
		insert = append(insert, &model.UIDSeat{AID: seat.AID, UID: seat.UID, Content: seat.Content})
	}
	sql := "insert into system_activity_seat (aid,uid,content) values "
	length := len(insert)
	values := ""
	var args []interface{}
	for k, v := range insert {
		if k == length-1 {
			values += "(?, ?, ?)"
		} else {
			values += "(?, ?, ?),"
		}
		args = append(args, v.AID, v.UID, v.Content)
	}
	sql += values
	if err = d.DB.Exec(sql, args...).Error; err != nil {
		log.Errorc(ctx, "InsertSignVipSeatList Error aid(%v) seats(%+v) err(%v)", aid, seats, err)
		return
	}
	return
}

// 获取座位人员列表
func (d *Dao) GetSeatUserList(ctx context.Context, aid int64) (res []*model.SystemActSeat, err error) {
	if err = d.DB.Where("aid = ?", aid).Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetSeatUserList List Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 分组获取投票所有数据
func (d *Dao) GetVoteDataList(ctx context.Context, aid int64) (res []*model.ActivitySystemVote, err error) {
	if err = d.DB.Select("sum(`score`) as score, `item_id`, `option_id`").Where("aid = ?", aid).Group("item_id, option_id").Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetVoteDataList Data Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 投票具体选项信息
func (d *Dao) GetVoteOptionDetail(ctx context.Context, aid, itemID, optionID int64) (res []*model.DetailUIDs, err error) {
	res = make([]*model.DetailUIDs, 0)
	if err = d.DB.Model(&model.DetailUIDs{}).Select("uid").Where("aid = ?", aid).Where("item_id = ?", itemID).Where("option_id = ?", optionID).Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetVoteDataList Data Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 获取投票所有数据
func (d *Dao) GetVoteDetailList(ctx context.Context, aid int64) (res []*model.ActivitySystemVote, err error) {
	if err = d.DB.Select("`uid`, `item_id`, `option_id`").Where("aid = ?", aid).Group("uid, item_id, option_id").Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(ctx, "GetVoteDetailList Data Error aid(%v) err(%v)", aid, err)
		return
	}
	return
}

// 按照题目获取提问数据
func (d *Dao) GetQuestionData(ctx context.Context, aid int64) (res []*model.ActivitySystemQuestion, err error) {
	if err = d.DB.Model(&model.ActivitySystemQuestion{}).Select("*").Where("aid = ?", aid).Order("id desc").Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		err = errors.Wrap(err, "d.DB.Select err")
		return
	}
	return
}

// 按照题目获取提问数据
func (d *Dao) GetQuestionItem(ctx context.Context, id int64) (res *model.ActivitySystemQuestion, err error) {
	res = new(model.ActivitySystemQuestion)
	if err = d.DB.Model(&model.ActivitySystemQuestion{}).Where("id = ?", id).First(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		err = errors.Wrap(err, "d.DB.Select err")
		return
	}
	if err == gorm.ErrRecordNotFound {
		err = nil
	}

	return
}

func (d *Dao) DeleteQuestionItem(ctx context.Context, id int64) (err error) {
	if err = d.DB.Model(&model.ActivitySystemQuestion{}).Where("id = ?", id).Update("state", -1).Error; err != nil {
		err = errors.Wrap(err, "DeleteQuestionItem err")
		return
	}
	return
}
