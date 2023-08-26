package model

import "go-gateway/app/app-svr/archive/service/api"

// ArcMsg reprensents the archive Notify-T message structure
type ArcMsg struct {
	Table string   `json:"table"`
	New   *Archive `json:"new"`
}

// ArchDatabus model ( we pick the fields that we need )
type Archive struct {
	Aid       int64 `json:"aid"`
	Attribute int32 `json:"attribute"`
	State     int32 `json:"state"`
}

// IsNormal check archive is normal
func (a *Archive) IsNormal() bool {
	return a.State >= api.StateOpen
}

// IsSteinsGate tells whether the archive is interactive
func (a *Archive) IsSteinsGate() bool {
	return a.AttrVal(api.AttrBitSteinsGate) == api.AttrYes
}

// AttrVal returns the attribute value
func (a *Archive) AttrVal(bit uint) int32 {
	return (a.Attribute >> bit) & int32(1)

}
