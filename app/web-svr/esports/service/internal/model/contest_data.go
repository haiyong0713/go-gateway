package model

type ContestDataModel struct {
	ID int64 `json:"id"`
	// 赛程id
	Cid int64 `json:"cid"`
	// 每BO局的url
	Url string `json:"url"`
	// 每BO局对应的三方id
	PointData int64 `json:"point_data"`
	// 每BO局对应的Av号
	AvCid int64 `json:"av_cid"`

	IsDeleted int64 `json:"is_deleted"`
}
