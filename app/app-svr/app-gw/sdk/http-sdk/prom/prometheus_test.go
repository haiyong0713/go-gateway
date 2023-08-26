package prom

import (
	"fmt"
	"testing"
)

func TestToFloat64(t *testing.T) {
	var (
		// BusinessResponseCodeCount for business respones count
		GateWayBusinessCount = New().WithCounter("go_gateway_business_count", []string{"zone", "code"}).WithState("go_gateway_business_state", []string{"zone", "code"})
	)
	GateWayBusinessCount.Incr("test1", "2")
	GateWayBusinessCount.Add("test1", 2, "2")
	GateWayBusinessCount.Incr("test2", "2")
	GateWayBusinessCount.Incr("test2", "2")
	label1 := []string{"test1", "2"}
	label2 := []string{"test2", "2"}
	fmt.Println(ToFloat64(GateWayBusinessCount.Counter.WithLabelValues(label1...)))
	fmt.Println(ToFloat64(GateWayBusinessCount.Counter.WithLabelValues(label2...)))
}
