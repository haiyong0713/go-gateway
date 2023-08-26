package model

import (
	"encoding/json"
	"math"
)

type KeyFrame struct {
	Aid       int64
	Cid       int64
	KeyFrames string
}

type Example struct {
	Ts     int64
	Action string
}

type Stat struct {
	X float64 `form:"x" json:"x"`
	Y float64 `form:"y" json:"y"`
	Z float64 `form:"z" json:"z"`
}

func (v *Stat) Euclidean() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v *Example) Euclidean() (float64, error) {
	s := new(Stat)
	err := json.Unmarshal([]byte(v.Action), &s)
	if err != nil {
		return 0.0, err
	}

	return s.Euclidean(), nil
}
