package search

// Upper is the struct of upper recommend result from search API
type Upper struct {
	Mid       int64  `json:"up_id"`
	RecReason string `json:"rec_reason"`
}
