package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/admin/model/system"
)

// 获取活动投票整体情况
func (s *Service) VoteSumList(ctx context.Context, aid int64) (res []*model.ActivitySystemVote, err error) {
	res = make([]*model.ActivitySystemVote, 0)
	// 获取投票所有数据
	res, err = s.dao.GetVoteDataList(ctx, aid)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetVoteDataList Err aid(%v) err(%v)", aid, err)
		return
	}
	return
}

func (s *Service) VoteOptionDetail(ctx context.Context, aid, itemID, optionID int64) (res []*model.VoteDetail, err error) {
	voteInfo, err := s.dao.GetVoteOptionDetail(ctx, aid, itemID, optionID)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetVoteDetail Err aid(%v) err(%v)", aid, err)
		return
	}
	if len(voteInfo) == 0 {
		return
	}
	uids := make([]string, 0)
	for _, v := range voteInfo {
		uids = append(uids, v.UID)
	}
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		log.Errorc(ctx, "VoteDetailList s.dao.GetUsersInfo Err uids(%v) err(%v)", uids, err)
		return
	}
	res = make([]*model.VoteDetail, 0)
	for _, userInfo := range usersInfo {
		res = append(res, &model.VoteDetail{
			UID:      userInfo.UID,
			NickName: userInfo.NickName,
			LastName: userInfo.LastName,
		})
	}
	return
}

func (s *Service) ExportVoteDetail(ctx context.Context, aid int64) (excelData [][]string, header []string, err error) {
	voteInfo, err := s.dao.GetVoteDetailList(ctx, aid)
	if err != nil {
		return
	}
	subject, err := s.dao.SystemActInfo(ctx, aid)
	if err != nil {
		return
	}
	if subject.Type != model.SystemActivityTypeVote {
		log.Errorc(ctx, "error subject type")
		err = errors.New("error subject type")
		return
	}
	config := new(model.VoteConfig)
	if err = json.Unmarshal([]byte(subject.Config), config); err != nil {
		log.Errorc(ctx, "ExportVoteDetail json.Unmarshal([]byte(%+v), %+v) Err err(%v)", subject.Config, config, err)
		return
	}
	// 拼接问卷数据
	itemDetail := make(map[int64]string)
	optionDetail := make(map[int64]map[int64]string)
	for itemIndex, item := range config.Items {
		itemDetail[int64(itemIndex)] = item.Title
		optionDetail[int64(itemIndex)] = make(map[int64]string)
		for k, v := range item.Options.Name {
			optionDetail[int64(itemIndex)][int64(k)] = v.Desc
		}
	}
	// db聚拢数据
	// 先按照uid维度
	kneadData := make(map[string]map[int64]string, 0)
	for _, row := range voteInfo {
		if _, ok := kneadData[row.UID]; !ok {
			kneadData[row.UID] = make(map[int64]string)
		}
		if _, ok := kneadData[row.UID][row.ItemID]; !ok {
			kneadData[row.UID][row.ItemID] = ""
		}
		if kneadData[row.UID][row.ItemID] != "" {
			kneadData[row.UID][row.ItemID] = kneadData[row.UID][row.ItemID] + "," + optionDetail[row.ItemID][row.OptionID]
		} else {
			kneadData[row.UID][row.ItemID] = optionDetail[row.ItemID][row.OptionID]
		}
	}
	// 提取uids
	uids := make([]string, 0)
	for uid := range kneadData {
		uids = append(uids, uid)
	}

	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		log.Errorc(ctx, "s.dao.GetUsersInfo Err uids(%+v) err(%v)", uids, err)
		return
	}

	excelData = make([][]string, 0)
	for uid, itemData := range kneadData {
		excelRow := make([]string, 0)
		excelRow = append(excelRow, uid)
		excelRow = append(excelRow, usersInfo[uid].LastName)
		excelRow = append(excelRow, usersInfo[uid].NickName)

		for _, options := range itemData {
			excelRow = append(excelRow, options)
		}

		excelData = append(excelData, excelRow)
	}

	header = []string{"uid", "姓名", "昵称"}
	itemLength := len(itemDetail)
	for i := 0; i < itemLength; i++ {
		header = append(header, itemDetail[int64(i)])
	}

	return
}

// 提问详情
func (s *Service) QuestionList(ctx context.Context, aid int64) (res map[int64][]*model.ActivitySystemQuestionList, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, err.Error())
		}
	}()

	res = make(map[int64][]*model.ActivitySystemQuestionList, 0)
	// 获取问题
	var data []*model.ActivitySystemQuestion
	if data, err = s.dao.GetQuestionData(ctx, aid); err != nil {
		err = errors.Wrap(err, "s.dao.GetQuestionData err")
		return
	}

	uids := make([]string, 0)
	for _, v := range data {
		uids = append(uids, v.UID)
	}

	// 获取用户信息
	usersInfo, err := s.dao.GetUsersInfo(ctx, uids)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetUsersInfo err")
		return
	}

	for _, v := range data {
		if _, ok := usersInfo[v.UID]; !ok {
			continue
		}
		userInfo := usersInfo[v.UID]

		item := &model.ActivitySystemQuestionList{
			*v,
			userInfo.NickName,
			userInfo.LastName,
		}

		if _, ok := res[v.QID]; ok {
			res[v.QID] = append(res[v.QID], item)
		} else {
			res[v.QID] = []*model.ActivitySystemQuestionList{item}
		}
	}
	return
}

// 提问详情
func (s *Service) QuestionState(ctx context.Context, id int64) (res int, err error) {
	var item *model.ActivitySystemQuestion
	item, err = s.dao.GetQuestionItem(ctx, id)
	if err != nil {
		err = errors.Wrap(err, "s.dao.GetQuestionItem err")
		return
	}
	if item.ID <= 0 {
		err = fmt.Errorf("数据不存在")
		return
	}
	if item.State == -1 {
		return
	}
	if err = s.dao.DeleteQuestionItem(ctx, id); err != nil {
		err = errors.Wrap(err, "s.dao.DeleteQuestionItem err")
		return
	}

	return
}
