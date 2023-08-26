package system

import (
	"context"
	model "go-gateway/app/web-svr/activity/interface/model/system"
)

func (s *Service) Party2021(ctx context.Context, sessionToken string) (res *model.Party2021Res, err error) {
	// 用户基本信息
	//var user *model.SystemUser
	//if user, err = s.dao.GetUserInfoWithCookie(ctx, sessionToken); err != nil {
	//	return
	//}
	//
	//// todo 年会拦截逻辑
	//if false {
	//	err = ecode.SystemNotIn2021PartyMembersErr
	//	return
	//}
	//
	//res = new(model.Party2021Res)
	//res.User.Avatar = user.Avatar
	//res.User.NickName = user.NickName
	//
	//// 座位表默认给部门名称
	//res.User.SeatContent = ""
	//department := strings.Split(user.DepartmentName, "-")
	//if len(department) > 0 {
	//	res.User.SeatContent = department[0]
	//}
	//
	//// 特殊座位表 配置文件获取
	//if v, ok := s.c.Party2021.SeatExtra[user.UID]; ok {
	//	res.User.SeatContent = v
	//}
	//
	//// 是否签到过
	//var signInfo *model.ActivitySign
	//if signInfo, err = s.dao.ActivitySigned(ctx, s.c.Party2021.AID, user.UID); err != nil {
	//	err = ecode.SystemNetWorkBuzyErr
	//	return
	//}
	//
	//res.Sign = 0
	//if signInfo != nil && signInfo.ID > 0 {
	//	res.Sign = 1
	//}
	//
	//// 签到二维码加密
	//var QRCode string
	//QRCode = tool.CFBEncrypt(user.UID, s.c.Party2021.AESKEY)
	//res.QRCode = QRCode

	return
}
