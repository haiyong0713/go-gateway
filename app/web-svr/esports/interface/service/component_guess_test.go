package service

import "testing"

// go test -v -count=1 component_guess_test.go
func TestComponentBiz(t *testing.T) {
	matchIDList := make([]int64, 0)
	for i := int64(0); i <= 666; i++ {
		matchIDList = append(matchIDList, i)
	}

	limit := 100
	listLen := len(matchIDList)
	lenAfterSplit := listLen / limit
	if d := listLen % limit; d > 0 {
		lenAfterSplit++
	}
	t.Log(lenAfterSplit, listLen%limit)
	for i := 0; i < lenAfterSplit; i++ {
		startIndex := limit * i
		endIndex := startIndex + 100
		if endIndex > listLen {
			endIndex = listLen
		}

		t.Log(startIndex, endIndex, matchIDList[startIndex:endIndex])
	}
}
