package model

type PageStructure struct {
	PageSize int64 `json:"ps" form:"ps" default:"10"`
	PageNum  int64 `json:"pn" form:"pn" default:"1"`
	Count    int64 `json:"count"`
}

type ScoreAnalysisRequest struct {
	AnalysisType int64 `form:"analysisType" validate:"min=1,max=3" default:"1"`
	SortType     int64 `form:"sortType" validate:"min=1,max=9" default:"1"`
	SortKey      int64 `form:"sortKey" validate:"min=1,max=2" default:"1"`
}

func NewPageStructure(size, num, count int64) (page PageStructure) {
	page = PageStructure{}
	{
		page.PageSize = size
		page.PageNum = num
		page.Count = count
	}

	return
}

func (page *PageStructure) CalculateStartAndEndIndex() (startIndex, endIndex int64, ok bool) {
	if page.Count == 0 {
		return
	}

	startIndex = (page.PageNum - 1) * page.PageSize
	endIndex = page.PageNum * page.PageSize
	if endIndex >= page.Count {
		endIndex = page.Count
	}

	if startIndex <= endIndex {
		ok = true
	}

	return
}
