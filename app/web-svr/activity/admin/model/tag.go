package model

const (
	// TagIsActivity 活动tag
	TagIsActivity = 1
	// TagIsNotActivity 非活动tag
	TagIsNotActivity = 2
)

// TagIsActivityRes .
type TagIsActivityRes struct {
	Status int `json:"status"`
}

// TagToActivityRes .
type TagToActivityRes struct {
}
