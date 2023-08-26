package model

const (
	_onlineRouter = "online_smooth_down"
	_effect       = 1
)

type TrafficControl struct {
	Router string       `json:"router"`
	Data   *ControlInfo `json:"data"`
}

type ControlInfo struct {
	Oid       int64      `json:"oid"`
	FlowState *FlowState `json:"flow_state"`
}

type FlowState struct {
	NoSearch         int64 `json:"nosearch"`
	NoRecommend      int64 `json:"norecommend"`
	PushBlog         int64 `json:"push_blog"`
	OnlineSmoothDown int64 `json:"online_smooth_down"`
}

func (tc *TrafficControl) LegalTCInfo() bool {
	if tc == nil || tc.Data == nil || tc.Data.FlowState == nil || tc.Data.Oid == 0 {
		return false
	}
	return tc.Router == _onlineRouter
}

func (tc *TrafficControl) ManualControl() bool {
	if !tc.LegalTCInfo() {
		return false
	}
	return tc.Data.FlowState.OnlineSmoothDown == _effect
}

func (tc *TrafficControl) AutoControl() bool {
	if !tc.LegalTCInfo() {
		return false
	}
	flowState := tc.Data.FlowState
	return flowState.NoRecommend == _effect || flowState.NoSearch == _effect || flowState.PushBlog == _effect
}

type SplitOnlineMsg struct {
	MobileApp string
	Cid       int64
}
