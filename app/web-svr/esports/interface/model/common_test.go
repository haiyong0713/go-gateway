package model

import "testing"

// go test -v -count=1 auto_subscribe.go common.go common_test.go component.go guess.go live.go match.go params.go pointdata.go s10.go score_analysis.go search.go season.go
func TestCommonBiz(t *testing.T) {
	list := make([]*MatchGuess, 0)
	resp := new(UserSeasonGuessResp)
	{
		resp.Data = make([]*MatchGuess, 0)
	}

	resp.PageStructure = NewPageStructure(10, 1, int64(len(list)))

	t.Run("Test page biz with empty value", pageBizWithEmptyValue)
	t.Run("Test page biz with one value", pageBizWithOneValue)
	t.Run("Test page biz with empty and second page", pageBizWithEmptyValueAndSecondPage)
	t.Run("Test page biz with values and first page", pageBizWithValuesAndFirstPage)
	t.Run("Test page biz with values and second page", pageBizWithValuesAndSecondPage)
}

func pageBizWithValuesAndFirstPage(t *testing.T) {
	list := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	page := new(PageStructure)
	{
		page.PageSize = 10
		page.PageNum = 1
		page.Count = 13
	}
	startIndex, endIndex, ok := page.CalculateStartAndEndIndex()
	t.Log(startIndex, endIndex, ok)
	if !ok || startIndex != 0 || endIndex != 10 {
		t.Error()

		return
	}

	t.Log(list[startIndex:endIndex])
}

func pageBizWithValuesAndSecondPage(t *testing.T) {
	list := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	page := new(PageStructure)
	{
		page.PageSize = 10
		page.PageNum = 2
		page.Count = 13
	}
	startIndex, endIndex, ok := page.CalculateStartAndEndIndex()
	t.Log(startIndex, endIndex, ok)
	if !ok || startIndex != 10 || endIndex != 13 {
		t.Error()

		return
	}

	t.Log(list[startIndex:endIndex])
}

func pageBizWithEmptyValueAndSecondPage(t *testing.T) {
	//list := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	page := new(PageStructure)
	{
		page.PageSize = 10
		page.PageNum = 2
		page.Count = 3
	}
	startIndex, endIndex, ok := page.CalculateStartAndEndIndex()
	t.Log(startIndex, endIndex, ok)
	if ok {
		t.Error()

		return
	}
}

func pageBizWithEmptyValue(t *testing.T) {
	page := new(PageStructure)
	startIndex, endIndex, ok := page.CalculateStartAndEndIndex()
	if ok {
		t.Error()

		return
	}

	t.Log(startIndex, endIndex, ok)
}

func pageBizWithOneValue(t *testing.T) {
	list := []int64{1}
	page := new(PageStructure)
	{
		page.PageSize = 10
		page.PageNum = 1
		page.Count = 1
	}
	startIndex, endIndex, ok := page.CalculateStartAndEndIndex()
	t.Log(startIndex, endIndex, ok)
	if !ok || startIndex != 0 || endIndex != 1 {
		t.Error()

		return
	}

	t.Log(list[startIndex:endIndex])
}
