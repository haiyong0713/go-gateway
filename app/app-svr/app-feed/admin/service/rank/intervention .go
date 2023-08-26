package rank

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// 编辑稿件干预
func (s Service) RankArchiveEdit(c *bm.Context, req *rankModel.RankArchiveIntervention) (err error) {
	_, uname := util.UserInfo(c)

	// 检查位置是否重复
	if err = s.dao.RankArchiveInterventionConflictCheck(req); err != nil {
		return
	}

	// 添加、编辑干预
	if err = s.dao.RankArchiveEdit(req, uname); err != nil {
		log.Error("service.rank.RankArchiveEdit error(%v)", err)
		return
	}

	return

}

// 手动添加稿件
func (s Service) RankArchiveAdd(rankid int64, avid int64) (err error) {

	if err = s.dao.RankArchiveAdd(rankid, avid); err != nil {
		log.Error("service.rank.RankArchiveEdit error(%v)", err)
		return
	}

	return

}
