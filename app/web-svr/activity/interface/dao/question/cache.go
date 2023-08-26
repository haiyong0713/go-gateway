package question

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/question"
)

func detailKey(id int64) string {
	return fmt.Sprintf("ques_%d", id)
}

func lastLogKey(mid, baseID int64) string {
	return fmt.Sprintf("ques_log_%d_%d", mid, baseID)
}

//go:generate kratos tool btsgen
type _bts interface {
	// get question detail info
	// bts: -struct_name=Dao
	Detail(c context.Context, id int64) (*question.Detail, error)
	// get question details info
	// bts: -struct_name=Dao
	Details(c context.Context, ids []int64) (map[int64]*question.Detail, error)
	// get user last answer log
	// bts: -struct_name=Dao
	LastQuesLog(c context.Context, mid int64, baseID int64) (*question.UserAnswerLog, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=detailKey -struct_name=Dao
	CacheDetail(c context.Context, id int64) (*question.Detail, error)
	// mc: -key=detailKey -expire=d.questionExpire -encode=pb -struct_name=Dao
	AddCacheDetail(c context.Context, id int64, data *question.Detail) error
	// mc: -key=detailKey -struct_name=Dao
	DelCacheDetail(c context.Context, id int64) error
	// mc: -key=detailKey -struct_name=Dao
	CacheDetails(c context.Context, ids []int64) (map[int64]*question.Detail, error)
	// mc: -key=detailKey -expire=d.questionExpire -encode=pb -struct_name=Dao
	AddCacheDetails(c context.Context, data map[int64]*question.Detail) error
	// mc: -key=lastLogKey -struct_name=Dao
	CacheLastQuesLog(c context.Context, mid int64, baseID int64) (*question.UserAnswerLog, error)
	// mc: -key=lastLogKey -expire=d.lastLogExpire -encode=pb -struct_name=Dao
	AddCacheLastQuesLog(c context.Context, mid int64, data *question.UserAnswerLog, baseID int64) error
}
