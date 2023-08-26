package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/admin/model/system"
	"strings"
)

func (s *Service) ImportSignVipList(ctx context.Context, aid int64, uids []string, seats []*model.UIDSeat) (err error) {
	// 处理vip名单
	if len(uids) == 0 {
		return
	}
	if err = s.dao.ClearVipList(ctx, aid); err != nil {
		return
	}
	if err = s.dao.InsertSignVipList(ctx, aid, uids); err != nil {
		return
	}

	// 处理桌号名单
	if len(seats) == 0 {
		if err = s.dao.ClearVipSeatList(ctx, aid); err != nil {
			return
		}
	} else {
		if err = s.dao.ClearVipSeatList(ctx, aid); err != nil {
			return
		}
		if err = s.dao.InsertSignVipSeatList(ctx, aid, seats); err != nil {
			return
		}
	}

	return
}

func (s *Service) GetSignList(ctx context.Context, aid int64, page int64, size int64) (res *model.GetSignListRes, err error) {
	res = new(model.GetSignListRes)
	signList, count, err := s.dao.GetSignUserList(ctx, aid, page, size)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetSignUserList Err aid(%v) page(%v) size(%v) err(%v)", aid, page, size, err)
		return
	}
	// 循环uid
	var uids []string
	for _, v := range signList {
		uids = append(uids, v.UID)
	}
	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetUsersInfo Err uids(%v) err(%v)", uids, err)
		return
	}
	for _, v := range signList {
		userInfo, ok := usersInfo[v.UID]
		if !ok {
			log.Errorc(ctx, "GetSignList Can`t Get UserInfo uid(%v) list(%+v)", v.UID, usersInfo)
			return
		}
		item := new(model.GetSignList)

		item.UID = v.UID
		item.NickName = userInfo.NickName
		item.LastName = userInfo.LastName
		item.Location = v.Location
		item.Time = v.Ctime

		res.List = append(res.List, item)
	}
	res.Page = model.ManagerPage{
		Num:   page,
		Size:  size,
		Total: count,
	}
	return
}

func (s *Service) GetSignVipList(ctx context.Context, aid int64, page int64, size int64) (res *model.GetSignVipListDetailRes, err error) {
	res = new(model.GetSignVipListDetailRes)
	// 分页获取白名单列表
	signList, count, err := s.dao.GetSignVipUserList(ctx, aid, page, size)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetSignVipUserList Err aid(%v) page(%v) size(%v) err(%v)", aid, page, size, err)
		return
	}
	if len(signList) == 0 {
		return
	}

	// 循环uid
	var uids []string
	for _, v := range signList {
		uids = append(uids, v.UID)
	}
	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetUsersInfo Err uids(%v) err(%v)", uids, err)
		return
	}
	// 获取VIP用户签到信息
	vipListSignInfo, err := s.dao.GetSignVipUserStateList(ctx, aid, uids)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetSignVipUserStateList Err aid(%v) uids(%v) err(%v)", aid, uids, err)
		return
	}
	// 标记vip用户是否签到过
	isVipListSigned := make(map[string]*model.SystemSignUser)
	for _, v := range vipListSignInfo {
		isVipListSigned[v.UID] = v
	}

	for _, v := range signList {
		userInfo, ok := usersInfo[v.UID]
		if !ok {
			log.Errorc(ctx, "GetSignList Can`t Get UserInfo uid(%v) list(%+v)", v.UID, usersInfo)
			return
		}
		item := new(model.GetSignListDetail)

		item.UID = v.UID
		item.NickName = userInfo.NickName
		item.LastName = userInfo.LastName

		// 如果签到过
		if val, ok := isVipListSigned[v.UID]; ok {
			item.Location = val.Location
			item.Time = val.Ctime
			item.IsSign = 1
		}

		res.List = append(res.List, item)
	}
	res.Page = model.ManagerPage{
		Num:   page,
		Size:  size,
		Total: count,
	}

	return
}

func (s *Service) SignUser(ctx context.Context, aid int64, uid string) (err error) {
	res, err := s.dao.GetUsersInfo(ctx, []string{uid})
	if _, ok := res[uid]; !ok {
		log.Errorc(ctx, "SignUser s.dao.GetUsersInfo No User aid(%v) uid(%v)", aid, uid)
		return errors.New("no exist uid")
	}
	return s.dao.SignUser(ctx, aid, uid)
}

func (s *Service) ExportSignList(ctx context.Context, data *model.GetSignListRes) (res [][]string) {
	res = make([][]string, 0)
	items := data.List

	for _, v := range items {
		t := ""
		if v.Time.Time().Unix() > 0 {
			t = v.Time.Time().Format("2006-01-02 15:04:05")
		}
		convItem := []string{v.UID, v.NickName, v.LastName, t, v.Location}
		res = append(res, convItem)
	}
	return
}

