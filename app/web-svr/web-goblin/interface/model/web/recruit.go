package web

// Params .
type Params struct {
	Route string `form:"route" validate:"required"`
	Mode  string `form:"mode" validate:"required"`
	JobID string `form:"job_id"`
}

// NewArchive new rank archive struct
type NewArchive struct {
	Aid   int64 `json:"aid"`
	Score int   `json:"score"`
}
