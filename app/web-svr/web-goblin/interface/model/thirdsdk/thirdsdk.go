package thirdsdk

type Author struct {
	Mid          int64  `json:"mid"`
	Invited      bool   `json:"invited"`
	BindState    int    `json:"bind_state"`
	CheckState   int    `json:"check_state"`
	RefuseReason string `json:"refuse_reason"`
}
type MgrUserBind struct {
	Mid                 int64  `json:"mid"`
	VendorID            int64  `json:"vendorId"`
	AuthorizationStatus string `json:"authorizationStatus"`
	BindStatus          string `json:"bindStatus"`
	VerificationStatus  string `json:"verificationStatus"`
	Reason              string `json:"reason"`
}
