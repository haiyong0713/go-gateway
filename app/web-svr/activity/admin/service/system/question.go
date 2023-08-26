package system

import (
	"context"
	"github.com/pkg/errors"
	"go-gateway/app/web-svr/activity/interface/model/system"
)

func (s *Service) GetQuestionList(ctx context.Context, aid int64) (res []*system.ActivitySystemQuestionExport, err error) {
	res = make([]*system.ActivitySystemQuestionExport, 0)
	questionList, err := s.dao.GetQuestionData(ctx, aid)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetQuestionData err")
		return
	}
	if len(questionList) == 0 {
		return
	}

	// 循环uid
	var uids []string
	for _, v := range questionList {
		uids = append(uids, v.UID)
	}
	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUsersInfo err")
		return
	}

	for _, v := range questionList {
		if _, ok := usersInfo[v.UID]; !ok {
			continue
		}
		tmp := &system.ActivitySystemQuestionExport{
			Question:       v.Question,
			NickName:       usersInfo[v.UID].NickName,
			UserName:       usersInfo[v.UID].LastName,
			DepartmentName: usersInfo[v.UID].DepartmentName,
			Ctime:          v.Ctime.Time().Format("2006-01-02 15:04:05"),
		}
		tmp.State = "展现"
		if v.State == -1 {
			tmp.State = "隐藏"
		}
		res = append(res, tmp)
	}

	return
}

func (s *Service) ExportQuestionList(ctx context.Context, data []*system.ActivitySystemQuestionExport) (res [][]string) {
	res = make([][]string, 0)

	for _, v := range data {
		convItem := []string{
			v.Question,
			v.NickName,
			v.UserName,
			v.DepartmentName,
			v.State,
			v.Ctime,
		}
		res = append(res, convItem)
	}

	return
}
