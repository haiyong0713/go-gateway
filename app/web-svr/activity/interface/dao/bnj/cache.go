package bnj

import "context"

func timeFinishKey() string {
	return "time_finish"
}

func timeLessKey() string {
	return "time_less"
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=timeFinishKey
	CacheTimeFinish(c context.Context) (int64, error)
	// mc: -key=timeFinishKey
	DelCacheTimeFinish(c context.Context) (int64, error)
	// mc: -key=timeLessKey
	DelCacheTimeLess(c context.Context) (int64, error)
}
