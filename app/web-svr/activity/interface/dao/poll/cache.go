package poll

import (
	"context"
	"fmt"
	"time"

	model "go-gateway/app/web-svr/activity/interface/model/poll"
)

func keyLastPollVoteUserStat(mid int64, pollID int64) string {
	return fmt.Sprintf("lpvus_%d_%d", mid, pollID)
}

func keyPollVoteUserStatByDate(mid int64, pollID int64, date time.Time) string {
	year, month, day := date.Date()
	return fmt.Sprintf("pvusd_%d_%d_%d%d%d", mid, pollID, year, month, day)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -nullcache=&model.PollVoteUserStat{Id:-1} -check_null_code=$!=nil&&$.Id==-1 -struct_name=Dao
	LastPollVoteUserStat(ctx context.Context, mid int64, pollID int64) (*model.PollVoteUserStat, error)

	// bts: -nullcache=&model.PollVoteUserStat{Id:-1} -check_null_code=$!=nil&&$.Id==-1 -struct_name=Dao
	PollVoteUserStatByDate(ctx context.Context, mid int64, pollID int64, date time.Time) (*model.PollVoteUserStat, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=keyLastPollVoteUserStat -struct_name=Dao
	CacheLastPollVoteUserStat(ctx context.Context, mid int64, pollID int64) (*model.PollVoteUserStat, error)

	// mc: -key=keyLastPollVoteUserStat -expire=172800 -encode=pb -struct_name=Dao
	AddCacheLastPollVoteUserStat(c context.Context, mid int64, value *model.PollVoteUserStat, pollID int64) error

	// mc: -key=keyLastPollVoteUserStat -struct_name=Dao
	DelCacheLastPollVoteUserStat(ctx context.Context, mid int64, pollID int64) error

	// mc: -key=keyPollVoteUserStatByDate -struct_name=Dao
	CachePollVoteUserStatByDate(ctx context.Context, mid int64, pollID int64, date time.Time) (*model.PollVoteUserStat, error)

	// mc: -key=keyPollVoteUserStatByDate -expire=172800 -encode=pb -struct_name=Dao
	AddCachePollVoteUserStatByDate(c context.Context, mid int64, value *model.PollVoteUserStat, pollID int64, date time.Time) error

	// mc: -key=keyPollVoteUserStatByDate -struct_name=Dao
	DelCachePollVoteUserStatByDate(ctx context.Context, mid int64, pollID int64, date time.Time) error
}
