package vote

import "go-gateway/app/web-svr/activity/interface/api"

const (

	//风控规则: 通用规则
	RiskControlRuleGeneric = "COMMON_RISK_CONTROL_RULE"
)

const (
	DSTypeMock      = "MOCK"
	DSTypeOperVideo = "OPER_VIDEO"
	DSTypeOperUp    = "OPER_UP"
	DSTypeUp        = "UP"
	DSTypeOperPic   = "OPER_PIC"
	DSTypeVideo     = "VIDEO"
)

type DataSourceItemVoteInfo struct {
	//投票的数据源下的稿件ID
	SourceItemId int64
	//最终得票数
	TotalVoteCount int64
}

type RankResultExternal struct {
	VoteRankVersion         int64       `json:"vote_rank_version"`
	VoteRankType            int64       `json:"vote_rank_type"`
	UserAvailVoteCount      int64       `json:"user_avail_vote_count"`
	UserExtraAvailVoteCount int64       `json:"user_extra_avail_vote_count"`
	DataSourceType          string      `json:"datasource_type"`
	DataSourceGroupId       int64       `json:"datasource_group_id"`
	List                    []*RankInfo `json:"list"`
	Page                    *Page       `json:"page"`
}

func (r *RankResultExternal) ToPB() (res *api.GetVoteActivityRankResp) {
	res = &api.GetVoteActivityRankResp{
		VoteRankVersion:       r.VoteRankVersion,
		VoteRankType:          r.VoteRankType,
		UserAvailVoteCount:    r.UserAvailVoteCount,
		UserAvailTmpVoteCount: r.UserExtraAvailVoteCount,
		DataSourceType:        r.DataSourceType,
		SourceGroupId:         r.DataSourceGroupId,
		Rank:                  make([]*api.ExternalRankInfo, 0, len(r.List)),
		Page:                  r.Page.ToPB(),
	}
	for _, t := range r.List {
		res.Rank = append(res.Rank, t.ToPB())
	}
	return
}

type RankSearchParams struct {
	DataSourceGroupId int64  `form:"datasource_group_id" validate:"min=1"`
	KeyWord           string `form:"keyword" validate:"required"`
	Limit             int    `form:"limit" validate:"min=1,max=50"`
}

type RankSearchResultExternal struct {
	VoteRankVersion         int64       `json:"vote_rank_version"`
	VoteRankType            int64       `json:"vote_rank_type"`
	UserAvailVoteCount      int64       `json:"user_avail_vote_count"`
	UserExtraAvailVoteCount int64       `json:"user_extra_avail_vote_count"`
	DataSourceType          string      `json:"datasource_type"`
	DataSourceGroupId       int64       `json:"datasource_group_id"`
	List                    []*RankInfo `json:"list"`
}

type Page struct {
	Pn    int64 `json:"num"`
	Ps    int64 `json:"size"`
	Total int64 `json:"total"`
}

func (p *Page) ToPB() (res *api.VotePage) {
	res = &api.VotePage{
		Num:   p.Pn,
		Ps:    p.Ps,
		Total: p.Total,
	}
	return
}

type RankInfo struct {
	Data               interface{} `json:"item"`
	DataSourceGroupId  int64       `json:"datasource_group_id"`
	DataSourceItemId   int64       `json:"datasource_item_id"`
	DataSourceItemName string      `json:"datasource_item_name"`
	Vote               int64       `json:"vote"`
	UserVoteCount      int64       `json:"user_vote_count"`
	UserVoteCountToday int64       `json:"user_vote_count_today"`
	UserCanVoteCount   int64       `json:"user_can_vote_count"`
}

func (i *RankInfo) ToPB() (res *api.ExternalRankInfo) {
	res = &api.ExternalRankInfo{
		SourceGroupId:      i.DataSourceGroupId,
		SourceItemId:       i.DataSourceItemId,
		SourceItemName:     i.DataSourceItemName,
		Vote:               i.Vote,
		UserVoteCount:      i.UserVoteCount,
		UserVoteCountToday: i.UserVoteCountToday,
		UserCanVoteCount:   i.UserCanVoteCount,
	}
	return
}

type DoVoteParams struct {
	ActivityId        int64 `json:"activity_id" form:"activity_id" validate:"min=1"`
	DataSourceGroupId int64 `json:"datasource_group_id" form:"datasource_group_id" validate:"min=1"`
	DataSourceItemId  int64 `json:"datasource_item_id" form:"datasource_item_id" validate:"min=1"`
	Vote              int64 `json:"vote" form:"vote" default:"1" validate:"min=1,max=10"`
}

type UndoVoteParams struct {
	ActivityId        int64 `json:"activity_id" form:"activity_id" validate:"min=1"`
	DataSourceGroupId int64 `json:"datasource_group_id" form:"datasource_group_id" validate:"min=1"`
	DataSourceItemId  int64 `json:"datasource_item_id" form:"datasource_item_id" validate:"min=1"`
}

type RankExternalParams struct {
	ActivityId        int64 `json:"activity_id" form:"activity_id" validate:"min=1"`
	DataSourceGroupId int64 `json:"datasource_group_id" form:"datasource_group_id" validate:"min=1"`
	Version           int64 `json:"version" form:"version"`
	Pn                int64 `json:"pn" form:"pn" default:"1" validate:"min=1"`
	Ps                int64 `json:"ps" form:"ps" default:"10" validate:"min=1,max=50"`
	Sort              int64 `json:"sort" form:"sort" default:"1" validate:"min=1,max=4"`
}

type InnerRankParams struct {
	Mid               int64
	ActivityId        int64
	DataSourceGroupId int64
	Version           int64
	Pn                int64
	Ps                int64
}

type UserVoteResponse struct {
	UserAvailVoteCount int64 `json:"user_avail_vote_count"`
}

type EsDataSourceItem struct {
	ID                string `json:"id"`
	DataSourceGroupId int64  `json:"datasource_group_id"`
	DataSourceItemId  int64  `json:"datasource_item_id"`
	SearchField1      string `json:"search_field_1"`
	SearchField2      string `json:"search_field_2"`
	SearchField3      string `json:"search_field_3"`
	DataVersion       int64  `json:"data_version"`
	WriteTime         string `json:"write_time"`
}

type EsDataSourceSearchResult struct {
	Result []*EsDataSourceItem `json:"result"`
	Page   *Page               `json:"page"`
}
