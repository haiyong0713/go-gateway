package space

type ContractResource struct {
	FollowShowType       int32                 `json:"follow_show_type"`
	ContractCard         *ContractCard         `json:"contract_card,omitempty"`
	FollowButtonDecorate *FollowButtonDecorate `json:"follow_button_decorate,omitempty"`
}

type ContractCard struct {
	Title    string `json:"title,omitempty"`
	SubTitle string `json:"subtitle,omitempty"`
	Icon     string `json:"icon,omitempty"`
}

type FollowButtonDecorate struct {
	WingLeft  string `json:"wing_left,omitempty"`
	WingRight string `json:"wing_right,omitempty"`
}
