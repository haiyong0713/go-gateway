package prometheus

const (
	Merging = "merging"
	Merged  = "merged"
	Closed  = "closed"
)

type CIInWaiting struct {
	Count  float64 `json:"count"`
	AppKey string  `json:"app_key"`
}
