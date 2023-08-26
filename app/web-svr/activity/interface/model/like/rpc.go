package like

// ArgMatch arg match
type ArgMatch struct {
	Sid int64
}

// ArgSubjectUp .
type ArgSubjectUp struct {
	Sid int64
}

// ArgLikeUp .
type ArgLikeUp struct {
	Lid int64
}

// ArgLikeItem .
type ArgLikeItem struct {
	ID   int64
	Sid  int64
	Type int64
}

// ArgActSubject .
type ArgActSubject struct {
	Sid int64
}

// ArgActProtocol .
type ArgActProtocol struct {
	Sid int64
}

// ArgSetReload .
type ArgSetReload struct {
	Lid int64
}

// ArgActLikes .
type ArgActLikes struct {
	Sid      int64 `json:"sid" validate:"min=1"`
	Mid      int64 `json:"mid"`
	SortType int   `json:"sort_type" default:"1" validate:"min=1"`
	Ps       int   `json:"ps" default:"1" validate:"min=1,max=100"`
	Pn       int   `json:"pn" default:"1" validate:"min=0"`
	Offset   int64 `json:"offset" default:"-1" validate:"min=-1"`
	Zone     int64 `json:"zone" default:"0"`
}

// ActLikes .
type ActLikes struct {
	Sub     *SubjectItem
	List    []*ItemObj
	Total   int64 `json:"total"`
	HasMore int32 `json:"has_more"`
	Offset  int64 `json:"offset"`
}
