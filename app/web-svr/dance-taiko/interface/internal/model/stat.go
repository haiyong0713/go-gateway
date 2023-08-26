package model

import (
	"fmt"
	"math"
	"strconv"

	"go-gateway/app/web-svr/dance-taiko/interface/api"
)

type StatCore struct {
	X float64 `form:"x" json:"x"`
	Y float64 `form:"y" json:"y"`
	Z float64 `form:"z" json:"z"`
}

type Stat struct {
	StatCore
	TS int64 `form:"ts" json:"ts"` // 毫秒
}

type ExamplesMap struct {
	Data map[int64]float64
	Aid  int64
}

type PlayerHonor struct {
	Mid   int64
	Score int64
}

type PlayerRank struct {
	PlayerHonor
	Face string
	Name string
}

func Round(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

// TreatFloat 收敛5位小数
func (v *Stat) TreatFloat(stime int64) {
	v.X = Round(v.X, 5)
	v.Y = Round(v.Y, 5)
	v.Z = Round(v.Z, 5)
	v.TS = v.TS - stime // 实际相对时间减去比赛的开始时间，毫秒
}

// Euclidean 返回欧式距离，开根号
func (v *Stat) Euclidean() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v *Stat) GenerateAcc() *api.StatAcc {
	return &api.StatAcc{
		Ts:  v.TS,
		Acc: v.Euclidean(),
	}
}
