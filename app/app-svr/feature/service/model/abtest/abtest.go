package abtest

type ABTest struct {
	ID        int64  `json:"id"`
	TreeID    int64  `json:"tre_id"`
	KeyName   string `json:"key_name"`
	AbType    string `json:"ab_type"`
	Bucket    int32  `json:"bucket"`
	Salt      string `json:"salt"`
	Config    string `json:"config"`
	Relations string `json:"relations"`
}
