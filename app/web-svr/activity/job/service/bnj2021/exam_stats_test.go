package bnj2021

import "testing"

// go test -v --count=1 exam_stats_test.go exam_stats.go service.go
func TestExamStats(t *testing.T) {
	if err := UpdateExamStatRule(); err != nil {
		t.Error(err)

		return
	}

	if err := fetchAndUpdateExamStats(); err != nil {
		t.Error(err)
	}
}