func (s *Service) ExportSignVipList(ctx context.Context, data *model.GetSignVipListDetailRes) (res [][]string) {
	res = make([][]string, 0)
	items := data.List

	for _, v := range items {
		sign := "未签到"
		if v.IsSign == 1 {
			sign = "已签到"
		}
		t := ""
		if v.Time.Time().Unix() > 0 {
			t = v.Time.Time().Format("2006-01-02 15:04:05")
		}
		convItem := []string{v.UID, v.NickName, v.LastName, sign, t, v.Location}
		res = append(res, convItem)
	}
	return
}

func (s *Service) SystemActAdd(ctx context.Context, req *model.SystemActAddArgs) (lastID int64, err error) {
	create := new(model.SystemActAddArgs)
	create.Name = strings.TrimSpace(req.Name)
	create.Stime = req.Stime
	create.Etime = req.Etime
	create.Type = req.Type
	create.Create = req.Create
	create.Update = req.Update
	// 参数检测
	if req.Type == model.SystemActTypeSign {
		// 签到活动
		config := new(model.SystemActSignConfig)
		if err = json.Unmarshal([]byte(req.Config), config); err != nil {
			err = fmt.Errorf("SystemActAdd Err req(%+v) err(%v)", req, err)
			log.Errorc(ctx, err.Error())
			return
		}
		create.Config = req.Config
	}
	if req.Type == model.SystemActivityTypeVote || req.Type == model.SystemActivityTypeQuestion {
		create.Config = req.Config
	}

	return s.dao.SystemActAdd(ctx, create)
}

func (s *Service) SystemActEdit(ctx context.Context, req *model.SystemActEditArgs) (err error) {
	update := new(model.SystemActEditArgs)
	update.ID = req.ID
	update.Name = strings.TrimSpace(req.Name)
	update.Stime = req.Stime
	update.Etime = req.Etime
	update.Type = req.Type
	update.Create = req.Create
	update.Update = req.Update

	update.Config = ""
	// 参数检测
	if req.Type == model.SystemActTypeSign {
		// 签到活动
		config := new(model.SystemActSignConfig)
		if err = json.Unmarshal([]byte(req.Config), config); err != nil {
			err = fmt.Errorf("SystemActAdd Err req(%+v) err(%v)", req, err)
			log.Errorc(ctx, err.Error())
			return
		}
		update.Config = req.Config
	}
	if req.Type == model.SystemActivityTypeVote || req.Type == model.SystemActivityTypeQuestion {
		update.Config = req.Config
	}

	return s.dao.SystemActEdit(ctx, update)
}

func (s *Service) SystemActState(ctx context.Context, id int64, state int64) (err error) {
	if state != model.SystemActStateOffline && state != model.SystemActStateDelete {
		err = fmt.Errorf("SystemActState illegal id (%v) state (%v)", id, state)
		log.Errorc(ctx, err.Error())
		return
	}

	return s.dao.SystemActState(ctx, id, state)
}

func (s *Service) SystemActInfo(ctx context.Context, id int64) (res *model.SystemActInfo, err error) {
	return s.dao.SystemActInfo(ctx, id)
}

func (s *Service) SystemActList(ctx context.Context, query string, page int64, size int64) (res *model.SystemActInfoList, err error) {
	data, count, err := s.dao.SystemActList(ctx, query, page, size)
	if err != nil {
		return
	}
	res = &model.SystemActInfoList{
		Page: model.ManagerPage{
			Num:   page,
			Size:  size,
			Total: count,
		},
		List: data,
	}
	return
}

func (s *Service) SystemSeatList(ctx context.Context, aid int64) (res []*model.GetSeatList, err error) {
	res = make([]*model.GetSeatList, 0)
	// 获取活动座位列表
	seatList, err := s.dao.GetSeatUserList(ctx, aid)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetSeatUserList Err aid(%v) err(%v)", aid, err)
		return
	}
	if len(seatList) == 0 {
		return
	}
	// 循环uid
	var uids []string
	for _, v := range seatList {
		uids = append(uids, v.UID)
	}
	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetUsersInfo Err uids(%v) err(%v)", uids, err)
		return
	}

	for _, v := range seatList {
		userInfo, ok := usersInfo[v.UID]
		if !ok {
			log.Errorc(ctx, "GetSignList Can`t Get UserInfo uid(%v) list(%+v)", v.UID, usersInfo)
			return
		}
		item := &model.GetSeatList{}

		item.UID = v.UID
		item.NickName = userInfo.NickName
		item.LastName = userInfo.LastName
		item.Content = v.Content

		res = append(res, item)
	}

	return
}
