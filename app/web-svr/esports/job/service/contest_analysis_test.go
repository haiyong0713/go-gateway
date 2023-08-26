package service

import (
	"context"
	"testing"
)

func TestContestAnalysisBiz(t *testing.T) {
	t.Run("test team analysis biz", teamAnalysisBiz)
}

func teamAnalysisBiz(t *testing.T) {
	svr := new(Service)
	svr.syncScoreAnalysis(context.Background(), 172)
}
